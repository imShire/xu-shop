// Package address 实现用户收货地址和行政区划相关逻辑。
package address

import "time"

// Address 用户收货地址。
type Address struct {
	ID           int64     `gorm:"primaryKey"`
	UserID       int64     `gorm:"column:user_id;not null;index"`
	Name         string    `gorm:"column:name;not null"`
	Phone        string    `gorm:"column:phone;not null"`
	ProvinceCode *string   `gorm:"column:province_code"`
	Province     *string   `gorm:"column:province"`
	CityCode     *string   `gorm:"column:city_code"`
	City         *string   `gorm:"column:city"`
	DistrictCode *string   `gorm:"column:district_code"`
	District     *string   `gorm:"column:district"`
	StreetCode   *string   `gorm:"column:street_code"`
	Street       *string   `gorm:"column:street"`
	Detail       string    `gorm:"column:detail;not null"`
	IsDefault    bool      `gorm:"column:is_default;default:false"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Address) TableName() string { return "address" }

// Region 行政区划（省/市/区县/街道）。
type Region struct {
	Code        string  `gorm:"primaryKey;column:code"`
	ParentCode  *string `gorm:"column:parent_code"`
	Name        string  `gorm:"column:name;not null"`
	Level       int     `gorm:"column:level;not null"` // 1=省 2=市 3=区县 4=街道
	Sort        int     `gorm:"column:sort;default:0"`
	HasChildren bool    `gorm:"column:has_children;->;-:migration"`
}

func (Region) TableName() string { return "region" }
