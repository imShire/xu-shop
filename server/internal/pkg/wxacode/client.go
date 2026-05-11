// Package wxacode 封装微信小程序码获取 API。
package wxacode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/xushop/xu-shop/internal/config"
)

const (
	atCacheKey  = "wxacode:at"
	atCacheTTL  = 6000 * time.Second
	httpTimeout = 5 * time.Second
)

// Client 小程序码客户端接口。
type Client interface {
	// GetUnlimited 调用 wxacode.getUnlimited 接口，返回图片 bytes (PNG)。
	// scene ≤32 字节（用 short_code），page 如 "pages/product/detail/index"。
	GetUnlimited(ctx context.Context, scene, page string) ([]byte, error)

	// GetAccessToken 获取小程序 access_token（带 Redis 缓存）。
	GetAccessToken(ctx context.Context) (string, error)
}

// RealClient 真实微信 API 客户端。
type RealClient struct {
	appID     string
	appSecret string
	rdb       *redis.Client
	httpCli   *http.Client
}

// NewRealClient 构造 RealClient，从 cfg.WxMP 读取 appid/secret。
func NewRealClient(cfg *config.Config, rdb *redis.Client) *RealClient {
	return &RealClient{
		appID:     cfg.WxMP.AppID,
		appSecret: cfg.WxMP.AppSecret,
		rdb:       rdb,
		httpCli:   &http.Client{Timeout: httpTimeout},
	}
}

type atResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// GetAccessToken 获取小程序 access_token，优先读取 Redis 缓存。
func (c *RealClient) GetAccessToken(ctx context.Context) (string, error) {
	if c.rdb != nil {
		if at, err := c.rdb.Get(ctx, atCacheKey).Result(); err == nil && at != "" {
			return at, nil
		}
	}

	resp, err := c.httpCli.PostForm("https://api.weixin.qq.com/cgi-bin/token", url.Values{
		"grant_type": {"client_credential"},
		"appid":      {c.appID},
		"secret":     {c.appSecret},
	})
	if err != nil {
		return "", fmt.Errorf("wxacode: fetch access_token: %w", err)
	}
	defer resp.Body.Close()

	var result atResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("wxacode: decode token response: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wxacode: token errcode=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("wxacode: empty access_token")
	}

	if c.rdb != nil {
		_ = c.rdb.Set(ctx, atCacheKey, result.AccessToken, atCacheTTL).Err()
	}
	return result.AccessToken, nil
}

type getUnlimitedReq struct {
	Scene string `json:"scene"`
	Page  string `json:"page"`
	Width int    `json:"width"`
}

type wxErrResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// GetUnlimited 获取小程序码（不限制数量）图片 PNG bytes。
func (c *RealClient) GetUnlimited(ctx context.Context, scene, page string) ([]byte, error) {
	at, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", at)
	body, err := json.Marshal(getUnlimitedReq{Scene: scene, Page: page, Width: 430})
	if err != nil {
		return nil, fmt.Errorf("wxacode: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("wxacode: build qrcode request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wxacode: fetch qrcode: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("wxacode: read qrcode response: %w", err)
	}

	// 微信返回 JSON 说明有错误（正常返回为二进制图片）
	if len(data) > 0 && data[0] == '{' {
		var e wxErrResp
		if jsonErr := json.Unmarshal(data, &e); jsonErr == nil && e.ErrCode != 0 {
			return nil, fmt.Errorf("wxacode: getUnlimited errcode=%d msg=%s", e.ErrCode, e.ErrMsg)
		}
	}
	return data, nil
}

// MockClient 测试用 mock，返回 1×1 透明 PNG bytes。
type MockClient struct{}

// transparentPNG 是 1×1 像素 RGBA 透明 PNG 的原始字节。
var transparentPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

// GetUnlimited mock 实现，返回 1×1 透明 PNG。
func (m *MockClient) GetUnlimited(_ context.Context, _, _ string) ([]byte, error) {
	return transparentPNG, nil
}

// GetAccessToken mock 实现，返回固定 token。
func (m *MockClient) GetAccessToken(_ context.Context) (string, error) {
	return "mock_access_token", nil
}
