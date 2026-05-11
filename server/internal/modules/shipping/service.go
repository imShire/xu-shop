package shipping

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/kdniao"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

const (
	// TaskExpressPull 快递轨迹兜底拉取任务名。
	TaskExpressPull = "express:pull_stale"
)

// OrderAccessor 发货服务对订单的操作接口。
type OrderAccessor interface {
	FindByID(ctx context.Context, id int64) (*OrderSnap, error)
	Transition(ctx context.Context, orderID int64, trigger, opType string, opID int64, reason string) error
}

// OrderSnap 发货服务所需的订单快照。
type OrderSnap struct {
	ID              int64
	OrderNo         string
	UserID          int64
	Status          string
	AddressSnapshot OrderAddrSnap
}

// OrderAddrSnap 地址快照。
type OrderAddrSnap struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	ProvinceCode string `json:"province_code,omitempty"`
	Province     string `json:"province"`
	CityCode     string `json:"city_code,omitempty"`
	City         string `json:"city"`
	DistrictCode string `json:"district_code,omitempty"`
	District     string `json:"district"`
	StreetCode   string `json:"street_code,omitempty"`
	Street       string `json:"street,omitempty"`
	Detail       string `json:"detail"`
}

// AsynqEnqueuer asynq 入队接口。
type AsynqEnqueuer interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// Service 发货服务。
type Service struct {
	senderRepo  SenderAddressRepo
	carrierRepo CarrierRepo
	shipRepo    ShipmentRepo
	kdniao      kdniao.Client
	orderDB     OrderAccessor
	rdb         *redis.Client
	enqueuer    AsynqEnqueuer
	notifyURL   string // 快递鸟回调 URL
}

// NewService 构造发货服务。
func NewService(
	senderRepo SenderAddressRepo,
	carrierRepo CarrierRepo,
	shipRepo ShipmentRepo,
	kdniaoClient kdniao.Client,
	orderDB OrderAccessor,
	rdb *redis.Client,
	enqueuer AsynqEnqueuer,
	notifyURL string,
) *Service {
	return &Service{
		senderRepo:  senderRepo,
		carrierRepo: carrierRepo,
		shipRepo:    shipRepo,
		kdniao:      kdniaoClient,
		orderDB:     orderDB,
		rdb:         rdb,
		enqueuer:    enqueuer,
		notifyURL:   notifyURL,
	}
}

// ---- 发货地址 CRUD ----

