package admin

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/modules/account"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/upload"
)

const settingEncPrefix = "enc:"

// fieldSchema 描述一个设置字段的元数据。
type fieldSchema struct {
	isSecret bool
}

// groupSchema 定义所有支持的设置分组及其字段。
var groupSchema = map[string]map[string]fieldSchema{
	"basic": {
		"shop_name":     {isSecret: false},
		"shop_logo":     {isSecret: false},
		"contact_phone": {isSecret: false},
		"icp_no":        {isSecret: false},
	},
	"wxpay": {
		"mch_id":      {isSecret: false},
		"app_id":      {isSecret: false},
		"api_v3_key":  {isSecret: true},
		"serial_no":   {isSecret: false},
		"cert_pem":    {isSecret: true},
		"key_pem":     {isSecret: true},
	},
	"wxlogin": {
		"mp_app_id":     {isSecret: false},
		"mp_app_secret": {isSecret: true},
	},
	"qywx": {
		"corp_id":        {isSecret: false},
		"agent_id":       {isSecret: false},
		"app_secret":     {isSecret: true},
		"robot_webhook":  {isSecret: true},
	},
	"kdniao": {
		"e_business_id": {isSecret: false},
		"api_key":       {isSecret: true},
		"env":           {isSecret: false},
	},
	"sms": {
		"provider":         {isSecret: false},
		"access_key":       {isSecret: false},
		"secret_key":       {isSecret: true},
		"sign_name":        {isSecret: false},
		"verify_template":  {isSecret: false},
	},
	"security": {
		"session_hours":       {isSecret: false},
		"max_login_attempts":  {isSecret: false},
		"admin_pw_min_len":    {isSecret: false},
	},
}

