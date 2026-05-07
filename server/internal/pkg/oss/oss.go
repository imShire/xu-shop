// Package oss 封装阿里云 OSS 上传操作。
package oss

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
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
	bucket    *osssdk.Bucket
	cdnDomain string
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
	return &Client{
		bucket:    bucket,
		cdnDomain: strings.TrimRight(cfg.OSS.CDNDomain, "/"),
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

// MockClient 用于测试，直接返回固定 URL。
type MockClient struct {
	BaseURL string
}

// UploadProductImage mock 实现，直接返回固定 URL。
func (m *MockClient) UploadProductImage(_ context.Context, _ io.Reader, filename string) (string, error) {
	return m.BaseURL + "/" + filename, nil
}
