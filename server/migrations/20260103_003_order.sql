-- +goose Up
-- +goose StatementBegin

CREATE TABLE cart_item (
  id                    bigint PRIMARY KEY,
  user_id               bigint NOT NULL,
  sku_id                bigint NOT NULL,
  qty                   int NOT NULL CHECK (qty >= 1 AND qty <= 999),
  snapshot_price_cents  bigint NOT NULL,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, sku_id)
);
CREATE INDEX idx_cart_user ON cart_item(user_id, updated_at DESC);

CREATE TABLE freight_template (
  id                     bigint PRIMARY KEY,
  name                   varchar(64) NOT NULL,
  free_threshold_cents   bigint NOT NULL DEFAULT 9900,
  default_fee_cents      bigint NOT NULL DEFAULT 1000,
  remote_threshold_cents bigint NOT NULL DEFAULT 19900,
  remote_fee_cents       bigint NOT NULL DEFAULT 2000,
  remote_provinces       jsonb NOT NULL DEFAULT '[]',
  is_default             bool NOT NULL DEFAULT false,
  created_at             timestamptz NOT NULL DEFAULT now(),
  updated_at             timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_freight_default ON freight_template((1)) WHERE is_default = true;

CREATE TABLE "order" (
  id                       bigint PRIMARY KEY,
  order_no                 varchar(32) NOT NULL UNIQUE,
  shop_id                  bigint NOT NULL DEFAULT 1,
  user_id                  bigint NOT NULL,
  status                   varchar(16) NOT NULL,
  goods_cents              bigint NOT NULL,
  freight_cents            bigint NOT NULL DEFAULT 0,
  discount_cents           bigint NOT NULL DEFAULT 0,
  coupon_discount_cents    bigint NOT NULL DEFAULT 0,
  total_cents              bigint NOT NULL,
  pay_cents                bigint NOT NULL,
  address_snapshot         jsonb NOT NULL,
  buyer_remark             varchar(200),
  cancel_request_pending   bool NOT NULL DEFAULT false,
  cancel_request_reason    varchar(200),
  cancel_request_at        timestamptz,
  cancel_reason            varchar(200),
  distributor_id           bigint,
  distribution_path        jsonb,
  group_buy_order_id       bigint,
  coupon_id                bigint,
  from_share_user_id       bigint,
  from_channel_code_id     bigint,
  idempotency_key          varchar(64),
  current_prepay_id        varchar(64),
  current_prepay_expire_at timestamptz,
  expire_at                timestamptz NOT NULL,
  paid_at                  timestamptz,
  shipped_at               timestamptz,
  completed_at             timestamptz,
  cancelled_at             timestamptz,
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_user_time ON "order"(user_id, created_at DESC);
CREATE INDEX idx_order_status_expire ON "order"(status, expire_at) WHERE status='pending';
CREATE INDEX idx_order_status_time ON "order"(status, created_at DESC);
CREATE UNIQUE INDEX uq_order_idem ON "order"(user_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;

CREATE TABLE order_item (
  id                bigint PRIMARY KEY,
  order_id          bigint NOT NULL,
  product_id        bigint NOT NULL,
  sku_id            bigint NOT NULL,
  product_snapshot  jsonb NOT NULL,
  price_cents       bigint NOT NULL,
  qty               int NOT NULL,
  weight_g          int NOT NULL,
  created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_orderitem_order ON order_item(order_id);

CREATE TABLE order_log (
  id            bigint PRIMARY KEY,
  order_id      bigint NOT NULL,
  from_status   varchar(16),
  to_status     varchar(16),
  reason        varchar(200),
  operator_type varchar(8),
  operator_id   bigint,
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_orderlog_order ON order_log(order_id, created_at);

CREATE TABLE order_remark (
  id         bigint PRIMARY KEY,
  order_id   bigint NOT NULL,
  admin_id   bigint NOT NULL,
  content    varchar(500) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_remark;
DROP INDEX IF EXISTS idx_orderlog_order;
DROP TABLE IF EXISTS order_log;
DROP INDEX IF EXISTS idx_orderitem_order;
DROP TABLE IF EXISTS order_item;
DROP INDEX IF EXISTS uq_order_idem;
DROP INDEX IF EXISTS idx_order_status_time;
DROP INDEX IF EXISTS idx_order_status_expire;
DROP INDEX IF EXISTS idx_order_user_time;
DROP TABLE IF EXISTS "order";
DROP INDEX IF EXISTS uq_freight_default;
DROP TABLE IF EXISTS freight_template;
DROP INDEX IF EXISTS idx_cart_user;
DROP TABLE IF EXISTS cart_item;
-- +goose StatementEnd
