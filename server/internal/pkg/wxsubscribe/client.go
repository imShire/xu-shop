// Package wxsubscribe 封装微信订阅消息发送（公众号/小程序）。
package wxsubscribe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	accessTokenKey = "wx:mp:access_token"
	tokenEndpoint  = "https://api.weixin.qq.com/cgi-bin/token"
	sendEndpoint   = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"
)

// SendReq 订阅消息发送请求。
type SendReq struct {
	ToUser           string         `json:"touser"`
	TemplateID       string         `json:"template_id"`
	Page             string         `json:"page,omitempty"`
	Data             map[string]any `json:"data"`
	MiniprogramState string         `json:"miniprogram_state,omitempty"`
}

// Client 微信订阅消息客户端接口。
type Client interface {
	Send(ctx context.Context, req SendReq) error
	GetAccessToken(ctx context.Context) (string, error)
}

type wxClient struct {
	appID     string
	appSecret string
	rdb       *redis.Client
	httpCli   *http.Client
}

// NewClient 构造真实微信订阅消息客户端。
func NewClient(appID, appSecret string, rdb *redis.Client) Client {
	return &wxClient{
		appID:     appID,
		appSecret: appSecret,
		rdb:       rdb,
		httpCli:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAccessToken 获取 access_token，优先从 Redis 缓存读取。
func (c *wxClient) GetAccessToken(ctx context.Context) (string, error) {
	if c.rdb != nil {
		if token, err := c.rdb.Get(ctx, accessTokenKey).Result(); err == nil && token != "" {
			return token, nil
		}
	}

	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s",
		tokenEndpoint, c.appID, c.appSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("wxsubscribe: build token request: %w", err)
	}
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("wxsubscribe: get token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("wxsubscribe: parse token response: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wxsubscribe: get token errcode=%d msg=%s", result.ErrCode, result.ErrMsg)
	}

	if c.rdb != nil && result.AccessToken != "" {
		ttl := time.Duration(result.ExpiresIn-300) * time.Second
		if ttl > 0 {
			_ = c.rdb.Set(ctx, accessTokenKey, result.AccessToken, ttl).Err()
		}
	}
	return result.AccessToken, nil
}

// Send 发送微信订阅消息。
func (c *wxClient) Send(ctx context.Context, req SendReq) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("wxsubscribe: marshal send req: %w", err)
	}

	apiURL := sendEndpoint + "?access_token=" + token
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("wxsubscribe: build send request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return fmt.Errorf("wxsubscribe: do send: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("wxsubscribe: parse send response: %w", err)
	}
	if result.ErrCode != 0 {
		return &WxAPIError{Code: result.ErrCode, Msg: result.ErrMsg}
	}
	return nil
}

// WxAPIError 微信 API 业务错误（携带 errcode 供 service 层判断）。
type WxAPIError struct {
	Code int
	Msg  string
}

func (e *WxAPIError) Error() string {
	return fmt.Sprintf("wxsubscribe: errcode=%d msg=%s", e.Code, e.Msg)
}
