-- +goose Up
-- +goose StatementBegin
ALTER TABLE customer_tag
  ADD COLUMN IF NOT EXISTS qywx_tag_id varchar(64);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customer_tag
  DROP COLUMN IF EXISTS qywx_tag_id;
-- +goose StatementEnd
