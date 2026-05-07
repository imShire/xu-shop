-- +goose Up
-- +goose StatementBegin

CREATE TABLE notification_template (
  id                    bigint PRIMARY KEY,
  code                  varchar(64) NOT NULL UNIQUE,
  channel               varchar(16) NOT NULL,
  template_id_external  varchar(128),
  fields                jsonb NOT NULL DEFAULT '{}',
  enabled               bool NOT NULL DEFAULT true,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE notification_task (
  id              bigint PRIMARY KEY,
  template_code   varchar(64) NOT NULL,
  target_type     varchar(16) NOT NULL,
  target          varchar(512) NOT NULL,
  params          jsonb NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'pending',
  retry_count     int NOT NULL DEFAULT 0,
  last_error      text,
  dedup_key       varchar(128),
  sent_at         timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_notif_dedup ON notification_task(dedup_key) WHERE dedup_key IS NOT NULL;

CREATE TABLE channel_code (
  id                bigint PRIMARY KEY,
  name              varchar(64) NOT NULL,
  qr_image_url      varchar(512),
  qywx_config_id    varchar(64),
  customer_servers  jsonb NOT NULL DEFAULT '[]',
  tag_ids           jsonb NOT NULL DEFAULT '[]',
  welcome_text      varchar(500),
  scan_count        int NOT NULL DEFAULT 0,
  add_count         int NOT NULL DEFAULT 0,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE customer_tag (
  id          bigint PRIMARY KEY,
  name        varchar(32) NOT NULL UNIQUE,
  source      varchar(16) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_tag (
  user_id     bigint NOT NULL,
  tag_id      bigint NOT NULL,
  source      varchar(16) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE share_attribution (
  id              bigint PRIMARY KEY,
  share_user_id   bigint NOT NULL,
  viewer_user_id  bigint,
  product_id      bigint,
  channel         varchar(16) NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_share_user ON share_attribution(share_user_id, created_at DESC);

CREATE TABLE share_short_code (
  id              bigint PRIMARY KEY,
  share_user_id   bigint NOT NULL,
  product_id      bigint,
  channel_code_id bigint,
  expire_at       timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (share_user_id, product_id, channel_code_id)
);

CREATE TABLE stats_daily (
  date                  date PRIMARY KEY,
  paid_order_count      int NOT NULL DEFAULT 0,
  paid_amount_cents     bigint NOT NULL DEFAULT 0,
  refund_amount_cents   bigint NOT NULL DEFAULT 0,
  net_amount_cents      bigint NOT NULL DEFAULT 0,
  paid_user_count       int NOT NULL DEFAULT 0,
  new_user_count        int NOT NULL DEFAULT 0,
  updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE stats_product_daily (
  date          date NOT NULL,
  product_id    bigint NOT NULL,
  qty           int NOT NULL DEFAULT 0,
  amount_cents  bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (date, product_id)
);

CREATE TABLE stats_channel_daily (
  date              date NOT NULL,
  channel_code_id   bigint NOT NULL,
  scan_count        int NOT NULL DEFAULT 0,
  add_count         int NOT NULL DEFAULT 0,
  order_count       int NOT NULL DEFAULT 0,
  amount_cents      bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (date, channel_code_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stats_channel_daily;
DROP TABLE IF EXISTS stats_product_daily;
DROP TABLE IF EXISTS stats_daily;
DROP TABLE IF EXISTS share_short_code;
DROP INDEX IF EXISTS idx_share_user;
DROP TABLE IF EXISTS share_attribution;
DROP TABLE IF EXISTS user_tag;
DROP TABLE IF EXISTS customer_tag;
DROP TABLE IF EXISTS channel_code;
DROP INDEX IF EXISTS uq_notif_dedup;
DROP TABLE IF EXISTS notification_task;
DROP TABLE IF EXISTS notification_template;
-- +goose StatementEnd
