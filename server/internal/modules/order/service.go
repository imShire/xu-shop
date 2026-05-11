package order

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/account"
	"github.com/xushop/xu-shop/internal/modules/address"
	"github.com/xushop/xu-shop/internal/modules/inventory"
	"github.com/xushop/xu-shop/internal/modules/product"
	pkgcsv "github.com/xushop/xu-shop/internal/pkg/csv"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/stock"
	pkgtypes "github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 状态机 ----

// transition 单条状态迁移规则。
type transition struct{ From, To, Trigger string }

var transitions = []transition{
	{StatusPending, StatusPaid, "pay_success"},
	{StatusPending, StatusCancelled, "user_cancel"},
	{StatusPending, StatusCancelled, "expire"},
	{StatusPending, StatusCancelled, "admin_close"},
	{StatusPaid, StatusShipped, "ship"},
	{StatusPaid, StatusRefunding, "refund_apply"},
	{StatusShipped, StatusCompleted, "user_confirm"},
	{StatusShipped, StatusCompleted, "auto_confirm"},
	{StatusShipped, StatusRefunding, "refund_apply"},
	{StatusRefunding, StatusRefunded, "refund_success"},
	{StatusRefunding, StatusPaid, "refund_failed"},
	{StatusRefunding, StatusShipped, "refund_failed"},
}

// findTransition 查找合法迁移，preferTo 用于 refund_failed 这类有多个目标的迁移。
// preferTo 为空时返回第一个匹配；非空时返回匹配 To 的那条。
func findTransition(from, trigger, preferTo string) *transition {
	var first *transition
	for i := range transitions {
		t := &transitions[i]
		if t.From != from || t.Trigger != trigger {
			continue
		}
		if first == nil {
			first = t
		}
		if preferTo != "" && t.To == preferTo {
			return t
		}
	}
	if preferTo == "" {
		return first
	}
	return nil
}

// ---- DTO ----

// CreateOrderReq 下单请求。
type CreateOrderReq struct {
	AddressID         pkgtypes.Int64Str  `json:"address_id"          binding:"required"`
	Items             []OrderItemReq     `json:"items"               binding:"required,min=1"`
	BuyerRemark       string             `json:"buyer_remark"`
	IdempotencyKey    string             `json:"idempotency_key"`
	FreightTemplateID *pkgtypes.Int64Str `json:"freight_template_id"`
	UseBalance        bool               `json:"use_balance"`
}

// OrderItemReq 单行商品请求。
type OrderItemReq struct {
	SkuID pkgtypes.Int64Str `json:"sku_id" binding:"required"`
	Qty   int               `json:"qty"    binding:"required,min=1,max=999"`
}

