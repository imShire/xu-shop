// Package oss 封装阿里云 OSS 上传操作及 STS 临时凭证获取。
package oss

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	osssdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"

	"github.com/xushop/xu-shop/internal/config"
)

// allowedImageMIME 允许上传的图片 MIME 类型。
var allowedImageMIME = map[string]string{
	"\xff\xd8\xff":    "image/jpeg",
	"\x89PNG\r\n":     "image/png",
	"RIFF":            "image/webp", // 需结合后续字节
	"GIF87a":          "image/gif",
	"GIF89a":          "image/gif",
	"\x00\x00\x00\x0c": "", // 可能是 HEIC，跳过
}

var mimeToExt = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/gif":  "gif",
}

// Client OSS 操作封装。
type Client struct {
	bucket          *osssdk.Bucket
	cdnDomain       string
	bucketName      string
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	stsRoleArn      string
	stsEndpoint     string
}

// Uploader 上传接口，便于 mock。
type Uploader interface {
	UploadProductImage(ctx context.Context, r io.Reader, filename string) (string, error)
}

// New 初始化 OSS Client。
func New(cfg *config.Config) (*Client, error) {
	if cfg.OSS.Endpoint == "" || cfg.OSS.Bucket == "" {
		return nil, fmt.Errorf("oss: endpoint and bucket are required")
	}
	cli, err := osssdk.New(cfg.OSS.Endpoint, cfg.OSS.AccessKeyID, cfg.OSS.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss: new client: %w", err)
	}
	bucket, err := cli.Bucket(cfg.OSS.Bucket)
	if err != nil {
		return nil, fmt.Errorf("oss: get bucket: %w", err)
	}
	stsEp := cfg.OSS.STSEndpoint
	if stsEp == "" {
		stsEp = "https://sts.aliyuncs.com"
	}
	return &Client{
		bucket:          bucket,
		cdnDomain:       strings.TrimRight(cfg.OSS.CDNDomain, "/"),
		bucketName:      cfg.OSS.Bucket,
		endpoint:        cfg.OSS.Endpoint,
		accessKeyID:     cfg.OSS.AccessKeyID,
		accessKeySecret: cfg.OSS.AccessKeySecret,
		stsRoleArn:      cfg.OSS.STSRoleArn,
		stsEndpoint:     stsEp,
	}, nil
}

// UploadProductImage 上传商品图，校验 mime（jpeg/png/webp/gif），返回 CDN URL。
func (c *Client) UploadProductImage(_ context.Context, r io.Reader, filename string) (string, error) {
	// 1. 读前 512 字节检测 mime
	header := make([]byte, 512)
	n, err := io.ReadFull(r, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("oss: read header: %w", err)
	}
	header = header[:n]

	mimeType := detectImageMIME(header)
	if mimeType == "" {
		return "", fmt.Errorf("oss: unsupported image type")
	}

	ext := mimeToExt[mimeType]
	if ext == "" {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	}

	// 2. 拼回完整内容
	fullReader := io.MultiReader(bytes.NewReader(header), r)

	// 3. 生成 OSS path: product/YYYY/MM/DD/{uuid}.{ext}
	now := time.Now()
	ossKey := fmt.Sprintf("product/%04d/%02d/%02d/%s.%s",
		now.Year(), now.Month(), now.Day(),
		uuid.New().String(), ext)

	// 4. 上传
	if err := c.bucket.PutObject(ossKey, fullReader); err != nil {
		return "", fmt.Errorf("oss: put object: %w", err)
	}

	return c.cdnDomain + "/" + ossKey, nil
}

// detectImageMIME 根据文件头魔数检测图片 MIME。
func detectImageMIME(header []byte) string {
	if len(header) < 4 {
		return ""
	}
	h := string(header)
	switch {
	case strings.HasPrefix(h, "\xff\xd8\xff"):
		return "image/jpeg"
	case strings.HasPrefix(h, "\x89PNG\r\n"):
		return "image/png"
	case strings.HasPrefix(h, "GIF87a") || strings.HasPrefix(h, "GIF89a"):
		return "image/gif"
	case strings.HasPrefix(h, "RIFF") && len(header) >= 12 && string(header[8:12]) == "WEBP":
		return "image/webp"
	default:
		return ""
	}
}

