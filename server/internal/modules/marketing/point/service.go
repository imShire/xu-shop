package point

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/marketing/shared"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// Service 积分服务。
type Service struct {
	repo Repo
	db   *gorm.DB
}

// NewService 构造服务。
func NewService(repo Repo, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// EarnReq 入账请求。
type EarnReq struct {
	UserID    int64
	Change    int64
	RefType   string
	RefID     *int64
	Reason    string
	IdemKey   string  // 必填，幂等键
	ExpireAt  *time.Time
	CreatedBy *int64
}

// SpendReq 消费请求。
type SpendReq struct {
	UserID  int64
	Change  int64 // 正数，函数内会取负
	RefType string
	RefID   *int64
	Reason  string
	IdemKey string
}

// Earn 积分入账（含订单返、签到、邀请、生日、注册等）。
//
// 幂等：通过 idem_key 唯一索引保障；重复请求返回首次结果。
func (s *Service) Earn(ctx context.Context, req EarnReq) (*Transaction, error) {
	if req.UserID <= 0 || req.Change <= 0 || req.IdemKey == "" {
		return nil, errs.ErrParam
	}
	// 先查幂等
	if existing, err := s.repo.FindTransactionByIdem(ctx, req.IdemKey); err == nil && existing != nil {
		return existing, nil
	}

	var created *Transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := s.repo.GetOrCreateAccount(ctx, tx, req.UserID)
		if err != nil {
			return err
		}
		idemPtr := req.IdemKey
		t := &Transaction{
			ID:           snowflake.NextID(),
			UserID:       req.UserID,
			Change:       req.Change,
			Type:         TxnTypeEarn,
			RefType:      strPtr(req.RefType),
			RefID:        req.RefID,
			BalanceAfter: acc.Balance + req.Change,
			ExpireAt:     req.ExpireAt,
			Reason:       req.Reason,
			CreatedBy:    req.CreatedBy,
			IdemKey:      &idemPtr,
		}
		if err := s.repo.InsertTransaction(ctx, tx, t); err != nil {
			return fmt.Errorf("insert txn: %w", err)
		}
		if err := s.repo.UpdateAccount(ctx, tx, req.UserID, req.Change, 0, req.Change, 0); err != nil {
			return err
		}
		created = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Spend 积分消费（FIFO，从最早可用 earn 流水扣减）。
func (s *Service) Spend(ctx context.Context, req SpendReq) (*Transaction, error) {
	if req.UserID <= 0 || req.Change <= 0 || req.IdemKey == "" {
		return nil, errs.ErrParam
	}
	if existing, err := s.repo.FindTransactionByIdem(ctx, req.IdemKey); err == nil && existing != nil {
		return existing, nil
	}

	var created *Transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := s.repo.GetOrCreateAccount(ctx, tx, req.UserID)
		if err != nil {
			return err
		}
		if acc.Balance < req.Change {
			return shared.ErrPointInsufficient
		}
		// FIFO 扫描可用 earn
		earns, err := s.repo.ListAvailableEarn(ctx, tx, req.UserID, 200)
		if err != nil {
			return err
		}
		// 简化策略：仅标记最早一条 consumed（如果其单条余额≥change）；
		// MVP 阶段一笔扣减通常由整笔 earn 抵消。后续可拆分流水。
		// 此处采取"累计 change 跨多条 earn 都标记 consumed"的折中实现：
		var (
			cum     int64
			markIDs []int64
		)
		for _, e := range earns {
			cum += e.Change
			markIDs = append(markIDs, e.ID)
			if cum >= req.Change {
				break
			}
		}
		if cum < req.Change {
			return shared.ErrPointInsufficient.WithMsg("可用积分不足以扣减")
		}
		if err := s.repo.MarkTransactionConsumed(ctx, tx, markIDs); err != nil {
			return err
		}
		idem := req.IdemKey
		t := &Transaction{
			ID:           snowflake.NextID(),
			UserID:       req.UserID,
			Change:       -req.Change,
			Type:         TxnTypeSpend,
			RefType:      strPtr(req.RefType),
			RefID:        req.RefID,
			BalanceAfter: acc.Balance - req.Change,
			Reason:       req.Reason,
			IdemKey:      &idem,
		}
		if err := s.repo.InsertTransaction(ctx, tx, t); err != nil {
			return err
		}
		if err := s.repo.UpdateAccount(ctx, tx, req.UserID, -req.Change, 0, 0, req.Change); err != nil {
			return err
		}
		created = t
		return nil
	})
	return created, err
}

