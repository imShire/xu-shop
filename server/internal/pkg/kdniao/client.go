// Package kdniao 封装快递鸟（KDNiao）物流 API，含全局 token bucket 限速。
package kdniao

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// Addr 地址结构。
type Addr struct {
	Company  string
	Name     string
	Phone    string
	Province string
	City     string
	District string
	Detail   string
}

// GoodsItem 货物信息。
type GoodsItem struct {
	Name   string
	Qty    int
	Weight float64
}

// CreateWaybillReq 创建电子面单请求。
type CreateWaybillReq struct {
	CarrierCode    string
	OrderNo        string
	MonthlyAccount string // 月结账号
	Sender         Addr
	Receiver       Addr
	Goods          []GoodsItem
	CallbackURL    string
}

// CreateWaybillResp 创建电子面单响应。
type CreateWaybillResp struct {
	TrackingNo  string
	PrintBase64 string // 面单 PDF base64
	Success     bool
	Reason      string
}

// TraceItem 单条轨迹记录。
type TraceItem struct {
	OccurredAt  time.Time
	Status      string
	Description string
}

// TrackResp 查询物流轨迹响应。
type TrackResp struct {
	Carrier  string
	TrackNo  string
	State    string
	Traces   []TraceItem
}

// PushResp 快递鸟 Webhook 推送解析结果。
type PushResp struct {
	CarrierCode string
	TrackingNo  string
	State       string
	Traces      []TraceItem
}

// Client 快递鸟客户端接口。
type Client interface {
	// CreateWaybill 创建电子面单。
	CreateWaybill(ctx context.Context, req CreateWaybillReq) (*CreateWaybillResp, error)
	// Track 主动查询物流轨迹。
	Track(ctx context.Context, carrier, no string) (*TrackResp, error)
	// Subscribe 订阅物流轨迹推送。
	Subscribe(ctx context.Context, carrier, no, callbackURL string) error
	// ParsePush 解析快递鸟 Webhook 推送数据。
	ParsePush(body []byte) (*PushResp, error)
}

// RealClient 带 token bucket 限速的快递鸟客户端（8 QPS，burst 8，最大并发 5）。
type RealClient struct {
	businessID string
	apiKey     string
	reqURL     string
	limiter    *rate.Limiter // 8 QPS，burst 8
	sem        chan struct{}  // 并发上限 5
}

// NewRealClient 初始化快递鸟真实客户端。
func NewRealClient(businessID, apiKey, reqURL string) *RealClient {
	if reqURL == "" {
		reqURL = "https://api.kdniao.com/Ebusiness/EbusinessOrderHandle.aspx"
	}
	return &RealClient{
		businessID: businessID,
		apiKey:     apiKey,
		reqURL:     reqURL,
		limiter:    rate.NewLimiter(rate.Limit(8), 8),
		sem:        make(chan struct{}, 5),
	}
}
