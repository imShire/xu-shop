-- v1.8 分享溯源 + 分销（配套 docs/arch/18-distribution.md 与 91-db-schema.md v1.8）
-- 必须在 20260521000001_v1_2_drop_legacy_reserved.sql 之后运行（旧 distributor / commission_record 已被 DROP）。

-- +goose Up
-- +goose StatementBegin

-- =========================================================
-- order 表分享/分销追踪三件套
-- =========================================================
ALTER TABLE "order"
  ADD COLUMN IF NOT EXISTS share_trace_id        varchar(64),
  ADD COLUMN IF NOT EXISTS share_link_id         bigint,
  ADD COLUMN IF NOT EXISTS distributor_user_id   bigint;
COMMENT ON COLUMN "order".share_trace_id      IS '分享溯源 trace_id（uuidv7），随用户 storage 透传';
COMMENT ON COLUMN "order".share_link_id       IS '触发归因的分享链接 id';
COMMENT ON COLUMN "order".distributor_user_id IS '本单佣金归属分销员 user_id（首次绑定即终身在有效期内）';
CREATE INDEX IF NOT EXISTS idx_order_distributor ON "order"(distributor_user_id, created_at DESC) WHERE distributor_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_order_share_link  ON "order"(share_link_id, created_at DESC) WHERE share_link_id IS NOT NULL;

-- =========================================================
-- user 表分销员标识
-- =========================================================
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS is_distributor bool NOT NULL DEFAULT false;

