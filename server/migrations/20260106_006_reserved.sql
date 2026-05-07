-- +goose Up
-- +goose StatementBegin

CREATE TABLE distributor (
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

CREATE TABLE commission_record (
  id              bigint PRIMARY KEY,
  order_id        bigint NOT NULL,
  distributor_id  bigint NOT NULL,
  level           smallint NOT NULL,
  amount_cents    bigint NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'locked',
  settle_at       timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE group_buy_activity (
  id                bigint PRIMARY KEY,
  product_id        bigint NOT NULL,
  sku_id            bigint NOT NULL,
  group_size        int NOT NULL,
  group_price_cents bigint NOT NULL,
  start_at          timestamptz NOT NULL,
  end_at            timestamptz NOT NULL,
  status            varchar(16) NOT NULL DEFAULT 'enabled',
  created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE group_buy_order (
  id              bigint PRIMARY KEY,
  activity_id     bigint NOT NULL,
  leader_user_id  bigint NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'pending',
  expire_at       timestamptz NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE coupon (
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

CREATE TABLE user_coupon (
  id              bigint PRIMARY KEY,
  user_id         bigint NOT NULL,
  coupon_id       bigint NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'unused',
  used_order_id   bigint,
  claimed_at      timestamptz NOT NULL DEFAULT now(),
  used_at         timestamptz,
  expire_at       timestamptz
);

CREATE TABLE points_log (
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

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS points_log;
DROP TABLE IF EXISTS user_coupon;
DROP TABLE IF EXISTS coupon;
DROP TABLE IF EXISTS group_buy_order;
DROP TABLE IF EXISTS group_buy_activity;
DROP TABLE IF EXISTS commission_record;
DROP TABLE IF EXISTS distributor;
-- +goose StatementEnd
