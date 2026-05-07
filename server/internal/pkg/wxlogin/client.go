// Package wxlogin 封装微信小程序/公众号登录相关 API 调用。
package wxlogin

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Code2SessionResp 小程序 code2session 响应。
type Code2SessionResp struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// OAuthResp 公众号 OAuth2 授权后的 token 响应。
type OAuthResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WxLoginClient 微信登录能力接口。
type WxLoginClient interface {
	// Code2Session 小程序登录，用 code 换取 openid + session_key。
	Code2Session(ctx context.Context, code string) (*Code2SessionResp, error)
	// DecryptUserData 解密微信加密数据（如手机号）。
	DecryptUserData(sessionKey, encryptedData, iv string) (map[string]any, error)
	// GetOAuthURL 生成公众号 OAuth2 授权 URL。
	GetOAuthURL(redirectURI, state string) string
	// OAuthCode2Token 公众号用 code 换 access_token + openid。
	OAuthCode2Token(ctx context.Context, code string) (*OAuthResp, error)
}

// Client 真实微信 API 客户端。
type Client struct {
	appID     string
	appSecret string
	httpCli   *http.Client
}

// NewClient 构造 Client，超时 5s。
func NewClient(appID, appSecret string) WxLoginClient {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		httpCli:   &http.Client{Timeout: 5 * time.Second},
	}
}

// Code2Session 调用微信 jscode2session 接口。
func (c *Client) Code2Session(ctx context.Context, code string) (*Code2SessionResp, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.appID, c.appSecret, url.QueryEscape(code),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: build request: %w", err)
	}
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: read body: %w", err)
	}

	var result Code2SessionResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("wxlogin: unmarshal: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wxlogin: code2session errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}
	return &result, nil
}

// DecryptUserData 使用 AES-128-CBC 解密微信加密数据。
func (c *Client) DecryptUserData(sessionKey, encryptedData, iv string) (map[string]any, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: decode session_key: %w", err)
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: decode encrypted_data: %w", err)
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: decode iv: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: new cipher: %w", err)
	}
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(cipherBytes, cipherBytes)

	// 去除 PKCS7 padding
	if len(cipherBytes) == 0 {
		return nil, fmt.Errorf("wxlogin: empty plaintext")
	}
	pad := int(cipherBytes[len(cipherBytes)-1])
	if pad > aes.BlockSize || pad == 0 {
		return nil, fmt.Errorf("wxlogin: invalid padding")
	}
	plaintext := cipherBytes[:len(cipherBytes)-pad]

	var result map[string]any
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("wxlogin: unmarshal decrypted: %w", err)
	}
	return result, nil
}

// GetOAuthURL 生成公众号网页授权 URL（snsapi_userinfo scope）。
func (c *Client) GetOAuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect",
		c.appID, url.QueryEscape(redirectURI), url.QueryEscape(state),
	)
}

// OAuthCode2Token 用 code 换取公众号 access_token。
func (c *Client) OAuthCode2Token(ctx context.Context, code string) (*OAuthResp, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		c.appID, c.appSecret, url.QueryEscape(code),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: build oauth request: %w", err)
	}
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: do oauth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("wxlogin: read oauth body: %w", err)
	}

	var result OAuthResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("wxlogin: unmarshal oauth: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wxlogin: oauth2 errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}
	return &result, nil
}
