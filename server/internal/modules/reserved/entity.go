package reserved

import "time"

// Distributor 分销商预留。
type Distributor struct {
	ID                  int64      `gorm:"primaryKey"`
	UserID              int64      `gorm:"not null;uniqueIndex"`
	Level               int16      `gorm:"not null;default:1"`
	ParentDistributorID *int64
	Code                *string    `gorm:"size:32;uniqueIndex"`
	Status              string     `gorm:"size:16;not null;default:'pending'"`
	ApplyAt             *time.Time
	ApprovedAt          *time.Time
	CreatedAt           time.Time
}

// CommissionRecord 分销佣金流水预留。
type CommissionRecord struct {
	ID            int64      `gorm:"primaryKey"`
	OrderID       int64      `gorm:"not null"`
	DistributorID int64      `gorm:"not null"`
	Level         int16      `gorm:"not null"`
	AmountCents   int64      `gorm:"not null"`
	Status        string     `gorm:"size:16;not null;default:'locked'"`
	SettleAt      *time.Time
	CreatedAt     time.Time
}

// GroupBuyActivity 拼团活动预留。
type GroupBuyActivity struct {
	ID              int64     `gorm:"primaryKey"`
	ProductID       int64     `gorm:"not null"`
	SkuID           int64     `gorm:"not null"`
	GroupSize       int       `gorm:"not null"`
	GroupPriceCents int64     `gorm:"not null"`
	StartAt         time.Time `gorm:"not null"`
	EndAt           time.Time `gorm:"not null"`
	Status          string    `gorm:"size:16;not null;default:'enabled'"`
	CreatedAt       time.Time
}

// GroupBuyOrder 拼团订单预留。
type GroupBuyOrder struct {
	ID           int64     `gorm:"primaryKey"`
	ActivityID   int64     `gorm:"not null"`
	LeaderUserID int64     `gorm:"not null"`
	Status       string    `gorm:"size:16;not null;default:'pending'"`
	ExpireAt     time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

// Coupon 优惠券预留。
type Coupon struct {
	ID             int64      `gorm:"primaryKey"`
	Name           string     `gorm:"size:64;not null"`
	Type           string     `gorm:"size:16;not null"`
	Value          int        `gorm:"not null"`
	MinAmountCents int64      `gorm:"not null;default:0"`
	ValidFrom      *time.Time
	ValidTo        *time.Time
	Total          int        `gorm:"not null;default:0"`
	Claimed        int        `gorm:"not null;default:0"`
	Status         string     `gorm:"size:16;not null;default:'enabled'"`
	CreatedAt      time.Time
}

// UserCoupon 用户优惠券预留。
type UserCoupon struct {
	ID          int64      `gorm:"primaryKey"`
	UserID      int64      `gorm:"not null"`
	CouponID    int64      `gorm:"not null"`
	Status      string     `gorm:"size:16;not null;default:'unused'"`
	UsedOrderID *int64
	ClaimedAt   time.Time  `gorm:"not null;autoCreateTime"`
	UsedAt      *time.Time
	ExpireAt    *time.Time
}

// PointsLog 积分流水预留。
type PointsLog struct {
	ID           int64     `gorm:"primaryKey"`
	UserID       int64     `gorm:"not null"`
	Change       int       `gorm:"not null"`
	Type         string    `gorm:"size:16;not null"`
	RefType      *string   `gorm:"size:16"`
	RefID        *int64
	BalanceAfter int       `gorm:"not null"`
	Reason       *string   `gorm:"size:200"`
	CreatedAt    time.Time
}
