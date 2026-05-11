package upload

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

const (
	driverLocal          = "local"
	driverS3             = "s3"
	defaultLocalDir      = "uploads"
	defaultMaxSizeMB     = 10
	defaultPublicBaseURL = "http://localhost:8080/uploads"
	encryptionPrefix     = "enc:"
)

var mimeToExt = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/gif":  "gif",
}

// allowedImageExtRegexp 文件名后缀白名单（大小写不敏感，禁止 SVG 等）。
var allowedImageExtRegexp = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|gif|webp)$`)

// Settings 上传配置。
type Settings struct {
	Driver           string
	PublicBaseURL    string
	MaxSizeMB        int
	LocalDir         string
	S3Vendor         string
	S3Endpoint       string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3Prefix         string
	S3ForcePathStyle bool
}

// PresignResult 直传签名结果。
type PresignResult struct {
	Method    string            `json:"method"`
	UploadURL string            `json:"upload_url"`
	Headers   map[string]string `json:"headers"`
	ObjectKey string            `json:"object_key"`
	FileURL   string            `json:"file_url"`
	ExpiresIn int64             `json:"expires_in"`
}

type systemSetting struct {
	Key       string    `gorm:"primaryKey;column:key"`
	Value     string    `gorm:"column:value"`
	IsSecret  bool      `gorm:"column:is_secret"`
	UpdatedBy *int64    `gorm:"column:updated_by"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (systemSetting) TableName() string { return "system_setting" }

// Manager 负责上传配置持久化与上传执行。
type Manager struct {
	db        *gorm.DB
	secretKey string
}

// NewManager 构造上传中心。
func NewManager(db *gorm.DB, secretKey string) *Manager {
	return &Manager{db: db, secretKey: secretKey}
}

// LoadSettings 读取上传配置。
func (m *Manager) LoadSettings(ctx context.Context) (Settings, error) {
	settings := defaultSettings()

	var rows []systemSetting
	if err := m.db.WithContext(ctx).
		Where(`key LIKE ?`, "upload.%").
		Find(&rows).Error; err != nil {
		return settings, err
	}

	for _, row := range rows {
		value := row.Value
		if row.IsSecret && value != "" {
			decoded, err := decryptString(m.secretKey, value)
			if err != nil {
				return settings, fmt.Errorf("decrypt %s: %w", row.Key, err)
			}
			value = decoded
		}
		switch row.Key {
		case "upload.driver":
			settings.Driver = value
		case "upload.public_base_url":
			settings.PublicBaseURL = value
		case "upload.max_size_mb":
			if n, err := strconv.Atoi(value); err == nil && n > 0 {
				settings.MaxSizeMB = n
			}
		case "upload.local.dir":
			settings.LocalDir = value
		case "upload.s3.vendor":
			settings.S3Vendor = value
		case "upload.s3.endpoint":
			settings.S3Endpoint = value
		case "upload.s3.region":
			settings.S3Region = value
		case "upload.s3.bucket":
			settings.S3Bucket = value
		case "upload.s3.access_key":
			settings.S3AccessKey = value
		case "upload.s3.secret_key":
			settings.S3SecretKey = value
		case "upload.s3.prefix":
			settings.S3Prefix = value
		case "upload.s3.force_path_style":
			settings.S3ForcePathStyle = value == "true"
		}
	}

	if settings.Driver == "" {
		settings.Driver = driverLocal
	}
	if settings.MaxSizeMB <= 0 {
		settings.MaxSizeMB = defaultMaxSizeMB
	}
	if settings.PublicBaseURL == "" {
		settings.PublicBaseURL = defaultPublicBaseURL
	}
	if settings.LocalDir == "" {
		settings.LocalDir = defaultLocalDir
	}

	return settings, nil
}

// SaveSettings 保存上传配置。
func (m *Manager) SaveSettings(ctx context.Context, settings Settings, adminID int64) error {
	if err := validateSettings(settings); err != nil {
		return err
	}

	updatedBy := adminID
	rows := []systemSetting{
		newSetting("upload.driver", settings.Driver, false, &updatedBy),
		newSetting("upload.public_base_url", strings.TrimRight(settings.PublicBaseURL, "/"), false, &updatedBy),
		newSetting("upload.max_size_mb", strconv.Itoa(settings.MaxSizeMB), false, &updatedBy),
		newSetting("upload.local.dir", settings.LocalDir, false, &updatedBy),
		newSetting("upload.s3.vendor", settings.S3Vendor, false, &updatedBy),
		newSetting("upload.s3.endpoint", settings.S3Endpoint, false, &updatedBy),
		newSetting("upload.s3.region", settings.S3Region, false, &updatedBy),
		newSetting("upload.s3.bucket", settings.S3Bucket, false, &updatedBy),
		newSetting("upload.s3.access_key", settings.S3AccessKey, true, &updatedBy),
		newSetting("upload.s3.secret_key", settings.S3SecretKey, true, &updatedBy),
		newSetting("upload.s3.prefix", settings.S3Prefix, false, &updatedBy),
		newSetting("upload.s3.force_path_style", strconv.FormatBool(settings.S3ForcePathStyle), false, &updatedBy),
	}

	for i := range rows {
		if rows[i].IsSecret {
			encrypted, err := encryptString(m.secretKey, rows[i].Value)
			if err != nil {
				return err
			}
			rows[i].Value = encrypted
		}
	}

	return m.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "is_secret", "updated_by", "updated_at"}),
		}).
		Create(&rows).Error
}

