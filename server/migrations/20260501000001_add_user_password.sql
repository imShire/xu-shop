-- +goose Up
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS password_hash varchar(255);
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS source varchar(32) NOT NULL DEFAULT 'mp';

-- +goose Down
ALTER TABLE "user" DROP COLUMN IF EXISTS password_hash;
ALTER TABLE "user" DROP COLUMN IF EXISTS source;
