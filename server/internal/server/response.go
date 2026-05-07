// Package server 提供统一 HTTP 响应工具函数。
package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// Response 统一响应结构。
type Response struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

// OK 返回成功响应（HTTP 200）。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Fail 返回业务错误响应并终止后续处理。
func Fail(c *gin.Context, err *errs.AppError) {
	c.AbortWithStatusJSON(err.HTTPStatus, Response{
		Code:      err.Code,
		Message:   err.Message,
		RequestID: getRequestID(c),
	})
}

// FailParam 将 ShouldBindJSON / ShouldBindQuery 的 err 转为可读的参数错误响应。
// 如果 err 是 validator.ValidationErrors，提取每个字段的具体原因；否则直接用 err.Error()。
func FailParam(c *gin.Context, err error) {
	msg := bindErrMsg(err)
	c.AbortWithStatusJSON(http.StatusBadRequest, Response{
		Code:      errs.ErrParam.Code,
		Message:   msg,
		RequestID: getRequestID(c),
	})
}

// bindErrMsg 将 binding error 转换为中文可读消息。
func bindErrMsg(err error) string {
	var ve validator.ValidationErrors
	if !errorAs(err, &ve) {
		// JSON 解析错误或其他 binding 错误，直接返回
		msg := err.Error()
		// 去掉冗长的 json 前缀使消息更干净
		if idx := strings.Index(msg, "cannot unmarshal"); idx >= 0 {
			return "请求体格式错误：" + msg[idx:]
		}
		return "请求参数格式错误"
	}

	parts := make([]string, 0, len(ve))
	for _, fe := range ve {
		parts = append(parts, fmt.Sprintf("%s %s", fieldLabel(fe.Field()), tagMsg(fe.Tag(), fe.Param())))
	}
	return strings.Join(parts, "；")
}

// fieldLabel 将英文字段名转为可读标签（未配置时原样返回小写）。
func fieldLabel(field string) string {
	labels := map[string]string{
		"Phone":          "手机号",
		"Password":       "密码",
		"Nickname":       "昵称",
		"Name":           "姓名",
		"Province":       "省",
		"City":           "市",
		"District":       "区/县",
		"Detail":         "详细地址",
		"AddressID":      "收货地址",
		"Items":          "商品列表",
		"SkuID":          "SKU",
		"Qty":            "数量",
		"Title":          "标题",
		"Price":          "价格",
		"Stock":          "库存",
		"CategoryID":     "分类",
		"IdempotencyKey": "幂等键",
		"Code":           "授权码",
		"Username":       "用户名",
		"Delta":          "调整数量",
		"OrderID":        "订单ID",
		"Scene":          "支付场景",
		"AmountCents":    "退款金额",
		"Reason":         "原因",
		"TrackingNo":     "运单号",
		"CarrierCode":    "快递公司",
		"Driver":         "上传驱动",
		"PublicBaseURL":  "访问前缀",
		"MaxSizeMB":      "大小限制",
		"LocalDir":       "本地上传目录",
		"S3Endpoint":     "S3 Endpoint",
		"S3Bucket":       "S3 Bucket",
		"S3AccessKey":    "S3 Access Key",
		"S3SecretKey":    "S3 Secret Key",
	}
	if l, ok := labels[field]; ok {
		return l
	}
	return field
}

// tagMsg 将 validator tag 转为中文描述。
func tagMsg(tag, param string) string {
	switch tag {
	case "required":
		return "不能为空"
	case "min":
		return fmt.Sprintf("最少 %s 个字符", param)
	case "max":
		return fmt.Sprintf("最多 %s 个字符", param)
	case "len":
		return fmt.Sprintf("长度必须为 %s", param)
	case "email":
		return "格式不正确（需为邮箱）"
	case "numeric":
		return "必须为数字"
	case "gte":
		return fmt.Sprintf("必须大于等于 %s", param)
	case "gt":
		return fmt.Sprintf("必须大于 %s", param)
	case "lte":
		return fmt.Sprintf("必须小于等于 %s", param)
	case "lt":
		return fmt.Sprintf("必须小于 %s", param)
	case "oneof":
		return fmt.Sprintf("必须是以下之一：%s", param)
	case "url":
		return "必须是有效的 URL"
	default:
		return fmt.Sprintf("校验失败（%s）", tag)
	}
}

// errorAs is errors.As but avoids importing errors package at call site.
func errorAs(err error, target any) bool {
	type asInterface interface {
		As(any) bool
	}
	// Use standard errors.As behaviour via type assertion chain
	type unwrapper interface{ Unwrap() error }
	for err != nil {
		if x, ok := target.(*validator.ValidationErrors); ok {
			if ve, ok2 := err.(validator.ValidationErrors); ok2 {
				*x = ve
				return true
			}
		}
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}

// getRequestID 从 gin.Context 取 request_id。
func getRequestID(c *gin.Context) string {
	return c.GetString("request_id")
}
