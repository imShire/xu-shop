-- v1.2 升级第一步：清理 v1.0 reserved/init 阶段的占位表与列。
-- 这些占位表在 docs/arch/91-db-schema.md v1.6/v1.7/v1.8 中被新结构取代。
-- 之前未投入使用，可以安全 DROP（数据为空）。
-- 同时 user 表的 points int 占位列被 v1.6 的 point_account 替代，删掉。

-- +goose Up
-- +goose StatementBegin

-- 1. 删除 14-reserved 旧 coupon / user_coupon / points_log（v1.6 全部重建）
DROP TABLE IF EXISTS points_log;
DROP TABLE IF EXISTS user_coupon;
DROP TABLE IF EXISTS coupon;

-- 2. 删除 14-reserved 旧 distributor / commission_record（v1.8 全部重建，结构差异巨大）
DROP TABLE IF EXISTS commission_record;
DROP TABLE IF EXISTS distributor;

-- 3. 删除 12-private_domain 旧 customer_tag / user_tag / share_attribution（v1.7/v1.8 重建）
DROP TABLE IF EXISTS user_tag;
DROP TABLE IF EXISTS customer_tag;
DROP INDEX IF EXISTS idx_share_user;
DROP TABLE IF EXISTS share_attribution;

-- 4. user 表占位 points 列删除（balance 改由 point_account 管理）
ALTER TABLE "user" DROP COLUMN IF EXISTS points;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE "user" ADD COLUMN IF NOT EXISTS points int NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS share_attribution (
  id              bigint PRIMARY KEY,
  share_user_id   bigint NOT NULL,
  viewer_user_id  bigint,
  product_id      bigint,
  channel         varchar(16) NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_share_user ON share_attribution(share_user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS customer_tag (
  id          bigint PRIMARY KEY,
  name        varchar(32) NOT NULL UNIQUE,
  source      varchar(16) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_tag (
  user_id     bigint NOT NULL,
  tag_id      bigint NOT NULL,
  source      varchar(16) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE IF NOT EXISTS distributor (
  id                       bigint PRIMARY KEY,
  user_id                  bigint NOT NULL UNIQUE,
  level                    smallint NOT NULL DEFAULT 1,
  parent_distributor_id    bigint,
  code                     varchar(32) UNIQUE,
  status                   varchar(16) NOT NULL DEFAULT 'pending',
  apply_at                 timestamptz,
  approved_at              timestamptz,
  created_at               timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS commission_record (
  id              bigint PRIMARY KEY,
  order_id        bigint NOT NULL,
  distributor_id  bigint NOT NULL,
  level           smallint NOT NULL,
  amount_cents    bigint NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'locked',
  settle_at       timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS coupon (
  id                   bigint PRIMARY KEY,
  name                 varchar(64) NOT NULL,
  type                 varchar(16) NOT NULL,
  value                int NOT NULL,
  min_amount_cents     bigint NOT NULL DEFAULT 0,
  valid_from           timestamptz,
  valid_to             timestamptz,
  total                int NOT NULL DEFAULT 0,
  claimed              int NOT NULL DEFAULT 0,
  status               varchar(16) NOT NULL DEFAULT 'enabled',
  created_at           timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_coupon (
  id              bigint PRIMARY KEY,
  user_id         bigint NOT NULL,
  coupon_id       bigint NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'unused',
  used_order_id   bigint,
  claimed_at      timestamptz NOT NULL DEFAULT now(),
  used_at         timestamptz,
  expire_at       timestamptz
);

CREATE TABLE IF NOT EXISTS points_log (
  id              bigint PRIMARY KEY,
  user_id         bigint NOT NULL,
  change          int NOT NULL,
  type            varchar(16) NOT NULL,
  ref_type        varchar(16),
  ref_id          bigint,
  balance_after   int NOT NULL,
  reason          varchar(200),
  created_at      timestamptz NOT NULL DEFAULT now()
);

-- +goose StatementEnd
