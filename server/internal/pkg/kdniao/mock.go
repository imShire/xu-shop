package kdniao

import (
	"context"
	"fmt"
	"time"
)

// MockClient 用于单元测试的快递鸟 mock 客户端。
type MockClient struct {
	CreateResult *CreateWaybillResp
	CreateErr    error
	TrackResult  *TrackResp
	TrackErr     error
	SubscribeErr error
	PushResult   *PushResp
	PushErr      error
}

// NewMockClient 返回带默认成功响应的 mock 客户端。
func NewMockClient() *MockClient {
	return &MockClient{
		CreateResult: &CreateWaybillResp{
			TrackingNo:  "MOCK123456789",
			PrintBase64: "bW9ja19wZGZfYmFzZTY0",
			Success:     true,
		},
		TrackResult: &TrackResp{
			Carrier: "SF",
			TrackNo: "MOCK123456789",
			State:   "in_transit",
			Traces: []TraceItem{
				{OccurredAt: time.Now().Add(-2 * time.Hour), Status: "picked", Description: "已揽收"},
				{OccurredAt: time.Now().Add(-1 * time.Hour), Status: "in_transit", Description: "运输中"},
			},
		},
		PushResult: &PushResp{
			CarrierCode: "SF",
			TrackingNo:  "MOCK123456789",
			State:       "delivered",
			Traces: []TraceItem{
				{OccurredAt: time.Now(), Status: "delivered", Description: "已签收"},
			},
		},
	}
}

func (m *MockClient) CreateWaybill(_ context.Context, _ CreateWaybillReq) (*CreateWaybillResp, error) {
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	if m.CreateResult == nil {
		return nil, fmt.Errorf("mock: no CreateResult configured")
	}
	return m.CreateResult, nil
}

func (m *MockClient) Track(_ context.Context, _, _ string) (*TrackResp, error) {
	return m.TrackResult, m.TrackErr
}

func (m *MockClient) Subscribe(_ context.Context, _, _, _ string) error {
	return m.SubscribeErr
}

func (m *MockClient) ParsePush(_ []byte) (*PushResp, error) {
	return m.PushResult, m.PushErr
}
