-- A8 通用日终对账差异表。
-- 旧 `reconciliation_diff` 仅服务于支付场景，本次升级为支付/库存/佣金通用差异表。
-- 旧表重命名为 `legacy_reconciliation_diff`，保留历史数据与既有 /admin/reconciliation 路由读取；
-- 新 `reconciliation_diff` 由 modules/reconciliation 服务，被三种 cron 作业写入。
-- 配套 docs/arch/91-db-schema.md v1.12 段、docs/arch/99-revisions.md v1.4 - A8。

-- +goose Up
-- +goose StatementBegin

ALTER TABLE reconciliation_diff RENAME TO legacy_reconciliation_diff;

CREATE TABLE reconciliation_diff (
  id              bigint PRIMARY KEY,
  job             varchar(32) NOT NULL,         -- 'payment' | 'inventory' | 'commission'
  biz_date        date NOT NULL,
  ref_type        varchar(32) NOT NULL,         -- 'order' | 'sku' | 'commission_record' | 'account'
  ref_id          varchar(64) NOT NULL,         -- 资源 id（沿用字符串红线）
  field           varchar(64) NOT NULL,
  expected_value  text,
  actual_value    text,
  diff_cents      bigint,
  severity        varchar(8) NOT NULL DEFAULT 'warn',  -- 'info' | 'warn' | 'critical'
  status          varchar(16) NOT NULL DEFAULT 'open', -- 'open' | 'acknowledged' | 'resolved'
  note            text,
  acked_by        bigint,
  acked_at        timestamptz,
  resolved_by     bigint,
  resolved_at     timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_recdiff_job_date_ref_field
  ON reconciliation_diff(job, biz_date, ref_type, ref_id, field);

CREATE INDEX idx_recdiff_job_date
  ON reconciliation_diff(job, biz_date DESC);

CREATE INDEX idx_recdiff_status_created
  ON reconciliation_diff(status, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_recdiff_status_created;
DROP INDEX IF EXISTS idx_recdiff_job_date;
DROP INDEX IF EXISTS uq_recdiff_job_date_ref_field;
DROP TABLE IF EXISTS reconciliation_diff;
ALTER TABLE legacy_reconciliation_diff RENAME TO reconciliation_diff;

-- +goose StatementEnd
