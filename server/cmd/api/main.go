// Package main 是 HTTP API 服务入口。
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/bootstrap"
	"github.com/xushop/xu-shop/internal/config"
	"github.com/xushop/xu-shop/internal/docs"
	"github.com/xushop/xu-shop/internal/middleware"
	"github.com/xushop/xu-shop/internal/modules/account"
	"github.com/xushop/xu-shop/internal/modules/address"
	adminmod "github.com/xushop/xu-shop/internal/modules/admin"
	"github.com/xushop/xu-shop/internal/modules/aftersale"
	"github.com/xushop/xu-shop/internal/modules/banner"
	"github.com/xushop/xu-shop/internal/modules/cart"
	"github.com/xushop/xu-shop/internal/modules/cms"
	"github.com/xushop/xu-shop/internal/modules/decorate"
	"github.com/xushop/xu-shop/internal/modules/distribution"
	"github.com/xushop/xu-shop/internal/modules/inventory"
	"github.com/xushop/xu-shop/internal/modules/marketing"
	nav_icon "github.com/xushop/xu-shop/internal/modules/nav_icon"
	"github.com/xushop/xu-shop/internal/modules/notification"
	"github.com/xushop/xu-shop/internal/modules/order"
	"github.com/xushop/xu-shop/internal/modules/payment"
	privatedomain "github.com/xushop/xu-shop/internal/modules/private_domain"
	"github.com/xushop/xu-shop/internal/modules/product"
	"github.com/xushop/xu-shop/internal/modules/recall"
	"github.com/xushop/xu-shop/internal/modules/shipping"
	"github.com/xushop/xu-shop/internal/modules/stats"
	"github.com/xushop/xu-shop/internal/modules/tag"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
	pkgkdniao "github.com/xushop/xu-shop/internal/pkg/kdniao"
	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
	pkgoss "github.com/xushop/xu-shop/internal/pkg/oss"
	"github.com/xushop/xu-shop/internal/pkg/poster"
	"github.com/xushop/xu-shop/internal/pkg/qywx"
	"github.com/xushop/xu-shop/internal/pkg/stock"
	"github.com/xushop/xu-shop/internal/pkg/tracer"
	"github.com/xushop/xu-shop/internal/pkg/wxacode"
	pkgupload "github.com/xushop/xu-shop/internal/pkg/upload"
	pkgvalidator "github.com/xushop/xu-shop/internal/pkg/validator"
	"github.com/xushop/xu-shop/internal/pkg/wxlogin"
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
	"github.com/xushop/xu-shop/internal/pkg/wxsubscribe"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic("config: " + err.Error())
	}

	// 2. 装配依赖（DB / Redis / Asynq）
	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		panic("bootstrap: " + err.Error())
	}
	defer app.Close()

	// 2.1 初始化 tracer（OTEL_EXPORTER_OTLP_ENDPOINT 为空则 noop）
	tracerCtx, tracerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, tracerShutdown, err := tracer.Init(tracerCtx, "xu-shop-api")
	tracerCancel()
	if err != nil {
		pkglogger.L().Warn("tracer init failed, fallback to noop", zap.Error(err))
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = tracerShutdown(ctx)
		}()
	}

	// 3. 初始化 gin
	if cfg.App.Env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 注册自定义 binding 校验器（mobile / strongpwd）
	pkgvalidator.Setup()

	if err := r.SetTrustedProxies([]string{
		"127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12",
	}); err != nil {
		pkglogger.L().Warn("set trusted proxies", zap.Error(err))
	}

	// 4. 全局中间件
	r.Use(middleware.Recovery())
	r.Use(otelgin.Middleware("xu-shop-api"))
	r.Use(middleware.Logging())
	r.Use(middleware.PrometheusMiddleware())
	r.Use(middleware.CORS([]string{"http://localhost:3000", "http://localhost:5273"}))

	// 5. 健康检查 & metrics & 文档
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// Swagger UI：/docs  OpenAPI YAML：/openapi.yaml
	docs.RegisterRoutes(r)

	// 6. JWT 配置
	jwtCfg := pkgjwt.Config{
		Secret:      cfg.JWT.Secret,
		UserExpiry:  cfg.JWT.UserTTL,
		AdminExpiry: cfg.JWT.AdminTTL,
	}

	// 7. 微信登录客户端
	var wxMP, wxOA wxlogin.WxLoginClient
	if wxlogin.IsMockMode() {
		wxMP = wxlogin.NewMockClient()
		wxOA = wxlogin.NewMockClient()
	} else {
		wxMP = wxlogin.NewClient(cfg.WxMP.AppID, cfg.WxMP.AppSecret)
		wxOA = wxlogin.NewClient(cfg.WxOA.AppID, cfg.WxOA.AppSecret)
	}

	// 8. 微信支付客户端
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

	// 9. 快递鸟客户端
	var kdniaoClient pkgkdniao.Client
	if os.Getenv("MOCK_KDNIAO") == "true" || cfg.KDNiao.BusinessID == "" {
		kdniaoClient = pkgkdniao.NewMockClient()
	} else {
		kdniaoClient = pkgkdniao.NewRealClient(cfg.KDNiao.BusinessID, cfg.KDNiao.APIKey, cfg.KDNiao.ReqURL)
	}

	// 10. 注册业务路由
	v1 := r.Group("/api/v1")
	{
		// account 模块
		accountUserRepo := account.NewUserRepo(app.DB)
		accountAdminRepo := account.NewAdminRepo(app.DB)
		accountRoleRepo := account.NewRoleRepo(app.DB)
		accountSvc := account.NewService(accountUserRepo, accountAdminRepo, accountRoleRepo, app.Redis, cfg, wxMP, wxOA)
		accountHandler := account.NewHandler(accountSvc, jwtCfg, cfg.App.Env == "prod", cfg.Auth.H5RedirectAllowList)
		account.RegisterRoutes(v1, accountHandler, app.Redis, app.DB, jwtCfg)
		uploadManager := pkgupload.NewManager(app.DB, cfg.App.SecretKey)
		adminSvc := adminmod.NewService(accountRoleRepo, uploadManager, app.DB)
		var ossClient *pkgoss.Client
		if cfg.OSS.AccessKeyID != "" && cfg.OSS.Bucket != "" {
			ossClient, err = pkgoss.New(cfg)
			if err != nil {
				pkglogger.L().Warn("oss client init failed", zap.Error(err))
			}
		}
		adminHandler := adminmod.NewHandler(adminSvc, ossClient)
		adminmod.RegisterRoutes(v1, adminHandler, app.Redis, app.DB, jwtCfg)
		r.GET("/uploads/*filepath", adminHandler.ServeLocalFile)

		// address 模块
		addrRepo := address.NewAddressRepo(app.DB)
		regionRepo := address.NewRegionRepo(app.DB)
		addrSvc := address.NewService(addrRepo, regionRepo, app.Redis, wxMP)
		addrHandler := address.NewHandler(addrSvc)
		address.RegisterRoutes(v1, addrHandler, app.Redis, app.DB, jwtCfg)

		// product 模块
		productCategoryRepo := product.NewCategoryRepo(app.DB)
		productRepo := product.NewProductRepo(app.DB)
		skuRepo := product.NewSKURepo(app.DB)
		favRepo := product.NewFavoriteRepo(app.DB)
		viewRepo := product.NewViewHistoryRepo(app.DB)
		stockClient := stock.New(app.Redis)
		productSvc := product.NewService(productRepo, productCategoryRepo, skuRepo, favRepo, viewRepo, app.Redis, uploadManager, app.AsynqClient)
		productHandler := product.NewHandler(productSvc)
		product.RegisterRoutes(v1, productHandler, app.Redis, app.DB, jwtCfg)

		// inventory 模块
		invRepo := inventory.NewInventoryRepo(app.DB)
		invSvc := inventory.NewService(invRepo, stockClient, app.Redis, app.AsynqClient)
		invHandler := inventory.NewHandler(invSvc)
		inventory.RegisterRoutes(v1, invHandler, app.Redis, app.DB, jwtCfg)

		// cart 模块
		cartRepo := cart.NewCartRepo(app.DB)
		cartSvc := cart.NewService(cartRepo, skuRepo, productRepo, stockClient)
		cartHandler := cart.NewHandler(cartSvc)
		cart.RegisterRoutes(v1, cartHandler, app.Redis, app.DB, jwtCfg)

		// order 模块
		orderRepo := order.NewOrderRepo(app.DB)
		orderSvc := order.NewService(orderRepo, skuRepo, productRepo, invRepo, addrRepo, stockClient, app.Redis, app.AsynqClient, accountUserRepo, cartRepo)
		orderHandler := order.NewHandler(orderSvc)
		order.RegisterRoutes(v1, orderHandler, app.Redis, app.DB, jwtCfg)

		// payment 模块
		paymentRepo := payment.NewPaymentRepo(app.DB)
		orderAccessor := &orderAccessorAdapter{repo: orderRepo, svc: orderSvc, db: app.DB}
		userOpenid := &userOpenidAdapter{userRepo: accountUserRepo, wxMP: cfg.WxMP.AppID, wxOA: cfg.WxOA.AppID}
		paySceneCfg := payment.WxPaySceneConfig{
			AppIDMP: cfg.WxMP.AppID,
			AppIDOA: cfg.WxOA.AppID,
			AppIDH5: cfg.WxOA.AppID,
			MchID:   cfg.WxPay.MchID,
		}
		payEnqueuer := &asynqEnqueuerAdapter{client: app.AsynqClient}
		paymentSvc := payment.NewService(paymentRepo, orderAccessor, userOpenid, wxpayClient, payEnqueuer, paySceneCfg).
			WithOrphanRetry(payment.NewOrphanRetryRepo(app.DB))
		paymentHandler := payment.NewHandler(paymentSvc)
		payment.RegisterRoutes(v1, paymentHandler, app.Redis, app.DB, jwtCfg)

		// shipping 模块
		senderRepo := shipping.NewSenderAddressRepo(app.DB)
		carrierRepo := shipping.NewCarrierRepo(app.DB)
		shipmentRepo := shipping.NewShipmentRepo(app.DB)
		shipOrderAccessor := &shipOrderAccessorAdapter{svc: orderSvc}
		shippingSvc := shipping.NewService(senderRepo, carrierRepo, shipmentRepo, kdniaoClient,
			shipOrderAccessor, app.Redis, payEnqueuer,
			cfg.WxPay.NotifyURL+"/express") // 复用 notifyURL 前缀
		shippingHandler := shipping.NewHandler(shippingSvc)
		shipping.RegisterRoutes(v1, shippingHandler, app.Redis, app.DB, jwtCfg)

		// aftersale 模块
		aftersaleRepo := aftersale.NewRepo(app.DB)
		aftersaleOrderAcc := &aftersaleOrderAccessor{db: app.DB}
		aftersaleLegacyRepo := &aftersaleLegacyOrderRepo{db: app.DB}
		aftersaleSvc := aftersale.NewService(aftersaleRepo, aftersaleOrderAcc, paymentSvc, app.DB, app.Redis).
			WithLegacyOrderRepo(aftersaleLegacyRepo)
		aftersaleHandler := aftersale.NewHandler(aftersaleSvc)
		aftersale.RegisterRoutes(v1, aftersaleHandler, app.Redis, app.DB, jwtCfg)

		// notification 模块
		var wxsubClient wxsubscribe.Client
		if os.Getenv("MOCK_WXSUBSCRIBE") == "true" || cfg.WxMP.AppSecret == "" {
			wxsubClient = wxsubscribe.NewMockClient()
		} else {
			wxsubClient = wxsubscribe.NewClient(cfg.WxMP.AppID, cfg.WxMP.AppSecret, app.Redis)
		}
		notifRepo := notification.NewNotificationRepo(app.DB)
		notifSvc := notification.NewService(notifRepo, wxsubClient, app.Redis)
		notifHandler := notification.NewHandler(notifSvc)
		notification.RegisterRoutes(v1, notifHandler, app.Redis, app.DB, jwtCfg)

		// private_domain 模块
		var qywxClient qywx.Client
		if os.Getenv("MOCK_QYWX") == "true" || cfg.QYWx.CorpID == "" {
			qywxClient = qywx.NewMockClient()
		} else {
			qywxClient = qywx.NewClient(cfg.QYWx.CorpID, cfg.QYWx.Secret, os.Getenv("QYWX_ENCODING_AES_KEY"))
		}
		channelRepo := privatedomain.NewChannelCodeRepo(app.DB)
		tagRepo := privatedomain.NewTagRepo(app.DB)
		userTagRepo := privatedomain.NewUserTagRepo(app.DB)
		shareRepo := privatedomain.NewShareRepo(app.DB)
		pdSvc := privatedomain.NewService(channelRepo, tagRepo, userTagRepo, shareRepo, qywxClient, app.Redis)
		// 注入海报依赖（poster + wxacode + oss + product finder）；任一为空将退回占位符
		posterRenderer := poster.NewImageRenderer()
		wxacodeClient := wxacode.NewRealClient(cfg, app.Redis)
		if ossClient != nil {
			pdSvc = pdSvc.WithPosterDeps(&posterProductFinderAdapter{repo: productRepo}, wxacodeClient, posterRenderer, ossClient)
		}
		pdHandler := privatedomain.NewHandler(pdSvc)
		privatedomain.RegisterRoutes(v1, pdHandler, app.Redis, app.DB, jwtCfg)

		// marketing 模块（coupon + point + member）
		mkSvc := marketing.NewService(app.DB)
		marketing.RegisterRoutes(v1, mkSvc, app.Redis, app.DB, jwtCfg)

		// tag 模块（用户标签）
		tagSvc := tag.NewService(tag.NewRepo(app.DB), app.DB)
		tagH := tag.NewHandler(tagSvc)
		tag.RegisterRoutes(v1, tagH, app.Redis, app.DB, jwtCfg)

		// recall 模块（召回活动）
		recallSvc := recall.NewService(recall.NewRepo(app.DB), app.DB, app.Redis, tagSvc, mkSvc.Coupon, notifSvc)
		recallH := recall.NewHandler(recallSvc)
		recall.RegisterRoutes(v1, recallH, app.Redis, app.DB, jwtCfg)

		// distribution 模块（分销 / 分享溯源 / 提现）
		// 微信支付转账客户端：复用 wxpayClient 的底层实现
		var transferClient pkgwxpay.TransferClient
		if tc, ok := wxpayClient.(pkgwxpay.TransferClient); ok {
			transferClient = tc
		} else {
			transferClient = &pkgwxpay.MockClient{}
		}
		distOpenid := &distOpenidAdapter{userRepo: accountUserRepo, wxMP: cfg.WxMP.AppID, wxOA: cfg.WxOA.AppID}
		distSms := &distSmsAdapter{rdb: app.Redis, userRepo: accountUserRepo}
		distCfg := distribution.Config{
			TransferAppID:  cfg.WxMP.AppID,
			TransferScene:  os.Getenv("WXPAY_TRANSFER_SCENE"),
			TransferNotify: cfg.WxPay.NotifyURL + "/transfer",
			ShareBaseURL:   os.Getenv("SHARE_BASE_URL"),
		}
		distSvc := distribution.NewService(distribution.NewRepo(app.DB), app.Redis, transferClient, distOpenid, distSms, nil, distCfg)
		distH := distribution.NewHandler(distSvc, transferClient)
		distribution.RegisterRoutes(v1, distH, app.Redis, app.DB, jwtCfg, r)

		// stats 模块
		statsRepo := stats.NewStatsRepo(app.DB)
		statsSvc := stats.NewService(statsRepo)
		statsHandler := stats.NewHandler(statsSvc)
		stats.RegisterRoutes(v1, statsHandler, app.Redis, app.DB, jwtCfg)

		// banner 模块
		bannerRepo := banner.NewBannerRepo(app.DB)
		bannerSvc := banner.NewService(bannerRepo)
		bannerHandler := banner.NewHandler(bannerSvc)
		banner.RegisterRoutes(v1, bannerHandler, app.Redis, app.DB, jwtCfg)

		// nav_icon 模块
		navIconRepo := nav_icon.NewNavIconRepo(app.DB)
		navIconSvc := nav_icon.NewService(navIconRepo)
		navIconHandler := nav_icon.NewHandler(navIconSvc)
		nav_icon.RegisterRoutes(v1, navIconHandler, app.Redis, app.DB, jwtCfg)

		// cms 模块
		cmsRepo := cms.NewArticleRepo(app.DB)
		cmsSvc := cms.NewService(cmsRepo)
		cmsHandler := cms.NewHandler(cmsSvc)
		cms.RegisterRoutes(v1, cmsHandler, app.Redis, app.DB, jwtCfg)

		// decorate 模块
		decorateRepo := decorate.NewPageConfigRepo(app.DB)
		decorateSvc := decorate.NewService(decorateRepo)
		decorateHandler := decorate.NewHandler(decorateSvc)
		decorate.RegisterRoutes(v1, decorateHandler, app.Redis, app.DB, jwtCfg)
	}

	// 11. 启动 HTTP 服务
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		pkglogger.L().Info("api server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkglogger.L().Fatal("listen and serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	pkglogger.L().Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		pkglogger.L().Error("server shutdown error", zap.Error(err))
	}
	pkglogger.L().Info("server stopped")
}

