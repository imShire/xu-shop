package wxpay

import (
	"context"
	"fmt"
	"os"
	"time"
)

// MockClient 用于单元测试的微信支付 mock 客户端。
type MockClient struct {
	PrepayResult        *PrepayResp
	PrepayErr           error
	QueryResult         *QueryResp
	QueryErr            error
	RefundErr           error
	NotifyResult        *NotifyResult
	NotifyErr           error
	RefundNotifyResult  *RefundNotifyResult
	RefundNotifyErr     error
	BillData            []byte
	BillErr             error
}

// NewMockClient 返回默认 mock 客户端（返回预设成功响应）。
func NewMockClient() *MockClient {
	now := time.Now()
	return &MockClient{
		PrepayResult: &PrepayResp{
			Scene:     "jsapi_mp",
			TimeStamp: fmt.Sprintf("%d", now.Unix()),
			NonceStr:  "mock_nonce_123456",
			Package:   "prepay_id=mock_prepay_id_001",
			SignType:   "RSA",
			PaySign:   "mock_pay_sign_base64",
		},
		NotifyResult: &NotifyResult{
			TransactionID: "mock_txn_001",
			OutTradeNo:    "mock_order_no",
			AmtCents:      100,
			PaidAt:        &now,
			Raw:           []byte(`{"mock":"notify"}`),
		},
		BillData: []byte("mock_bill_data"),
	}
}

// IsMockMode 判断是否开启微信支付 mock 模式。
func IsMockMode() bool {
	return os.Getenv("MOCK_WXPAY") == "true"
}

func (m *MockClient) Prepay(_ context.Context, _ PrepayReq) (*PrepayResp, error) {
	return m.PrepayResult, m.PrepayErr
}

func (m *MockClient) QueryByOutTradeNo(_ context.Context, _ string) (*QueryResp, error) {
	return m.QueryResult, m.QueryErr
}

func (m *MockClient) Refund(_ context.Context, _ RefundReq) error {
	return m.RefundErr
}

func (m *MockClient) VerifyNotify(_ context.Context, _ []byte, _ map[string]string) (*NotifyResult, error) {
	return m.NotifyResult, m.NotifyErr
}

func (m *MockClient) VerifyRefundNotify(_ context.Context, _ []byte, _ map[string]string) (*RefundNotifyResult, error) {
	return m.RefundNotifyResult, m.RefundNotifyErr
}

func (m *MockClient) DownloadBill(_ context.Context, _ string) ([]byte, error) {
	return m.BillData, m.BillErr
}
