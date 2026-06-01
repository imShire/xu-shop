-- A2 后台审计日志中间件配套表。
-- 中间件捕获 /admin/ 下所有 POST/PUT/PATCH/DELETE 请求；
-- 写库异步执行，不阻塞主响应。
-- 与既有 audit_log 表（按模块手工记录）并存：
--   * audit_log：业务模块手动 diff 留痕（用于"实名快照"）
--   * admin_audit_log：HTTP 层全量请求/响应留痕
-- 配套 docs/arch/91-db-schema.md v1.11 段、docs/arch/10-admin.md 审计章节。

-- +goose Up
-- +goose StatementBegin

CREATE TABLE admin_audit_log (
  id                bigint PRIMARY KEY,
  admin_id          bigint NOT NULL DEFAULT 0,   -- 0 表示未登录（如登录失败）
  action            varchar(64) NOT NULL,        -- 业务动作（route meta 注入 或 method:path 兜底）
  target_type       varchar(64),
  target_id         varchar(64),
  method            varchar(8) NOT NULL,
  path              varchar(255) NOT NULL,
  query             text,
  request_body      jsonb,
  response_status   int NOT NULL,
  response_excerpt  text,
  client_ip         inet,
  user_agent        varchar(255),
  trace_id          varchar(64),
  duration_ms       int NOT NULL DEFAULT 0,
  created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_audit_admin_time
  ON admin_audit_log(admin_id, created_at DESC);
CREATE INDEX idx_admin_audit_action_time
  ON admin_audit_log(action, created_at DESC);
CREATE INDEX idx_admin_audit_target
  ON admin_audit_log(target_type, target_id)
  WHERE target_type IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_admin_audit_target;
DROP INDEX IF EXISTS idx_admin_audit_action_time;
DROP INDEX IF EXISTS idx_admin_audit_admin_time;
DROP TABLE IF EXISTS admin_audit_log;

-- +goose StatementEnd