// ---- 适配器 ----

// orderAccessorAdapter 将 order.OrderRepo + order.Service 适配为 payment.OrderAccessor。
type orderAccessorAdapter struct {
	repo order.OrderRepo
	svc  *order.Service
	db   *gorm.DB
}

func (a *orderAccessorAdapter) DB() *gorm.DB { return a.db }

func (a *orderAccessorAdapter) FindByOrderNo(ctx context.Context, orderNo string) (*payment.OrderSnapshot, error) {
	o, err := a.repo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	return orderToPaySnapshot(o), nil
}

func (a *orderAccessorAdapter) FindByID(ctx context.Context, id int64) (*payment.OrderSnapshot, error) {
	o, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return orderToPaySnapshot(o), nil
}

func (a *orderAccessorAdapter) SetPrepayID(ctx context.Context, orderID int64, prepayID string, expireAt time.Time) error {
	return a.db.WithContext(ctx).Table(`"order"`).Where("id = ?", orderID).
		Updates(map[string]any{
			"current_prepay_id":        prepayID,
			"current_prepay_expire_at": expireAt,
			"updated_at":               time.Now(),
		}).Error
}

func (a *orderAccessorAdapter) Transition(ctx context.Context, orderID int64, trigger, opType string, opID int64, reason string) error {
	_, err := a.svc.Transition(ctx, orderID, trigger, opType, opID, reason)
	return err
}

