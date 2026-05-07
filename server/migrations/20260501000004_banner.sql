-- +goose Up
-- +goose StatementBegin

CREATE TABLE banner (
  id          bigint PRIMARY KEY,
  title       varchar(128) NOT NULL DEFAULT '',
  image_url   text NOT NULL,
  link_url    text NOT NULL DEFAULT '',
  sort        int NOT NULL DEFAULT 0,
  is_active   bool NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_banner_sort ON banner(sort ASC) WHERE is_active = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_banner_sort;
DROP TABLE IF EXISTS banner;

-- +goose StatementEnd
