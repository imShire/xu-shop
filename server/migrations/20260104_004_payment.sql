-- +goose Up
-- +goose StatementBegin

CREATE TABLE payment (
  id              bigint PRIMARY KEY,
  order_id        bigint NOT NULL,
  channel         varchar(16) NOT NULL DEFAULT 'wxpay',
  trade_type      varchar(16) NOT NULL,
  appid           varchar(64),
  prepay_id       varchar(64),
  transaction_id  varchar(64) UNIQUE,
  amount_cents    bigint NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'pending',
  raw_notify      jsonb,
  paid_at         timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_payment_order ON payment(order_id);

CREATE TABLE refund (
  id              bigint PRIMARY KEY,
  order_id        bigint NOT NULL,
  payment_id      bigint NOT NULL,
  refund_no       varchar(32) NOT NULL UNIQUE,
  amount_cents    bigint NOT NULL,
  reason          varchar(200),
  status          varchar(16) NOT NULL DEFAULT 'pending',
  raw_notify      jsonb,
  refunded_at     timestamptz,
  operator_id     bigint,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE reconciliation_diff (
  id                bigint PRIMARY KEY,
  bill_date         date NOT NULL,
  transaction_id    varchar(64),
  order_no          varchar(32),
  our_amount_cents  bigint,
  wx_amount_cents   bigint,
  diff_type         varchar(32),
  status            varchar(16) NOT NULL DEFAULT 'unresolved',
  resolved_by       bigint,
  resolved_at       timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sender_address (
  id          bigint PRIMARY KEY,
  company     varchar(64),
  name        varchar(32) NOT NULL,
  phone       varchar(20) NOT NULL,
  province    varchar(32) NOT NULL,
  city        varchar(32) NOT NULL,
  district    varchar(32) NOT NULL,
  detail      varchar(200) NOT NULL,
  is_default  bool NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_sender_default ON sender_address((1)) WHERE is_default = true;

CREATE TABLE carrier (
  code             varchar(16) PRIMARY KEY,
  name             varchar(32) NOT NULL,
  kdniao_code      varchar(16) NOT NULL,
  monthly_account  varchar(64),
  enabled          bool NOT NULL DEFAULT true,
  sort             int NOT NULL DEFAULT 0
);

CREATE TABLE shipment (
  id                  bigint PRIMARY KEY,
  order_id            bigint NOT NULL,
  carrier_code        varchar(16) NOT NULL,
  tracking_no         varchar(64) NOT NULL,
  waybill_pdf_url     varchar(512),
  sender_snapshot     jsonb NOT NULL,
  receiver_snapshot   jsonb NOT NULL,
  status              varchar(16) NOT NULL DEFAULT 'picked',
  last_track_at       timestamptz,
  delivered_at        timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  UNIQUE(carrier_code, tracking_no)
);
CREATE INDEX idx_shipment_order ON shipment(order_id);
CREATE INDEX idx_shipment_status ON shipment(status, last_track_at);

CREATE TABLE shipment_track (
  id           bigint PRIMARY KEY,
  shipment_id  bigint NOT NULL,
  status       varchar(16) NOT NULL,
  description  text,
  occurred_at  timestamptz NOT NULL,
  raw          jsonb,
  created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_track_shipment_time ON shipment_track(shipment_id, occurred_at DESC);

CREATE TABLE audit_log (
  id                bigint PRIMARY KEY,
  admin_id          bigint NOT NULL,
  admin_username    varchar(64),
  admin_real_name   varchar(64),
  module            varchar(32) NOT NULL,
  action            varchar(32) NOT NULL,
  target_id         varchar(64),
  diff              jsonb,
  ip                inet,
  ua                text,
  created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_admin_time ON audit_log(admin_id, created_at DESC);
CREATE INDEX idx_audit_module_time ON audit_log(module, created_at DESC);

CREATE TABLE system_setting (
  key         varchar(64) PRIMARY KEY,
  value       text NOT NULL,
  is_secret   bool NOT NULL DEFAULT false,
  updated_by  bigint,
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system_setting;
DROP INDEX IF EXISTS idx_audit_module_time;
DROP INDEX IF EXISTS idx_audit_admin_time;
DROP TABLE IF EXISTS audit_log;
DROP INDEX IF EXISTS idx_track_shipment_time;
DROP TABLE IF EXISTS shipment_track;
DROP INDEX IF EXISTS idx_shipment_status;
DROP INDEX IF EXISTS idx_shipment_order;
DROP TABLE IF EXISTS shipment;
DROP TABLE IF EXISTS carrier;
DROP INDEX IF EXISTS uq_sender_default;
DROP TABLE IF EXISTS sender_address;
DROP TABLE IF EXISTS reconciliation_diff;
DROP TABLE IF EXISTS refund;
DROP INDEX IF EXISTS idx_payment_order;
DROP TABLE IF EXISTS payment;
-- +goose StatementEnd