func (a *orderAccessorAdapter) TransitionInTx(ctx context.Context, tx *gorm.DB, orderID int64, trigger, opType string, opID int64, reason string) error {
	return a.svc.TransitionInTx(ctx, tx, orderID, trigger, opType, opID, reason)
}

func (a *orderAccessorAdapter) DeductStock(ctx context.Context, orderID int64, orderNo string) {
	a.svc.DeductStock(ctx, orderID, orderNo)
}

func orderToPaySnapshot(o *order.Order) *payment.OrderSnapshot {
	return &payment.OrderSnapshot{
		ID:                    o.ID,
		OrderNo:               o.OrderNo,
		UserID:                o.UserID,
		Status:                o.Status,
		PayCents:              o.PayCents,
		ExpireAt:              o.ExpireAt,
		CurrentPrepayID:       o.CurrentPrepayID,
		CurrentPrepayExpireAt: o.CurrentPrepayExpireAt,
	}
}

// userOpenidAdapter 将 account.UserRepo 适配为 payment.UserOpenidGetter。
type userOpenidAdapter struct {
	userRepo account.UserRepo
	wxMP     string
	wxOA     string
}

func (a *userOpenidAdapter) GetOpenid(ctx context.Context, userID int64, scene string) (string, error) {
	u, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	switch scene {
	case "jsapi_mp":
		if u.OpenidMP != nil {
			return *u.OpenidMP, nil
		}
	case "jsapi_oa":
		if u.OpenidH5 != nil {
			return *u.OpenidH5, nil
		}
	}
	return "", nil
}

