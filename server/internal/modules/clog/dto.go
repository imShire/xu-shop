package clog

// SubmitReq 单条上报请求。
//
// 故意不接受前端传入 user_id / admin_id：
// 服务端只信任由 UserOptionalAuth / AdminOptionalAuth 解析出的身份，
// 避免任意来源伪造身份污染日志表（CR Blocker #1）。
type SubmitReq struct {
	Source    string         `json:"source"     binding:"required,oneof=admin client_h5 client_weapp"`
	Level     string         `json:"level"      binding:"required,oneof=error warn info"`
	Message   string         `json:"message"    binding:"required"`
	Stack     string         `json:"stack"`
	URL       string         `json:"url"`
	UserAgent string         `json:"user_agent"`
	Release   string         `json:"release"`
	Extra     map[string]any `json:"extra,omitempty"`
}

// BatchSubmitReq 批量上报请求。
type BatchSubmitReq struct {
	Logs []SubmitReq `json:"logs" binding:"required,min=1,max=50,dive"`
}
