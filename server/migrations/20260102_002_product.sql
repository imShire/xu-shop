-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE category (
  id          bigint PRIMARY KEY,
  parent_id   bigint NOT NULL DEFAULT 0,
  name        varchar(64) NOT NULL,
  icon        varchar(512),
  sort        int NOT NULL DEFAULT 0,
  status      varchar(16) NOT NULL DEFAULT 'enabled',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
);
CREATE INDEX idx_category_parent ON category(parent_id, sort);

CREATE TABLE product (
  id           bigint PRIMARY KEY,
  category_id  bigint NOT NULL,
  title        varchar(60) NOT NULL,
  subtitle     varchar(120),
  main_image   varchar(512) NOT NULL,
  images       jsonb NOT NULL DEFAULT '[]',
  video_url    varchar(512),
  detail_html  text,
  detail_nodes jsonb,
  status       varchar(16) NOT NULL DEFAULT 'draft',
  sales        int NOT NULL DEFAULT 0,
  sort         int NOT NULL DEFAULT 0,
  tags         jsonb NOT NULL DEFAULT '[]',
  price_min_cents bigint NOT NULL DEFAULT 0,
  price_max_cents bigint NOT NULL DEFAULT 0,
  on_sale_at   timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  deleted_at   timestamptz
);
CREATE INDEX idx_product_status_sort ON product(status, sort DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_category ON product(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_title_trgm ON product USING gin(title gin_trgm_ops);
CREATE INDEX idx_product_subtitle_trgm ON product USING gin(subtitle gin_trgm_ops);

CREATE TABLE product_spec (
  id          bigint PRIMARY KEY,
  product_id  bigint NOT NULL,
  name        varchar(32) NOT NULL,
  sort        int NOT NULL DEFAULT 0
);
CREATE INDEX idx_spec_product ON product_spec(product_id);

CREATE TABLE product_spec_value (
  id        bigint PRIMARY KEY,
  spec_id   bigint NOT NULL,
  value     varchar(32) NOT NULL,
  sort      int NOT NULL DEFAULT 0
);

CREATE TABLE sku (
  id                   bigint PRIMARY KEY,
  product_id           bigint NOT NULL,
  attrs                jsonb NOT NULL DEFAULT '{}',
  price_cents          bigint NOT NULL,
  original_price_cents bigint,
  stock                int NOT NULL DEFAULT 0,
  locked_stock         int NOT NULL DEFAULT 0,
  weight_g             int NOT NULL DEFAULT 0,
  sku_code             varchar(64),
  image                varchar(512),
  status               varchar(16) NOT NULL DEFAULT 'active',
  low_stock_threshold  int NOT NULL DEFAULT 0,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_sku_product ON sku(product_id);
CREATE UNIQUE INDEX uq_sku_code ON sku(sku_code) WHERE sku_code IS NOT NULL;

CREATE TABLE user_view_history (
  user_id    bigint NOT NULL,
  product_id bigint NOT NULL,
  viewed_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, product_id)
);
CREATE INDEX idx_view_user_time ON user_view_history(user_id, viewed_at DESC);

CREATE TABLE user_favorite (
  user_id    bigint NOT NULL,
  product_id bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, product_id)
);

CREATE TABLE inventory_log (
  id              bigint PRIMARY KEY,
  sku_id          bigint NOT NULL,
  change          int NOT NULL,
  type            varchar(16) NOT NULL,
  ref_type        varchar(16),
  ref_id          bigint,
  balance_before  int NOT NULL,
  balance_after   int NOT NULL,
  locked_before   int NOT NULL,
  locked_after    int NOT NULL,
  operator_type   varchar(8),
  operator_id     bigint,
  reason          varchar(200),
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_inv_sku_time ON inventory_log(sku_id, created_at DESC);

CREATE TABLE low_stock_alert (
  id                 bigint PRIMARY KEY,
  sku_id             bigint NOT NULL,
  threshold_at_alert int NOT NULL,
  current_stock      int NOT NULL,
  status             varchar(8) NOT NULL DEFAULT 'unread',
  read_by            bigint,
  read_at            timestamptz,
  created_at         timestamptz NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS low_stock_alert;
DROP INDEX IF EXISTS idx_inv_sku_time;
DROP TABLE IF EXISTS inventory_log;
DROP TABLE IF EXISTS user_favorite;
DROP INDEX IF EXISTS idx_view_user_time;
DROP TABLE IF EXISTS user_view_history;
DROP INDEX IF EXISTS uq_sku_code;
DROP INDEX IF EXISTS idx_sku_product;
DROP TABLE IF EXISTS sku;
DROP TABLE IF EXISTS product_spec_value;
DROP INDEX IF EXISTS idx_spec_product;
DROP TABLE IF EXISTS product_spec;
DROP INDEX IF EXISTS idx_product_subtitle_trgm;
DROP INDEX IF EXISTS idx_product_title_trgm;
DROP INDEX IF EXISTS idx_product_category;
DROP INDEX IF EXISTS idx_product_status_sort;
DROP TABLE IF EXISTS product;
DROP INDEX IF EXISTS idx_category_parent;
DROP TABLE IF EXISTS category;
-- +goose StatementEnd