// asynqEnqueuerAdapter 将 *asynq.Client 适配为 payment.AsynqEnqueuer/shipping.AsynqEnqueuer。
type asynqEnqueuerAdapter struct {
	client *asynq.Client
}

func (a *asynqEnqueuerAdapter) EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return a.client.EnqueueContext(ctx, task, opts...)
}

// shipOrderAccessorAdapter 将 order.Service 适配为 shipping.OrderAccessor。
type shipOrderAccessorAdapter struct {
	svc *order.Service
}

func (a *shipOrderAccessorAdapter) FindByID(ctx context.Context, id int64) (*shipping.OrderSnap, error) {
	o, err := a.svc.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	snap := &shipping.OrderSnap{
		ID:      o.ID,
		OrderNo: o.OrderNo,
		UserID:  o.UserID,
		Status:  o.Status,
		AddressSnapshot: shipping.OrderAddrSnap{
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
		},
	}
	return snap, nil
}

func (a *shipOrderAccessorAdapter) Transition(ctx context.Context, orderID int64, trigger, opType string, opID int64, reason string) error {
	_, err := a.svc.Transition(ctx, orderID, trigger, opType, opID, reason)
	return err
}

// aftersaleOrderAccessor 实现 aftersale.OrderAccessor。
// 直接走 *gorm.DB，避免与 order 模块的内部 DTO 耦合。
// RefundedCents 通过 refund 表 status='success' 求和得到。
type aftersaleOrderAccessor struct {
	db *gorm.DB
}

