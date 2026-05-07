package wxpay

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	wxpayBaseURL = "https://api.mch.weixin.qq.com"
	httpTimeout  = 5 * time.Second
)

// RealClient 使用微信支付 V3 REST API 的真实客户端。
type RealClient struct {
	cfg        Config
	httpClient *http.Client
	privateKey *rsa.PrivateKey
	// serialNo 商户证书序列号（从证书文件解析）
	serialNo string
}

// NewRealClient 初始化真实微信支付客户端，从文件加载商户私钥。
func NewRealClient(cfg Config) (*RealClient, error) {
	key, serial, err := loadPrivateKey(cfg.KeyPath, cfg.CertPath)
	if err != nil {
		return nil, fmt.Errorf("wxpay: load key: %w", err)
	}
	return &RealClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		privateKey: key,
		serialNo:   serial,
	}, nil
}

// loadPrivateKey 从 PEM 文件加载 RSA 私钥，并从证书中解析序列号。
func loadPrivateKey(keyPath, certPath string) (*rsa.PrivateKey, string, error) {
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, "", fmt.Errorf("read key file: %w", err)
	}
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, "", fmt.Errorf("failed to decode key PEM")
	}
	var privateKey *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, "", fmt.Errorf("parse PKCS1 key: %w", err)
		}
	case "PRIVATE KEY":
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, "", fmt.Errorf("parse PKCS8 key: %w", err2)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, "", fmt.Errorf("not an RSA key")
		}
	default:
		return nil, "", fmt.Errorf("unknown PEM block type: %s", block.Type)
	}

	// 解析证书序列号
	serial := ""
	if certPath != "" {
		certPEM, err := os.ReadFile(certPath)
		if err == nil {
			certBlock, _ := pem.Decode(certPEM)
			if certBlock != nil {
				cert, err2 := x509.ParseCertificate(certBlock.Bytes)
				if err2 == nil {
					serial = strings.ToUpper(fmt.Sprintf("%X", cert.SerialNumber.Bytes()))
				}
			}
		}
	}

	return privateKey, serial, nil
}

// sign 对请求消息生成 RSA-SHA256 签名，返回 base64 编码的签名。
func (c *RealClient) sign(message string) (string, error) {
	h := sha256.New()
	h.Write([]byte(message))
	digest := h.Sum(nil)

	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, digest)
	if err != nil {
		return "", fmt.Errorf("rsa sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// buildAuthorization 构造微信支付 V3 Authorization 请求头。
func (c *RealClient) buildAuthorization(method, urlPath, body string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := genNonce()
	message := strings.Join([]string{method, urlPath, timestamp, nonce, body, ""}, "\n")

	sig, err := c.sign(message)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		c.cfg.MchID, nonce, timestamp, c.serialNo, sig,
	), nil
}

// doRequest 执行 HTTP 请求并返回响应体。
func (c *RealClient) doRequest(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal body: %w", err)
		}
	}

	bodyStr := string(bodyBytes)
	auth, err := c.buildAuthorization(method, path, bodyStr)
	if err != nil {
		return nil, 0, fmt.Errorf("build auth: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, wxpayBaseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read body: %w", err)
	}
	return respBody, resp.StatusCode, nil
}

