// Package main 是 asynq worker 服务入口。
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/bootstrap"
	"github.com/xushop/xu-shop/internal/config"
	"github.com/xushop/xu-shop/internal/jobs"
	"github.com/xushop/xu-shop/internal/modules/account"
	"github.com/xushop/xu-shop/internal/modules/address"
	"github.com/xushop/xu-shop/internal/modules/inventory"
	"github.com/xushop/xu-shop/internal/modules/notification"
	"github.com/xushop/xu-shop/internal/modules/order"
	"github.com/xushop/xu-shop/internal/modules/payment"
	"github.com/xushop/xu-shop/internal/modules/product"
	"github.com/xushop/xu-shop/internal/modules/shipping"
	"github.com/xushop/xu-shop/internal/modules/stats"
	pkgkdniao "github.com/xushop/xu-shop/internal/pkg/kdniao"
	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/stock"
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
	"github.com/xushop/xu-shop/internal/pkg/wxsubscribe"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("config: " + err.Error())
	}

	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		panic("bootstrap: " + err.Error())
	}
	defer app.Close()

	// ---- 装配依赖 ----

	skuRepo := product.NewSKURepo(app.DB)
	productRepo := product.NewProductRepo(app.DB)
	invRepo := inventory.NewInventoryRepo(app.DB)
	addrRepo := address.NewAddressRepo(app.DB)
	stockClient := stock.New(app.Redis)
	orderRepo := order.NewOrderRepo(app.DB)
	userRepo := account.NewUserRepo(app.DB)
	orderSvc := order.NewService(orderRepo, skuRepo, productRepo, invRepo, addrRepo, stockClient, app.Redis, app.AsynqClient, userRepo, nil)

	// 微信支付客户端
	var wxpayClient pkgwxpay.Client
	if pkgwxpay.IsMockMode() {
		wxpayClient = pkgwxpay.NewMockClient()
	} else {
		wxpayClient, err = pkgwxpay.NewRealClient(pkgwxpay.Config{
			MchID:           cfg.WxPay.MchID,
			APIKeyV3:        cfg.WxPay.APIKeyV3,
			CertPath:        cfg.WxPay.CertPath,
			KeyPath:         cfg.WxPay.KeyPath,
			AppIDMP:         cfg.WxMP.AppID,
			AppIDOA:         cfg.WxOA.AppID,
			AppIDH5:         cfg.WxOA.AppID,
			NotifyURL:       cfg.WxPay.NotifyURL,
			RefundNotifyURL: cfg.WxPay.NotifyURL + "/refund",
		})
		if err != nil {
			pkglogger.L().Warn("wxpay real client init failed, using mock", zap.Error(err))
			wxpayClient = pkgwxpay.NewMockClient()
		}
	}

	// 快递鸟客户端
	var kdniaoClient pkgkdniao.Client
	if os.Getenv("MOCK_KDNIAO") == "true" || cfg.KDNiao.BusinessID == "" {
		kdniaoClient = pkgkdniao.NewMockClient()
	} else {
		kdniaoClient = pkgkdniao.NewRealClient(cfg.KDNiao.BusinessID, cfg.KDNiao.APIKey, cfg.KDNiao.ReqURL)
	}

	// payment 服务（worker 只需 ApplyRefund）
	paymentRepo := payment.NewPaymentRepo(app.DB)
	payEnqueuer := &workerAsynqEnqueuer{client: app.AsynqClient}
	workerOrderAccessor := &workerOrderAccessorAdapter{repo: orderRepo, svc: orderSvc, db: app.DB}
	paymentSvc := payment.NewService(paymentRepo, workerOrderAccessor, nil, wxpayClient, payEnqueuer,
		payment.WxPaySceneConfig{
			AppIDMP: cfg.WxMP.AppID,
			AppIDOA: cfg.WxOA.AppID,
			AppIDH5: cfg.WxOA.AppID,
			MchID:   cfg.WxPay.MchID,
		})

	// shipping 服务（worker 只需 PullStaleShipments）
	senderRepo := shipping.NewSenderAddressRepo(app.DB)
	carrierRepo := shipping.NewCarrierRepo(app.DB)
	shipmentRepo := shipping.NewShipmentRepo(app.DB)
	shipOrderAccessor := &workerShipOrderAccessor{svc: orderSvc}
	shippingSvc := shipping.NewService(senderRepo, carrierRepo, shipmentRepo, kdniaoClient,
		shipOrderAccessor, app.Redis, payEnqueuer, cfg.WxPay.NotifyURL+"/express")

	// 订单任务适配器
	closeAdapter := &orderCloseAdapter{svc: orderSvc}
	confirmAdapter := &orderConfirmAdapter{svc: orderSvc}

	// notification 服务（worker 需要 HandleSend）
	var wxsubClient wxsubscribe.Client
	if os.Getenv("MOCK_WXSUBSCRIBE") == "true" || cfg.WxMP.AppSecret == "" {
		wxsubClient = wxsubscribe.NewMockClient()
	} else {
		wxsubClient = wxsubscribe.NewClient(cfg.WxMP.AppID, cfg.WxMP.AppSecret, app.Redis)
	}
	notifRepo := notification.NewNotificationRepo(app.DB)
	notifSvc := notification.NewService(notifRepo, wxsubClient, app.Redis)

	// stats 服务（worker 需要 AggregateDaily）
	statsRepo := stats.NewStatsRepo(app.DB)
	statsSvc := stats.NewService(statsRepo)

	// ---- 注册 asynq mux ----
	mux := asynq.NewServeMux()
	mux.Handle(jobs.TaskOrderClose, jobs.NewOrderCloseHandler(closeAdapter))
	mux.Handle(jobs.TaskOrderAutoConfirm, jobs.NewOrderAutoConfirmHandler(confirmAdapter))
	mux.Handle(jobs.TaskPaymentAutoRefund, jobs.NewPaymentAutoRefundHandler(paymentSvc))
	mux.Handle(jobs.TaskExpressPull, jobs.NewExpressPullHandler(shippingSvc))
	mux.Handle(jobs.TaskNotificationSend, jobs.NewNotificationSendHandler(notifSvc))
	mux.Handle(jobs.TaskStatsAggregate, jobs.NewStatsAggregateHandler(statsSvc))
	mux.Handle(jobs.TaskPaymentActiveQuery, jobs.NewPaymentActiveQueryHandler(orderRepo, paymentSvc, wxpayClient))

	// ---- asynq Scheduler：定时投递 periodic 任务 ----
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr: cfg.Asynq.RedisAddr,
			DB:   cfg.Asynq.RedisDB,
		},
		nil,
	)
	// 每 2 分钟触发一次主动查单（payload 为空）
	if _, err := scheduler.Register("*/2 * * * *", asynq.NewTask(jobs.TaskPaymentActiveQuery, nil)); err != nil {
		pkglogger.L().Fatal("scheduler register payment:active_query failed", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		pkglogger.L().Info("worker starting")
		if err := app.AsynqServer.Run(mux); err != nil {
			pkglogger.L().Fatal("worker run error", zap.Error(err))
		}
	}()

	go func() {
		pkglogger.L().Info("scheduler starting")
		if err := scheduler.Run(); err != nil {
			pkglogger.L().Fatal("scheduler run error", zap.Error(err))
		}
	}()

	<-quit
	pkglogger.L().Info("worker shutting down...")
	scheduler.Shutdown()
	app.AsynqServer.Shutdown()
	pkglogger.L().Info("worker stopped")
}

