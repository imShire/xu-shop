module github.com/xushop/xu-shop

go 1.26.1

// 阶段 0 由 Codex 执行 `go mod tidy` 后自动填充依赖。
// 关键依赖（参考 docs/arch/00-overview.md 技术栈表）：
//   github.com/gin-gonic/gin
//   gorm.io/gorm + gorm.io/driver/postgres
//   github.com/pressly/goose/v3
//   github.com/redis/go-redis/v9
//   github.com/hibiken/asynq
//   github.com/go-playground/validator/v10
//   github.com/golang-jwt/jwt/v5
//   golang.org/x/crypto/bcrypt
//   github.com/wechatpay-apiv3/wechatpay-go
//   github.com/microcosm-cc/bluemonday
//   github.com/go-redis/redis_rate/v10
//   github.com/bwmarrin/snowflake
//   go.uber.org/zap
//   github.com/spf13/viper
//   github.com/spf13/cobra
//   go.opentelemetry.io/otel
//   github.com/prometheus/client_golang
//   github.com/aliyun/aliyun-oss-go-sdk
//   github.com/pdfcpu/pdfcpu

require (
	github.com/alicebob/miniredis/v2 v2.37.0
	github.com/bwmarrin/snowflake v0.3.0
	github.com/dchest/captcha v1.1.0
	github.com/gin-gonic/gin v1.12.0
	github.com/go-playground/validator/v10 v10.30.2
	github.com/go-redis/redis_rate/v10 v10.0.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/hibiken/asynq v0.26.0
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/go-redis/v9 v9.19.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	go.uber.org/zap v1.28.0
	golang.org/x/crypto v0.50.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

require golang.org/x/image v0.40.0 // indirect

require (
	github.com/aliyun/aliyun-oss-go-sdk v3.0.2+incompatible
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.15.0 // indirect
	github.com/bytedance/sonic/loader v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/minio/crc64nvme v1.0.2 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.95
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tinylib/msgp v1.3.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.1 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.mongodb.org/mongo-driver/v2 v2.5.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.22.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.14.0
	google.golang.org/protobuf v1.36.11 // indirect
)
