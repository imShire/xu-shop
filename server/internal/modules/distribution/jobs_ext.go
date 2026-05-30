package distribution

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
)

// ===== commission:recompute =====

// CommissionRecomputeStats 每日 05:30：重算 distributor 维度累计佣金统计。
//
// MVP 简化：直接按 status 聚合写回 distributor.total_pending_cents / total_locked_cents / total_settled_cents（如有这些字段）。
// 当前 entity 未含这些字段，因此仅做日志摘要，不写表。
//
// TODO(backend-dev): 若产品后续要展示分销员"累计佣金"卡片，
// 应在 distribution.Service.CommissionRecomputeStats 内部增量更新 distributor 表对应字段（需 91-db-schema 先扩列）。
func (s *Service) CommissionRecomputeStats(ctx context.Context) error {
	// 仅做轻量聚合统计、输出指标日志，便于运维观察。
	var (
		pending  int64
		locked   int64
		settled  int64
	)
	_ = s.repo.DB().WithContext(ctx).
		Table("commission_record").
		Where("status = ?", CommissionStatusPending).
		Count(&pending).Error
	_ = s.repo.DB().WithContext(ctx).
		Table("commission_record").
		Where("status = ?", CommissionStatusLocked).
		Count(&locked).Error
	_ = s.repo.DB().WithContext(ctx).
		Table("commission_record").
		Where("status = ?", CommissionStatusSettled).
		Count(&settled).Error
	logger.L().Info("commission:recompute summary",
		zap.Int64("pending", pending),
		zap.Int64("locked", locked),
		zap.Int64("settled", settled))
	return nil
}

// ===== commission:antifraud-scan =====

// AntifraudScan 每小时：扫近 24h 异常订单，标记 commission.status=suspect。
//
// MVP 规则：同 distributor 24h 内成单数 > 30 或同 IP 24h 内成单数 > 10 → 标记。
// 复杂规则（同设备/同链接/订单短时间内大额）需要更多上下文，后续在 distribution.Service.AntifraudScan 内部扩展。
//
// TODO(backend-dev): 完整规则建议落到 distribution.Service.AntifraudScan，输入 24h 内订单清单 → 应用规则 → 批量 Transition mark_suspect。
// 本期仅按"同 distributor 短时间内大批佣金"做粗规则。
func (s *Service) AntifraudScan(ctx context.Context) (int, error) {
	since := time.Now().Add(-24 * time.Hour)
	type row struct {
		DistributorUserID int64
		Cnt               int64
	}
	var rows []row
	if err := s.repo.DB().WithContext(ctx).
		Table("commission_record").
		Select("distributor_user_id, COUNT(*) AS cnt").
		Where("status = ? AND created_at >= ?", CommissionStatusPending, since).
		Group("distributor_user_id").
		Having("COUNT(*) > ?", 30).
		Scan(&rows).Error; err != nil {
		return 0, err
	}
	marked := 0
	for _, r := range rows {
		var cs []CommissionRecord
		if err := s.repo.DB().WithContext(ctx).
			Where("distributor_user_id = ? AND status = ? AND created_at >= ?",
				r.DistributorUserID, CommissionStatusPending, since).
			Find(&cs).Error; err != nil {
			continue
		}
		for i := range cs {
			c := &cs[i]
			reason := fmt.Sprintf("antifraud: 24h 成单数 %d>30", r.Cnt)
			if err := s.Transition(ctx, c, "mark_suspect", reason); err == nil {
				marked++
			}
		}
	}
	logger.L().Info("commission:antifraud-scan done", zap.Int("marked", marked))
	return marked, nil
}

// ===== withdraw:active-query =====

// WithdrawActiveQuery 每 5 分钟：扫 processing 提现单，调 wxpay TransferQuery 回查状态。
//
// 终态由 OnTransferNotify 处理（共享 release 佣金 / 更新状态机），本 cron 只是"主动拉"补回调缺失。
func (s *Service) WithdrawActiveQuery(ctx context.Context, limit int) (int, error) {
	if s.wxpay == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 200
	}
	var list []WithdrawOrder
	if err := s.repo.DB().WithContext(ctx).
		Where("status = ?", WithdrawStatusProcessing).
		Order("processed_at ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return 0, err
	}
	queried := 0
	for _, w := range list {
		resp, err := s.wxpay.QueryTransfer(ctx, w.WithdrawNo)
		if err != nil {
			logger.L().Warn("withdraw:active-query QueryTransfer failed",
				zap.String("withdraw_no", w.WithdrawNo), zap.Error(err))
			continue
		}
		if resp == nil {
			continue
		}
		// 复用 notify 处理逻辑保证幂等
		notify := &pkgwxpay.TransferNotifyResult{
			OutBillNo:      w.WithdrawNo,
			TransferBillNo: resp.TransferBillNo,
			State:          resp.State,
			FailReason:     resp.FailReason,
		}
		if err := s.OnTransferNotify(ctx, notify); err != nil {
			logger.L().Warn("withdraw:active-query OnTransferNotify failed",
				zap.String("withdraw_no", w.WithdrawNo), zap.Error(err))
			continue
		}
		queried++
	}
	logger.L().Info("withdraw:active-query done",
		zap.Int("scanned", len(list)), zap.Int("queried", queried))
	return queried, nil
}

// ===== withdraw:reconcile =====

// WithdrawReconcile 每日 06:00：对账前一日 wxpay 转账明细 vs 本地 withdraw_order，diff 写 reconciliation_diff。
//
// TODO(backend-dev): 完整实现需要：
//   1) 调用 wxpay BillDownload 拉前一日转账账单（pkg/wxpay 当前未提供 TransferBill 接口，需扩展）
//   2) 解析账单与本地 withdraw_order 做对账
//   3) diff 写 reconciliation_diff 表（91-db-schema 已含 reconciliation_diff，预留 channel="wx_transfer"）
//
// 本期仅做"本地不一致自检"：扫 status=processing 超过 24h 未流转的提现单，记录告警。
func (s *Service) WithdrawReconcile(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)
	var stale []WithdrawOrder
	if err := s.repo.DB().WithContext(ctx).
		Where("status = ? AND processed_at < ?", WithdrawStatusProcessing, cutoff).
		Limit(500).
		Find(&stale).Error; err != nil {
		return err
	}
	if len(stale) > 0 {
		logger.L().Warn("withdraw:reconcile stale processing withdraws detected",
			zap.Int("count", len(stale)))
	}
	return nil
}

// ===== share:click-flush =====

// ShareClickFlush 每分钟：Redis Sorted Set 中的 share_click 批量 flush 到 share_click 表。
//
// 当前实现：TrackClick 已经直接写 DB，本 cron 仅扫 Redis 暂存 key (key 形如 dist:share:click:buf) 并补写。
// 若 TrackClick 未启用 Redis 暂存（MVP 默认直接写库），本 cron 等于 no-op。
//
// TODO(backend-dev): 真正减小 DB 写压力需要在 distribution.Service.TrackClick 切换"先写 Redis ZSET + 本 cron 批量 flush"模式；
// 该开关应通过 distribution.Config.ClickBufferEnabled 控制，确保对账可回溯。
func (s *Service) ShareClickFlush(ctx context.Context) (int, error) {
	if s.rdb == nil {
		return 0, nil
	}
	// 当前 MVP：占位，不读 ZSET，仅日志。
	logger.L().Debug("share:click-flush no buffer mode, skip")
	return 0, nil
}