// ---- 适配器 ----

type orderCloseAdapter struct{ svc *order.Service }

func (a *orderCloseAdapter) GetRaw(ctx context.Context, orderID int64) (jobs.OrderInfo, error) {
	o, err := a.svc.GetRaw(ctx, orderID)
	if err != nil {
		return jobs.OrderInfo{}, err
	}
	return jobs.OrderInfo{
		Status:   o.Status,
		OrderNo:  o.OrderNo,
		ExpireAt: o.ExpireAt,
	}, nil
}

func (a *orderCloseAdapter) Transition(ctx context.Context, orderID int64, trigger, operatorType string, operatorID int64, reason string) error {
	_, err := a.svc.Transition(ctx, orderID, trigger, operatorType, operatorID, reason)
	return err
}

func (a *orderCloseAdapter) ReleaseStock(ctx context.Context, orderID int64, orderNo string) {
	a.svc.ReleaseStock(ctx, orderID, orderNo)
}

func (a *orderCloseAdapter) RefundUserBalance(ctx context.Context, orderID int64) error {
	return a.svc.RefundUserBalance(ctx, orderID)
}

type orderConfirmAdapter struct{ svc *order.Service }

func (a *orderConfirmAdapter) GetRaw(ctx context.Context, orderID int64) (jobs.OrderInfo, error) {
	o, err := a.svc.GetRaw(ctx, orderID)
	if err != nil {
		return jobs.OrderInfo{}, err
	}
	return jobs.OrderInfo{
		Status:   o.Status,
		OrderNo:  o.OrderNo,
		ExpireAt: time.Time{},
	}, nil
}

