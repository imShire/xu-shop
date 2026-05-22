package wxpay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ===== 商家转账 V3 类型 =====

// TransferReq 商家转账到零钱请求。
//
// 参考微信支付 V3 文档：商家转账（mch-transfer/transfer-bills）。
// 注意：单笔金额 ≤ 2000 元时无需 UserName（实名校验）。本 SDK 仅支持免实名场景。
type TransferReq struct {
	OutBillNo            string                  // 商户单号（withdraw_order.withdraw_no）
	AppID                string                  // 接收侧 appid（按 OpenID 所属应用决定）
	OpenID               string                  // 收款用户 openid
	TransferSceneID      string                  // 商家后台配置的场景 id（必传，如 "1000"）
	TransferAmountCents  int64                   // 金额（单位：分）
	TransferRemark       string                  // 转账备注（用户可见）
	NotifyURL            string                  // 回调地址
	UserRecvPerception   string                  // 用户感知（可选，如 "劳务报酬"）
	SceneReportInfos     []TransferSceneReport   // 场景报备信息（部分场景必填）
}

// TransferSceneReport 转账场景报备信息（如佣金报酬：报备字段名+内容）。
type TransferSceneReport struct {
	InfoType    string `json:"info_type"`
	InfoContent string `json:"info_content"`
}

// TransferResp 转账下单响应。
type TransferResp struct {
	OutBillNo      string // 商户单号回显
	TransferBillNo string // 微信侧单号
	State          string // ACCEPTED / PROCESSING / WAIT_USER_CONFIRM / TRANSFERRING / SUCCESS / FAIL / CANCELING / CANCELLED
	CreateTime     *time.Time
	PackageInfo    string // 部分场景需要前端拉起 jsapi 确认页（小程序确认收款）
}

// TransferQueryResp 查单响应。
type TransferQueryResp struct {
	OutBillNo      string
	TransferBillNo string
	State          string
	TransferAmount int64
	FailReason     string
	CreateTime     *time.Time
	UpdateTime     *time.Time
}

// TransferNotifyResult 转账回调解密结果。
type TransferNotifyResult struct {
	OutBillNo      string
	TransferBillNo string
	State          string
	TransferAmount int64
	OpenID         string
	FailReason     string
	UpdateTime     *time.Time
	Raw            []byte
}

// TransferClient 商家转账客户端接口。RealClient 与 MockClient 均实现此接口。
//
// 与 Client 解耦：业务模块（distribution/withdraw）只依赖此接口，方便测试与 mock。
type TransferClient interface {
	Transfer(ctx context.Context, req TransferReq) (*TransferResp, error)
	QueryTransfer(ctx context.Context, outBillNo string) (*TransferQueryResp, error)
	VerifyTransferNotify(ctx context.Context, body []byte, headers map[string]string) (*TransferNotifyResult, error)
}

// ===== Transfer 状态常量 =====

const (
	TransferStateAccepted        = "ACCEPTED"
	TransferStateProcessing      = "PROCESSING"
	TransferStateWaitUserConfirm = "WAIT_USER_CONFIRM"
	TransferStateTransferring    = "TRANSFERRING"
	TransferStateSuccess         = "SUCCESS"
	TransferStateFail            = "FAIL"
	TransferStateCanceling       = "CANCELING"
	TransferStateCancelled       = "CANCELLED"
)

// IsTransferTerminal 判断转账状态是否终态。
func IsTransferTerminal(state string) bool {
	return state == TransferStateSuccess ||
		state == TransferStateFail ||
		state == TransferStateCancelled
}

// ===== RealClient.Transfer =====

// Transfer 发起商家转账到零钱。
//
// 接口：POST /v3/fund-app/mch-transfer/transfer-bills
// 注意：本实现暂不传 user_name（仅支持单笔 ≤ 2000 元免实名场景）。
// 如未来需支持大额转账，需引入平台公钥加载与 RSA-OAEP 敏感字段加密。
func (c *RealClient) Transfer(ctx context.Context, req TransferReq) (*TransferResp, error) {
	if req.OutBillNo == "" || req.OpenID == "" || req.TransferSceneID == "" || req.AppID == "" {
		return nil, fmt.Errorf("wxpay transfer: missing required field")
	}
	if req.TransferAmountCents <= 0 {
		return nil, fmt.Errorf("wxpay transfer: invalid amount")
	}

	body := map[string]any{
		"appid":             req.AppID,
		"out_bill_no":       req.OutBillNo,
		"transfer_scene_id": req.TransferSceneID,
		"openid":            req.OpenID,
		"transfer_amount":   req.TransferAmountCents,
		"transfer_remark":   req.TransferRemark,
	}
	if req.NotifyURL != "" {
		body["notify_url"] = req.NotifyURL
	}
	if req.UserRecvPerception != "" {
		body["user_recv_perception"] = req.UserRecvPerception
	}
	if len(req.SceneReportInfos) > 0 {
		body["transfer_scene_report_infos"] = req.SceneReportInfos
	}

	respBody, statusCode, err := c.doRequest(ctx, http.MethodPost, "/v3/fund-app/mch-transfer/transfer-bills", body)
	if err != nil {
		return nil, fmt.Errorf("wxpay transfer http: %w", err)
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("wxpay transfer status=%d body=%s", statusCode, string(respBody))
	}

	var raw struct {
		OutBillNo      string `json:"out_bill_no"`
		TransferBillNo string `json:"transfer_bill_no"`
		CreateTime     string `json:"create_time"`
		State          string `json:"state"`
		PackageInfo    string `json:"package_info"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("wxpay transfer unmarshal: %w", err)
	}

	out := &TransferResp{
		OutBillNo:      raw.OutBillNo,
		TransferBillNo: raw.TransferBillNo,
		State:          raw.State,
		PackageInfo:    raw.PackageInfo,
	}
	if raw.CreateTime != "" {
		if t, err := time.Parse(time.RFC3339, raw.CreateTime); err == nil {
			out.CreateTime = &t
		}
	}
	return out, nil
}

// QueryTransfer 通过商户单号查询转账状态。
func (c *RealClient) QueryTransfer(ctx context.Context, outBillNo string) (*TransferQueryResp, error) {
	if outBillNo == "" {
		return nil, fmt.Errorf("wxpay query transfer: empty out_bill_no")
	}
	path := fmt.Sprintf("/v3/fund-app/mch-transfer/transfer-bills/out-bill-no/%s", outBillNo)
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("wxpay query transfer http: %w", err)
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("wxpay query transfer status=%d body=%s", statusCode, string(respBody))
	}

	var raw struct {
		OutBillNo      string `json:"out_bill_no"`
		TransferBillNo string `json:"transfer_bill_no"`
		State          string `json:"state"`
		TransferAmount int64  `json:"transfer_amount"`
		FailReason     string `json:"fail_reason"`
		CreateTime     string `json:"create_time"`
		UpdateTime     string `json:"update_time"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("wxpay query transfer unmarshal: %w", err)
	}

	out := &TransferQueryResp{
		OutBillNo:      raw.OutBillNo,
		TransferBillNo: raw.TransferBillNo,
		State:          raw.State,
		TransferAmount: raw.TransferAmount,
		FailReason:     raw.FailReason,
	}
	if t, err := time.Parse(time.RFC3339, raw.CreateTime); err == nil {
		out.CreateTime = &t
	}
	if t, err := time.Parse(time.RFC3339, raw.UpdateTime); err == nil {
		out.UpdateTime = &t
	}
	return out, nil
}