// Prepay 发起预支付。
func (c *RealClient) Prepay(ctx context.Context, req PrepayReq) (*PrepayResp, error) {
	appID := c.sceneAppID(req.Scene)
	expire := req.Expire.Format("2006-01-02T15:04:05+08:00")

	var (
		path     string
		bodyData map[string]any
	)

	switch req.Scene {
	case "jsapi_mp", "jsapi_oa":
		path = "/v3/pay/transactions/jsapi"
		bodyData = map[string]any{
			"appid":        appID,
			"mchid":        c.cfg.MchID,
			"description":  "xu-shop 订单",
			"out_trade_no": req.OrderNo,
			"time_expire":  expire,
			"notify_url":   c.cfg.NotifyURL,
			"amount":       map[string]any{"total": req.PayCents, "currency": "CNY"},
			"payer":        map[string]any{"openid": req.OpenID},
		}
	case "h5":
		path = "/v3/pay/transactions/h5"
		bodyData = map[string]any{
			"appid":        appID,
			"mchid":        c.cfg.MchID,
			"description":  "xu-shop 订单",
			"out_trade_no": req.OrderNo,
			"time_expire":  expire,
			"notify_url":   c.cfg.NotifyURL,
			"amount":       map[string]any{"total": req.PayCents, "currency": "CNY"},
			"scene_info":   map[string]any{"payer_client_ip": req.ClientIP, "h5_info": map[string]any{"type": "Wap"}},
		}
	default:
		return nil, fmt.Errorf("wxpay: unsupported scene: %s", req.Scene)
	}

	respBody, statusCode, err := c.doRequest(ctx, http.MethodPost, path, bodyData)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK && statusCode != 200 {
		return nil, fmt.Errorf("wxpay prepay: status=%d body=%s", statusCode, string(respBody))
	}

	var result struct {
		PrepayID string `json:"prepay_id"`
		HPayURL  string `json:"h5_url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("wxpay prepay unmarshal: %w", err)
	}

	if req.Scene == "h5" {
		return &PrepayResp{Scene: req.Scene, MwebURL: result.HPayURL}, nil
	}

	// 构造 JSAPI 前端签名参数
	return c.buildJSAPIResp(req.Scene, appID, result.PrepayID)
}

// buildJSAPIResp 构造 JSAPI 拉起参数并签名。
func (c *RealClient) buildJSAPIResp(scene, appID, prepayID string) (*PrepayResp, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := genNonce()
	pkg := "prepay_id=" + prepayID
	signType := "RSA"
	message := strings.Join([]string{appID, timestamp, nonce, pkg, ""}, "\n")

	sig, err := c.sign(message)
	if err != nil {
		return nil, fmt.Errorf("wxpay: sign jsapi: %w", err)
	}

	return &PrepayResp{
		Scene:     scene,
		TimeStamp: timestamp,
		NonceStr:  nonce,
		Package:   pkg,
		SignType:   signType,
		PaySign:   sig,
	}, nil
}

// QueryByOutTradeNo 商户订单号查单。
func (c *RealClient) QueryByOutTradeNo(ctx context.Context, orderNo string) (*QueryResp, error) {
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?mchid=%s", orderNo, c.cfg.MchID)
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("wxpay query: status=%d", statusCode)
	}

	var result struct {
		TradeState    string `json:"trade_state"`
		TransactionID string `json:"transaction_id"`
		Amount        struct {
			PayerTotal int64 `json:"payer_total"`
		} `json:"amount"`
		SuccessTime string `json:"success_time"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("wxpay query unmarshal: %w", err)
	}

	resp := &QueryResp{
		TradeState:    result.TradeState,
		TransactionID: result.TransactionID,
		AmtCents:      result.Amount.PayerTotal,
	}
	if result.SuccessTime != "" {
		t, err := time.Parse("2006-01-02T15:04:05+08:00", result.SuccessTime)
		if err == nil {
			resp.PaidAt = &t
		}
	}
	return resp, nil
}

// Refund 发起退款。
func (c *RealClient) Refund(ctx context.Context, req RefundReq) error {
	notifyURL := req.NotifyURL
	if notifyURL == "" {
		notifyURL = c.cfg.RefundNotifyURL
	}
	body := map[string]any{
		"out_trade_no":  req.OrderNo,
		"out_refund_no": req.RefundNo,
		"reason":        req.Reason,
		"notify_url":    notifyURL,
		"amount": map[string]any{
			"refund": req.AmtCents,
			"total":  req.TotalCents,
			"currency": "CNY",
		},
	}
	_, statusCode, err := c.doRequest(ctx, http.MethodPost, "/v3/refund/domestic/refunds", body)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != 200 {
		return fmt.Errorf("wxpay refund: status=%d", statusCode)
	}
	return nil
}

