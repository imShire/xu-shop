// Package wxpay 封装微信支付 V3 API，与业务模块完全解耦。
package wxpay

import (
	"context"
	"time"
)

// Config 微信支付配置。
type Config struct {
	MchID           string
	APIKeyV3        string // 32 字节 API V3 密钥
	CertPath        string // 商户 API 证书路径（apiclient_cert.pem）
	KeyPath         string // 商户 API 私钥路径（apiclient_key.pem）
	AppIDMP         string // 小程序 AppID
	AppIDOA         string // 公众号 AppID
	AppIDH5         string // H5 AppID（通常同公众号）
	NotifyURL       string
	RefundNotifyURL string
}

// PrepayReq 预支付请求。
type PrepayReq struct {
	Scene    string    // jsapi_mp / jsapi_oa / h5
	OrderNo  string
	PayCents int64
	OpenID   string    // jsapi 模式必须
	ClientIP string    // h5 模式
	Expire   time.Time
}

// PrepayResp 预支付响应（前端拉起支付所需参数）。
type PrepayResp struct {
	Scene     string
	TimeStamp string
	NonceStr  string
	Package   string
	SignType   string
	PaySign   string
	MwebURL   string // h5 模式
}

// RefundReq 退款请求。
type RefundReq struct {
	OrderNo    string
	RefundNo   string
	AmtCents   int64
	TotalCents int64
	Reason     string
	NotifyURL  string
}

// QueryResp 查单结果。
type QueryResp struct {
	TradeState    string
	TransactionID string
	AmtCents      int64
	PaidAt        *time.Time
}

// NotifyResult 支付回调解析结果。
type NotifyResult struct {
	TransactionID string
	OutTradeNo    string
	AmtCents      int64
	PaidAt        *time.Time
	Raw           []byte
}

// RefundNotifyResult 退款回调解析结果。
type RefundNotifyResult struct {
	OutTradeNo  string
	OutRefundNo string
	Status      string // SUCCESS / ABNORMAL
	AmtCents    int64
	Raw         []byte
}

// Client 微信支付客户端接口。
type Client interface {
	// Prepay 发起预支付，返回前端拉起所需参数。
	Prepay(ctx context.Context, req PrepayReq) (*PrepayResp, error)
	// QueryByOutTradeNo 通过商户订单号查单。
	QueryByOutTradeNo(ctx context.Context, orderNo string) (*QueryResp, error)
	// Refund 发起退款（异步，结果通过退款回调通知）。
	Refund(ctx context.Context, req RefundReq) error
	// VerifyNotify 校验签名并解密支付回调。
	VerifyNotify(ctx context.Context, body []byte, headers map[string]string) (*NotifyResult, error)
	// VerifyRefundNotify 校验签名并解密退款回调。
	VerifyRefundNotify(ctx context.Context, body []byte, headers map[string]string) (*RefundNotifyResult, error)
	// DownloadBill 下载对账单（date 格式 2006-01-02）。
	DownloadBill(ctx context.Context, date string) ([]byte, error)
}
