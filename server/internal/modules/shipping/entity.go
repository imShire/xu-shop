// Package shipping 实现发货、物流追踪和快递鸟封装。
package shipping

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 状态常量 ----

const (
	ShipStatusPicked    = "picked"
	ShipStatusInTransit = "in_transit"
	ShipStatusDelivered = "delivered"
	ShipStatusException = "exception"
	ShipStatusReturning = "returning"
	ShipStatusReturned  = "returned"
)

// ---- 自定义 JSONB 类型 ----

// RawJSON 通用 JSONB 字段。
type RawJSON []byte

func (j RawJSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "null", nil
	}
	return string(j), nil
}

func (j *RawJSON) Scan(value any) error {
	if value == nil {
		*j = RawJSON("null")
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = RawJSON(v)
	default:
		return fmt.Errorf("RawJSON: unsupported type %T", value)
	}
	return nil
}

func (j RawJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *RawJSON) UnmarshalJSON(data []byte) error {
	*j = append((*j)[0:0], data...)
	return nil
}

// ---- GORM 模型 ----

// SenderAddress 发货地址。
type SenderAddress struct {
	ID           int64     `gorm:"primaryKey"             json:"id"`
	Company      *string   `gorm:"size:64"                json:"company,omitempty"`
	Name         string    `gorm:"size:32;not null"       json:"name"`
	Phone        string    `gorm:"size:20;not null"       json:"phone"`
	ProvinceCode *string   `gorm:"size:12"                json:"province_code,omitempty"`
	Province     string    `gorm:"size:32;not null"       json:"province"`
	CityCode     *string   `gorm:"size:12"                json:"city_code,omitempty"`
	City         string    `gorm:"size:32;not null"       json:"city"`
	DistrictCode *string   `gorm:"size:12"                json:"district_code,omitempty"`
	District     string    `gorm:"size:32;not null"       json:"district"`
	StreetCode   *string   `gorm:"size:12"                json:"street_code,omitempty"`
	Street       string    `gorm:"size:32"                json:"street"`
	Detail       string    `gorm:"size:200;not null"      json:"detail"`
	IsDefault    bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SenderAddress) TableName() string { return "sender_address" }

// Carrier 快递公司。
type Carrier struct {
	Code           string  `gorm:"primaryKey;size:16"    json:"code"`
	Name           string  `gorm:"size:32;not null"      json:"name"`
	KdniaoCode     string  `gorm:"size:16;not null"      json:"kdniao_code"`
	MonthlyAccount *string `gorm:"size:64"               json:"monthly_account,omitempty"`
	Enabled        bool    `gorm:"not null;default:true" json:"enabled"`
	Sort           int     `gorm:"not null;default:0"    json:"sort"`
}

func (Carrier) TableName() string { return "carrier" }

