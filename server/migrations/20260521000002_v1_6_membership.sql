-- v1.6 会员体系：优惠券 + 积分 + 等级
-- 配套 docs/arch/16-membership.md 与 91-db-schema.md v1.6
-- 必须先运行 20260521000001_v1_2_drop_legacy_reserved.sql

-- +goose Up
-- +goose StatementBegin

-- =========================================================
-- order 表会员相关字段
-- =========================================================
ALTER TABLE "order"
  ADD COLUMN IF NOT EXISTS coupon_amount_cents    bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS point_used             bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS point_deduct_cents     bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS point_earned           bigint NOT NULL DEFAULT 0;
-- coupon_id 在 v1.0 已有；类型不变。

COMMENT ON COLUMN "order".coupon_id            IS '使用的用户券 id (user_coupon.id)，可空';
COMMENT ON COLUMN "order".coupon_amount_cents  IS '券抵扣金额（分）';
COMMENT ON COLUMN "order".point_used           IS '本单使用积分数';
COMMENT ON COLUMN "order".point_deduct_cents   IS '积分抵扣金额（分）';
COMMENT ON COLUMN "order".point_earned         IS '实际入账积分数（T+N 后回填）';
CREATE INDEX IF NOT EXISTS idx_order_coupon ON "order"(coupon_id) WHERE coupon_id IS NOT NULL;

-- =========================================================
-- user 表会员相关字段
-- =========================================================
ALTER TABLE "user"
  ADD COLUMN IF NOT EXISTS member_level_code        varchar(16) NOT NULL DEFAULT 'normal',
  ADD COLUMN IF NOT EXISTS member_level_expire_at   timestamptz,
  ADD COLUMN IF NOT EXISTS recent_365d_gmv_cents    bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS register_channel_code    varchar(64),
  ADD COLUMN IF NOT EXISTS register_share_link_id   bigint;

COMMENT ON COLUMN "user".member_level_code      IS '当前会员等级 code，关联 member_level.code';
COMMENT ON COLUMN "user".member_level_expire_at IS '保级到期时间，到期后才允许下调';
COMMENT ON COLUMN "user".recent_365d_gmv_cents  IS '近 365 天累计实付 GMV（分），每日重算缓存';
COMMENT ON COLUMN "user".register_channel_code  IS '注册来源渠道码';
COMMENT ON COLUMN "user".register_share_link_id IS '注册时归因的 share_link.id（可空）';
CREATE INDEX IF NOT EXISTS idx_user_member_level ON "user"(member_level_code);

-- =========================================================
-- 优惠券模板
-- =========================================================
CREATE TABLE coupon_template (
  id                          bigint PRIMARY KEY,
  name                        varchar(128) NOT NULL,
  description                 varchar(500) NOT NULL DEFAULT '',
  type                        varchar(16) NOT NULL,
  value_cents                 bigint NOT NULL DEFAULT 0,
  discount_rate               decimal(5,4),
  max_discount_cents          bigint NOT NULL DEFAULT 0,
  min_amount_cents            bigint NOT NULL DEFAULT 0,
  scope_type                  varchar(16) NOT NULL DEFAULT 'all',
  scope_targets               jsonb NOT NULL DEFAULT '[]'::jsonb,
  exclude_promotion_items     bool NOT NULL DEFAULT false,
  include_freight             bool NOT NULL DEFAULT false,
  validity_mode               varchar(16) NOT NULL DEFAULT 'absolute',
  valid_from                  timestamptz,
  valid_to                    timestamptz,
  valid_days                  int,
  total_quota                 bigint NOT NULL DEFAULT 0,
  claimed_count               bigint NOT NULL DEFAULT 0,
  used_count                  bigint NOT NULL DEFAULT 0,
  per_user_limit              int NOT NULL DEFAULT 1,
  per_order_limit             int NOT NULL DEFAULT 1,
  stack_with_points           bool NOT NULL DEFAULT true,
  claim_start_at              timestamptz,
  claim_end_at                timestamptz,
  status                      varchar(16) NOT NULL DEFAULT 'draft',
  created_by                  bigint NOT NULL,
  created_at                  timestamptz NOT NULL DEFAULT now(),
  updated_at                  timestamptz NOT NULL DEFAULT now(),
  deleted_at                  timestamptz,
  CONSTRAINT chk_coupon_tpl_type CHECK (type IN ('amount','discount','no_threshold','exchange')),
  CONSTRAINT chk_coupon_tpl_status CHECK (status IN ('draft','online','offline'))
);
CREATE INDEX idx_coupon_tpl_status ON coupon_template(status, claim_end_at) WHERE deleted_at IS NULL;
COMMENT ON TABLE coupon_template IS '优惠券模板';

