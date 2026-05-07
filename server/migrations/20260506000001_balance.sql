-- +goose Up

-- 用户余额字段
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS balance_cents BIGINT NOT NULL DEFAULT 0;

-- 余额流水表
CREATE TABLE IF NOT EXISTS balance_log (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user"(id),
    change_cents BIGINT NOT NULL,
    type VARCHAR(16) NOT NULL,
    ref_type VARCHAR(16),
    ref_id BIGINT,
    balance_before_cents BIGINT NOT NULL,
    balance_after_cents BIGINT NOT NULL,
    operator_id BIGINT,
    remark VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_balance_log_user_id ON balance_log(user_id);

-- 订单余额支付字段
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS balance_pay_cents BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE "order" DROP COLUMN IF EXISTS balance_pay_cents;
DROP TABLE IF EXISTS balance_log;
ALTER TABLE "user" DROP COLUMN IF EXISTS balance_cents;