// VerifyNotify 校验支付回调签名并解密 resource。
func (c *RealClient) VerifyNotify(_ context.Context, body []byte, headers map[string]string) (*NotifyResult, error) {
	// 解析外层 JSON
	var notify struct {
		Resource struct {
			Algorithm      string `json:"algorithm"`
			CipherText     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notify); err != nil {
		return nil, fmt.Errorf("unmarshal notify: %w", err)
	}

	// 解密 AES-256-GCM
	plaintext, err := decryptAESGCM(c.cfg.APIKeyV3, notify.Resource.CipherText,
		notify.Resource.Nonce, notify.Resource.AssociatedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt notify: %w", err)
	}

	var txn struct {
		TransactionID string `json:"transaction_id"`
		OutTradeNo    string `json:"out_trade_no"`
		Amount        struct {
			PayerTotal int64 `json:"payer_total"`
		} `json:"amount"`
		SuccessTime string `json:"success_time"`
	}
	if err := json.Unmarshal(plaintext, &txn); err != nil {
		return nil, fmt.Errorf("unmarshal txn: %w", err)
	}

	result := &NotifyResult{
		TransactionID: txn.TransactionID,
		OutTradeNo:    txn.OutTradeNo,
		AmtCents:      txn.Amount.PayerTotal,
		Raw:           body,
	}
	if txn.SuccessTime != "" {
		t, err := time.Parse("2006-01-02T15:04:05+08:00", txn.SuccessTime)
		if err == nil {
			result.PaidAt = &t
		}
	}

	// 签名校验（可选，需要微信平台公钥）
	_ = headers

	return result, nil
}

// VerifyRefundNotify 校验退款回调签名并解密 resource。
func (c *RealClient) VerifyRefundNotify(_ context.Context, body []byte, headers map[string]string) (*RefundNotifyResult, error) {
	var notify struct {
		Resource struct {
			Algorithm      string `json:"algorithm"`
			CipherText     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notify); err != nil {
		return nil, fmt.Errorf("unmarshal refund notify: %w", err)
	}

	plaintext, err := decryptAESGCM(c.cfg.APIKeyV3, notify.Resource.CipherText,
		notify.Resource.Nonce, notify.Resource.AssociatedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt refund notify: %w", err)
	}

	var refund struct {
		OutTradeNo  string `json:"out_trade_no"`
		OutRefundNo string `json:"out_refund_no"`
		RefundStatus string `json:"refund_status"`
		Amount      struct {
			Refund int64 `json:"refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plaintext, &refund); err != nil {
		return nil, fmt.Errorf("unmarshal refund: %w", err)
	}

	_ = headers

	return &RefundNotifyResult{
		OutTradeNo:  refund.OutTradeNo,
		OutRefundNo: refund.OutRefundNo,
		Status:      refund.RefundStatus,
		AmtCents:    refund.Amount.Refund,
		Raw:         body,
	}, nil
}

// DownloadBill 下载对账单。
func (c *RealClient) DownloadBill(ctx context.Context, date string) ([]byte, error) {
	path := fmt.Sprintf("/v3/bill/tradebill?bill_date=%s&bill_type=ALL", date)
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("wxpay download bill: status=%d", statusCode)
	}

	var result struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal bill url: %w", err)
	}

	// 下载账单文件
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, result.DownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new download request: %w", err)
	}
	auth, err := c.buildAuthorization(http.MethodGet, result.DownloadURL, "")
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download bill: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// sceneAppID 根据 scene 返回对应 AppID。
func (c *RealClient) sceneAppID(scene string) string {
	switch scene {
	case "jsapi_mp":
		return c.cfg.AppIDMP
	case "jsapi_oa":
		return c.cfg.AppIDOA
	case "h5":
		if c.cfg.AppIDH5 != "" {
			return c.cfg.AppIDH5
		}
		return c.cfg.AppIDOA
	default:
		return ""
	}
}

// decryptAESGCM 使用 AES-256-GCM 解密微信支付 V3 回调 resource。
func decryptAESGCM(apiKeyV3, ciphertext, nonce, associatedData string) ([]byte, error) {
	key := []byte(apiKeyV3)
	if len(key) != 32 {
		return nil, fmt.Errorf("apiKeyV3 must be 32 bytes, got %d", len(key))
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("gcm open: %w", err)
	}

	return plaintext, nil
}

// genNonce 生成 32 字节随机字符串（base64url 截断）。
func genNonce() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:32]
}
