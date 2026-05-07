package address

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/wxlogin"
)

const maxAddressPerUser = 20

// Service 地址服务。
type Service struct {
	addrRepo   AddressRepo
	regionRepo RegionRepo
	rdb        *redis.Client
	wxMP       wxlogin.WxLoginClient
}

// NewService 构造 Service。
func NewService(
	addrRepo AddressRepo,
	regionRepo RegionRepo,
	rdb *redis.Client,
	wxMP wxlogin.WxLoginClient,
) *Service {
	return &Service{
		addrRepo:   addrRepo,
		regionRepo: regionRepo,
		rdb:        rdb,
		wxMP:       wxMP,
	}
}

// List 查询用户所有地址（默认地址排第一）。
func (s *Service) List(ctx context.Context, userID int64) ([]AddressResp, error) {
	list, err := s.addrRepo.List(ctx, userID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	resp := make([]AddressResp, len(list))
	for i, a := range list {
		resp[i] = toAddressResp(&a)
	}
	return resp, nil
}

// Get 查询单条地址（校验归属）。
func (s *Service) Get(ctx context.Context, id, userID int64) (*AddressResp, error) {
	addr, err := s.addrRepo.FindByID(ctx, id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	resp := toAddressResp(addr)
	return &resp, nil
}

// Create 新增地址（每用户限 20 条；is_default 时先清除旧默认）。
func (s *Service) Create(ctx context.Context, userID int64, req CreateAddressReq) (*AddressResp, error) {
	cnt, err := s.addrRepo.Count(ctx, userID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	if cnt >= maxAddressPerUser {
		return nil, errs.ErrConflict.WithMsg("地址数量已达上限（20 条）")
	}

	if req.IsDefault {
		_ = s.addrRepo.ClearDefault(ctx, userID)
	}

	addr := &Address{
		ID:           snowflake.NextID(),
		UserID:       userID,
		Name:         req.Name,
		Phone:        req.Phone,
		ProvinceCode: req.ProvinceCode,
		Province:     req.Province,
		CityCode:     req.CityCode,
		City:         req.City,
		DistrictCode: req.DistrictCode,
		District:     req.District,
		StreetCode:   req.StreetCode,
		Street:       req.Street,
		Detail:       req.Detail,
		IsDefault:    req.IsDefault,
	}

	if err := s.addrRepo.Create(ctx, addr); err != nil {
		return nil, errs.ErrInternal
	}
	resp := toAddressResp(addr)
	return &resp, nil
}

// Update 更新地址。
func (s *Service) Update(ctx context.Context, id, userID int64, req UpdateAddressReq) error {
	// 先校验归属
	if _, err := s.addrRepo.FindByID(ctx, id, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.ProvinceCode != nil {
		updates["province_code"] = *req.ProvinceCode
	}
	if req.Province != nil {
		updates["province"] = *req.Province
	}
	if req.CityCode != nil {
		updates["city_code"] = *req.CityCode
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.DistrictCode != nil {
		updates["district_code"] = *req.DistrictCode
	}
	if req.District != nil {
		updates["district"] = *req.District
	}
	if req.StreetCode != nil {
		updates["street_code"] = *req.StreetCode
	}
	if req.Street != nil {
		updates["street"] = *req.Street
	}
	if req.Detail != nil {
		updates["detail"] = *req.Detail
	}
	if req.IsDefault != nil && *req.IsDefault {
		_ = s.addrRepo.ClearDefault(ctx, userID)
		updates["is_default"] = true
	}

	if len(updates) == 0 {
		return nil
	}
	return s.addrRepo.Update(ctx, id, userID, updates)
}

// Delete 删除地址（校验归属）。
func (s *Service) Delete(ctx context.Context, id, userID int64) error {
	if _, err := s.addrRepo.FindByID(ctx, id, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	return s.addrRepo.Delete(ctx, id, userID)
}

// SetDefault 设置默认地址（校验归属）。
func (s *Service) SetDefault(ctx context.Context, id, userID int64) error {
	if _, err := s.addrRepo.FindByID(ctx, id, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	return s.addrRepo.SetDefault(ctx, id, userID)
}

// DecryptWxAddress 解密微信小程序地址数据，不落库，直接返回结构化地址。
func (s *Service) DecryptWxAddress(ctx context.Context, userID int64, encryptedData, iv string) (map[string]any, error) {
	skKey := fmt.Sprintf("wx:sk:%d", userID)
	sessionKey, err := s.rdb.Get(ctx, skKey).Result()
	if err != nil {
		return nil, errs.ErrSessionExpired
	}
	data, err := s.wxMP.DecryptUserData(sessionKey, encryptedData, iv)
	if err != nil {
		return nil, errs.ErrParam.WithMsg("解密失败")
	}
	return data, nil
}

// ListRegions 查询行政区划列表（按父级 code 查子级，为空则查省级）。
func (s *Service) ListRegions(ctx context.Context, parentCode string) ([]RegionResp, error) {
	list, err := s.regionRepo.ListByParent(ctx, parentCode)
	if err != nil {
		return nil, errs.ErrInternal
	}
	resp := make([]RegionResp, len(list))
	for i, r := range list {
		resp[i] = toRegionResp(&r)
	}
	return resp, nil
}
