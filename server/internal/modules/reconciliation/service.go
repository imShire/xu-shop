package reconciliation

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// bizLoc 业务自然日时区（Asia/Shanghai）。与 jobs/reconciliation 保持一致。
var bizLoc = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}()

// Service 对账差异业务服务。
type Service struct {
	repo Repository
}

// NewService 构造服务。
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// RecordDiff 由对账作业调用，写入或更新一条差异。
// 自动填充 ID / 业务日去时（biz_date 仅取日期部分）/ 默认 severity / status=open。
func (s *Service) RecordDiff(ctx context.Context, d *Diff) error {
	if d.Job == "" || d.RefType == "" || d.RefID == "" || d.Field == "" {
		return errs.ErrParam.WithMsg("reconciliation: job/ref_type/ref_id/field 必填")
	}
	if d.ID == 0 {
		d.ID = snowflake.NextID()
	}
	if d.BizDate.IsZero() {
		d.BizDate = time.Now().In(bizLoc).AddDate(0, 0, -1)
	}
	bd := d.BizDate.In(bizLoc)
	d.BizDate = time.Date(bd.Year(), bd.Month(), bd.Day(),
		0, 0, 0, 0, bizLoc)
	if d.Severity == "" {
		d.Severity = SeverityWarn
	}
	if d.Status == "" {
		d.Status = StatusOpen
	}
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	return s.repo.Upsert(ctx, d)
}

// List 后台查询差异列表。
func (s *Service) List(ctx context.Context, f Filter) ([]Diff, int64, error) {
	return s.repo.List(ctx, f)
}

// Acknowledge 标记差异为已确认（人工已知晓）。
func (s *Service) Acknowledge(ctx context.Context, id, operatorID int64) error {
	d, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound.WithMsg("差异记录不存在")
		}
		return err
	}
	if d.Status == StatusResolved {
		return errs.ErrParam.WithMsg("差异已解决，无需再次确认")
	}
	return s.repo.UpdateStatus(ctx, id, StatusAcknowledged, operatorID, nil)
}

// Resolve 标记差异为已解决，可附处理备注。
func (s *Service) Resolve(ctx context.Context, id, operatorID int64, note *string) error {
	d, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound.WithMsg("差异记录不存在")
		}
		return err
	}
	if d.Status == StatusResolved {
		return errs.ErrParam.WithMsg("差异已解决")
	}
	return s.repo.UpdateStatus(ctx, id, StatusResolved, operatorID, note)
}