func (a *aftersaleOrderAccessor) FindOrder(ctx context.Context, orderID int64) (*aftersale.OrderInfo, error) {
	var row struct {
		ID          int64      `gorm:"column:id"`
		OrderNo     string     `gorm:"column:order_no"`
		UserID      int64      `gorm:"column:user_id"`
		Status      string     `gorm:"column:status"`
		PayCents    int64      `gorm:"column:pay_cents"`
		ShippedAt   *time.Time `gorm:"column:shipped_at"`
		PaidAt      *time.Time `gorm:"column:paid_at"`
	}
	if err := a.db.WithContext(ctx).Table(`"order"`).
		Select("id", "order_no", "user_id", "status", "pay_cents", "shipped_at", "paid_at").
		Where("id = ?", orderID).Take(&row).Error; err != nil {
		return nil, err
	}
	var refunded int64
	if err := a.db.WithContext(ctx).Table("refund").
		Where("order_id = ? AND status = ?", orderID, "success").
		Select("COALESCE(SUM(amount_cents), 0)").Scan(&refunded).Error; err != nil {
		return nil, err
	}
	return &aftersale.OrderInfo{
		ID:            row.ID,
		OrderNo:       row.OrderNo,
		UserID:        row.UserID,
		Status:        row.Status,
		PayCents:      row.PayCents,
		RefundedCents: refunded,
		DeliveredAt:   row.ShippedAt,
		PaidAt:        row.PaidAt,
	}, nil
}

