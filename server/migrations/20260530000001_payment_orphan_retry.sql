-- +goose Up
-- +goose StatementBegin

-- payment_orphan_retry: enqueue 自动退款失败时落表兜底，待 reconciler 扫描重试。
-- 字段语义详见 docs/arch/91-db-schema.md「v1.3 修订」与 docs/arch/07-payment.md。
CREATE TABLE payment_orphan_retry (
  id              bigint PRIMARY KEY,
  payment_id      bigint NOT NULL,
  wx_txid         varchar(64) NOT NULL,
  amount_cents    bigint NOT NULL,
  reason          varchar(200) NOT NULL DEFAULT '',
  retry_count     int NOT NULL DEFAULT 0,
  next_retry_at   timestamptz NOT NULL DEFAULT now(),
  last_error      text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_orphan_retry_pending ON payment_orphan_retry(next_retry_at);
CREATE INDEX idx_orphan_retry_payment ON payment_orphan_retry(payment_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS payment_orphan_retry;

-- +goose StatementEnd