func (s *Service) ListSenderAddresses(ctx context.Context) ([]SenderAddressResp, error) {
	list, err := s.senderRepo.List(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	resp := make([]SenderAddressResp, len(list))
	for i := range list {
		resp[i] = toSenderAddressResp(&list[i])
	}
	return resp, nil
}

func (s *Service) CreateSenderAddress(ctx context.Context, a *SenderAddress) (*SenderAddressResp, error) {
	if err := s.senderRepo.Create(ctx, a); err != nil {
		return nil, errs.ErrInternal
	}
	r := toSenderAddressResp(a)
	return &r, nil
}

func (s *Service) UpdateSenderAddress(ctx context.Context, id int64, req SenderAddress) error {
	a, err := s.senderRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	req.ID = a.ID
	req.CreatedAt = a.CreatedAt
	if err := s.senderRepo.Update(ctx, &req); err != nil {
		return errs.ErrInternal
	}
	return nil
}

func (s *Service) DeleteSenderAddress(ctx context.Context, id int64) error {
	if err := s.senderRepo.Delete(ctx, id); err != nil {
		return errs.ErrInternal
	}
	return nil
}

func (s *Service) SetDefaultSenderAddress(ctx context.Context, id int64) error {
	if _, err := s.senderRepo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if err := s.senderRepo.SetDefault(ctx, id); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// ---- 快递公司 ----

func (s *Service) ListCarriers(ctx context.Context) ([]Carrier, error) {
	return s.carrierRepo.List(ctx)
}

func (s *Service) UpdateCarrier(ctx context.Context, code string, req Carrier) error {
	c, err := s.carrierRepo.FindByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	c.Name = req.Name
	c.KdniaoCode = req.KdniaoCode
	c.MonthlyAccount = req.MonthlyAccount
	c.Enabled = req.Enabled
	c.Sort = req.Sort
	if err := s.carrierRepo.Update(ctx, c); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// ---- 单单发货 ----

// Ship 单单发货全流程。
func (s *Service) Ship(ctx context.Context, orderID, adminID int64, req ShipReq) (*ShipmentResp, error) {
	// 1. 加载订单
	o, err := s.orderDB.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound.WithMsg("订单不存在")
		}
		return nil, errs.ErrInternal
	}
	if o.Status != "paid" {
		return nil, errs.ErrParam.WithMsg(fmt.Sprintf("订单状态 %s 不可发货", o.Status))
	}

	// 2. 加载快递公司
	carrier, err := s.carrierRepo.FindByCode(ctx, req.CarrierCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound.WithMsg("快递公司不存在")
		}
		return nil, errs.ErrInternal
	}

	// 3. 加载默认发件地址
	sender, err := s.senderRepo.FindDefault(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrParam.WithMsg("未配置默认发货地址")
		}
		return nil, errs.ErrInternal
	}

	trackingNo := req.TrackingNo
	var waybillPdfURL *string

	// 4. 如需创建面单
	if req.CreateWaybill || trackingNo == "" {
		goods := make([]kdniao.GoodsItem, len(req.Goods))
		for i, g := range req.Goods {
			goods[i] = kdniao.GoodsItem{Name: g.Name, Qty: g.Qty, Weight: g.Weight}
		}
		monthlyAccount := ""
		if carrier.MonthlyAccount != nil {
			monthlyAccount = *carrier.MonthlyAccount
		}
		waybillResp, err := s.kdniao.CreateWaybill(ctx, kdniao.CreateWaybillReq{
			CarrierCode:    carrier.KdniaoCode,
			OrderNo:        o.OrderNo,
			MonthlyAccount: monthlyAccount,
			Sender:         senderAddrToKD(*sender),
			Receiver:       orderAddrToKD(o.AddressSnapshot),
			Goods:          goods,
			CallbackURL:    s.notifyURL,
		})
		if err != nil {
			return nil, errs.ErrInternal.WithMsg("快递鸟建单失败: " + err.Error())
		}
		if !waybillResp.Success {
			return nil, errs.ErrInternal.WithMsg("快递鸟建单失败: " + waybillResp.Reason)
		}
		trackingNo = waybillResp.TrackingNo
		if waybillResp.PrintBase64 != "" {
			pdfURL := "base64://" + waybillResp.PrintBase64[:min(len(waybillResp.PrintBase64), 20)]
			waybillPdfURL = &pdfURL
		}
	}

	// 5. 订阅轨迹
	if trackingNo != "" {
		if err := s.kdniao.Subscribe(ctx, carrier.KdniaoCode, trackingNo, s.notifyURL); err != nil {
			logger.Ctx(ctx).Warn("subscribe track failed", zap.Error(err))
		}
	}

	// 6. 构建发货单快照
	senderSnap := AddressSnapshot{
		Company:      ptrStr(sender.Company),
		Name:         sender.Name,
		Phone:        sender.Phone,
		ProvinceCode: ptrStr(sender.ProvinceCode),
		Province:     sender.Province,
		CityCode:     ptrStr(sender.CityCode),
		City:         sender.City,
		DistrictCode: ptrStr(sender.DistrictCode),
		District:     sender.District,
		StreetCode:   ptrStr(sender.StreetCode),
		Street:       sender.Street,
		Detail:       sender.Detail,
	}
	receiverSnap := AddressSnapshot{
		Name:         o.AddressSnapshot.Name,
		Phone:        o.AddressSnapshot.Phone,
		ProvinceCode: o.AddressSnapshot.ProvinceCode,
		Province:     o.AddressSnapshot.Province,
		CityCode:     o.AddressSnapshot.CityCode,
		City:         o.AddressSnapshot.City,
		DistrictCode: o.AddressSnapshot.DistrictCode,
		District:     o.AddressSnapshot.District,
		StreetCode:   o.AddressSnapshot.StreetCode,
		Street:       o.AddressSnapshot.Street,
		Detail:       o.AddressSnapshot.Detail,
	}

	now := time.Now()
	shipment := &Shipment{
		ID:               snowflake.NextID(),
		OrderID:          orderID,
		CarrierCode:      req.CarrierCode,
		TrackingNo:       trackingNo,
		WaybillPdfURL:    waybillPdfURL,
		SenderSnapshot:   marshalSnapshot(senderSnap),
		ReceiverSnapshot: marshalSnapshot(receiverSnap),
		Status:           ShipStatusPicked,
		LastTrackAt:      &now,
	}
	if err := s.shipRepo.Create(ctx, shipment); err != nil {
		return nil, errs.ErrInternal
	}

	// 7. 触发订单状态机 ship
	if err := s.orderDB.Transition(ctx, orderID, "ship", "admin", adminID, "管理员发货"); err != nil {
		logger.Ctx(ctx).Error("order ship transition failed", zap.Int64("order_id", orderID), zap.Error(err))
	}

	shipResp := toShipmentResp(shipment)
	return &shipResp, nil
}

