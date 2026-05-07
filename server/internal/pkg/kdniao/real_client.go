package kdniao

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const kdniaoHTTPTimeout = 10 * time.Second

// kdniaoResponse 快递鸟通用响应结构。
type kdniaoResponse struct {
	EBusinessID  string `json:"EBusinessID"`
	ResultCode   string `json:"ResultCode"`
	Reason       string `json:"Reason"`
	Success      bool   `json:"Success"`
}

// acquire 获取限速 + 并发令牌。
func (c *RealClient) acquire(ctx context.Context) error {
	// token bucket 限速
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("kdniao: rate limit wait: %w", err)
	}
	// 并发信号量
	select {
	case c.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("kdniao: semaphore wait canceled: %w", ctx.Err())
	}
}

// release 释放并发令牌。
func (c *RealClient) release() {
	<-c.sem
}

// sign 生成快递鸟签名（MD5(RequestData + APIKey) Base64）。
func (c *RealClient) sign(requestData string) string {
	raw := requestData + c.apiKey
	h := md5.Sum([]byte(raw))
	return base64.StdEncoding.EncodeToString(h[:])
}

// doPost 向快递鸟 API 发起 POST 请求。
func (c *RealClient) doPost(ctx context.Context, requestType, requestData string) ([]byte, error) {
	if err := c.acquire(ctx); err != nil {
		return nil, err
	}
	defer c.release()

	sign := c.sign(requestData)

	form := url.Values{}
	form.Set("RequestType", requestType)
	form.Set("EBusinessID", c.businessID)
	form.Set("RequestData", requestData)
	form.Set("DataSign", sign)
	form.Set("DataType", "2") // JSON

	client := &http.Client{Timeout: kdniaoHTTPTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.reqURL,
		bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("kdniao: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kdniao: http do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kdniao: read body: %w", err)
	}
	return body, nil
}

// CreateWaybill 创建电子面单。
func (c *RealClient) CreateWaybill(ctx context.Context, req CreateWaybillReq) (*CreateWaybillResp, error) {
	goods := make([]map[string]any, len(req.Goods))
	for i, g := range req.Goods {
		goods[i] = map[string]any{"GoodsName": g.Name, "Quantity": g.Qty, "Weight": g.Weight}
	}

	reqData, err := json.Marshal(map[string]any{
		"ShipperCode":     req.CarrierCode,
		"OrderCode":       req.OrderNo,
		"MonthCode":       req.MonthlyAccount,
		"IsSendMessage":   "0",
		"PayType":         "1", // 月结
		"ExpType":         "1",
		"Sender":          addrToMap(req.Sender),
		"Receiver":        addrToMap(req.Receiver),
		"Commodity":       goods,
		"Remark":          "",
		"IsNotice":        "0",
		"CallBackUrl":     req.CallbackURL,
	})
	if err != nil {
		return nil, fmt.Errorf("kdniao: marshal waybill req: %w", err)
	}

	body, err := c.doPost(ctx, "1007", string(reqData))
	if err != nil {
		return nil, err
	}

	var result struct {
		kdniaoResponse
		Order struct {
			LogisticCode string `json:"LogisticCode"`
		} `json:"Order"`
		PrintTemplate string `json:"PrintTemplate"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kdniao: unmarshal waybill resp: %w", err)
	}

	return &CreateWaybillResp{
		TrackingNo:  result.Order.LogisticCode,
		PrintBase64: result.PrintTemplate,
		Success:     result.Success,
		Reason:      result.Reason,
	}, nil
}

// Track 查询物流轨迹。
func (c *RealClient) Track(ctx context.Context, carrier, no string) (*TrackResp, error) {
	reqData, _ := json.Marshal(map[string]any{
		"OrderCode":    "",
		"ShipperCode":  carrier,
		"LogisticCode": no,
	})

	body, err := c.doPost(ctx, "1002", string(reqData))
	if err != nil {
		return nil, err
	}

	var result struct {
		kdniaoResponse
		ShipperCode  string `json:"ShipperCode"`
		LogisticCode string `json:"LogisticCode"`
		State        string `json:"State"`
		Traces       []struct {
			AcceptTime    string `json:"AcceptTime"`
			AcceptStation string `json:"AcceptStation"`
			Remark        string `json:"Remark"`
		} `json:"Traces"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kdniao: unmarshal track resp: %w", err)
	}

	traces := make([]TraceItem, 0, len(result.Traces))
	for _, t := range result.Traces {
		ts, _ := time.ParseInLocation("2006/01/02 15:04:05", t.AcceptTime, time.Local)
		desc := t.AcceptStation
		if t.Remark != "" {
			desc += " " + t.Remark
		}
		traces = append(traces, TraceItem{
			OccurredAt:  ts,
			Status:      mapKDState(result.State),
			Description: strings.TrimSpace(desc),
		})
	}

	return &TrackResp{
		Carrier:  result.ShipperCode,
		TrackNo:  result.LogisticCode,
		State:    mapKDState(result.State),
		Traces:   traces,
	}, nil
}

