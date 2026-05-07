package inventory

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/stock"
)

const dirtyKeyPrefix = "inv:dirty:"

// Service 库存服务。
type Service struct {
	repo        InventoryRepo
	stockClient *stock.Client
	rdb         *redis.Client
	asynqClient *asynq.Client
}

// NewService 构造 Service。
func NewService(
	repo InventoryRepo,
	stockClient *stock.Client,
	rdb *redis.Client,
	asynqClient *asynq.Client,
) *Service {
	return &Service{
		repo:        repo,
		stockClient: stockClient,
		rdb:         rdb,
		asynqClient: asynqClient,
	}
}

// Adjust 手动调整库存（changeType: in/out/set）。
// DB 事务提交后再更新 Redis；Redis 失败写 inv:dirty:{sku_id}。
func (s *Service) Adjust(ctx context.Context, adminID, skuID int64, changeType string, change int, reason string) error {
	stock, locked, err := s.repo.GetSKUStock(ctx, skuID)
	if err != nil {
		return errs.ErrNotFound.WithMsg("SKU 不存在")
	}

	switch changeType {
	case "in":
		if err := s.repo.AdjustDB(ctx, skuID, change, "in", "", 0, adminID, reason); err != nil {
			logger.Ctx(ctx).Error("adjust in db", zap.Error(err))
			return errs.ErrInternal
		}
		if redisErr := s.stockClient.AdjustIn(ctx, skuID, change); redisErr != nil {
			logger.Ctx(ctx).Warn("adjust in redis failed, mark dirty",
				zap.Int64("sku_id", skuID), zap.Error(redisErr))
			_ = s.rdb.Set(ctx, dirtyKeyPrefix+fmt.Sprintf("%d", skuID), "1", 0).Err()
		}

	case "out":
		newStock := stock - change
		if newStock < locked {
			return errs.ErrParam.WithMsg(fmt.Sprintf("库存不足：当前 %d，锁定 %d，不能减到 %d", stock, locked, newStock))
		}
		if newStock < 0 {
			return errs.ErrParam.WithMsg("库存不能为负")
		}
		if err := s.repo.AdjustDB(ctx, skuID, -change, "out", "", 0, adminID, reason); err != nil {
			logger.Ctx(ctx).Error("adjust out db", zap.Error(err))
			return errs.ErrInternal
		}
		result, redisErr := s.stockClient.AdjustOut(ctx, skuID, change)
		if redisErr != nil || result == "insufficient" {
			logger.Ctx(ctx).Warn("adjust out redis failed, mark dirty",
				zap.Int64("sku_id", skuID), zap.Error(redisErr))
			_ = s.rdb.Set(ctx, dirtyKeyPrefix+fmt.Sprintf("%d", skuID), "1", 0).Err()
		}

	case "set":
		// 新库存必须 >= locked_stock
		if change < locked {
			return errs.ErrParam.WithMsg(fmt.Sprintf("设置库存 %d 不能低于锁定数量 %d", change, locked))
		}
		delta := change - stock
		if err := s.repo.AdjustDB(ctx, skuID, delta, "set", "", 0, adminID, reason); err != nil {
			logger.Ctx(ctx).Error("adjust set db", zap.Error(err))
			return errs.ErrInternal
		}
		available := change - locked
		if redisErr := s.stockClient.Set(ctx, skuID, available); redisErr != nil {
			logger.Ctx(ctx).Warn("adjust set redis failed, mark dirty",
				zap.Int64("sku_id", skuID), zap.Error(redisErr))
			_ = s.rdb.Set(ctx, dirtyKeyPrefix+fmt.Sprintf("%d", skuID), "1", 0).Err()
		}

	default:
		return errs.ErrParam.WithMsg("changeType 仅支持 in/out/set")
	}

	// 检查低库存预警
	newStock, _, _ := s.repo.GetSKUStock(ctx, skuID)
	_ = s.CheckLowStock(ctx, skuID, newStock)

	return nil
}

// CheckLowStock 检查低库存并创建预警（已有未读预警时跳过）。
func (s *Service) CheckLowStock(ctx context.Context, skuID int64, currentStock int) error {
	threshold, err := s.repo.GetSKUThreshold(ctx, skuID)
	if err != nil || threshold <= 0 {
		return nil
	}
	if currentStock > threshold {
		return nil
	}
	// 已有未读预警则跳过
	has, err := s.repo.HasUnreadAlert(ctx, skuID)
	if err != nil || has {
		return nil
	}
	alert := &LowStockAlert{
		SkuID:            skuID,
		ThresholdAtAlert: threshold,
		CurrentStock:     currentStock,
		Status:           "unread",
	}
	if createErr := s.repo.CreateAlert(ctx, alert); createErr != nil {
		logger.Ctx(ctx).Error("create low stock alert failed",
			zap.Int64("sku_id", skuID),
			zap.Error(createErr))
	}
	return nil
}

// ListLogs 库存日志分页。
func (s *Service) ListLogs(ctx context.Context, filter LogFilter, page, size int) ([]InventoryLog, int64, error) {
	list, total, err := s.repo.ListLogs(ctx, filter, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	return list, total, nil
}

// ListAlerts 预警列表分页。
func (s *Service) ListAlerts(ctx context.Context, status string, page, size int) ([]LowStockAlert, int64, error) {
	list, total, err := s.repo.ListAlerts(ctx, status, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	return list, total, nil
}

// MarkAlertRead 标记预警已读。
func (s *Service) MarkAlertRead(ctx context.Context, id, adminID int64) error {
	if err := s.repo.MarkAlertRead(ctx, id, adminID); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// MarkAllAlertsRead 批量标记所有未读预警为已读。
func (s *Service) MarkAllAlertsRead(ctx context.Context, adminID int64) error {
	if err := s.repo.MarkAllAlertsRead(ctx, adminID); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// Reconcile 对账：对比 DB 和 Redis 库存，差异写告警。
func (s *Service) Reconcile(ctx context.Context) error {
	rows, err := s.repo.FindAllSKUStocks(ctx)
	if err != nil {
		return errs.ErrInternal
	}

	totalSKUs := len(rows)
	if totalSKUs == 0 {
		return nil
	}

	hitCount := 0
	diffCount := 0

	for _, row := range rows {
		redisStock, getErr := s.stockClient.Get(ctx, row.ID)
		if getErr != nil {
			continue
		}
		hitCount++
		dbAvailable := row.Stock - row.LockedStock
		if redisStock != dbAvailable {
			diffCount++
			logger.Ctx(ctx).Warn("inventory reconcile diff",
				zap.Int64("sku_id", row.ID),
				zap.Int("redis", redisStock),
				zap.Int("db_available", dbAvailable),
			)
			// 写 dirty 标记等待同步
			_ = s.rdb.Set(ctx, dirtyKeyPrefix+fmt.Sprintf("%d", row.ID), "1", 0).Err()
		}
	}

	hitRate := float64(hitCount) / float64(totalSKUs)
	logger.Ctx(ctx).Info("inventory reconcile",
		zap.Float64("hit_rate", hitRate),
		zap.Int("total", totalSKUs),
		zap.Int("hit", hitCount),
		zap.Int("diff", diffCount),
	)

	// <80% 命中率不报警，只尝试加载
	if hitRate < 0.8 {
		for _, row := range rows {
			available := row.Stock - row.LockedStock
			if available < 0 {
				available = 0
			}
			_ = s.stockClient.Load(ctx, row.ID, available)
		}
	}

	return nil
}