// Refund 退款时把已抵扣积分返还（不返还已过期部分）。
func (s *Service) Refund(ctx context.Context, userID int64, change int64, refID int64, idemKey string) error {
	if change <= 0 {
		return nil
	}
	if existing, _ := s.repo.FindTransactionByIdem(ctx, idemKey); existing != nil {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := s.repo.GetOrCreateAccount(ctx, tx, userID)
		if err != nil {
			return err
		}
		idem := idemKey
		ref := refID
		refType := "order_refund"
		t := &Transaction{
			ID:           snowflake.NextID(),
			UserID:       userID,
			Change:       change,
			Type:         TxnTypeRefund,
			RefType:      &refType,
			RefID:        &ref,
			BalanceAfter: acc.Balance + change,
			Reason:       "订单退款返还积分",
			IdemKey:      &idem,
		}
		if err := s.repo.InsertTransaction(ctx, tx, t); err != nil {
			return err
		}
		return s.repo.UpdateAccount(ctx, tx, userID, change, 0, 0, -change)
	})
}

// AdjustTicketCreate 后台申请调整工单。
func (s *Service) AdjustTicketCreate(ctx context.Context, userID, change int64, reason string, applicantAdminID int64) (*AdjustTicket, error) {
	if userID <= 0 || change == 0 || reason == "" {
		return nil, errs.ErrParam
	}
	hasPending, err := s.repo.HasPendingTicket(ctx, userID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, shared.ErrPointAdjustPending
	}
	t := &AdjustTicket{
		ID:               snowflake.NextID(),
		UserID:           userID,
		Change:           change,
		Reason:           reason,
		Status:           TicketStatusPending,
		ApplicantAdminID: applicantAdminID,
	}
	if err := s.repo.CreateTicket(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// AdjustTicketApprove 审批通过 → 实际执行 Earn/Spend（admin_adjust 类型）。
func (s *Service) AdjustTicketApprove(ctx context.Context, ticketID int64, approverAdminID int64) error {
	t, err := s.repo.FindTicket(ctx, ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if t.Status != TicketStatusPending {
		return shared.ErrInvalidStateTransition.WithMsg(fmt.Sprintf("工单当前状态:%s", t.Status))
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := s.repo.UpdateTicketStatus(ctx, tx, ticketID, TicketStatusPending, TicketStatusApproved, approverAdminID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return shared.ErrInvalidStateTransition
		}
		acc, err := s.repo.GetOrCreateAccount(ctx, tx, t.UserID)
		if err != nil {
			return err
		}
		idem := fmt.Sprintf("adjust:%d", ticketID)
		ref := ticketID
		refType := "adjust_ticket"
		txn := &Transaction{
			ID:           snowflake.NextID(),
			UserID:       t.UserID,
			Change:       t.Change,
			Type:         TxnTypeAdminAdjust,
			RefType:      &refType,
			RefID:        &ref,
			BalanceAfter: acc.Balance + t.Change,
			Reason:       t.Reason,
			IdemKey:      &idem,
			CreatedBy:    &approverAdminID,
		}
		if err := s.repo.InsertTransaction(ctx, tx, txn); err != nil {
			return err
		}
		if t.Change > 0 {
			return s.repo.UpdateAccount(ctx, tx, t.UserID, t.Change, 0, t.Change, 0)
		}
		return s.repo.UpdateAccount(ctx, tx, t.UserID, t.Change, 0, 0, -t.Change)
	})
}

// AdjustTicketReject 审批拒绝。
func (s *Service) AdjustTicketReject(ctx context.Context, ticketID int64, approverAdminID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := s.repo.UpdateTicketStatus(ctx, tx, ticketID, TicketStatusPending, TicketStatusRejected, approverAdminID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return shared.ErrInvalidStateTransition
		}
		return nil
	})
}

// ExpireScan 扫描过期 earn 流水，写一笔 expire 流水抵消。
func (s *Service) ExpireScan(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	now := time.Now()
	list, err := s.repo.ScanExpired(ctx, now, batchSize)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, e := range list {
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// CAS：如果已 consumed 则跳过
			res := tx.Model(&Transaction{}).Where("id = ? AND consumed = false", e.ID).
				Update("consumed", true)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			acc, err := s.repo.GetOrCreateAccount(ctx, tx, e.UserID)
			if err != nil {
				return err
			}
			idem := fmt.Sprintf("expire:%d", e.ID)
			ref := e.ID
			refType := "earn_expire"
			t := &Transaction{
				ID:           snowflake.NextID(),
				UserID:       e.UserID,
				Change:       -e.Change,
				Type:         TxnTypeExpire,
				RefType:      &refType,
				RefID:        &ref,
				BalanceAfter: acc.Balance - e.Change,
				Reason:       "积分过期",
				IdemKey:      &idem,
			}
			if err := s.repo.InsertTransaction(ctx, tx, t); err != nil {
				return err
			}
			if err := s.repo.UpdateAccount(ctx, tx, e.UserID, -e.Change, 0, 0, 0); err != nil {
				return err
			}
			processed++
			return nil
		})
		if err != nil {
			return processed, err
		}
	}
	return processed, nil
}

// Balance 当前余额。
func (s *Service) Balance(ctx context.Context, userID int64) (int64, error) {
	var acc Account
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return acc.Balance, nil
}

// History 流水历史。
func (s *Service) History(ctx context.Context, userID int64, page, size int) ([]Transaction, int64, error) {
	return s.repo.ListTransactions(ctx, userID, page, size)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