func (a *aftersaleOrderAccessor) FindOrderItem(ctx context.Context, itemID int64) (*aftersale.OrderItemInfo, error) {
	var row struct {
		ID         int64 `gorm:"column:id"`
		OrderID    int64 `gorm:"column:order_id"`
		PriceCents int64 `gorm:"column:price_cents"`
		Qty        int   `gorm:"column:qty"`
	}
	if err := a.db.WithContext(ctx).Table("order_item").
		Select("id", "order_id", "price_cents", "qty").
		Where("id = ?", itemID).Take(&row).Error; err != nil {
		return nil, err
	}
	return &aftersale.OrderItemInfo{
		ID:            row.ID,
		OrderID:       row.OrderID,
		SubtotalCents: row.PriceCents * int64(row.Qty),
	}, nil
}

// aftersaleLegacyOrderRepo 实现 aftersale.LegacyOrderRepo，
// 服务旧 cancel_request_pending 流程的后台处理。
type aftersaleLegacyOrderRepo struct {
	db *gorm.DB
}

func (a *aftersaleLegacyOrderRepo) ListAftersale(ctx context.Context, page, size int) ([]aftersale.LegacyOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var rows []struct {
		ID                   int64      `gorm:"column:id"`
		OrderNo              string     `gorm:"column:order_no"`
		UserID               int64      `gorm:"column:user_id"`
		Status               string     `gorm:"column:status"`
		PayCents             int64      `gorm:"column:pay_cents"`
		CancelRequestPending bool       `gorm:"column:cancel_request_pending"`
		CancelRequestReason  *string    `gorm:"column:cancel_request_reason"`
		CancelRequestAt      *time.Time `gorm:"column:cancel_request_at"`
	}
	q := a.db.WithContext(ctx).Table(`"order"`).
		Where(`cancel_request_pending = true OR status IN ('refunding','refunded')`)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]aftersale.LegacyOrder, len(rows))
	for i, r := range rows {
		result[i] = aftersale.LegacyOrder{
			ID:                   r.ID,
			OrderNo:              r.OrderNo,
			UserID:               r.UserID,
			Status:               r.Status,
			PayCents:             r.PayCents,
			CancelRequestPending: r.CancelRequestPending,
			CancelRequestReason:  r.CancelRequestReason,
			CancelRequestAt:      r.CancelRequestAt,
		}
	}
	return result, total, nil
}