// UploadBytes 以指定 key 和 content-type 上传原始字节，返回 CDN URL。
func (c *Client) UploadBytes(_ context.Context, key string, data []byte, contentType string) (string, error) {
	allowed := []string{"uploads/", "poster/", "waybill/", "stats/"}
	valid := false
	for _, p := range allowed {
		if strings.HasPrefix(key, p) {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("oss: key prefix not allowed: %s", key)
	}
	opt := osssdk.ContentType(contentType)
	if err := c.bucket.PutObject(key, bytes.NewReader(data), opt); err != nil {
		return "", fmt.Errorf("oss: put object: %w", err)
	}
	return c.cdnDomain + "/" + key, nil
}

// MockClient 用于测试，直接返回固定 URL。
type MockClient struct {
	BaseURL string
}

// UploadProductImage mock 实现，直接返回固定 URL。
func (m *MockClient) UploadProductImage(_ context.Context, _ io.Reader, filename string) (string, error) {
	return m.BaseURL + "/" + filename, nil
}

// STSCredentials 是阿里云 STS 临时凭证。
type STSCredentials struct {
	AccessKeyId     string    `json:"access_key_id"`
	AccessKeySecret string    `json:"access_key_secret"`
	SecurityToken   string    `json:"security_token"`
	Expiration      time.Time `json:"expiration"`
	Bucket          string    `json:"bucket"`
	Endpoint        string    `json:"endpoint"`
}

// GetSTSToken 调用阿里云 STS AssumeRole 接口，返回有效期 3600s 的临时凭证。
// Policy 限制只允许向 uploads/ 前缀写入。
//
// 依赖：accessKeyID / accessKeySecret / stsRoleArn 必须在配置中提供。
// STS API 使用 HMAC-SHA1 签名方案（阿里云 API v1.0）。
func (c *Client) GetSTSToken(ctx context.Context) (*STSCredentials, error) {
	if c.accessKeyID == "" || c.accessKeySecret == "" {
		return nil, fmt.Errorf("oss: sts requires AccessKeyID and AccessKeySecret")
	}
	if c.stsRoleArn == "" {
		return nil, fmt.Errorf("oss: sts_role_arn is required (set OSS_STS_ROLE_ARN)")
	}

	policy := fmt.Sprintf(`{
    "Version":"1",
    "Statement":[{
        "Effect":"Allow",
        "Action":["oss:PutObject"],
        "Resource":["acs:oss:*:*:%s/uploads/*"],
        "Condition":{
            "StringLike":{
                "oss:RequestObjectContentType":["image/jpeg","image/png","image/webp","image/gif"]
            }
        }
    }]
}`, c.bucketName)

	now := time.Now().UTC()
	params := map[string]string{
		"Action":           "AssumeRole",
		"Format":           "JSON",
		"Version":          "2015-04-01",
		"AccessKeyId":      c.accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"Timestamp":        now.Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce":   uuid.New().String(),
		"RoleArn":          c.stsRoleArn,
		"RoleSessionName":  "xu-shop-upload",
		"DurationSeconds":  "3600",
		"Policy":           policy,
	}

	// 按 key 字母序排序，构建规范查询串。
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, stsEncode(k)+"="+stsEncode(params[k]))
	}
	canonicalQuery := strings.Join(pairs, "&")

	// 待签字符串：GET&%2F&{二次编码的规范查询串}
	stringToSign := "GET&%2F&" + stsEncode(canonicalQuery)

	// HMAC-SHA1，密钥 = AccessKeySecret + "&"
	mac := hmac.New(sha1.New, []byte(c.accessKeySecret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	finalURL := c.stsEndpoint + "?" + canonicalQuery + "&Signature=" + url.QueryEscape(sig)

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("oss: sts build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oss: sts request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Credentials struct {
			AccessKeyId     string `json:"AccessKeyId"`
			AccessKeySecret string `json:"AccessKeySecret"`
			SecurityToken   string `json:"SecurityToken"`
			Expiration      string `json:"Expiration"`
		} `json:"Credentials"`
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("oss: sts decode response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || (result.Code != "" && result.Code != "200") {
		return nil, fmt.Errorf("oss: sts error %s: %s", result.Code, result.Message)
	}

	exp, err := time.Parse(time.RFC3339, result.Credentials.Expiration)
	if err != nil {
		exp = now.Add(3600 * time.Second)
	}
	return &STSCredentials{
		AccessKeyId:     result.Credentials.AccessKeyId,
		AccessKeySecret: result.Credentials.AccessKeySecret,
		SecurityToken:   result.Credentials.SecurityToken,
		Expiration:      exp,
		Bucket:          c.bucketName,
		Endpoint:        c.endpoint,
	}, nil
}

// stsEncode 对字符串进行 RFC 3986 百分号编码（空格 → %20，~ 不编码）。
func stsEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
