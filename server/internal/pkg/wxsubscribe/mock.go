package wxsubscribe

import "context"

// MockClient 测试用 mock 客户端。
type MockClient struct {
	SendErr  error
	nextErr  error
}

// NewMockClient 返回 MockClient。
func NewMockClient() *MockClient { return &MockClient{} }

// SetNextError 设置下一次 Send 调用返回的错误。
func (m *MockClient) SetNextError(err error) { m.nextErr = err }

func (m *MockClient) GetAccessToken(_ context.Context) (string, error) {
	return "mock_access_token", nil
}

func (m *MockClient) Send(_ context.Context, _ SendReq) error {
	if m.nextErr != nil {
		err := m.nextErr
		m.nextErr = nil
		return err
	}
	return m.SendErr
}
