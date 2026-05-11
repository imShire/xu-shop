-- +goose NO TRANSACTION
-- +goose Up
-- 在 status='active' 且手机号非空的活跃用户上建立部分唯一索引，防止同一手机号绑定多个活跃账号。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_user_phone_active
    ON "user" (phone)
    WHERE status = 'active'
      AND phone IS NOT NULL
      AND phone != '';

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS uq_user_phone_active;