type settingRow struct {
	Key       string    `gorm:"primaryKey;column:key"`
	Value     string    `gorm:"column:value"`
	IsSecret  bool      `gorm:"column:is_secret"`
	UpdatedBy *int64    `gorm:"column:updated_by"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (settingRow) TableName() string { return "system_setting" }

// Service 系统设置服务。
type Service struct {
	roleRepo account.RoleRepo
	upload   *upload.Manager
	db       *gorm.DB
}

// NewService 构造系统设置服务。
func NewService(roleRepo account.RoleRepo, uploadManager *upload.Manager, db *gorm.DB) *Service {
	return &Service{roleRepo: roleRepo, upload: uploadManager, db: db}
}

// GetUploadSettings 获取上传设置。
func (s *Service) GetUploadSettings(ctx context.Context, adminID int64) (*UploadSettingsResp, error) {
	settings, err := s.upload.LoadSettings(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	resp := &UploadSettingsResp{
		Driver:           settings.Driver,
		PublicBaseURL:    settings.PublicBaseURL,
		MaxSizeMB:        settings.MaxSizeMB,
		LocalDir:         settings.LocalDir,
		S3Vendor:         settings.S3Vendor,
		S3Endpoint:       settings.S3Endpoint,
		S3Region:         settings.S3Region,
		S3Bucket:         settings.S3Bucket,
		S3Prefix:         settings.S3Prefix,
		S3ForcePathStyle: settings.S3ForcePathStyle,
		S3AccessKeySet:   settings.S3AccessKey != "",
		S3SecretKeySet:   settings.S3SecretKey != "",
	}
	return resp, nil
}

// UpdateUploadSettings 更新上传设置。
func (s *Service) UpdateUploadSettings(ctx context.Context, adminID int64, req UpdateUploadSettingsReq) error {
	settings, err := s.mergedSettings(ctx, req)
	if err != nil {
		return err
	}
	if err := s.upload.SaveSettings(ctx, settings, adminID); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			return appErr
		}
		return errs.ErrInternal
	}
	return nil
}

// TestUploadSettings 测试上传设置。
func (s *Service) TestUploadSettings(ctx context.Context, adminID int64, req UpdateUploadSettingsReq) error {
	settings, err := s.mergedSettings(ctx, req)
	if err != nil {
		return err
	}
	return s.upload.TestSettings(ctx, settings)
}

// ProbeUploadSettings 使用当前生效配置试传文件。
func (s *Service) ProbeUploadSettings(ctx context.Context, adminID int64, r io.Reader, filename string) (string, error) {
	return s.upload.UploadImage(ctx, "probe", r, filename)
}

func (s *Service) mergedSettings(ctx context.Context, req UpdateUploadSettingsReq) (upload.Settings, error) {
	current, err := s.upload.LoadSettings(ctx)
	if err != nil {
		return upload.Settings{}, errs.ErrInternal
	}

	current.Driver = req.Driver
	current.PublicBaseURL = req.PublicBaseURL
	current.MaxSizeMB = req.MaxSizeMB
	current.LocalDir = req.LocalDir
	current.S3Vendor = req.S3Vendor
	current.S3Endpoint = req.S3Endpoint
	current.S3Region = req.S3Region
	current.S3Bucket = req.S3Bucket
	current.S3Prefix = req.S3Prefix
	current.S3ForcePathStyle = req.S3ForcePathStyle
	if req.S3AccessKey != "" {
		current.S3AccessKey = req.S3AccessKey
	}
	if req.S3SecretKey != "" {
		current.S3SecretKey = req.S3SecretKey
	}
	return current, nil
}

// GetSettings 读取指定分组的所有设置项，机密字段非空时返回空字符串（掩码）。
func (s *Service) GetSettings(ctx context.Context, group string) (map[string]string, error) {
	fields, ok := groupSchema[group]
	if !ok {
		return nil, errs.ErrParam.WithMsg("unknown group")
	}

	prefix := group + "."
	var rows []settingRow
	if err := s.db.WithContext(ctx).
		Where(`key LIKE ?`, prefix+"%").
		Find(&rows).Error; err != nil {
		return nil, errs.ErrInternal
	}

	// 将 DB 行索引为 short key → raw value
	stored := make(map[string]string, len(rows))
	for _, row := range rows {
		shortKey := strings.TrimPrefix(row.Key, prefix)
		stored[shortKey] = row.Value
	}

	result := make(map[string]string, len(fields))
	for fieldName, meta := range fields {
		raw := stored[fieldName]
		if meta.isSecret {
			// 机密字段：只要 DB 中有值（含加密值）即掩码
			result[fieldName] = ""
		} else {
			result[fieldName] = raw
		}
	}
	return result, nil
}

// UpdateSettings 批量 upsert 指定分组的设置项。
// 机密字段空字符串 = 跳过（保留现有值）；非机密字段空字符串 = 保存为空。
func (s *Service) UpdateSettings(ctx context.Context, group string, data map[string]string, adminID int64) error {
	fields, ok := groupSchema[group]
	if !ok {
		return errs.ErrParam.WithMsg("unknown group")
	}

	secretKey := os.Getenv("APP_SECRET_KEY")
	updatedBy := adminID
	now := time.Now()

	var rows []settingRow
	for fieldName, meta := range fields {
		val, provided := data[fieldName]
		if meta.isSecret && val == "" {
			// 机密字段：空 = 跳过
			continue
		}
		if !provided {
			// 未提供的字段也跳过
			continue
		}

		storedVal := val
		if meta.isSecret && val != "" {
			encrypted, err := settingEncrypt(secretKey, val)
			if err != nil {
				return errs.ErrInternal
			}
			storedVal = encrypted
		}

		rows = append(rows, settingRow{
			Key:       group + "." + fieldName,
			Value:     storedVal,
			IsSecret:  meta.isSecret,
			UpdatedBy: &updatedBy,
			UpdatedAt: now,
		})
	}

	if len(rows) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "is_secret", "updated_by", "updated_at"}),
		}).
		Create(&rows).Error
}

// settingEncrypt AES-GCM 加密，与 upload 包使用相同算法。
func settingEncrypt(secretKey, value string) (string, error) {
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
	return settingEncPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// settingDecrypt AES-GCM 解密（供将来扩展使用）。
func settingDecrypt(secretKey, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, settingEncPrefix) {
		return value, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, settingEncPrefix))
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
