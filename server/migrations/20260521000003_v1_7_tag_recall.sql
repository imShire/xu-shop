-- v1.7 用户标签 + 召回（配套 docs/arch/17-tag-recall.md 与 91-db-schema.md v1.7）

-- +goose Up
-- +goose StatementBegin

-- =========================================================
-- 标签字典
-- =========================================================
CREATE TABLE user_tag (
  code              varchar(64) PRIMARY KEY,
  name              varchar(64) NOT NULL,
  category          varchar(16) NOT NULL,
  parent_code       varchar(64),
  color             varchar(16),
  description       varchar(500) NOT NULL DEFAULT '',
  source            varchar(8) NOT NULL DEFAULT 'auto',
  config            jsonb NOT NULL DEFAULT '{}'::jsonb,
  enabled           bool NOT NULL DEFAULT true,
  sort              int NOT NULL DEFAULT 0,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_user_tag_category CHECK (category IN ('rfm','lifecycle','category_pref','price_band','source','business','member','system')),
  CONSTRAINT chk_user_tag_source CHECK (source IN ('auto','manual'))
);
CREATE INDEX idx_user_tag_category ON user_tag(category, enabled);
COMMENT ON TABLE user_tag IS '用户标签字典';

-- =========================================================
-- 用户标签关系
-- =========================================================
CREATE TABLE user_tag_relation (
  id                bigint PRIMARY KEY,
  user_id           bigint NOT NULL,
  tag_code          varchar(64) NOT NULL,
  score             int NOT NULL DEFAULT 1,
  source            varchar(8) NOT NULL DEFAULT 'auto',
  source_ref        varchar(128),
  expire_at         timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_user_tag_rel_source CHECK (source IN ('auto','manual'))
);
CREATE UNIQUE INDEX uq_user_tag_rel ON user_tag_relation(user_id, tag_code);
CREATE INDEX idx_user_tag_rel_tag ON user_tag_relation(tag_code, source);
CREATE INDEX idx_user_tag_rel_user ON user_tag_relation(user_id);

-- =========================================================
-- 月度全量快照
-- =========================================================
CREATE TABLE user_tag_snapshot (
  id                bigint PRIMARY KEY,
  snapshot_date     date NOT NULL,
  user_id           bigint NOT NULL,
  tags              jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_user_tag_snapshot ON user_tag_snapshot(snapshot_date, user_id);
CREATE INDEX idx_user_tag_snapshot_date ON user_tag_snapshot(snapshot_date);

-- =========================================================
-- 召回活动
-- =========================================================
CREATE TABLE recall_campaign (
  id                       bigint PRIMARY KEY,
  name                     varchar(128) NOT NULL,
  goal                     varchar(500) NOT NULL DEFAULT '',
  audience_filter          jsonb NOT NULL DEFAULT '{}'::jsonb,
  actions                  jsonb NOT NULL DEFAULT '[]'::jsonb,
  trigger_type             varchar(16) NOT NULL,
  trigger_config           jsonb NOT NULL DEFAULT '{}'::jsonb,
  effective_from           timestamptz,
  effective_to             timestamptz,
  throttle_per_user_days   int NOT NULL DEFAULT 7,
  daily_quota              bigint NOT NULL DEFAULT 0,
  total_quota              bigint NOT NULL DEFAULT 0,
  attribution_window_days  int NOT NULL DEFAULT 7,
  status                   varchar(16) NOT NULL DEFAULT 'draft',
  created_by               bigint NOT NULL,
  approver_admin_id        bigint,
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_recall_trigger CHECK (trigger_type IN ('cron','event','immediate')),
  CONSTRAINT chk_recall_status CHECK (status IN ('draft','online','paused','closed'))
);
CREATE INDEX idx_recall_camp_status ON recall_campaign(status, trigger_type);

-- =========================================================
-- 召回日志
-- =========================================================
CREATE TABLE recall_log (
  id                       bigint PRIMARY KEY,
  campaign_id              bigint NOT NULL,
  user_id                  bigint NOT NULL,
  triggered_at             timestamptz NOT NULL DEFAULT now(),
  audience_snapshot        jsonb NOT NULL DEFAULT '{}'::jsonb,
  actions_result           jsonb NOT NULL DEFAULT '[]'::jsonb,
  opened_at                timestamptz,
  converted_order_id       bigint,
  converted_at             timestamptz,
  converted_gmv_cents      bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX uq_recall_log_campaign_user_day
  ON recall_log(campaign_id, user_id, (triggered_at::date));
CREATE INDEX idx_recall_log_user_time ON recall_log(user_id, triggered_at DESC);
CREATE INDEX idx_recall_log_camp_time ON recall_log(campaign_id, triggered_at DESC);
CREATE INDEX idx_recall_log_converted ON recall_log(campaign_id) WHERE converted_order_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_recall_log_converted;
DROP INDEX IF EXISTS idx_recall_log_camp_time;
DROP INDEX IF EXISTS idx_recall_log_user_time;
DROP INDEX IF EXISTS uq_recall_log_campaign_user_day;
DROP TABLE IF EXISTS recall_log;
DROP INDEX IF EXISTS idx_recall_camp_status;
DROP TABLE IF EXISTS recall_campaign;
DROP INDEX IF EXISTS idx_user_tag_snapshot_date;
DROP INDEX IF EXISTS uq_user_tag_snapshot;
DROP TABLE IF EXISTS user_tag_snapshot;
DROP INDEX IF EXISTS idx_user_tag_rel_user;
DROP INDEX IF EXISTS idx_user_tag_rel_tag;
DROP INDEX IF EXISTS uq_user_tag_rel;
DROP TABLE IF EXISTS user_tag_relation;
DROP INDEX IF EXISTS idx_user_tag_category;
DROP TABLE IF EXISTS user_tag;

-- +goose StatementEnd
