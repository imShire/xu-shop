-- +goose Up
-- +goose StatementBegin
ALTER TABLE banner ADD COLUMN IF NOT EXISTS link_config jsonb;
ALTER TABLE nav_icon ADD COLUMN IF NOT EXISTS link_config jsonb;

-- 历史数据迁移：将旧 link_url 转成 custom 类型的 link_config
UPDATE banner SET link_config = jsonb_build_object(
    'type', 'custom',
    'target_id', '',
    'target_name', '',
    'url', link_url
) WHERE link_url IS NOT NULL AND link_url != '' AND link_config IS NULL;

UPDATE nav_icon SET link_config = jsonb_build_object(
    'type', 'custom',
    'target_id', '',
    'target_name', '',
    'url', link_url
) WHERE link_url IS NOT NULL AND link_url != '' AND link_config IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE banner DROP COLUMN IF EXISTS link_config;
ALTER TABLE nav_icon DROP COLUMN IF EXISTS link_config;
-- +goose StatementEnd