// VerifyTransferNotify 校验商家转账回调签名并解密 resource。
//
// 与支付/退款回调结构一致：外层 resource 为 AES-256-GCM，使用 APIKeyV3 解密。
func (c *RealClient) VerifyTransferNotify(_ context.Context, body []byte, headers map[string]string) (*TransferNotifyResult, error) {
	var notify struct {
		Resource struct {
			Algorithm      string `json:"algorithm"`
			CipherText     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notify); err != nil {
		return nil, fmt.Errorf("unmarshal transfer notify: %w", err)
	}

	plaintext, err := decryptAESGCM(c.cfg.APIKeyV3, notify.Resource.CipherText,
		notify.Resource.Nonce, notify.Resource.AssociatedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt transfer notify: %w", err)
	}

	var raw struct {
		OutBillNo      string `json:"out_bill_no"`
		TransferBillNo string `json:"transfer_bill_no"`
		State          string `json:"state"`
		TransferAmount int64  `json:"transfer_amount"`
		OpenID         string `json:"openid"`
		FailReason     string `json:"fail_reason"`
		UpdateTime     string `json:"update_time"`
	}
	if err := json.Unmarshal(plaintext, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal transfer notify resource: %w", err)
	}

	_ = headers // 上层中间件统一做签名校验（与 VerifyNotify 处理方式一致）

	out := &TransferNotifyResult{
		OutBillNo:      raw.OutBillNo,
		TransferBillNo: raw.TransferBillNo,
		State:          raw.State,
		TransferAmount: raw.TransferAmount,
		OpenID:         raw.OpenID,
		FailReason:     raw.FailReason,
		Raw:            body,
	}
	if t, err := time.Parse(time.RFC3339, raw.UpdateTime); err == nil {
		out.UpdateTime = &t
	}
	return out, nil
}

// ===== MockClient.Transfer =====

// Transfer mock：返回预设响应。
func (m *MockClient) Transfer(_ context.Context, req TransferReq) (*TransferResp, error) {
	if m.TransferErr != nil {
		return nil, m.TransferErr
	}
	if m.TransferResult != nil {
		return m.TransferResult, nil
	}
	now := time.Now()
	return &TransferResp{
		OutBillNo:      req.OutBillNo,
		TransferBillNo: "mock_transfer_bill_" + req.OutBillNo,
		State:          TransferStateProcessing,
		CreateTime:     &now,
	}, nil
}

// QueryTransfer mock。
func (m *MockClient) QueryTransfer(_ context.Context, outBillNo string) (*TransferQueryResp, error) {
	if m.QueryTransferErr != nil {
		return nil, m.QueryTransferErr
	}
	if m.QueryTransferResult != nil {
		return m.QueryTransferResult, nil
	}
	now := time.Now()
	return &TransferQueryResp{
		OutBillNo:      outBillNo,
		TransferBillNo: "mock_transfer_bill_" + outBillNo,
		State:          TransferStateSuccess,
		UpdateTime:     &now,
	}, nil
}

// VerifyTransferNotify mock。
func (m *MockClient) VerifyTransferNotify(_ context.Context, _ []byte, _ map[string]string) (*TransferNotifyResult, error) {
	if m.TransferNotifyErr != nil {
		return nil, m.TransferNotifyErr
	}
	if m.TransferNotifyResult != nil {
		return m.TransferNotifyResult, nil
	}
	now := time.Now()
	return &TransferNotifyResult{
		OutBillNo:      "mock_out_bill_no",
		TransferBillNo: "mock_transfer_bill_no",
		State:          TransferStateSuccess,
		TransferAmount: 100,
		UpdateTime:     &now,
		Raw:            []byte(`{"mock":"transfer_notify"}`),
	}, nil
}