// OrderResp 订单列表/详情响应 DTO。
type OrderResp struct {
	ID                   pkgtypes.Int64Str `json:"id"`
	OrderNo              string            `json:"order_no"`
	UserID               pkgtypes.Int64Str `json:"user_id"`
	Status               string            `json:"status"`
	GoodsCents           int64             `json:"goods_cents"`
	FreightCents         int64             `json:"freight_cents"`
	DiscountCents        int64             `json:"discount_cents"`
	CouponDiscountCents  int64             `json:"coupon_discount_cents"`
	TotalCents           int64             `json:"total_cents"`
	PayCents             int64             `json:"pay_cents"`
	BalancePayCents      int64             `json:"balance_pay_cents"`
	AddressSnapshot      AddressSnapshot   `json:"address_snapshot"`
	BuyerRemark          *string           `json:"buyer_remark,omitempty"`
	CancelRequestPending bool              `json:"cancel_request_pending"`
	CancelRequestReason  *string           `json:"cancel_request_reason,omitempty"`
	CancelRequestAt      *time.Time        `json:"cancel_request_at,omitempty"`
	CancelReason         *string           `json:"cancel_reason,omitempty"`
	ExpireAt             time.Time         `json:"expire_at"`
	PaidAt               *time.Time        `json:"paid_at,omitempty"`
	ShippedAt            *time.Time        `json:"shipped_at,omitempty"`
	CompletedAt          *time.Time        `json:"completed_at,omitempty"`
	CancelledAt          *time.Time        `json:"cancelled_at,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

// OrderItemResp 订单商品行响应 DTO。
type OrderItemResp struct {
	ID              pkgtypes.Int64Str `json:"id"`
	OrderID         pkgtypes.Int64Str `json:"order_id"`
	ProductID       pkgtypes.Int64Str `json:"product_id"`
	SkuID           pkgtypes.Int64Str `json:"sku_id"`
	ProductSnapshot RawJSON           `json:"product_snapshot"`
	PriceCents      int64             `json:"price_cents"`
	Qty             int               `json:"qty"`
	WeightG         int               `json:"weight_g"`
	CreatedAt       time.Time         `json:"created_at"`
}

// OrderDetailResp 订单详情响应。
type OrderDetailResp struct {
	OrderResp
	Items   []OrderItemResp   `json:"items"`
	Logs    []OrderLogResp    `json:"logs"`
	Remarks []OrderRemarkResp `json:"remarks"`
}

// AdminOrderListResp 后台订单列表响应（含商品行快照）。
type AdminOrderListResp struct {
	OrderResp
	Items []OrderItemResp `json:"items"`
}

// AddedItem 再次购买响应中添加到购物车的条目。
type AddedItem struct {
	SkuID pkgtypes.Int64Str `json:"sku_id"`
	Qty   int               `json:"qty"`
}

// toOrderResp entity → OrderResp。
func toOrderResp(o *Order) OrderResp {
	return OrderResp{
		ID:                   pkgtypes.Int64Str(o.ID),
		OrderNo:              o.OrderNo,
		UserID:               pkgtypes.Int64Str(o.UserID),
		Status:               o.Status,
		GoodsCents:           o.GoodsCents,
		FreightCents:         o.FreightCents,
		DiscountCents:        o.DiscountCents,
		CouponDiscountCents:  o.CouponDiscountCents,
		TotalCents:           o.TotalCents,
		PayCents:             o.PayCents,
		BalancePayCents:      o.BalancePayCents,
		AddressSnapshot:      o.AddressSnapshot,
		BuyerRemark:          o.BuyerRemark,
		CancelRequestPending: o.CancelRequestPending,
		CancelRequestReason:  o.CancelRequestReason,
		CancelRequestAt:      o.CancelRequestAt,
		CancelReason:         o.CancelReason,
		ExpireAt:             o.ExpireAt,
		PaidAt:               o.PaidAt,
		ShippedAt:            o.ShippedAt,
		CompletedAt:          o.CompletedAt,
		CancelledAt:          o.CancelledAt,
		CreatedAt:            o.CreatedAt,
		UpdatedAt:            o.UpdatedAt,
	}
}

// toOrderItemResp entity → OrderItemResp。
func toOrderItemResp(it *OrderItem) OrderItemResp {
	return OrderItemResp{
		ID:              pkgtypes.Int64Str(it.ID),
		OrderID:         pkgtypes.Int64Str(it.OrderID),
		ProductID:       pkgtypes.Int64Str(it.ProductID),
		SkuID:           pkgtypes.Int64Str(it.SkuID),
		ProductSnapshot: it.ProductSnapshot,
		PriceCents:      it.PriceCents,
		Qty:             it.Qty,
		WeightG:         it.WeightG,
		CreatedAt:       it.CreatedAt,
	}
}

// FreightTemplateReq 运费模板请求。
type FreightTemplateReq struct {
	Name                 string  `json:"name"                   binding:"required"`
	FreeThresholdCents   int64   `json:"free_threshold_cents"   binding:"required"`
	DefaultFeeCents      int64   `json:"default_fee_cents"      binding:"required"`
	RemoteThresholdCents int64   `json:"remote_threshold_cents" binding:"required"`
	RemoteFeeCents       int64   `json:"remote_fee_cents"       binding:"required"`
	RemoteProvinces      RawJSON `json:"remote_provinces"`
	IsDefault            bool    `json:"is_default"`
}

// ---- Service ----

// Service 订单服务。
type Service struct {
	repo        OrderRepo
	skuRepo     product.SKURepo
	productRepo product.ProductRepo
	invRepo     inventory.InventoryRepo
	addrRepo    address.AddressRepo
	stockClient *stock.Client
	rdb         *redis.Client
	asynqClient *asynq.Client
	userRepo    account.UserRepo
}

// NewService 构造 Service。
func NewService(
	repo OrderRepo,
	skuRepo product.SKURepo,
	productRepo product.ProductRepo,
	invRepo inventory.InventoryRepo,
	addrRepo address.AddressRepo,
	stockClient *stock.Client,
	rdb *redis.Client,
	asynqClient *asynq.Client,
	userRepo account.UserRepo,
) *Service {
	return &Service{
		repo:        repo,
		skuRepo:     skuRepo,
		productRepo: productRepo,
		invRepo:     invRepo,
		addrRepo:    addrRepo,
		stockClient: stockClient,
		rdb:         rdb,
		asynqClient: asynqClient,
		userRepo:    userRepo,
	}
}

// ---- 状态机辅助 ----

// findPreRefundStatus 查询进入 refunding 之前的状态（paid 或 shipped），用于 refund_failed 回滚。
func (s *Service) findPreRefundStatus(ctx context.Context, orderID int64) string {
	logs, err := s.repo.ListLogsByOrder(ctx, orderID)
	if err != nil {
		return ""
	}
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i].ToStatus != nil && *logs[i].ToStatus == StatusRefunding {
			if logs[i].FromStatus != nil {
				return *logs[i].FromStatus
			}
		}
	}
	return ""
}

// ---- 状态机 ----

// Transition 状态机变更（SELECT FOR UPDATE + 校验 + UPDATE + 写 order_log）。
func (s *Service) Transition(
	ctx context.Context,
	orderID int64,
	trigger, operatorType string,
	operatorID int64,
	reason string,
) (*Order, error) {
	var result *Order
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o, err := s.repo.FindByIDForUpdate(ctx, tx, orderID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errs.ErrNotFound
			}
			return errs.ErrInternal
		}

		// refund_failed 需要回滚到退款前的状态（paid 或 shipped），通过查 order_log 确定
		preferTo := ""
		if trigger == "refund_failed" && o.Status == StatusRefunding {
			preferTo = s.findPreRefundStatus(ctx, o.ID)
		}
		tr := findTransition(o.Status, trigger, preferTo)
		if tr == nil {
			return errs.ErrParam.WithMsg(
				fmt.Sprintf("非法状态迁移：%s -[%s]->", o.Status, trigger))
		}

		// 更新状态字段
		now := time.Now()
		o.Status = tr.To
		o.UpdatedAt = now
		switch tr.To {
		case StatusPaid:
			o.PaidAt = &now
		case StatusShipped:
			o.ShippedAt = &now
		case StatusCompleted:
			o.CompletedAt = &now
		case StatusCancelled:
			o.CancelledAt = &now
		}

		if err := tx.Save(o).Error; err != nil {
			return err
		}

		// 写变更日志
		fromStatus := tr.From
		toStatus := tr.To
		reasonPtr := &reason
		opType := operatorType
		var opID *int64
		if operatorID > 0 {
			opID = &operatorID
		}
		log := &OrderLog{
			ID:           snowflake.NextID(),
			OrderID:      orderID,
			FromStatus:   &fromStatus,
			ToStatus:     &toStatus,
			Reason:       reasonPtr,
			OperatorType: &opType,
			OperatorID:   opID,
		}
		if err := tx.Create(log).Error; err != nil {
			return err
		}

		result = o
		return nil
	})
	return result, err
}

// ---- 下单 ----

// CreateOrder 完整下单流程。
func (s *Service) CreateOrder(ctx context.Context, userID int64, req CreateOrderReq) (*Order, error) {
	// 1. 幂等键检查（Redis 缓存 24h，在 handler 层 Idempotency-Key 中间件已做初步拦截；
	//    此处补充 service 层按 user 维度的幂等）
	idemKey := req.IdempotencyKey
	if idemKey != "" {
		redisKey := fmt.Sprintf("order:idem:%d:%s", userID, idemKey)
		exists, err := s.rdb.Exists(ctx, redisKey).Result()
		if err == nil && exists > 0 {
			return nil, errs.ErrConflict.WithMsg("重复下单（幂等键已存在）")
		}
	}

	// 2. 加载地址（校验归属）
	addr, err := s.addrRepo.FindByID(ctx, req.AddressID.Int64(), userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound.WithMsg("收货地址不存在")
		}
		return nil, errs.ErrInternal
	}

	// 3. 加载 SKU（批量 IN 查询）
	skuIDs := make([]int64, len(req.Items))
	for i, it := range req.Items {
		skuIDs[i] = it.SkuID.Int64()
	}
	skus, err := s.skuRepo.FindByIDs(ctx, skuIDs)
	if err != nil {
		return nil, errs.ErrInternal
	}
	skuMap := make(map[int64]product.SKU, len(skus))
	for _, sk := range skus {
		skuMap[sk.ID] = sk
	}

	// 批量加载商品信息（用于快照标题）
	productIDSet := make(map[int64]struct{}, len(skus))
	for _, sk := range skus {
		productIDSet[sk.ProductID] = struct{}{}
	}
	productIDs := make([]int64, 0, len(productIDSet))
	for pid := range productIDSet {
		productIDs = append(productIDs, pid)
	}
	products, _ := s.productRepo.FindByIDs(ctx, productIDs)
	productMap := make(map[int64]product.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	// 4. 校验商品/SKU 状态
	for _, it := range req.Items {
		sk, ok := skuMap[it.SkuID.Int64()]
		if !ok {
			return nil, errs.ErrNotFound.WithMsg(fmt.Sprintf("SKU %d 不存在", it.SkuID.Int64()))
		}
		if sk.Status != "active" {
			return nil, errs.ErrParam.WithMsg(fmt.Sprintf("SKU %d 已下架", it.SkuID.Int64()))
		}
		p, ok := productMap[sk.ProductID]
		if !ok {
			return nil, errs.ErrNotFound.WithMsg(fmt.Sprintf("商品 %d 不存在", sk.ProductID))
		}
		if p.Status != "onsale" {
			return nil, errs.ErrParam.WithMsg(fmt.Sprintf("商品 %d 已下架", sk.ProductID))
		}
	}

	// 5. 计算商品金额
	var goodsCents int64
	for _, it := range req.Items {
		sk := skuMap[it.SkuID.Int64()]
		goodsCents += sk.PriceCents * int64(it.Qty)
	}

	// 加载运费模板
	var freightTpl *FreightTemplate
	if req.FreightTemplateID != nil {
		freightTpl, err = s.repo.GetFreightByID(ctx, req.FreightTemplateID.Int64())
	} else {
		freightTpl, err = s.repo.FindDefaultFreight(ctx)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 无运费模板，使用默认
			freightTpl = &FreightTemplate{
				FreeThresholdCents:   9900,
				DefaultFeeCents:      1000,
				RemoteThresholdCents: 19900,
				RemoteFeeCents:       2000,
				RemoteProvinces:      RawJSON("[]"),
			}
		} else {
			return nil, errs.ErrInternal
		}
	}

	// 5. 构建地址快照
	addrSnap := AddressSnapshot{
		Name:   addr.Name,
		Phone:  addr.Phone,
		Detail: addr.Detail,
	}
	if addr.Province != nil {
		addrSnap.Province = *addr.Province
	}
	if addr.ProvinceCode != nil {
		addrSnap.ProvinceCode = *addr.ProvinceCode
	}
	if addr.City != nil {
		addrSnap.City = *addr.City
	}
	if addr.CityCode != nil {
		addrSnap.CityCode = *addr.CityCode
	}
	if addr.District != nil {
		addrSnap.District = *addr.District
	}
	if addr.DistrictCode != nil {
		addrSnap.DistrictCode = *addr.DistrictCode
	}
	if addr.Street != nil {
		addrSnap.Street = *addr.Street
	}
	if addr.StreetCode != nil {
		addrSnap.StreetCode = *addr.StreetCode
	}

	// 计算运费
	freightCents := CalcFreight(goodsCents, addrSnap, freightTpl)

	totalCents := goodsCents + freightCents
	payCents := totalCents
	var balancePayCents int64

	// 余额抵扣（预估，实际扣减在 DB 事务后）
	if req.UseBalance && s.userRepo != nil {
		userBalance, balErr := s.userRepo.GetBalance(ctx, userID)
		if balErr == nil && userBalance > 0 {
			if userBalance >= payCents {
				balancePayCents = payCents
			} else {
				balancePayCents = userBalance
			}
			payCents = payCents - balancePayCents
		}
	}

	// 6. 生成 orderNo（yyyyMMddHHmmss + 6位随机）
	orderNo := genOrderNo()

	// 7. Redis Lua 锁库存
	lockItems := make([]stock.LockItem, len(req.Items))
	for i, it := range req.Items {
		lockItems[i] = stock.LockItem{SKUID: it.SkuID.Int64(), Qty: it.Qty}
	}
	// 预热 Redis 库存 key（SETNX，不覆盖已存在的 key）
	for _, it := range req.Items {
		skuID := it.SkuID.Int64()
		dbStock, dbLocked, stockErr := s.invRepo.GetSKUStock(ctx, skuID)
		if stockErr == nil {
			available := dbStock - dbLocked
			if available < 0 {
				available = 0
			}
			_ = s.stockClient.Load(ctx, skuID, available)
		}
	}
	lockResult, lockErr := s.stockClient.Lock(ctx, orderNo, lockItems)
	if lockErr != nil {
		logger.Ctx(ctx).Error("stock lock redis error", zap.Error(lockErr))
		return nil, errs.ErrInternal
	}
	if lockResult != "ok" {
		return nil, errs.ErrParam.WithMsg("库存不足，下单失败")
	}

	// 8. DB 事务
	expireAt := time.Now().Add(15 * time.Minute)
	orderID := snowflake.NextID()

	var buyerRemark *string
	if req.BuyerRemark != "" {
		buyerRemark = &req.BuyerRemark
	}
	var idemKeyPtr *string
	if idemKey != "" {
		idemKeyPtr = &idemKey
	}

	o := &Order{
		ID:              orderID,
		OrderNo:         orderNo,
		UserID:          userID,
		Status:          StatusPending,
		GoodsCents:      goodsCents,
		FreightCents:    freightCents,
		TotalCents:      totalCents,
		PayCents:        payCents,
		BalancePayCents: balancePayCents,
		AddressSnapshot: addrSnap,
		BuyerRemark:     buyerRemark,
		IdempotencyKey:  idemKeyPtr,
		ExpireAt:        expireAt,
	}

	// 构建订单行
	orderItems := make([]OrderItem, len(req.Items))
	for i, it := range req.Items {
		sk := skuMap[it.SkuID.Int64()]
		prod := productMap[sk.ProductID]
		snapBytes, _ := json.Marshal(map[string]any{
			"product_id": sk.ProductID,
			"sku_id":     sk.ID,
			"sku_code":   sk.SkuCode,
			"title":      prod.Title,
			"main_image": prod.MainImage,
			"price":      sk.PriceCents,
			"image":      sk.Image,
			"attrs":      sk.Attrs,
		})
		orderItems[i] = OrderItem{
			ID:              snowflake.NextID(),
			OrderID:         orderID,
			ProductID:       sk.ProductID,
			SkuID:           sk.ID,
			ProductSnapshot: RawJSON(snapBytes),
			PriceCents:      sk.PriceCents,
			Qty:             it.Qty,
			WeightG:         sk.WeightG,
		}
	}

	toStatus := StatusPending
	opType := "system"
	reasonStr := "用户下单"
	initLog := OrderLog{
		ID:           snowflake.NextID(),
		OrderID:      orderID,
		ToStatus:     &toStatus,
		OperatorType: &opType,
		Reason:       &reasonStr,
	}

	dbErr := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}
		if err := tx.Create(&initLog).Error; err != nil {
			return err
		}
		// DB 层锁库存（locked_stock += qty）
		for _, it := range req.Items {
			sk := skuMap[it.SkuID.Int64()]
			newLocked := sk.LockedStock + it.Qty
			if err := tx.Table("sku").Where("id = ?", it.SkuID.Int64()).
				Update("locked_stock", newLocked).Error; err != nil {
				return err
			}
		}
		return nil
	})

	// 9. DB 失败则释放 Redis 锁
	if dbErr != nil {
		logger.Ctx(ctx).Error("create order db tx failed, releasing redis lock",
			zap.String("order_no", orderNo), zap.Error(dbErr))
		if _, relErr := s.stockClient.Release(ctx, orderNo); relErr != nil {
			logger.Ctx(ctx).Error("release stock lock failed", zap.Error(relErr))
		}
		return nil, errs.ErrInternal
	}

	// 10. 扣减余额（DB 事务外；若扣减失败不回滚订单，重置 balance_pay_cents 并记录告警）
	if balancePayCents > 0 && s.userRepo != nil {
		if deductErr := s.userRepo.DeductBalance(ctx, userID, balancePayCents, "order", orderID, "订单支付"); deductErr != nil {
			logger.Ctx(ctx).Error("deduct balance failed, order still created",
				zap.Int64("order_id", orderID),
				zap.Int64("balance_pay_cents", balancePayCents),
				zap.Error(deductErr))
			// 余额扣减失败：重置订单 balance_pay_cents 并补回 pay_cents
			_ = s.repo.DB().WithContext(ctx).Model(&Order{}).Where("id = ?", orderID).
				Updates(map[string]any{
					"balance_pay_cents": 0,
					"pay_cents":         totalCents,
				}).Error
			o.BalancePayCents = 0
			o.PayCents = totalCents
		}
	}

	// 11. 投递超时关单任务（15 min）
	payload, _ := json.Marshal(map[string]int64{"order_id": orderID})
	task := asynq.NewTask("order:close", payload)
	if _, enqErr := s.asynqClient.EnqueueContext(ctx, task,
		asynq.ProcessIn(15*time.Minute),
		asynq.MaxRetry(3),
	); enqErr != nil {
		logger.Ctx(ctx).Warn("enqueue order:close failed", zap.Error(enqErr))
	}

	// 11. 存幂等缓存
	if idemKey != "" {
		redisKey := fmt.Sprintf("order:idem:%d:%s", userID, idemKey)
		_ = s.rdb.SetEx(ctx, redisKey, orderID, 24*time.Hour)
	}

	return o, nil
}

// CalcFreight 计算运费。
func CalcFreight(goodsCents int64, addr AddressSnapshot, tpl *FreightTemplate) int64 {
	isRemote := containsProvince(tpl.RemoteProvinces, addr.Province)
	threshold := tpl.FreeThresholdCents
	fee := tpl.DefaultFeeCents
	if isRemote {
		threshold = tpl.RemoteThresholdCents
		fee = tpl.RemoteFeeCents
	}
	if goodsCents >= threshold {
		return 0
	}
	return fee
}

// containsProvince 检查省份是否在偏远地区列表中（JSON 数组）。
func containsProvince(provincesJSON RawJSON, province string) bool {
	if len(provincesJSON) == 0 || province == "" {
		return false
	}
	var provinces []string
	if err := json.Unmarshal(provincesJSON, &provinces); err != nil {
		return false
	}
	for _, p := range provinces {
		if p == province {
			return true
		}
	}
	return false
}

// genOrderNo 生成订单号：yyyyMMddHHmmss + 6位随机数字。
func genOrderNo() string {
	now := time.Now().Format("20060102150405")
	n := rand.Intn(900000) + 100000 //nolint:gosec
	return fmt.Sprintf("%s%06d", now, n)
}

// ---- C 端业务 ----

// GetOrder 获取订单详情。
func (s *Service) GetOrder(ctx context.Context, orderID, userID int64) (*OrderDetailResp, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	if o.UserID != userID {
		return nil, errs.ErrForbidden
	}

	var items []OrderItem
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, errs.ErrInternal
	}
	var rawLogs []OrderLog
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at ASC").Find(&rawLogs).Error; err != nil {
		return nil, errs.ErrInternal
	}
	logs := make([]OrderLogResp, len(rawLogs))
	for i := range rawLogs {
		logs[i] = toOrderLogResp(&rawLogs[i])
	}

	itemResps := make([]OrderItemResp, len(items))
	for i := range items {
		itemResps[i] = toOrderItemResp(&items[i])
	}
	return &OrderDetailResp{OrderResp: toOrderResp(o), Items: itemResps, Logs: logs}, nil
}

// ListOrders C 端订单列表。
func (s *Service) ListOrders(ctx context.Context, userID int64, status string, page, size int) ([]OrderResp, int64, error) {
	list, total, err := s.repo.ListByUser(ctx, userID, status, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resps := make([]OrderResp, len(list))
	for i := range list {
		resps[i] = toOrderResp(&list[i])
	}
	return resps, total, nil
}

// CancelOrder 取消订单。
// - pending：直接取消（状态机 expire/user_cancel + 释放 Redis 锁）
// - paid：设置 cancel_request_pending=true
func (s *Service) CancelOrder(ctx context.Context, orderID, userID int64, reason string) error {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if o.UserID != userID {
		return errs.ErrForbidden
	}

	switch o.Status {
	case StatusPending:
		_, transErr := s.Transition(ctx, orderID, "user_cancel", "user", userID, reason)
		if transErr != nil {
			return transErr
		}
		// 释放 Redis 库存锁
		if _, relErr := s.stockClient.Release(ctx, o.OrderNo); relErr != nil {
			logger.Ctx(ctx).Warn("release stock on cancel failed",
				zap.String("order_no", o.OrderNo), zap.Error(relErr))
		}
		// DB 解锁库存
		s.releaseDBStock(ctx, orderID)
		return nil

	case StatusPaid:
		now := time.Now()
		o.CancelRequestPending = true
		o.CancelRequestReason = &reason
		o.CancelRequestAt = &now
		return s.repo.Update(ctx, o)

	default:
		return errs.ErrParam.WithMsg(fmt.Sprintf("当前状态 %s 不允许取消", o.Status))
	}
}

// WithdrawCancelRequest 撤回取消申请。
func (s *Service) WithdrawCancelRequest(ctx context.Context, orderID, userID int64) error {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if o.UserID != userID {
		return errs.ErrForbidden
	}
	if !o.CancelRequestPending {
		return errs.ErrParam.WithMsg("没有待处理的取消申请")
	}
	o.CancelRequestPending = false
	o.CancelRequestReason = nil
	o.CancelRequestAt = nil
	return s.repo.Update(ctx, o)
}

// ConfirmReceived 确认收货。
func (s *Service) ConfirmReceived(ctx context.Context, orderID, userID int64) error {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if o.UserID != userID {
		return errs.ErrForbidden
	}
	_, err = s.Transition(ctx, orderID, "user_confirm", "user", userID, "用户确认收货")
	return err
}

// BuyAgain 再次购买（把订单 SKU 加购物车，返回 AddedItem 列表供前端展示）。
func (s *Service) BuyAgain(ctx context.Context, orderID, userID int64) ([]AddedItem, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	if o.UserID != userID {
		return nil, errs.ErrForbidden
	}

	var items []OrderItem
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, errs.ErrInternal
	}

	added := make([]AddedItem, 0, len(items))
	for _, it := range items {
		added = append(added, AddedItem{SkuID: pkgtypes.Int64Str(it.SkuID), Qty: it.Qty})
	}
	return added, nil
}

// ---- 后台业务 ----

// AdminCancelOrder 后台强制关单。
func (s *Service) AdminCancelOrder(ctx context.Context, orderID, adminID int64, reason string) error {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	_, transErr := s.Transition(ctx, orderID, "admin_close", "admin", adminID, reason)
	if transErr != nil {
		return transErr
	}
	if o.Status == StatusPending {
		if _, relErr := s.stockClient.Release(ctx, o.OrderNo); relErr != nil {
			logger.Ctx(ctx).Warn("release stock on admin cancel",
				zap.String("order_no", o.OrderNo), zap.Error(relErr))
		}
		s.releaseDBStock(ctx, orderID)
	}
	return nil
}

// AdminListOrders 后台订单列表。
func (s *Service) AdminListOrders(ctx context.Context, filter AdminOrderFilter) ([]AdminOrderListResp, int64, error) {
	list, total, err := s.repo.ListByAdmin(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	if len(list) == 0 {
		return nil, total, nil
	}
	orderIDs := make([]int64, len(list))
	for i, o := range list {
		orderIDs[i] = o.ID
	}
	allItems, _ := s.repo.FindItemsByOrderIDs(ctx, orderIDs)
	itemsByOrder := make(map[int64][]OrderItemResp, len(list))
	for i := range allItems {
		it := &allItems[i]
		itemsByOrder[it.OrderID] = append(itemsByOrder[it.OrderID], toOrderItemResp(it))
	}
	resps := make([]AdminOrderListResp, len(list))
	for i := range list {
		resps[i] = AdminOrderListResp{
			OrderResp: toOrderResp(&list[i]),
			Items:     itemsByOrder[list[i].ID],
		}
	}
	return resps, total, nil
}

// AdminExportOrders 游标导出 CSV（防注入）。
func (s *Service) AdminExportOrders(ctx context.Context, filter AdminOrderFilter) ([]byte, error) {
	filter.Size = 500
	filter.Page = 1

	var buf bytes.Buffer
	w := pkgcsv.NewSafeWriter(&buf)
	csvWriter := csv.NewWriter(&buf)

	// 写表头（直接用原始 writer 写，无需防注入）
	_ = csvWriter.Write([]string{
		"订单号", "用户ID", "状态", "商品金额(分)", "运费(分)", "实付(分)",
		"创建时间", "收货人", "手机号",
	})
	csvWriter.Flush()

	for {
		list, _, err := s.repo.ListByAdmin(ctx, filter)
		if err != nil {
			return nil, errs.ErrInternal
		}
		if len(list) == 0 {
			break
		}
		for _, o := range list {
			row := []string{
				o.OrderNo,
				fmt.Sprintf("%d", o.UserID),
				o.Status,
				fmt.Sprintf("%d", o.GoodsCents),
				fmt.Sprintf("%d", o.FreightCents),
				fmt.Sprintf("%d", o.PayCents),
				o.CreatedAt.Format("2006-01-02 15:04:05"),
				o.AddressSnapshot.Name,
				o.AddressSnapshot.Phone,
			}
			_ = w.Write(row)
		}
		w.Flush()
		if len(list) < filter.Size {
			break
		}
		filter.Page++
	}

	return buf.Bytes(), nil
}

// AddRemark 管理员备注。
func (s *Service) AddRemark(ctx context.Context, orderID, adminID int64, content string) error {
	remark := &OrderRemark{
		OrderID: orderID,
		AdminID: adminID,
		Content: content,
	}
	return s.repo.AddRemark(ctx, remark)
}

// ListRemarks 获取管理员备注列表。
func (s *Service) ListRemarks(ctx context.Context, orderID int64) ([]OrderRemarkResp, error) {
	list, err := s.repo.ListRemarks(ctx, orderID)
	if err != nil {
		return nil, err
	}
	resp := make([]OrderRemarkResp, len(list))
	for i := range list {
		resp[i] = toOrderRemarkResp(&list[i])
	}
	return resp, nil
}

// ---- 运费模板 ----

// CreateFreightTemplate 创建运费模板。
func (s *Service) CreateFreightTemplate(ctx context.Context, req FreightTemplateReq) (*FreightTemplate, error) {
	provinces := req.RemoteProvinces
	if len(provinces) == 0 {
		provinces = RawJSON("[]")
	}
	t := &FreightTemplate{
		Name:                 req.Name,
		FreeThresholdCents:   req.FreeThresholdCents,
		DefaultFeeCents:      req.DefaultFeeCents,
		RemoteThresholdCents: req.RemoteThresholdCents,
		RemoteFeeCents:       req.RemoteFeeCents,
		RemoteProvinces:      provinces,
		IsDefault:            req.IsDefault,
	}
	if err := s.repo.CreateFreightTemplate(ctx, t); err != nil {
		return nil, errs.ErrInternal
	}
	return t, nil
}

// UpdateFreightTemplate 更新运费模板。
func (s *Service) UpdateFreightTemplate(ctx context.Context, id int64, req FreightTemplateReq) error {
	t, err := s.repo.GetFreightByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	t.Name = req.Name
	t.FreeThresholdCents = req.FreeThresholdCents
	t.DefaultFeeCents = req.DefaultFeeCents
	t.RemoteThresholdCents = req.RemoteThresholdCents
	t.RemoteFeeCents = req.RemoteFeeCents
	if len(req.RemoteProvinces) > 0 {
		t.RemoteProvinces = req.RemoteProvinces
	}
	t.IsDefault = req.IsDefault
	return s.repo.UpdateFreightTemplate(ctx, t)
}

// ListFreightTemplates 运费模板列表。
func (s *Service) ListFreightTemplates(ctx context.Context) ([]FreightTemplate, error) {
	return s.repo.ListFreightTemplates(ctx)
}

// DeleteFreightTemplate 删除运费模板。
func (s *Service) DeleteFreightTemplate(ctx context.Context, id int64) error {
	return s.repo.DeleteFreightTemplate(ctx, id)
}

// ---- 内部辅助 ----

// releaseDBStock 关单后 DB 解锁库存。
func (s *Service) releaseDBStock(ctx context.Context, orderID int64) {
	var items []OrderItem
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		logger.Ctx(ctx).Error("releaseDBStock find items", zap.Error(err))
		return
	}
	for _, it := range items {
		if err := s.invRepo.UnlockDB(ctx, int(it.SkuID), it.Qty, orderID); err != nil {
			logger.Ctx(ctx).Error("releaseDBStock unlock",
				zap.Int64("sku_id", it.SkuID), zap.Error(err))
		}
	}
}

// GetRaw 供 job 使用的裸查询（不校验归属）。
func (s *Service) GetRaw(ctx context.Context, orderID int64) (*Order, error) {
	return s.repo.FindByID(ctx, orderID)
}

// ReleaseStock 供 job 使用的库存释放。
func (s *Service) ReleaseStock(ctx context.Context, orderID int64, orderNo string) {
	if _, err := s.stockClient.Release(ctx, orderNo); err != nil {
		logger.Ctx(ctx).Error("job release stock redis",
			zap.String("order_no", orderNo), zap.Error(err))
	}
	s.releaseDBStock(ctx, orderID)
}

// AdminGetOrder 后台获取订单详情。
func (s *Service) AdminGetOrder(ctx context.Context, orderID int64) (*OrderDetailResp, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	var items []OrderItem
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, errs.ErrInternal
	}
	var rawLogs []OrderLog
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at ASC").Find(&rawLogs).Error; err != nil {
		return nil, errs.ErrInternal
	}
	logs := make([]OrderLogResp, len(rawLogs))
	for i := range rawLogs {
		logs[i] = toOrderLogResp(&rawLogs[i])
	}
	rawRemarks, _ := s.repo.ListRemarks(ctx, orderID)
	remarks := make([]OrderRemarkResp, len(rawRemarks))
	for i := range rawRemarks {
		remarks[i] = toOrderRemarkResp(&rawRemarks[i])
	}
	itemResps := make([]OrderItemResp, len(items))
	for i := range items {
		itemResps[i] = toOrderItemResp(&items[i])
	}
	return &OrderDetailResp{OrderResp: toOrderResp(o), Items: itemResps, Logs: logs, Remarks: remarks}, nil
}

// suppress unused import warning for strings
var _ = strings.Contains