func (a *orderConfirmAdapter) Transition(ctx context.Context, orderID int64, trigger, operatorType string, operatorID int64, reason string) error {
	_, err := a.svc.Transition(ctx, orderID, trigger, operatorType, operatorID, reason)
	return err
}

type workerAsynqEnqueuer struct{ client *asynq.Client }

func (a *workerAsynqEnqueuer) EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return a.client.EnqueueContext(ctx, task, opts...)
}

type workerOrderAccessorAdapter struct {
	repo order.OrderRepo
	svc  *order.Service
	db   *gorm.DB
}

func (a *workerOrderAccessorAdapter) DB() *gorm.DB { return a.db }

func (a *workerOrderAccessorAdapter) FindByOrderNo(ctx context.Context, orderNo string) (*payment.OrderSnapshot, error) {
	o, err := a.repo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	return &payment.OrderSnapshot{
		ID: o.ID, OrderNo: o.OrderNo, UserID: o.UserID,
		Status: o.Status, PayCents: o.PayCents, ExpireAt: o.ExpireAt,
		CurrentPrepayID: o.CurrentPrepayID, CurrentPrepayExpireAt: o.CurrentPrepayExpireAt,
	}, nil
}

func (a *workerOrderAccessorAdapter) FindByID(ctx context.Context, id int64) (*payment.OrderSnapshot, error) {
	o, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &payment.OrderSnapshot{
		ID: o.ID, OrderNo: o.OrderNo, UserID: o.UserID,
		Status: o.Status, PayCents: o.PayCents, ExpireAt: o.ExpireAt,
		CurrentPrepayID: o.CurrentPrepayID, CurrentPrepayExpireAt: o.CurrentPrepayExpireAt,
	}, nil
}

func (a *workerOrderAccessorAdapter) SetPrepayID(ctx context.Context, orderID int64, prepayID string, expireAt time.Time) error {
	return a.db.WithContext(ctx).Table(`"order"`).Where("id = ?", orderID).
		Updates(map[string]any{
			"current_prepay_id":        prepayID,
			"current_prepay_expire_at": expireAt,
			"updated_at":               time.Now(),
		}).Error
}

func (a *workerOrderAccessorAdapter) Transition(ctx context.Context, orderID int64, trigger, opType string, opID int64, reason string) error {
	_, err := a.svc.Transition(ctx, orderID, trigger, opType, opID, reason)
	return err
}

func (a *workerOrderAccessorAdapter) TransitionInTx(ctx context.Context, tx *gorm.DB, orderID int64, trigger, opType string, opID int64, reason string) error {
	return a.svc.TransitionInTx(ctx, tx, orderID, trigger, opType, opID, reason)
}

func (a *workerOrderAccessorAdapter) DeductStock(ctx context.Context, orderID int64, orderNo string) {
	a.svc.DeductStock(ctx, orderID, orderNo)
}

type workerShipOrderAccessor struct{ svc *order.Service }

func (a *workerShipOrderAccessor) FindByID(ctx context.Context, id int64) (*shipping.OrderSnap, error) {
	o, err := a.svc.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	return &shipping.OrderSnap{
		ID: o.ID, OrderNo: o.OrderNo, UserID: o.UserID, Status: o.Status,
		AddressSnapshot: shipping.OrderAddrSnap{
			Name: o.AddressSnapshot.Name, Phone: o.AddressSnapshot.Phone,
			Province: o.AddressSnapshot.Province, City: o.AddressSnapshot.City,
			District: o.AddressSnapshot.District, Detail: o.AddressSnapshot.Detail,
		},
	}, nil
}

func (a *workerShipOrderAccessor) Transition(ctx context.Context, orderID int64, trigger, opType string, opID int64, reason string) error {
	_, err := a.svc.Transition(ctx, orderID, trigger, opType, opID, reason)
	return err
}