// Subscribe 订阅物流轨迹推送。
func (c *RealClient) Subscribe(ctx context.Context, carrier, no, callbackURL string) error {
	reqData, _ := json.Marshal(map[string]any{
		"OrderCode":    "",
		"ShipperCode":  carrier,
		"LogisticCode": no,
		"CallBackUrl":  callbackURL,
	})

	body, err := c.doPost(ctx, "1008", string(reqData))
	if err != nil {
		return err
	}

	var result kdniaoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("kdniao: unmarshal subscribe resp: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("kdniao: subscribe failed: %s", result.Reason)
	}
	return nil
}

// ParsePush 解析快递鸟推送 body（application/x-www-form-urlencoded）。
func (c *RealClient) ParsePush(body []byte) (*PushResp, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("kdniao: parse push body: %w", err)
	}

	requestData := values.Get("RequestData")
	if requestData == "" {
		return nil, fmt.Errorf("kdniao: empty RequestData in push")
	}

	var data struct {
		ShipperCode  string `json:"ShipperCode"`
		LogisticCode string `json:"LogisticCode"`
		State        string `json:"State"`
		Traces       []struct {
			AcceptTime    string `json:"AcceptTime"`
			AcceptStation string `json:"AcceptStation"`
			Remark        string `json:"Remark"`
		} `json:"Traces"`
	}
	if err := json.Unmarshal([]byte(requestData), &data); err != nil {
		// 可能是数组包裹
		var arr []json.RawMessage
		if err2 := json.Unmarshal([]byte(requestData), &arr); err2 != nil || len(arr) == 0 {
			return nil, fmt.Errorf("kdniao: unmarshal push data: %w", err)
		}
		if err3 := json.Unmarshal(arr[0], &data); err3 != nil {
			return nil, fmt.Errorf("kdniao: unmarshal push data[0]: %w", err3)
		}
	}

	traces := make([]TraceItem, 0, len(data.Traces))
	for _, t := range data.Traces {
		ts, _ := time.ParseInLocation("2006/01/02 15:04:05", t.AcceptTime, time.Local)
		desc := t.AcceptStation
		if t.Remark != "" {
			desc += " " + t.Remark
		}
		traces = append(traces, TraceItem{
			OccurredAt:  ts,
			Status:      mapKDState(data.State),
			Description: strings.TrimSpace(desc),
		})
	}

	return &PushResp{
		CarrierCode: data.ShipperCode,
		TrackingNo:  data.LogisticCode,
		State:       mapKDState(data.State),
		Traces:      traces,
	}, nil
}

// addrToMap 将 Addr 转为快递鸟接口所需 map。
func addrToMap(a Addr) map[string]any {
	return map[string]any{
		"Company":      a.Company,
		"Name":         a.Name,
		"Tel":          a.Phone,
		"ProvinceName": a.Province,
		"CityName":     a.City,
		"ExpAreaName":  a.District,
		"Address":      a.Detail,
	}
}

// mapKDState 将快递鸟 State 数字映射到内部状态字符串。
func mapKDState(state string) string {
	switch state {
	case "1":
		return "picked"
	case "2":
		return "in_transit"
	case "3":
		return "delivered"
	case "4":
		return "problem"
	case "5":
		return "returned"
	case "6":
		return "returning"
	default:
		return "unknown"
	}
}