// TestSettings 测试配置是否可用。
func (m *Manager) TestSettings(ctx context.Context, settings Settings) error {
	if err := validateSettings(settings); err != nil {
		return err
	}
	switch settings.Driver {
	case driverLocal:
		return os.MkdirAll(settings.LocalDir, 0o755)
	case driverS3:
		client, secure, endpoint, err := buildMinioClient(settings)
		if err != nil {
			return err
		}
		exists, err := client.BucketExists(ctx, settings.S3Bucket)
		if err != nil {
			scheme := "https"
			if !secure {
				scheme = "http"
			}
			return errs.ErrParam.WithMsg(fmt.Sprintf("S3 连接失败：%s://%s: %v", scheme, endpoint, err))
		}
		if !exists {
			return errs.ErrParam.WithMsg("S3 bucket 不存在")
		}
		return nil
	default:
		return errs.ErrParam.WithMsg("不支持的上传驱动")
	}
}

// UploadProductImage 兼容商品模块上传接口。
func (m *Manager) UploadProductImage(ctx context.Context, r io.Reader, filename string) (string, error) {
	return m.UploadImage(ctx, "product", r, filename)
}

// AllowedImageURLPrefix 返回详情图允许的 URL 前缀。
func (m *Manager) AllowedImageURLPrefix(ctx context.Context) (string, error) {
	settings, err := m.LoadSettings(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(settings.PublicBaseURL, "/"), nil
}

// UploadImage 上传图片。
func (m *Manager) UploadImage(ctx context.Context, biz string, r io.Reader, filename string) (string, error) {
	settings, err := m.LoadSettings(ctx)
	if err != nil {
		return "", errs.ErrInternal
	}
	if err := validateSettings(settings); err != nil {
		return "", err
	}

	maxBytes := int64(settings.MaxSizeMB) * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return "", errs.ErrParam.WithMsg("读取上传文件失败")
	}
	if int64(len(body)) > maxBytes {
		return "", errs.ErrParam.WithMsg(fmt.Sprintf("图片不能超过 %dMB", settings.MaxSizeMB))
	}

	// 校验文件名后缀（大小写不敏感）
	if !allowedImageExtRegexp.MatchString(filename) {
		return "", errs.ErrParam.WithMsg("仅支持 jpg/png/webp/gif 图片")
	}

	// 取前 512 字节检测真实 MIME，防止伪造后缀
	mimeType := detectImageMIME(body)
	if mimeType == "" {
		return "", errs.ErrParam.WithMsg("仅支持 jpg/png/webp/gif 图片")
	}

	objectKey := buildObjectKey(biz, filename, mimeType, settings.S3Prefix)
	switch settings.Driver {
	case driverLocal:
		if err := writeLocalFile(settings.LocalDir, objectKey, body); err != nil {
			return "", errs.ErrInternal.WithMsg("本地上传失败")
		}
	case driverS3:
		if err := putS3Object(ctx, settings, objectKey, body, mimeType); err != nil {
			return "", err
		}
	default:
		return "", errs.ErrParam.WithMsg("不支持的上传驱动")
	}

	return strings.TrimRight(settings.PublicBaseURL, "/") + "/" + objectKey, nil
}

// ResolveLocalFile 将 URL path 转成本地文件路径。
func (m *Manager) ResolveLocalFile(ctx context.Context, relativePath string) (string, error) {
	settings, err := m.LoadSettings(ctx)
	if err != nil {
		return "", err
	}
	if settings.Driver != driverLocal {
		return "", errs.ErrNotFound
	}

	clean := path.Clean("/" + strings.TrimPrefix(relativePath, "/"))
	if strings.Contains(clean, "..") {
		return "", errs.ErrNotFound
	}
	absRoot, err := filepath.Abs(settings.LocalDir)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(absRoot, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
	if !strings.HasPrefix(fullPath, absRoot+string(os.PathSeparator)) && fullPath != absRoot {
		return "", errs.ErrNotFound
	}
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errs.ErrNotFound
		}
		return "", err
	}
	return fullPath, nil
}

func defaultSettings() Settings {
	return Settings{
		Driver:        driverLocal,
		PublicBaseURL: defaultPublicBaseURL,
		MaxSizeMB:     defaultMaxSizeMB,
		LocalDir:      defaultLocalDir,
		S3Vendor:      "generic",
	}
}