func (a *aftersaleLegacyOrderRepo) FindByID(ctx context.Context, id int64) (*aftersale.LegacyOrder, error) {
	var row struct {
		ID                   int64      `gorm:"column:id"`
		OrderNo              string     `gorm:"column:order_no"`
		UserID               int64      `gorm:"column:user_id"`
		Status               string     `gorm:"column:status"`
		PayCents             int64      `gorm:"column:pay_cents"`
		CancelRequestPending bool       `gorm:"column:cancel_request_pending"`
		CancelRequestReason  *string    `gorm:"column:cancel_request_reason"`
		CancelRequestAt      *time.Time `gorm:"column:cancel_request_at"`
	}
	err := a.db.WithContext(ctx).Table(`"order"`).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &aftersale.LegacyOrder{
		ID:                   row.ID,
		OrderNo:              row.OrderNo,
		UserID:               row.UserID,
		Status:               row.Status,
		PayCents:             row.PayCents,
		CancelRequestPending: row.CancelRequestPending,
		CancelRequestReason:  row.CancelRequestReason,
		CancelRequestAt:      row.CancelRequestAt,
	}, nil
}

func (a *aftersaleLegacyOrderRepo) UpdateCancelRequest(ctx context.Context, id int64, pending bool, _ *time.Time) error {
	return a.db.WithContext(ctx).Table(`"order"`).Where("id = ?", id).
		Updates(map[string]any{
			"cancel_request_pending": pending,
			"updated_at":             time.Now(),
		}).Error
}

// suppress unused imports
var _ = fmt.Sprintf

// distOpenidAdapter 实现 distribution.UserOpenidGetter。
// 优先返回 MP openid 与公众平台 appID（WxMP），无则回落 H5（WxOA）。
type distOpenidAdapter struct {
	userRepo account.UserRepo
	wxMP     string
	wxOA     string
}

func (a *distOpenidAdapter) GetOpenidForTransfer(ctx context.Context, userID int64) (string, string, error) {
	u, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if u.OpenidMP != nil && *u.OpenidMP != "" {
		return *u.OpenidMP, a.wxMP, nil
	}
	if u.OpenidH5 != nil && *u.OpenidH5 != "" {
		return *u.OpenidH5, a.wxOA, nil
	}
	return "", "", fmt.Errorf("user %d has no openid for transfer", userID)
}

// distSmsAdapter 实现 distribution.SmsSender。
// 开发阶段：仅打日志（参考 account.SendSmsCode 的 dev 行为）。
type distSmsAdapter struct {
	rdb      *redis.Client
	userRepo account.UserRepo
}

func (a *distSmsAdapter) SendWithdrawSms(ctx context.Context, userID int64, code string) error {
	u, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	phone := ""
	if u.Phone != nil {
		phone = *u.Phone
	}
	pkglogger.L().Info("withdraw sms (dev)",
		zap.Int64("user_id", userID),
		zap.String("phone", phone),
		zap.String("code", code),
	)
	return nil
}

// posterProductFinderAdapter 将 product.ProductRepo 适配为 private_domain.PosterProductFinder。
type posterProductFinderAdapter struct {
	repo product.ProductRepo
}

func (a *posterProductFinderAdapter) FindProductForPoster(ctx context.Context, id int64) (*privatedomain.PosterProductInfo, error) {
	p, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &privatedomain.PosterProductInfo{
		Title:         p.Title,
		MainImage:     p.MainImage,
		PriceMinCents: p.PriceMinCents,
	}, nil
}