-- =========================================================
-- 用户券实例
-- =========================================================
CREATE TABLE user_coupon (
  id                  bigint PRIMARY KEY,
  user_id             bigint NOT NULL,
  coupon_template_id  bigint NOT NULL,
  source              varchar(32) NOT NULL,
  source_ref          jsonb NOT NULL DEFAULT '{}'::jsonb,
  status              varchar(16) NOT NULL DEFAULT 'unused',
  order_id            bigint,
  claimed_at          timestamptz NOT NULL DEFAULT now(),
  locked_at           timestamptz,
  used_at             timestamptz,
  expire_at           timestamptz NOT NULL,
  snapshot            jsonb NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_user_coupon_status CHECK (status IN ('unused','locked','used','expired'))
);
CREATE INDEX idx_user_coupon_user_status ON user_coupon(user_id, status, expire_at);
CREATE INDEX idx_user_coupon_order ON user_coupon(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_user_coupon_template ON user_coupon(coupon_template_id);
CREATE UNIQUE INDEX uq_user_coupon_lock ON user_coupon(order_id, coupon_template_id) WHERE status IN ('locked','used');
COMMENT ON TABLE user_coupon IS '用户券实例';

-- =========================================================
-- 兑换码
-- =========================================================
CREATE TABLE coupon_redeem_code (
  id                  bigint PRIMARY KEY,
  template_id         bigint NOT NULL,
  batch_id            bigint NOT NULL,
  code                varchar(32) NOT NULL UNIQUE,
  status              varchar(16) NOT NULL DEFAULT 'unused',
  used_by_user_id     bigint,
  used_at             timestamptz,
  expire_at           timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_redeem_code_status CHECK (status IN ('unused','used','voided'))
);
CREATE INDEX idx_redeem_code_batch ON coupon_redeem_code(batch_id, status);
COMMENT ON TABLE coupon_redeem_code IS '优惠券兑换码';

-- =========================================================
-- 发放任务
-- =========================================================
CREATE TABLE coupon_grant_task (
  id                  bigint PRIMARY KEY,
  template_id         bigint NOT NULL,
  filter              jsonb NOT NULL DEFAULT '{}'::jsonb,
  estimate_count      bigint NOT NULL DEFAULT 0,
  granted_count       bigint NOT NULL DEFAULT 0,
  failed_count        bigint NOT NULL DEFAULT 0,
  failed_detail_oss   text,
  status              varchar(16) NOT NULL DEFAULT 'pending',
  created_by          bigint NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  started_at          timestamptz,
  finished_at         timestamptz,
  CONSTRAINT chk_grant_task_status CHECK (status IN ('pending','running','done','failed'))
);
CREATE INDEX idx_grant_task_status ON coupon_grant_task(status, created_at DESC);

-- =========================================================
-- 积分账户
-- =========================================================
CREATE TABLE point_account (
  user_id           bigint PRIMARY KEY,
  balance           bigint NOT NULL DEFAULT 0,
  locked            bigint NOT NULL DEFAULT 0,
  total_earned      bigint NOT NULL DEFAULT 0,
  total_spent       bigint NOT NULL DEFAULT 0,
  updated_at        timestamptz NOT NULL DEFAULT now()
);
COMMENT ON COLUMN point_account.balance IS '可用余额；退款冲销可使其为负，UI 展示为 0';

-- =========================================================
-- 积分流水
-- =========================================================
CREATE TABLE point_transaction (
  id                bigint PRIMARY KEY,
  user_id           bigint NOT NULL,
  change            bigint NOT NULL,
  type              varchar(16) NOT NULL,
  ref_type          varchar(32),
  ref_id            bigint,
  balance_after     bigint NOT NULL,
  expire_at         timestamptz,
  consumed          bool NOT NULL DEFAULT false,
  reason            varchar(500) NOT NULL DEFAULT '',
  created_by        bigint,
  idem_key          varchar(128),
  created_at        timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_point_txn_type CHECK (type IN ('earn','spend','expire','refund','admin_adjust','freeze','unfreeze'))
);
CREATE INDEX idx_point_txn_user_time ON point_transaction(user_id, created_at DESC);
CREATE INDEX idx_point_txn_ref ON point_transaction(ref_type, ref_id);
CREATE INDEX idx_point_txn_expire ON point_transaction(expire_at) WHERE type='earn' AND consumed=false;
CREATE UNIQUE INDEX uq_point_txn_idem ON point_transaction(idem_key) WHERE idem_key IS NOT NULL;

-- =========================================================
-- 积分规则
-- =========================================================
CREATE TABLE point_rule (
  code              varchar(64) PRIMARY KEY,
  enabled           bool NOT NULL DEFAULT true,
  config            jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_by        bigint,
  updated_at        timestamptz NOT NULL DEFAULT now()
);

-- 默认规则种子
INSERT INTO point_rule(code, enabled, config) VALUES
  ('order_earn',   true, '{"rate":0.01,"earn_delay_days":7,"expire_days":365}'),
  ('order_deduct', true, '{"max_deduct_rate":0.30,"cents_per_point":1}'),
  ('sign_in',      true, '{"daily":1,"weekly_bonus":7}'),
  ('invite',       false,'{"per_invite":50}'),
  ('review',       false,'{"per_review":10,"daily_cap":30}'),
  ('birthday',     true, '{"fixed":100}'),
  ('register',     true, '{"fixed":50}')
ON CONFLICT (code) DO NOTHING;

-- =========================================================
-- 积分调整工单
-- =========================================================
CREATE TABLE point_adjust_ticket (
  id                bigint PRIMARY KEY,
  user_id           bigint NOT NULL,
  change            bigint NOT NULL,
  reason            varchar(500) NOT NULL,
  status            varchar(16) NOT NULL DEFAULT 'pending',
  applicant_admin_id bigint NOT NULL,
  approver_admin_id  bigint,
  created_at        timestamptz NOT NULL DEFAULT now(),
  approved_at       timestamptz,
  CONSTRAINT chk_point_ticket_status CHECK (status IN ('pending','approved','rejected'))
);
CREATE INDEX idx_point_ticket_status ON point_adjust_ticket(status, created_at DESC);

-- =========================================================
-- 会员等级
-- =========================================================
CREATE TABLE member_level (
  code                       varchar(16) PRIMARY KEY,
  name                       varchar(32) NOT NULL,
  threshold_amount_cents     bigint NOT NULL,
  points_multiplier          decimal(3,1) NOT NULL DEFAULT 1.0,
  benefits                   jsonb NOT NULL DEFAULT '{}'::jsonb,
  sort                       int NOT NULL DEFAULT 0,
  is_active                  bool NOT NULL DEFAULT true,
  updated_at                 timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE member_level IS '会员等级配置；用户当前等级写在 user.member_level_code';

-- 默认等级种子
INSERT INTO member_level(code, name, threshold_amount_cents, points_multiplier, sort, is_active) VALUES
  ('normal',  '普通会员', 0,        1.0, 0, true),
  ('silver',  '银卡会员', 50000,    1.2, 1, true),
  ('gold',    '金卡会员', 200000,   1.5, 2, true),
  ('diamond', '钻石会员', 1000000,  2.0, 3, true)
ON CONFLICT (code) DO NOTHING;

-- =========================================================
-- 通知模板（会员体系新增）
-- =========================================================
INSERT INTO notification_template(id, code, channel, fields, enabled)
VALUES
  -- 雪花用 floor(now ts) 占位，改用固定大值避免冲突
  (162001000000001, 'point.earned',       'wxmp', '{"thing1":"reason","amount2":"amount","time3":"time"}', true),
  (162001000000002, 'point.expire_warn',  'wxmp', '{"thing1":"reason","amount2":"amount","time3":"expire_time"}', true),
  (162001000000003, 'coupon.granted',     'wxmp', '{"thing1":"name","amount2":"value","time3":"expire_time"}', true),
  (162001000000004, 'coupon.expire_warn', 'wxmp', '{"thing1":"name","time3":"expire_time"}', true),
  (162001000000005, 'coupon.birthday',    'wxmp', '{"thing1":"greeting","amount2":"value"}', true),
  (162001000000006, 'member.level_up',    'wxmp', '{"thing1":"new_level","time2":"time"}', true)
ON CONFLICT (code) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM notification_template WHERE code IN (
  'point.earned','point.expire_warn','coupon.granted','coupon.expire_warn','coupon.birthday','member.level_up'
);

DROP TABLE IF EXISTS member_level;
DROP TABLE IF EXISTS point_adjust_ticket;
DROP TABLE IF EXISTS point_rule;
DROP INDEX IF EXISTS uq_point_txn_idem;
DROP INDEX IF EXISTS idx_point_txn_expire;
DROP INDEX IF EXISTS idx_point_txn_ref;
DROP INDEX IF EXISTS idx_point_txn_user_time;
DROP TABLE IF EXISTS point_transaction;
DROP TABLE IF EXISTS point_account;
DROP TABLE IF EXISTS coupon_grant_task;
DROP TABLE IF EXISTS coupon_redeem_code;
DROP INDEX IF EXISTS uq_user_coupon_lock;
DROP INDEX IF EXISTS idx_user_coupon_template;
DROP INDEX IF EXISTS idx_user_coupon_order;
DROP INDEX IF EXISTS idx_user_coupon_user_status;
DROP TABLE IF EXISTS user_coupon;
DROP INDEX IF EXISTS idx_coupon_tpl_status;
DROP TABLE IF EXISTS coupon_template;

DROP INDEX IF EXISTS idx_user_member_level;
ALTER TABLE "user"
  DROP COLUMN IF EXISTS register_share_link_id,
  DROP COLUMN IF EXISTS register_channel_code,
  DROP COLUMN IF EXISTS recent_365d_gmv_cents,
  DROP COLUMN IF EXISTS member_level_expire_at,
  DROP COLUMN IF EXISTS member_level_code;

DROP INDEX IF EXISTS idx_order_coupon;
ALTER TABLE "order"
  DROP COLUMN IF EXISTS point_earned,
  DROP COLUMN IF EXISTS point_deduct_cents,
  DROP COLUMN IF EXISTS point_used,
  DROP COLUMN IF EXISTS coupon_amount_cents;

-- +goose StatementEnd
