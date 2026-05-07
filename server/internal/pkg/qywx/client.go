// Package qywx 封装企业微信机器人 + 客户联系 API。
package qywx

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AddContactWayReq 创建联系我（渠道码）请求。
type AddContactWayReq struct {
	Type          int      `json:"type"`           // 1=单人 2=多人
	Scene         int      `json:"scene"`          // 1=客户扫码 2=公众号
	Remark        string   `json:"remark,omitempty"`
	SkipVerify    bool     `json:"skip_verify"`
	State         string   `json:"state,omitempty"`
	User          []string `json:"user,omitempty"`
	Party         []int    `json:"party,omitempty"`
	IsTemp        bool     `json:"is_temp,omitempty"`
	ExpiresIn     int      `json:"expires_in,omitempty"`
	ChatExpiresIn int      `json:"chat_expires_in,omitempty"`
	UnionID       string   `json:"unionid,omitempty"`
	Conclusions   any      `json:"conclusions,omitempty"`
}

// AddContactWayResp 创建联系我响应。
type AddContactWayResp struct {
	ConfigID string `json:"config_id"`
	QrCode   string `json:"qr_code"`
}

// WelcomeMsg 欢迎语消息体。
type WelcomeMsg struct {
	Text        *WelcomeText        `json:"text,omitempty"`
	Attachments []WelcomeAttachment `json:"attachments,omitempty"`
}

// WelcomeText 欢迎语文本内容。
type WelcomeText struct {
	Content string `json:"content"`
}

// WelcomeAttachment 欢迎语附件（小程序/图片等）。
type WelcomeAttachment struct {
	Msgtype    string      `json:"msgtype"`
	Miniprogram *MiniProgram `json:"miniprogram,omitempty"`
}

// MiniProgram 小程序附件。
type MiniProgram struct {
	Title        string `json:"title"`
	PicMediaID   string `json:"pic_media_id"`
	AppID        string `json:"appid"`
	Page         string `json:"page"`
}

// MarkTagReq 打企业标签请求。
type MarkTagReq struct {
	UserID         string   `json:"userid"`
	ExternalUserID string   `json:"external_userid"`
	AddTag         []string `json:"add_tag,omitempty"`
	RemoveTag      []string `json:"remove_tag,omitempty"`
}

// CallbackEvent 企微回调事件（解密后）。
type CallbackEvent struct {
	MsgType  string
	Event    string
	ToUserID string
	FromUser string
	// 客户联系相关
	ExternalUserID string
	State          string
	WelcomeCode    string
	// 其余原始 XML（供业务层自行解析）
	Raw []byte
}

// callbackXML 回调 XML 外层结构。
type callbackXML struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
}

// callbackInnerXML 解密后内层 XML 结构。
type callbackInnerXML struct {
	XMLName        xml.Name `xml:"xml"`
	ToUserName     string   `xml:"ToUserName"`
	FromUserName   string   `xml:"FromUserName"`
	MsgType        string   `xml:"MsgType"`
	Event          string   `xml:"Event"`
	ChangeType     string   `xml:"ChangeType"`
	ExternalUserID string   `xml:"ExternalUserID"`
	State          string   `xml:"State"`
	WelcomeCode    string   `xml:"WelcomeCode"`
}

// Client 企业微信客户端接口。
type Client interface {
	SendRobotMarkdown(ctx context.Context, webhookURL, content string) error
	AddContactWay(ctx context.Context, req AddContactWayReq) (*AddContactWayResp, error)
	SendWelcomeMsg(ctx context.Context, welcomeCode string, msg WelcomeMsg) error
	MarkCustomerTag(ctx context.Context, req MarkTagReq) error
	AddCorpTag(ctx context.Context, name, groupID string) (tagID string, err error)
	DeleteCorpTag(ctx context.Context, tagID string) error
	DecryptCallback(ctx context.Context, body []byte, params map[string]string) (*CallbackEvent, error)
}

type qywxClient struct {
	corpID        string
	secret        string
	encodingAESKey string
	httpCli       *http.Client
}

