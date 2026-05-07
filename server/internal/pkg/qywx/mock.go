package qywx

import "context"

// MockClient 测试用 mock 客户端。
type MockClient struct {
	AddContactWayResp *AddContactWayResp
	AddCorpTagID      string
	Err               error
}

// NewMockClient 返回 MockClient。
func NewMockClient() Client { return &MockClient{} }

func (m *MockClient) SendRobotMarkdown(_ context.Context, _, _ string) error { return m.Err }

func (m *MockClient) AddContactWay(_ context.Context, _ AddContactWayReq) (*AddContactWayResp, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.AddContactWayResp != nil {
		return m.AddContactWayResp, nil
	}
	return &AddContactWayResp{ConfigID: "mock_config_id", QrCode: "https://mock.qr/code"}, nil
}

func (m *MockClient) SendWelcomeMsg(_ context.Context, _ string, _ WelcomeMsg) error {
	return m.Err
}

func (m *MockClient) MarkCustomerTag(_ context.Context, _ MarkTagReq) error { return m.Err }

func (m *MockClient) AddCorpTag(_ context.Context, _, _ string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.AddCorpTagID != "" {
		return m.AddCorpTagID, nil
	}
	return "mock_tag_id_001", nil
}

func (m *MockClient) DeleteCorpTag(_ context.Context, _ string) error { return m.Err }

func (m *MockClient) DecryptCallback(_ context.Context, body []byte, _ map[string]string) (*CallbackEvent, error) {
	return &CallbackEvent{
		MsgType: "event",
		Event:   "change_external_contact",
		Raw:     body,
	}, m.Err
}
