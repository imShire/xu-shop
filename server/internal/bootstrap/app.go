// Package bootstrap 负责依赖装配（手动 DI）。
package bootstrap

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/xushop/xu-shop/internal/config"
	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// App 是所有依赖的容器。
type App struct {
	DB          *gorm.DB
	Redis       *redis.Client
	AsynqClient *asynq.Client
	AsynqServer *asynq.Server
	Cfg         *config.Config
}

// NewApp 初始化所有依赖。
func NewApp(cfg *config.Config) (*App, error) {
	// 初始化 logger
	pkglogger.Init(cfg.Log.Level)

	// 初始化 snowflake
	snowflake.Init(cfg.App.InstanceID)

	// 初始化 PostgreSQL（GORM + pgx driver）
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	if cfg.App.Env == "dev" {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: open db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("bootstrap: get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// GORM OTel 插件（noop tracer 下零开销）
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, fmt.Errorf("bootstrap: install otelgorm plugin: %w", err)
	}

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	// Redis OTel tracing（noop tracer 下零开销）
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("bootstrap: install redisotel tracing: %w", err)
	}

	// 初始化 asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr: cfg.Asynq.RedisAddr,
		DB:   cfg.Asynq.RedisDB,
	})

	// 初始化 asynq server（Concurrency=10，可按需调整）
	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: cfg.Asynq.RedisAddr,
			DB:   cfg.Asynq.RedisDB,
		},
		asynq.Config{Concurrency: 10},
	)

	return &App{
		DB:          db,
		Redis:       rdb,
		AsynqClient: asynqClient,
		AsynqServer: asynqServer,
		Cfg:         cfg,
	}, nil
}

// Close 释放所有资源。
func (a *App) Close() {
	if a.AsynqClient != nil {
		_ = a.AsynqClient.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
	if a.DB != nil {
		sqlDB, _ := a.DB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}
