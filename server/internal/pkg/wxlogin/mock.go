package wxlogin

import (
	"context"
	"os"
)

// MockClient 在 MOCK_WX=true 时返回固定测试数据。
type MockClient struct{}

// NewMockClient 返回 mock 客户端。
func NewMockClient() WxLoginClient {
	return &MockClient{}
}

// IsMockMode 判断是否开启微信 mock 模式。
func IsMockMode() bool {
	return os.Getenv("MOCK_WX") == "true"
}

func (m *MockClient) Code2Session(_ context.Context, _ string) (*Code2SessionResp, error) {
	return &Code2SessionResp{
		OpenID:     "mock_openid_mp_001",
		UnionID:    "mock_unionid_001",
		SessionKey: "mock_session_key_base64==",
	}, nil
}

func (m *MockClient) DecryptUserData(_, _, _ string) (map[string]any, error) {
	return map[string]any{
		"phoneNumber":     "13800138000",
		"countryCode":     "86",
		"purePhoneNumber": "13800138000",
	}, nil
}

func (m *MockClient) GetOAuthURL(redirectURI, state string) string {
	return redirectURI + "?code=mock_h5_code&state=" + state
}

func (m *MockClient) OAuthCode2Token(_ context.Context, _ string) (*OAuthResp, error) {
	return &OAuthResp{
		OpenID:  "mock_openid_h5_001",
		UnionID: "mock_unionid_001",
	}, nil
}
