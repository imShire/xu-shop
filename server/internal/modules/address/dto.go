package address

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// CreateAddressReq 新增地址请求。
type CreateAddressReq struct {
	Name         string  `json:"name"          binding:"required,max=32"`
	Phone        string  `json:"phone"         binding:"required,mobile"`
	ProvinceCode *string `json:"province_code"`
	Province     *string `json:"province"`
	CityCode     *string `json:"city_code"`
	City         *string `json:"city"`
	DistrictCode *string `json:"district_code"`
	District     *string `json:"district"`
	StreetCode   *string `json:"street_code"`
	Street       *string `json:"street"`
	Detail       string  `json:"detail"        binding:"required,max=200"`
	IsDefault    bool    `json:"is_default"`
}

// UpdateAddressReq 更新地址请求（所有字段可选）。
type UpdateAddressReq struct {
	Name         *string `json:"name"          binding:"omitempty,max=32"`
	Phone        *string `json:"phone"         binding:"omitempty,mobile"`
	ProvinceCode *string `json:"province_code"`
	Province     *string `json:"province"`
	CityCode     *string `json:"city_code"`
	City         *string `json:"city"`
	DistrictCode *string `json:"district_code"`
	District     *string `json:"district"`
	StreetCode   *string `json:"street_code"`
	Street       *string `json:"street"`
	Detail       *string `json:"detail"        binding:"omitempty,max=200"`
	IsDefault    *bool   `json:"is_default"`
}

// DecryptWxAddressReq 解密微信地址请求。
type DecryptWxAddressReq struct {
	EncryptedData string `json:"encrypted_data" binding:"required"`
	IV            string `json:"iv"             binding:"required"`
}

// AddressResp 地址响应 DTO。
type AddressResp struct {
	ID           types.Int64Str `json:"id"`
	Name         string         `json:"name"`
	Phone        string         `json:"phone"`
	ProvinceCode *string        `json:"province_code,omitempty"`
	Province     *string        `json:"province,omitempty"`
	CityCode     *string        `json:"city_code,omitempty"`
	City         *string        `json:"city,omitempty"`
	DistrictCode *string        `json:"district_code,omitempty"`
	District     *string        `json:"district,omitempty"`
	StreetCode   *string        `json:"street_code,omitempty"`
	Street       *string        `json:"street,omitempty"`
	Detail       string         `json:"detail"`
	IsDefault    bool           `json:"is_default"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// RegionResp 行政区划响应 DTO。
type RegionResp struct {
	Code        string  `json:"code"`
	ParentCode  *string `json:"parent_code,omitempty"`
	Name        string  `json:"name"`
	Level       int     `json:"level"`
	HasChildren bool    `json:"has_children"`
}

// toAddressResp 将 entity 转为响应 DTO。
func toAddressResp(a *Address) AddressResp {
	return AddressResp{
		ID:           types.Int64Str(a.ID),
		Name:         a.Name,
		Phone:        a.Phone,
		ProvinceCode: a.ProvinceCode,
		Province:     a.Province,
		CityCode:     a.CityCode,
		City:         a.City,
		DistrictCode: a.DistrictCode,
		District:     a.District,
		StreetCode:   a.StreetCode,
		Street:       a.Street,
		Detail:       a.Detail,
		IsDefault:    a.IsDefault,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func toRegionResp(r *Region) RegionResp {
	return RegionResp{
		Code:        r.Code,
		ParentCode:  r.ParentCode,
		Name:        r.Name,
		Level:       r.Level,
		HasChildren: r.HasChildren,
	}
}
