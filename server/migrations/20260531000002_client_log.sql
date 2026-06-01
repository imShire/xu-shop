-- D1 前端日志上报落地表。
-- 接收 admin / client_h5 / client_weapp 的 JS error / unhandled rejection / API 失败。
-- 入库前服务端做长度截断；查询展示侧负责 XSS 防护。
-- 配套 docs/arch/91-db-schema.md v1.11 段。

-- +goose Up
-- +goose StatementBegin

CREATE TABLE client_log (
  id          bigint PRIMARY KEY,
  source      varchar(16) NOT NULL,                 -- admin / client_h5 / client_weapp
  level       varchar(8)  NOT NULL,                 -- error / warn / info
  message     text NOT NULL,
  stack       text,
  url         varchar(512),
  user_agent  varchar(255),
  release     varchar(64),
  user_id     bigint,
  admin_id    bigint,
  extra       jsonb,
  client_ip   inet,
  trace_id    varchar(64),
  created_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_client_log_source
    CHECK (source IN ('admin','client_h5','client_weapp')),
  CONSTRAINT chk_client_log_level
    CHECK (level IN ('error','warn','info'))
);

CREATE INDEX idx_client_log_source_level_time
  ON client_log(source, level, created_at DESC);
CREATE INDEX idx_client_log_release_time
  ON client_log(release, created_at DESC)
  WHERE release IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_client_log_release_time;
DROP INDEX IF EXISTS idx_client_log_source_level_time;
DROP TABLE IF EXISTS client_log;

-- +goose StatementEnd
