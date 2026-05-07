-- +goose Up
CREATE TABLE IF NOT EXISTS article (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    cover VARCHAR(512),
    content TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    sort INT NOT NULL DEFAULT 0,
    created_by BIGINT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_article_status ON article(status);

-- 首页装修配置
CREATE TABLE IF NOT EXISTS page_config (
    id BIGINT PRIMARY KEY,
    page_key VARCHAR(32) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    modules JSONB NOT NULL DEFAULT '[]',
    is_active BOOL NOT NULL DEFAULT FALSE,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_page_config_key_active ON page_config(page_key, is_active);

-- +goose Down
DROP TABLE IF EXISTS page_config;
DROP TABLE IF EXISTS article;