// AddressSnapshot 地址快照（面单发件人/收件人）。
type AddressSnapshot struct {
	Company      string `json:"company,omitempty"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	ProvinceCode string `json:"province_code,omitempty"`
	Province     string `json:"province"`
	CityCode     string `json:"city_code,omitempty"`
	City         string `json:"city"`
	DistrictCode string `json:"district_code,omitempty"`
	District     string `json:"district"`
	StreetCode   string `json:"street_code,omitempty"`
	Street       string `json:"street,omitempty"`
	Detail       string `json:"detail"`
}

func marshalSnapshot(a AddressSnapshot) RawJSON {
	b, _ := json.Marshal(a)
	return RawJSON(b)
}

// Shipment 发货单。
type Shipment struct {
	ID               int64   `gorm:"primaryKey"`
	OrderID          int64   `gorm:"not null"`
	CarrierCode      string  `gorm:"size:16;not null"`
	TrackingNo       string  `gorm:"size:64;not null"`
	WaybillPdfURL    *string `gorm:"size:512"`
	SenderSnapshot   RawJSON `gorm:"type:jsonb;not null"`
	ReceiverSnapshot RawJSON `gorm:"type:jsonb;not null"`
	Status           string  `gorm:"size:16;not null;default:picked"`
	LastTrackAt      *time.Time
	DeliveredAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (Shipment) TableName() string { return "shipment" }

// ShipmentTrack 物流轨迹。
type ShipmentTrack struct {
	ID          int64     `gorm:"primaryKey"`
	ShipmentID  int64     `gorm:"not null"`
	Status      string    `gorm:"size:16;not null"`
	Description *string   `gorm:"type:text"`
	OccurredAt  time.Time `gorm:"not null"`
	Raw         RawJSON   `gorm:"type:jsonb"`
	CreatedAt   time.Time
}

func (ShipmentTrack) TableName() string { return "shipment_track" }

// ---- DTO ----

// ShipReq 发货请求。
type ShipReq struct {
	CarrierCode   string          `json:"carrier_code"    binding:"required"`
	TrackingNo    string          `json:"tracking_no"`    // 手动填入（不用面单时）
	CreateWaybill bool            `json:"create_waybill"` // 是否调用快递鸟建单
	Goods         []ShipGoodsItem `json:"goods"`
}

// ShipGoodsItem 货物明细。
type ShipGoodsItem struct {
	Name   string  `json:"name"`
	Qty    int     `json:"qty"`
	Weight float64 `json:"weight"`
}

// BatchShipReq 批量发货请求。
type BatchShipReq struct {
	Orders []BatchShipItem `json:"orders" binding:"required,min=1"`
}

// BatchShipItem 单条批量发货条目。
type BatchShipItem struct {
	OrderID     types.Int64Str `json:"order_id"`
	CarrierCode string         `json:"carrier_code"`
	TrackingNo  string         `json:"tracking_no"`
}

// BatchShipStatus 批量发货任务状态。
type BatchShipStatus struct {
	TaskID string   `json:"task_id"`
	Total  int      `json:"total"`
	Done   int      `json:"done"`
	Failed int      `json:"failed"`
	Errors []string `json:"errors,omitempty"`
	PDFURL string   `json:"pdf_url,omitempty"`
}

// UpdateShipmentReq 修改发货单请求。
type UpdateShipmentReq struct {
	TrackingNo  string `json:"tracking_no"`
	CarrierCode string `json:"carrier_code"`
}

// ShipmentFilter 后台发货单列表过滤。
type ShipmentFilter struct {
	OrderID int64
	Status  string
	Page    int
	Size    int
}

// ShipmentResp 发货单响应 DTO。
type ShipmentResp struct {
	ID               types.Int64Str `json:"id"`
	OrderID          types.Int64Str `json:"order_id"`
	CarrierCode      string         `json:"carrier_code"`
	TrackingNo       string         `json:"tracking_no"`
	WaybillPdfURL    *string        `json:"waybill_pdf_url,omitempty"`
	SenderSnapshot   RawJSON        `json:"sender_snapshot"`
	ReceiverSnapshot RawJSON        `json:"receiver_snapshot"`
	Status           string         `json:"status"`
	LastTrackAt      *time.Time     `json:"last_track_at,omitempty"`
	DeliveredAt      *time.Time     `json:"delivered_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// toShipmentResp entity → ShipmentResp。
func toShipmentResp(s *Shipment) ShipmentResp {
	return ShipmentResp{
		ID:               types.Int64Str(s.ID),
		OrderID:          types.Int64Str(s.OrderID),
		CarrierCode:      s.CarrierCode,
		TrackingNo:       s.TrackingNo,
		WaybillPdfURL:    s.WaybillPdfURL,
		SenderSnapshot:   s.SenderSnapshot,
		ReceiverSnapshot: s.ReceiverSnapshot,
		Status:           s.Status,
		LastTrackAt:      s.LastTrackAt,
		DeliveredAt:      s.DeliveredAt,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
