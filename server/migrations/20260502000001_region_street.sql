-- +goose Up
-- +goose StatementBegin
ALTER TABLE address
  ADD COLUMN IF NOT EXISTS street_code varchar(12),
  ADD COLUMN IF NOT EXISTS street varchar(32);

ALTER TABLE sender_address
  ADD COLUMN IF NOT EXISTS province_code varchar(12),
  ADD COLUMN IF NOT EXISTS city_code varchar(12),
  ADD COLUMN IF NOT EXISTS district_code varchar(12),
  ADD COLUMN IF NOT EXISTS street_code varchar(12),
  ADD COLUMN IF NOT EXISTS street varchar(32);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sender_address
  DROP COLUMN IF EXISTS street,
  DROP COLUMN IF EXISTS street_code,
  DROP COLUMN IF EXISTS district_code,
  DROP COLUMN IF EXISTS city_code,
  DROP COLUMN IF EXISTS province_code;

ALTER TABLE address
  DROP COLUMN IF EXISTS street,
  DROP COLUMN IF EXISTS street_code;
-- +goose StatementEnd
