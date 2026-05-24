-- +goose Up
-- +goose StatementBegin
ALTER TABLE banner
  ADD COLUMN start_at timestamptz,
  ADD COLUMN end_at   timestamptz;

COMMENT ON COLUMN banner.start_at IS '展示开始时间（NULL 表示立即）';
COMMENT ON COLUMN banner.end_at   IS '展示结束时间（NULL 表示永久）';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE banner
  DROP COLUMN IF EXISTS start_at,
  DROP COLUMN IF EXISTS end_at;
-- +goose StatementEnd