func newSetting(key, value string, secret bool, updatedBy *int64) systemSetting {
	return systemSetting{
		Key:       key,
		Value:     value,
		IsSecret:  secret,
		UpdatedBy: updatedBy,
		UpdatedAt: time.Now(),
	}
}

func validateSettings(settings Settings) error {
	if settings.Driver != driverLocal && settings.Driver != driverS3 {
		return errs.ErrParam.WithMsg("上传驱动仅支持 local 或 s3")
	}
	if strings.TrimSpace(settings.PublicBaseURL) == "" {
		return errs.ErrParam.WithMsg("请填写上传访问地址")
	}
	if settings.MaxSizeMB < 1 || settings.MaxSizeMB > 50 {
		return errs.ErrParam.WithMsg("上传大小限制需在 1-50MB 之间")
	}
	if settings.Driver == driverLocal {
		if strings.TrimSpace(settings.LocalDir) == "" {
			return errs.ErrParam.WithMsg("请填写本地上传目录")
		}
		return nil
	}
	if strings.TrimSpace(settings.S3Endpoint) == "" {
		return errs.ErrParam.WithMsg("请填写 S3 Endpoint")
	}
	if strings.TrimSpace(settings.S3Bucket) == "" {
		return errs.ErrParam.WithMsg("请填写 S3 Bucket")
	}
	if strings.TrimSpace(settings.S3AccessKey) == "" {
		return errs.ErrParam.WithMsg("请填写 S3 Access Key")
	}
	if strings.TrimSpace(settings.S3SecretKey) == "" {
		return errs.ErrParam.WithMsg("请填写 S3 Secret Key")
	}
	return nil
}

func buildObjectKey(biz, filename, mimeType, prefix string) string {
	now := time.Now()
	ext := mimeToExt[mimeType]
	parts := make([]string, 0, 6)
	if strings.TrimSpace(prefix) != "" {
		parts = append(parts, strings.Trim(prefix, "/"))
	}
	parts = append(parts,
		biz,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		uuid.NewString()+"."+ext,
	)
	return strings.Join(parts, "/")
}

func writeLocalFile(rootDir, objectKey string, body []byte) error {
	fullPath := filepath.Join(rootDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, body, 0o644)
}

func putS3Object(ctx context.Context, settings Settings, objectKey string, body []byte, mimeType string) error {
	client, _, _, err := buildMinioClient(settings)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body)
	_, err = client.PutObject(ctx, settings.S3Bucket, objectKey, reader, int64(len(body)), minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return errs.ErrParam.WithMsg("S3 上传失败：" + err.Error())
	}
	return nil
}

func buildMinioClient(settings Settings) (*minio.Client, bool, string, error) {
	endpoint := strings.TrimSpace(settings.S3Endpoint)
	secure := true
	if strings.Contains(endpoint, "://") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, false, "", errs.ErrParam.WithMsg("S3 Endpoint 格式错误")
		}
		secure = u.Scheme != "http"
		endpoint = u.Host
	}
	if endpoint == "" {
		return nil, false, "", errs.ErrParam.WithMsg("S3 Endpoint 格式错误")
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(settings.S3AccessKey, settings.S3SecretKey, ""),
		Secure:       secure,
		Region:       settings.S3Region,
		BucketLookup: bucketLookupType(settings.S3ForcePathStyle),
	})
	if err != nil {
		return nil, false, "", errs.ErrParam.WithMsg("初始化 S3 客户端失败：" + err.Error())
	}
	return client, secure, endpoint, nil
}

func bucketLookupType(forcePathStyle bool) minio.BucketLookupType {
	if forcePathStyle {
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

// detectImageMIME 读取前 512 字节，用 http.DetectContentType 检测真实 MIME，
// 仅返回白名单类型（image/jpeg、image/png、image/gif、image/webp），其余返回空字符串。
func detectImageMIME(body []byte) string {
	sample := body
	if len(sample) > 512 {
		sample = sample[:512]
	}
	detected := http.DetectContentType(sample)
	// 去掉可能的参数部分（如 "text/plain; charset=utf-8"）
	if idx := strings.IndexByte(detected, ';'); idx >= 0 {
		detected = strings.TrimSpace(detected[:idx])
	}
	// 仅允许白名单 MIME，image/svg+xml 等不在 mimeToExt 中，自动拒绝
	if _, ok := mimeToExt[detected]; ok {
		return detected
	}
	return ""
}

func encryptString(secretKey, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	key := sha256.Sum256([]byte(secretKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return encryptionPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptString(secretKey, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, encryptionPrefix) {
		return value, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, encryptionPrefix))
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(secretKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("cipher text too short")
	}
	nonce := raw[:gcm.NonceSize()]
	plain, err := gcm.Open(nil, nonce, raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
