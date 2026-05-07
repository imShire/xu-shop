// Package stats 实现经营数据看板聚合查询。
package stats

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// StatsDaily 每日汇总数据。
type StatsDaily struct {
	Date               string    `gorm:"column:date;primaryKey" json:"date"`
	PaidOrderCount     int       `gorm:"column:paid_order_count" json:"paid_order_count"`
	PaidAmountCents    int64     `gorm:"column:paid_amount_cents" json:"paid_amount_cents"`
	RefundAmountCents  int64     `gorm:"column:refund_amount_cents" json:"refund_amount_cents"`
	NetAmountCents     int64     `gorm:"column:net_amount_cents" json:"net_amount_cents"`
	PaidUserCount      int       `gorm:"column:paid_user_count" json:"paid_user_count"`
	NewUserCount       int       `gorm:"column:new_user_count" json:"new_user_count"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StatsDaily) TableName() string { return "stats_daily" }

// StatsProductDaily 商品每日数据。
type StatsProductDaily struct {
	Date        string `gorm:"column:date;primaryKey" json:"date"`
	ProductID   int64  `gorm:"column:product_id;primaryKey" json:"product_id"`
	Qty         int    `gorm:"column:qty" json:"qty"`
	AmountCents int64  `gorm:"column:amount_cents" json:"amount_cents"`
}

func (StatsProductDaily) TableName() string { return "stats_product_daily" }

// StatsChannelDaily 渠道每日数据。
type StatsChannelDaily struct {
	Date          string `gorm:"column:date;primaryKey" json:"date"`
	ChannelCodeID int64  `gorm:"column:channel_code_id;primaryKey" json:"channel_code_id"`
	ScanCount     int    `gorm:"column:scan_count" json:"scan_count"`
	AddCount      int    `gorm:"column:add_count" json:"add_count"`
	OrderCount    int    `gorm:"column:order_count" json:"order_count"`
	AmountCents   int64  `gorm:"column:amount_cents" json:"amount_cents"`
}

func (StatsChannelDaily) TableName() string { return "stats_channel_daily" }

// ---- DTO ----

// OverviewResp 总览数据（含同比）。
type OverviewResp struct {
	PaidOrderCount    int     `json:"paid_order_count"`
	PaidAmountCents   int64   `json:"paid_amount_cents"`
	NetAmountCents    int64   `json:"net_amount_cents"`
	RefundAmountCents int64   `json:"refund_amount_cents"`
	PaidUserCount     int     `json:"paid_user_count"`
	NewUserCount      int     `json:"new_user_count"`
	// YoY 同比（百分比，如 0.1 = 10%）
	OrderCountYoY  float64 `json:"order_count_yoy"`
	AmountYoY      float64 `json:"amount_yoy"`
}

// DailyPoint 折线图数据点。
type DailyPoint struct {
	Date        string `json:"date"`
	AmountCents int64  `json:"amount_cents"`
	OrderCount  int    `json:"order_count"`
}

// CategoryShare 分类占比。
type CategoryShare struct {
	CategoryID   types.Int64Str `json:"category_id"`
	CategoryName string         `json:"category_name"`
	AmountCents  int64          `json:"amount_cents"`
	Percent      float64        `json:"percent"`
}

// TopProductRow 商品销售排行。
type TopProductRow struct {
	ProductID   types.Int64Str `json:"product_id"`
	ProductName string         `json:"product_name"`
	Qty         int            `json:"qty"`
	AmountCents int64          `json:"amount_cents"`
}

// UserStatsResp 用户统计。
type UserStatsResp struct {
	NewUserCount  int `json:"new_user_count"`
	PaidUserCount int `json:"paid_user_count"`
}

// ChannelStatsRow 渠道统计行。
type ChannelStatsRow struct {
	ChannelCodeID   types.Int64Str `json:"channel_code_id"`
	ChannelCodeName string         `json:"channel_code_name"`
	ScanCount       int    `json:"scan_count"`
	AddCount        int    `json:"add_count"`
	OrderCount      int    `json:"order_count"`
	AmountCents     int64  `json:"amount_cents"`
}

// WorkbenchResp 工作台概览数据。
type WorkbenchResp struct {
	TodayOrderCount int   `json:"today_order_count"`
	TodaySales      int64 `json:"today_sales"`
	PendingShip     int   `json:"pending_ship"`
	AftersalePending int  `json:"aftersale_pending"`
}