const batchShipRedisPrefix = "batch_ship:"

// BatchShip 异步批量发货，返回 task_id。
func (s *Service) BatchShip(ctx context.Context, adminID int64, req BatchShipReq) (string, error) {
	taskID := uuid.New().String()

	// 存储初始状态到 Redis
	status := BatchShipStatus{
		TaskID: taskID,
		Total:  len(req.Orders),
	}
	statusBytes, _ := json.Marshal(status)
	if s.rdb != nil {
		_ = s.rdb.SetEx(ctx, batchShipRedisPrefix+taskID, string(statusBytes), 24*time.Hour)
	}

	// 入队异步任务
	if s.enqueuer != nil {
		payload, _ := json.Marshal(map[string]any{
			"task_id":  taskID,
			"admin_id": adminID,
			"orders":   req.Orders,
		})
		task := asynq.NewTask("shipping:batch_ship", payload)
		if _, err := s.enqueuer.EnqueueContext(ctx, task, asynq.MaxRetry(1)); err != nil {
			logger.Ctx(ctx).Error("enqueue batch ship failed", zap.Error(err))
		}
	}

	return taskID, nil
}

// GetBatchShipStatus 查询批量发货任务状态。
func (s *Service) GetBatchShipStatus(ctx context.Context, taskID string) (*BatchShipStatus, error) {
	if s.rdb == nil {
		return nil, errs.ErrNotFound
	}
	val, err := s.rdb.Get(ctx, batchShipRedisPrefix+taskID).Result()
	if err != nil {
		return nil, errs.ErrNotFound.WithMsg("任务不存在")
	}
	var status BatchShipStatus
	if err := json.Unmarshal([]byte(val), &status); err != nil {
		return nil, errs.ErrInternal
	}
	return &status, nil
}

// GetBatchShipPDF 获取批量发货合并面单 URL。
func (s *Service) GetBatchShipPDF(ctx context.Context, taskID string) (string, error) {
	status, err := s.GetBatchShipStatus(ctx, taskID)
	if err != nil {
		return "", err
	}
	if status.PDFURL == "" {
		return "", errs.ErrNotFound.WithMsg("面单尚未生成")
	}
	return status.PDFURL, nil
}

// ---- 发货单列表 & 修改 ----

// ListShipments 后台发货单列表。
func (s *Service) ListShipments(ctx context.Context, filter ShipmentFilter) ([]ShipmentResp, int64, error) {
	list, total, err := s.shipRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resps := make([]ShipmentResp, len(list))
	for i := range list {
		resps[i] = toShipmentResp(&list[i])
	}
	return resps, total, nil
}

// UpdateShipment 修改发货单（改运单号/快递公司）。
func (s *Service) UpdateShipment(ctx context.Context, id int64, req UpdateShipmentReq) error {
	ship, err := s.shipRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if req.TrackingNo != "" {
		ship.TrackingNo = req.TrackingNo
	}
	if req.CarrierCode != "" {
		ship.CarrierCode = req.CarrierCode
	}
	ship.UpdatedAt = time.Now()
	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// ListTracks C 端查询订单物流轨迹。
func (s *Service) ListTracks(ctx context.Context, orderID int64) ([]ShipmentTrack, error) {
	ship, err := s.shipRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []ShipmentTrack{}, nil
		}
		return nil, errs.ErrInternal
	}
	tracks, err := s.shipRepo.ListTracks(ctx, ship.ID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	return tracks, nil
}

// ---- 快递鸟 Webhook ----

