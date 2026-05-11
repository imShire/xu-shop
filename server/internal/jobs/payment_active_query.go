package jobs

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/xushop/xu-shop/internal/modules/order"
	"github.com/xushop/xu-shop/internal/modules/payment"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/wxpay"
)

// TaskPaymentActiveQuery asynq 任务类型名（由 periodic scheduler 每 2 分钟触发一次空 payload 任务）。
const TaskPaymentActiveQuery = "payment:active_query"

// pendingQueryBatchSize 单批最多扫描的 pending 订单数。
const pendingQueryBatchSize = 300

// pendingQueryWindowMax 主动查单最大回溯时长（超过此时长认为永久未支付）。
const pendingQueryWindowMax = 24 * time.Hour

// pendingQueryWindowMin 主动查单最小回溯时长（太近的订单跳过，等待正常回调）。
const pendingQueryWindowMin = 2 * time.Minute

// NewPaymentActiveQueryHandler 构造主动查单 Handler。
// periodic scheduler 注册示例：
//
//	scheduler.Register("*/2 * * * *", asynq.NewTask(jobs.TaskPaymentActiveQuery, nil))
func NewPaymentActiveQueryHandler(
	orderRepo order.OrderRepo,
	paymentSvc *payment.Service,
	wxpayClient wxpay.Client,
) asynq.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(5), 5)
	return func(ctx context.Context, t *asynq.Task) error {
		return handlePaymentActiveQuery(ctx, orderRepo, paymentSvc, wxpayClient, limiter)
	}
}

func handlePaymentActiveQuery(
	ctx context.Context,
	orderRepo order.OrderRepo,
	paymentSvc *payment.Service,
	wxpayClient wxpay.Client,
	limiter *rate.Limiter,
) error {
	now := time.Now()
	winStart := now.Add(-pendingQueryWindowMax)
	winEnd := now.Add(-pendingQueryWindowMin)

	orders, err := orderRepo.FindPendingForActiveQuery(ctx, winStart, winEnd, pendingQueryBatchSize)
	if err != nil {
		logger.L().Error("payment:active_query find pending orders failed", zap.Error(err))
		return err
	}

	logger.L().Info("payment:active_query start", zap.Int("count", len(orders)))

	var successCount int
	for _, o := range orders {
		// 5 QPS token bucket：等待令牌，context 取消时退出
		if err := limiter.Wait(ctx); err != nil {
			logger.L().Warn("payment:active_query rate limiter interrupted", zap.Error(err))
			return err
		}

		resp, err := wxpayClient.QueryByOutTradeNo(ctx, o.OrderNo)
		if err != nil {
			logger.L().Warn("payment:active_query wxpay query failed",
				zap.String("order_no", o.OrderNo), zap.Error(err))
			continue
		}

		if resp.TradeState != "SUCCESS" {
			continue
		}

		// 补单：幂等由 payment.Service.HandleQuerySuccess 内部保证
		if err := paymentSvc.HandleQuerySuccess(ctx, o.OrderNo, resp.TransactionID, resp.AmtCents, resp.PaidAt); err != nil {
			logger.L().Warn("payment:active_query handle success failed",
				zap.String("order_no", o.OrderNo), zap.Error(err))
			continue
		}
		successCount++
	}

	logger.L().Info("payment:active_query done",
		zap.Int("scanned", len(orders)),
		zap.Int("patched", successCount))
	return nil
}
