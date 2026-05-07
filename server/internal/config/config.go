// Package config 实现 12-factor 配置，通过 viper 从环境变量加载。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是全局配置结构。
type Config struct {
	App    AppConfig
	DB     DBConfig
	Redis  RedisConfig
	Asynq  AsynqConfig
	JWT    JWTConfig
	WxMP   WxAppConfig // 小程序
	WxOA   WxAppConfig // 公众号
	WxPay  WxPayConfig
	OSS    OSSConfig
	KDNiao KDNiaoConfig
	QYWx   QYWxConfig
	Log    LogConfig
}

type AppConfig struct {
	Env        string
	Port       string
	InstanceID int64
	SecretKey  string
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AsynqConfig struct {
	RedisAddr string
	RedisDB   int
}

type JWTConfig struct {
	Secret        string
	UserTTL       time.Duration // access token
	AdminTTL      time.Duration
	UserRefreshTTL  time.Duration
	AdminRefreshTTL time.Duration
}

type WxAppConfig struct {
	AppID     string
	AppSecret string
}

type WxPayConfig struct {
	MchID     string
	APIKeyV3  string
	CertPath  string
	KeyPath   string
	NotifyURL string
	Env       string
}

type OSSConfig struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	AccessKeySecret string
	CDNDomain       string
}

type KDNiaoConfig struct {
	BusinessID string
	APIKey     string
	ReqURL     string
}

type QYWxConfig struct {
	CorpID  string
	AgentID string
	Secret  string
}

type LogConfig struct {
	Level string
}

// Load 从环境变量读取配置，校验必填项。
func Load() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	v.SetDefault("APP_ENV", "dev")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("INSTANCE_ID", 1)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("ASYNQ_REDIS_DB", 1)
	v.SetDefault("JWT_USER_TTL_HOURS", 720)
	v.SetDefault("JWT_ADMIN_TTL_HOURS", 8)
	v.SetDefault("JWT_USER_REFRESH_TTL_HOURS", 2160)  // 90 天
	v.SetDefault("JWT_ADMIN_REFRESH_TTL_HOURS", 168)  // 7 天

	cfg := &Config{}

	cfg.App.Env = v.GetString("APP_ENV")
	cfg.App.Port = v.GetString("APP_PORT")
	cfg.App.InstanceID = v.GetInt64("INSTANCE_ID")
	cfg.App.SecretKey = v.GetString("APP_SECRET_KEY")

	cfg.DB.DSN = v.GetString("DB_DSN")

	cfg.Redis.Addr = v.GetString("REDIS_ADDR")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = v.GetInt("REDIS_DB")

	cfg.Asynq.RedisAddr = v.GetString("ASYNQ_REDIS_ADDR")
	if cfg.Asynq.RedisAddr == "" {
		cfg.Asynq.RedisAddr = cfg.Redis.Addr
	}
	cfg.Asynq.RedisDB = v.GetInt("ASYNQ_REDIS_DB")

	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.UserTTL = time.Duration(v.GetInt("JWT_USER_TTL_HOURS")) * time.Hour
	cfg.JWT.AdminTTL = time.Duration(v.GetInt("JWT_ADMIN_TTL_HOURS")) * time.Hour
	cfg.JWT.UserRefreshTTL = time.Duration(v.GetInt("JWT_USER_REFRESH_TTL_HOURS")) * time.Hour
	cfg.JWT.AdminRefreshTTL = time.Duration(v.GetInt("JWT_ADMIN_REFRESH_TTL_HOURS")) * time.Hour

	cfg.WxMP.AppID = v.GetString("WXMP_APPID")
	cfg.WxMP.AppSecret = v.GetString("WXMP_APPSECRET")
	cfg.WxOA.AppID = v.GetString("WXOA_APPID")
	cfg.WxOA.AppSecret = v.GetString("WXOA_APPSECRET")

	cfg.WxPay.MchID = v.GetString("WXPAY_MCHID")
	cfg.WxPay.APIKeyV3 = v.GetString("WXPAY_APIKEYV3")
	cfg.WxPay.CertPath = v.GetString("WXPAY_CERTPATH")
	cfg.WxPay.KeyPath = v.GetString("WXPAY_KEYPATH")
	cfg.WxPay.NotifyURL = v.GetString("WXPAY_NOTIFYURL")
	cfg.WxPay.Env = v.GetString("WXPAY_ENV")

	cfg.OSS.Endpoint = v.GetString("OSS_ENDPOINT")
	cfg.OSS.Region = v.GetString("OSS_REGION")
	cfg.OSS.Bucket = v.GetString("OSS_BUCKET")
	cfg.OSS.AccessKeyID = v.GetString("OSS_AK")
	cfg.OSS.AccessKeySecret = v.GetString("OSS_SK")
	cfg.OSS.CDNDomain = v.GetString("OSS_CDN_DOMAIN")

	cfg.KDNiao.BusinessID = v.GetString("KDNIAO_BUSINESS_ID")
	cfg.KDNiao.APIKey = v.GetString("KDNIAO_API_KEY")
	cfg.KDNiao.ReqURL = v.GetString("KDNIAO_REQ_URL")

	cfg.QYWx.CorpID = v.GetString("QYWX_CORPID")
	cfg.QYWx.AgentID = v.GetString("QYWX_AGENTID")
	cfg.QYWx.Secret = v.GetString("QYWX_SECRET")

	cfg.Log.Level = v.GetString("LOG_LEVEL")

	// 校验必填项
	if cfg.DB.DSN == "" {
		return nil, fmt.Errorf("config: DB_DSN is required")
	}
	if cfg.Redis.Addr == "" {
		return nil, fmt.Errorf("config: REDIS_ADDR is required")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}

	return cfg, nil
}
