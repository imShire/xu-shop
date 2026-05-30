-- v1.4 售后单独立链路（aftersale_order + aftersale_negotiation）。
-- 配套 docs/arch/91-db-schema.md v1.10 段、docs/arch/09-aftersale.md v1.4、
-- docs/arch/99-revisions.md v1.4 段。
-- 与旧 order.cancel_request_pending 字段并存：旧链路不写本表。

-- +goose Up
-- +goose StatementBegin

CREATE TABLE aftersale_order (
  id                  bigint PRIMARY KEY,
  aftersale_no        varchar(32) NOT NULL UNIQUE,
  order_id            bigint NOT NULL,
  order_item_id       bigint,
  user_id             bigint NOT NULL,
  type                varchar(16) NOT NULL,
  reason              varchar(200) NOT NULL,
  refund_amount_cents bigint NOT NULL DEFAULT 0,
  status              varchar(24) NOT NULL DEFAULT 'applying',
  buyer_evidence      jsonb NOT NULL DEFAULT '[]',
  buyer_express       jsonb,
  seller_remark       varchar(500) NOT NULL DEFAULT '',
  refund_id           bigint,
  applied_at          timestamptz NOT NULL DEFAULT now(),
  agreed_at           timestamptz,
  returned_at         timestamptz,
  received_at         timestamptz,
  completed_at        timestamptz,
  closed_at           timestamptz,
  auto_close_at       timestamptz NOT NULL,
  operator_admin_id   bigint,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_aftersale_type
    CHECK (type IN ('refund_only','refund_return','exchange')),
  CONSTRAINT chk_aftersale_status
    CHECK (status IN ('applying','seller_agreed','buyer_returned','seller_received',
                      'completed','seller_rejected','cancelled','closed'))
);

CREATE INDEX idx_aftersale_order ON aftersale_order(order_id);
CREATE INDEX idx_aftersale_user_status ON aftersale_order(user_id, status, applied_at DESC);
CREATE INDEX idx_aftersale_scan ON aftersale_order(status, auto_close_at)
  WHERE status IN ('applying','seller_agreed','buyer_returned','seller_received');

CREATE UNIQUE INDEX uq_aftersale_active_item ON aftersale_order(order_item_id)
  WHERE order_item_id IS NOT NULL
    AND status NOT IN ('completed','seller_rejected','cancelled','closed');

CREATE UNIQUE INDEX uq_aftersale_active_order ON aftersale_order(order_id)
  WHERE order_item_id IS NULL
    AND status NOT IN ('completed','seller_rejected','cancelled','closed');

CREATE TABLE aftersale_negotiation (
  id           bigint PRIMARY KEY,
  aftersale_id bigint NOT NULL,
  role         varchar(8) NOT NULL,
  admin_id     bigint,
  content      varchar(1000) NOT NULL DEFAULT '',
  evidence     jsonb NOT NULL DEFAULT '[]',
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_aftersale_neg_role CHECK (role IN ('buyer','seller','system'))
);
CREATE INDEX idx_aftersale_neg ON aftersale_negotiation(aftersale_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS aftersale_negotiation;
DROP INDEX IF EXISTS uq_aftersale_active_order;
DROP INDEX IF EXISTS uq_aftersale_active_item;
DROP INDEX IF EXISTS idx_aftersale_scan;
DROP INDEX IF EXISTS idx_aftersale_user_status;
DROP INDEX IF EXISTS idx_aftersale_order;
DROP TABLE IF EXISTS aftersale_order;

-- +goose StatementEnd
