-- +goose Up
-- +goose StatementBegin

CREATE TABLE nav_icon (
  id          bigint PRIMARY KEY,
  title       varchar(16) NOT NULL DEFAULT '',
  icon_url    text NOT NULL,
  link_url    text NOT NULL DEFAULT '',
  sort        int NOT NULL DEFAULT 0,
  is_active   bool NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_nav_icon_sort ON nav_icon(sort ASC) WHERE is_active = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_nav_icon_sort;
DROP TABLE IF EXISTS nav_icon;

-- +goose StatementEnd