// HandleExpressPush 处理快递鸟推送。
func (s *Service) HandleExpressPush(ctx context.Context, body []byte) error {
	push, err := s.kdniao.ParsePush(body)
	if err != nil {
		logger.Ctx(ctx).Error("parse express push failed", zap.Error(err))
		return errs.ErrParam.WithMsg("解析快递推送失败")
	}

	ship, err := s.shipRepo.FindByCarrierAndNo(ctx, push.CarrierCode, push.TrackingNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Ctx(ctx).Warn("express push: shipment not found",
				zap.String("carrier", push.CarrierCode),
				zap.String("tracking_no", push.TrackingNo))
			return nil
		}
		return errs.ErrInternal
	}

	now := time.Now()
	ship.LastTrackAt = &now
	newStatus := mapPushState(push.State)
	ship.Status = newStatus
	if newStatus == ShipStatusDelivered {
		ship.DeliveredAt = &now
	}

	// 写入轨迹（幂等：按 occurred_at + description 去重）
	for _, trace := range push.Traces {
		t := trace // 避免循环变量捕获
		desc := t.Description
		exists, err := s.shipRepo.ExistsTrack(ctx, ship.ID, t.OccurredAt, desc)
		if err != nil {
			logger.Ctx(ctx).Warn("check track exists failed", zap.Error(err))
			continue
		}
		if exists {
			continue
		}
		track := &ShipmentTrack{
			ID:          snowflake.NextID(),
			ShipmentID:  ship.ID,
			Status:      t.Status,
			Description: &desc,
			OccurredAt:  t.OccurredAt,
		}
		if err := s.shipRepo.CreateTrack(ctx, track); err != nil {
			logger.Ctx(ctx).Warn("create track failed", zap.Error(err))
		}
	}

	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return errs.ErrInternal
	}

	// 已签收 → 触发订单状态（auto_confirm 或 delivered）
	if newStatus == ShipStatusDelivered {
		logger.Ctx(ctx).Info("express delivered", zap.Int64("shipment_id", ship.ID))
	}

	return nil
}

// PullStaleShipments 兜底拉取超时未更新的物流信息。
func (s *Service) PullStaleShipments(ctx context.Context) error {
	staleThreshold := time.Now().Add(-24 * time.Hour)
	excludeStatuses := []string{ShipStatusDelivered, ShipStatusException, ShipStatusReturned}
	shipments, err := s.shipRepo.ListStale(ctx, staleThreshold, excludeStatuses)
	if err != nil {
		return errs.ErrInternal
	}

	for _, ship := range shipments {
		ship := ship
		trackResp, err := s.kdniao.Track(ctx, ship.CarrierCode, ship.TrackingNo)
		if err != nil {
			logger.Ctx(ctx).Warn("pull stale track failed",
				zap.Int64("shipment_id", ship.ID), zap.Error(err))
			continue
		}

		now := time.Now()
		ship.LastTrackAt = &now
		newStatus := mapPushState(trackResp.State)
		ship.Status = newStatus
		if newStatus == ShipStatusDelivered && ship.DeliveredAt == nil {
			ship.DeliveredAt = &now
		}

		for _, trace := range trackResp.Traces {
			t := trace
			desc := t.Description
			exists, _ := s.shipRepo.ExistsTrack(ctx, ship.ID, t.OccurredAt, desc)
			if exists {
				continue
			}
			track := &ShipmentTrack{
				ShipmentID:  ship.ID,
				Status:      t.Status,
				Description: &desc,
				OccurredAt:  t.OccurredAt,
			}
			_ = s.shipRepo.CreateTrack(ctx, track)
		}

		if err := s.shipRepo.Update(ctx, &ship); err != nil {
			logger.Ctx(ctx).Warn("update shipment failed", zap.Error(err))
		}
	}

	return nil
}

// ---- 内部辅助 ----

func senderAddrToKD(a SenderAddress) kdniao.Addr {
	company := ""
	if a.Company != nil {
		company = *a.Company
	}
	return kdniao.Addr{
		Company:  company,
		Name:     a.Name,
		Phone:    a.Phone,
		Province: a.Province,
		City:     a.City,
		District: a.District,
		Detail:   joinAddressDetail(a.Street, a.Detail),
	}
}

func orderAddrToKD(a OrderAddrSnap) kdniao.Addr {
	return kdniao.Addr{
		Name:     a.Name,
		Phone:    a.Phone,
		Province: a.Province,
		City:     a.City,
		District: a.District,
		Detail:   joinAddressDetail(a.Street, a.Detail),
	}
}

func joinAddressDetail(street, detail string) string {
	if street == "" {
		return detail
	}
	return street + detail
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func mapPushState(state string) string {
	switch state {
	case "delivered":
		return ShipStatusDelivered
	case "in_transit":
		return ShipStatusInTransit
	case "problem":
		return ShipStatusException
	case "returning", "returned":
		return state
	default:
		return ShipStatusPicked
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