// NewClient 构造真实企微客户端。
func NewClient(corpID, secret, encodingAESKey string) Client {
	return &qywxClient{
		corpID:         corpID,
		secret:         secret,
		encodingAESKey: encodingAESKey,
		httpCli:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *qywxClient) getAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		c.corpID, c.secret)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("qywx: get token: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("qywx: parse token: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("qywx: token errcode=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return result.AccessToken, nil
}

func (c *qywxClient) post(ctx context.Context, url string, payload any) ([]byte, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	apiURL := url + "?access_token=" + token
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("qywx: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qywx: do request: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func checkErrCode(data []byte) error {
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("qywx: parse response: %w", err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("qywx: errcode=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// SendRobotMarkdown 机器人发 markdown 消息。
func (c *qywxClient) SendRobotMarkdown(ctx context.Context, webhookURL, content string) error {
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{"content": content},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("qywx: marshal markdown: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("qywx: build webhook req: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("qywx: send robot markdown: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return checkErrCode(data)
}

// AddContactWay 创建联系我（渠道码）。
func (c *qywxClient) AddContactWay(ctx context.Context, req AddContactWayReq) (*AddContactWayResp, error) {
	data, err := c.post(ctx, "https://qyapi.weixin.qq.com/cgi-bin/externalcontact/add_contact_way", req)
	if err != nil {
		return nil, err
	}
	var result struct {
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
		ConfigID string `json:"config_id"`
		QrCode   string `json:"qr_code"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("qywx: parse add_contact_way: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("qywx: add_contact_way errcode=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return &AddContactWayResp{ConfigID: result.ConfigID, QrCode: result.QrCode}, nil
}

// SendWelcomeMsg 发送欢迎语。
func (c *qywxClient) SendWelcomeMsg(ctx context.Context, welcomeCode string, msg WelcomeMsg) error {
	payload := map[string]any{
		"welcome_code": welcomeCode,
		"text":         msg.Text,
		"attachments":  msg.Attachments,
	}
	data, err := c.post(ctx, "https://qyapi.weixin.qq.com/cgi-bin/externalcontact/send_welcome_msg", payload)
	if err != nil {
		return err
	}
	return checkErrCode(data)
}

// MarkCustomerTag 打企业标签。
func (c *qywxClient) MarkCustomerTag(ctx context.Context, req MarkTagReq) error {
	data, err := c.post(ctx, "https://qyapi.weixin.qq.com/cgi-bin/externalcontact/mark_tag", req)
	if err != nil {
		return err
	}
	return checkErrCode(data)
}

// AddCorpTag 创建企业标签。
func (c *qywxClient) AddCorpTag(ctx context.Context, name, groupID string) (string, error) {
	payload := map[string]any{
		"group_id": groupID,
		"tag":      []map[string]string{{"name": name}},
	}
	data, err := c.post(ctx, "https://qyapi.weixin.qq.com/cgi-bin/externalcontact/add_corp_tag", payload)
	if err != nil {
		return "", err
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TagGroup struct {
			Tag []struct {
				ID string `json:"id"`
			} `json:"tag"`
		} `json:"tag_group"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("qywx: parse add_corp_tag: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("qywx: add_corp_tag errcode=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	if len(result.TagGroup.Tag) == 0 {
		return "", fmt.Errorf("qywx: add_corp_tag no tag returned")
	}
	return result.TagGroup.Tag[0].ID, nil
}

// DeleteCorpTag 删除企业标签。
func (c *qywxClient) DeleteCorpTag(ctx context.Context, tagID string) error {
	payload := map[string]any{"tag_id": []string{tagID}}
	data, err := c.post(ctx, "https://qyapi.weixin.qq.com/cgi-bin/externalcontact/del_corp_tag", payload)
	if err != nil {
		return err
	}
	return checkErrCode(data)
}

// DecryptCallback 解密企微回调消息（AES-256-CBC）。
func (c *qywxClient) DecryptCallback(_ context.Context, body []byte, _ map[string]string) (*CallbackEvent, error) {
	var outer callbackXML
	if err := xml.Unmarshal(body, &outer); err != nil {
		return nil, fmt.Errorf("qywx: parse callback xml: %w", err)
	}

	if c.encodingAESKey == "" {
		return &CallbackEvent{Raw: body}, nil
	}

	plaintext, err := decryptAES(c.encodingAESKey, outer.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("qywx: decrypt callback: %w", err)
	}

	var inner callbackInnerXML
	if err := xml.Unmarshal(plaintext, &inner); err != nil {
		return nil, fmt.Errorf("qywx: parse inner xml: %w", err)
	}

	event := inner.Event
	if inner.ChangeType != "" {
		event = inner.ChangeType
	}

	return &CallbackEvent{
		MsgType:        inner.MsgType,
		Event:          event,
		ToUserID:       inner.ToUserName,
		FromUser:       inner.FromUserName,
		ExternalUserID: inner.ExternalUserID,
		State:          inner.State,
		WelcomeCode:    inner.WelcomeCode,
		Raw:            plaintext,
	}, nil
}

// decryptAES 解密企微 AES-256-CBC 加密内容。
func decryptAES(encodingAESKey, encryptedMsg string) ([]byte, error) {
	aesKey, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("decode aes key: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted msg: %w", err)
	}

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("invalid aes key length: %d", len(aesKey))
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// remove PKCS7 padding
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("empty plaintext")
	}
	pad := int(ciphertext[len(ciphertext)-1])
	if pad > aes.BlockSize || pad == 0 || pad > len(ciphertext) {
		return nil, fmt.Errorf("invalid padding")
	}
	plaintext := ciphertext[:len(ciphertext)-pad]

	// 格式：16字节random + 4字节消息长度(big-endian) + 消息 + appid
	if len(plaintext) < 20 {
		return nil, fmt.Errorf("plaintext too short")
	}
	msgLen := int(binary.BigEndian.Uint32(plaintext[16:20]))
	if 20+msgLen > len(plaintext) {
		return nil, fmt.Errorf("invalid msg length")
	}
	return plaintext[20 : 20+msgLen], nil
}
