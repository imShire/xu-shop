package private_domain

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelCodeRepo 渠道码仓库接口。
type ChannelCodeRepo interface {
	Create(ctx context.Context, c *ChannelCode) error
	FindByID(ctx context.Context, id int64) (*ChannelCode, error)
	Update(ctx context.Context, c *ChannelCode) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, size int) ([]ChannelCode, int64, error)
}

// TagRepo 标签仓库接口。
type TagRepo interface {
	Create(ctx context.Context, t *CustomerTag) error
	FindByID(ctx context.Context, id int64) (*CustomerTag, error)
	Update(ctx context.Context, t *CustomerTag) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]CustomerTag, error)
}

// UserTagRepo 用户标签关联仓库接口。
type UserTagRepo interface {
	Add(ctx context.Context, ut *UserTag) error
	Remove(ctx context.Context, userID, tagID int64) error
	ListByUser(ctx context.Context, userID int64) ([]CustomerTag, error)
}

// ShareRepo 分享仓库接口。
type ShareRepo interface {
	UpsertShortCode(ctx context.Context, sc *ShareShortCode) (*ShareShortCode, error)
	FindShortCode(ctx context.Context, id int64) (*ShareShortCode, error)
	CreateAttribution(ctx context.Context, a *ShareAttribution) error
	ExistsAttribution(ctx context.Context, shareUserID, viewerUserID, productID int64, since time.Time) (bool, error)
}

// ---- 实现 ----

type channelCodeRepoImpl struct{ db *gorm.DB }

func NewChannelCodeRepo(db *gorm.DB) ChannelCodeRepo { return &channelCodeRepoImpl{db: db} }

func (r *channelCodeRepoImpl) Create(ctx context.Context, c *ChannelCode) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *channelCodeRepoImpl) FindByID(ctx context.Context, id int64) (*ChannelCode, error) {
	var c ChannelCode
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *channelCodeRepoImpl) Update(ctx context.Context, c *ChannelCode) error {
	c.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *channelCodeRepoImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&ChannelCode{}, id).Error
}

func (r *channelCodeRepoImpl) List(ctx context.Context, page, size int) ([]ChannelCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var total int64
	var list []ChannelCode
	q := r.db.WithContext(ctx).Model(&ChannelCode{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

type tagRepoImpl struct{ db *gorm.DB }

func NewTagRepo(db *gorm.DB) TagRepo { return &tagRepoImpl{db: db} }

func (r *tagRepoImpl) Create(ctx context.Context, t *CustomerTag) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *tagRepoImpl) FindByID(ctx context.Context, id int64) (*CustomerTag, error) {
	var t CustomerTag
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tagRepoImpl) Update(ctx context.Context, t *CustomerTag) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *tagRepoImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&CustomerTag{}, id).Error
}

func (r *tagRepoImpl) List(ctx context.Context) ([]CustomerTag, error) {
	var list []CustomerTag
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

type userTagRepoImpl struct{ db *gorm.DB }

func NewUserTagRepo(db *gorm.DB) UserTagRepo { return &userTagRepoImpl{db: db} }

func (r *userTagRepoImpl) Add(ctx context.Context, ut *UserTag) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(ut).Error
}

func (r *userTagRepoImpl) Remove(ctx context.Context, userID, tagID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND tag_id = ?", userID, tagID).
		Delete(&UserTag{}).Error
}

func (r *userTagRepoImpl) ListByUser(ctx context.Context, userID int64) ([]CustomerTag, error) {
	var tags []CustomerTag
	err := r.db.WithContext(ctx).
		Joins("JOIN user_tag ON user_tag.tag_id = customer_tag.id").
		Where("user_tag.user_id = ?", userID).
		Find(&tags).Error
	return tags, err
}

type shareRepoImpl struct{ db *gorm.DB }

func NewShareRepo(db *gorm.DB) ShareRepo { return &shareRepoImpl{db: db} }

func (r *shareRepoImpl) UpsertShortCode(ctx context.Context, sc *ShareShortCode) (*ShareShortCode, error) {
	result := r.db.WithContext(ctx).
		Where(clause.OnConflict{
			Columns:   []clause.Column{{Name: "share_user_id"}, {Name: "product_id"}, {Name: "channel_code_id"}},
			DoNothing: true,
		}).
		Create(sc)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		var existing ShareShortCode
		q := r.db.WithContext(ctx).Where("share_user_id = ?", sc.ShareUserID)
		if sc.ProductID != nil {
			q = q.Where("product_id = ?", *sc.ProductID)
		} else {
			q = q.Where("product_id IS NULL")
		}
		if sc.ChannelCodeID != nil {
			q = q.Where("channel_code_id = ?", *sc.ChannelCodeID)
		} else {
			q = q.Where("channel_code_id IS NULL")
		}
		if err := q.First(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}
	return sc, nil
}

func (r *shareRepoImpl) FindShortCode(ctx context.Context, id int64) (*ShareShortCode, error) {
	var sc ShareShortCode
	if err := r.db.WithContext(ctx).First(&sc, id).Error; err != nil {
		return nil, err
	}
	return &sc, nil
}

func (r *shareRepoImpl) CreateAttribution(ctx context.Context, a *ShareAttribution) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *shareRepoImpl) ExistsAttribution(ctx context.Context, shareUserID, viewerUserID, productID int64, since time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ShareAttribution{}).
		Where("share_user_id = ? AND viewer_user_id = ? AND product_id = ? AND created_at >= ?",
			shareUserID, viewerUserID, productID, since).
		Count(&count).Error
	return count > 0, err
}