-- =========================================================
-- 分享链接
-- =========================================================
CREATE TABLE share_link (
  id              bigint PRIMARY KEY,
  user_id         bigint NOT NULL,
  scene           varchar(16) NOT NULL,
  target_id       bigint,
  channel_code    varchar(32) NOT NULL DEFAULT 'other',
  short_token     varchar(16) NOT NULL UNIQUE,
  expire_at       timestamptz NOT NULL,
  click_count     bigint NOT NULL DEFAULT 0,
  register_count  bigint NOT NULL DEFAULT 0,
  order_count     bigint NOT NULL DEFAULT 0,
  gmv_cents       bigint NOT NULL DEFAULT 0,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_share_link_user ON share_link(user_id, created_at DESC);
CREATE INDEX idx_share_link_scene ON share_link(scene, target_id);

-- =========================================================
-- 分享点击
-- =========================================================
CREATE TABLE share_click (
  id                    bigint PRIMARY KEY,
  trace_id              varchar(64) NOT NULL,
  share_link_id         bigint NOT NULL,
  visitor_fingerprint   varchar(64),
  ts                    timestamptz NOT NULL DEFAULT now(),
  ua                    varchar(500),
  ip                    inet,
  device                varchar(8),
  referer               varchar(500)
);
CREATE INDEX idx_share_click_link_time ON share_click(share_link_id, ts DESC);
CREATE INDEX idx_share_click_trace ON share_click(trace_id);
CREATE INDEX idx_share_click_fp ON share_click(visitor_fingerprint, ts DESC);

-- =========================================================
-- 分享归因
-- =========================================================
CREATE TABLE share_attribution (
  id                       bigint PRIMARY KEY,
  user_id                  bigint,
  share_link_id            bigint NOT NULL,
  trace_id                 varchar(64) NOT NULL,
  first_touch_ts           timestamptz NOT NULL DEFAULT now(),
  last_touch_ts            timestamptz NOT NULL DEFAULT now(),
  attribution_window_days  int NOT NULL DEFAULT 7
);
CREATE UNIQUE INDEX uq_share_attribution_trace ON share_attribution(trace_id);
CREATE INDEX idx_share_attribution_user ON share_attribution(user_id) WHERE user_id IS NOT NULL;

-- =========================================================
-- 分销员
-- =========================================================
CREATE TABLE distributor (
  id                  bigint PRIMARY KEY,
  user_id             bigint NOT NULL UNIQUE,
  level               varchar(16) NOT NULL DEFAULT 'normal',
  rate                decimal(5,4) NOT NULL DEFAULT 0.0500,
  rate_override       decimal(5,4),
  status              varchar(16) NOT NULL DEFAULT 'pending',
  apply_at            timestamptz NOT NULL DEFAULT now(),
  approved_at         timestamptz,
  approver_admin_id   bigint,
  suspended_at        timestamptz,
  suspended_reason    varchar(500),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_distributor_status CHECK (status IN ('pending','active','disabled'))
);
CREATE INDEX idx_distributor_status ON distributor(status, level);

-- =========================================================
-- 分销邀请关系
-- =========================================================
CREATE TABLE distributor_relation (
  id                 bigint PRIMARY KEY,
  invitee_user_id    bigint NOT NULL UNIQUE,
  inviter_user_id    bigint NOT NULL,
  share_link_id      bigint NOT NULL,
  bound_at           timestamptz NOT NULL DEFAULT now(),
  expire_at          timestamptz NOT NULL,
  last_renewed_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_dist_rel_inviter ON distributor_relation(inviter_user_id, bound_at DESC);

-- =========================================================
-- 佣金记录
-- =========================================================
CREATE TABLE commission_record (
  id                  bigint PRIMARY KEY,
  order_id            bigint NOT NULL,
  distributor_user_id bigint NOT NULL,
  level               varchar(16) NOT NULL,
  rate                decimal(5,4) NOT NULL,
  base_amount_cents   bigint NOT NULL,
  amount_cents        bigint NOT NULL,
  status              varchar(16) NOT NULL DEFAULT 'pending',
  suspect_reason      varchar(500),
  freeze_until        timestamptz NOT NULL,
  settlement_id       bigint,
  settled_at          timestamptz,
  canceled_reason     varchar(500),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_commission_status CHECK (status IN ('pending','locked','settled','canceled','suspect'))
);
CREATE UNIQUE INDEX uq_commission_order_dist ON commission_record(order_id, distributor_user_id);
CREATE INDEX idx_commission_dist_status ON commission_record(distributor_user_id, status, freeze_until);
CREATE INDEX idx_commission_status_freeze ON commission_record(status, freeze_until) WHERE status='pending';

-- =========================================================
-- 结算单
-- =========================================================
CREATE TABLE commission_settlement (
  id                       bigint PRIMARY KEY,
  distributor_user_id      bigint NOT NULL,
  period_yyyymm            varchar(8),
  request_amount_cents     bigint NOT NULL,
  records                  jsonb NOT NULL DEFAULT '[]'::jsonb,
  withdraw_order_id        bigint,
  status                   varchar(16) NOT NULL DEFAULT 'pending',
  channel                  varchar(16) NOT NULL DEFAULT 'wx_transfer',
  fail_reason              varchar(500),
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_settlement_status CHECK (status IN ('pending','processing','success','failed'))
);
CREATE INDEX idx_settlement_dist ON commission_settlement(distributor_user_id, created_at DESC);

-- =========================================================
-- 提现工单
-- =========================================================
CREATE TABLE withdraw_order (
  id                   bigint PRIMARY KEY,
  distributor_user_id  bigint NOT NULL,
  withdraw_no          varchar(64) NOT NULL UNIQUE,
  amount_cents         bigint NOT NULL,
  channel              varchar(16) NOT NULL DEFAULT 'wx_transfer',
  status               varchar(16) NOT NULL DEFAULT 'pending',
  wx_transfer_no       varchar(64),
  wx_transfer_state    varchar(32),
  fail_reason          varchar(500),
  applied_at           timestamptz NOT NULL DEFAULT now(),
  processed_at         timestamptz,
  finished_at          timestamptz,
  idem_key             varchar(128) NOT NULL,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_withdraw_status CHECK (status IN ('pending','processing','success','failed','canceled'))
);
CREATE UNIQUE INDEX uq_withdraw_idem ON withdraw_order(idem_key);
CREATE INDEX idx_withdraw_dist ON withdraw_order(distributor_user_id, applied_at DESC);
CREATE INDEX idx_withdraw_status ON withdraw_order(status, applied_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_withdraw_status;
DROP INDEX IF EXISTS idx_withdraw_dist;
DROP INDEX IF EXISTS uq_withdraw_idem;
DROP TABLE IF EXISTS withdraw_order;
DROP INDEX IF EXISTS idx_settlement_dist;
DROP TABLE IF EXISTS commission_settlement;
DROP INDEX IF EXISTS idx_commission_status_freeze;
DROP INDEX IF EXISTS idx_commission_dist_status;
DROP INDEX IF EXISTS uq_commission_order_dist;
DROP TABLE IF EXISTS commission_record;
DROP INDEX IF EXISTS idx_dist_rel_inviter;
DROP TABLE IF EXISTS distributor_relation;
DROP INDEX IF EXISTS idx_distributor_status;
DROP TABLE IF EXISTS distributor;
DROP INDEX IF EXISTS idx_share_attribution_user;
DROP INDEX IF EXISTS uq_share_attribution_trace;
DROP TABLE IF EXISTS share_attribution;
DROP INDEX IF EXISTS idx_share_click_fp;
DROP INDEX IF EXISTS idx_share_click_trace;
DROP INDEX IF EXISTS idx_share_click_link_time;
DROP TABLE IF EXISTS share_click;
DROP INDEX IF EXISTS idx_share_link_scene;
DROP INDEX IF EXISTS idx_share_link_user;
DROP TABLE IF EXISTS share_link;

ALTER TABLE "user" DROP COLUMN IF EXISTS is_distributor;

DROP INDEX IF EXISTS idx_order_share_link;
DROP INDEX IF EXISTS idx_order_distributor;
ALTER TABLE "order"
  DROP COLUMN IF EXISTS distributor_user_id,
  DROP COLUMN IF EXISTS share_link_id,
  DROP COLUMN IF EXISTS share_trace_id;

-- +goose StatementEnd
