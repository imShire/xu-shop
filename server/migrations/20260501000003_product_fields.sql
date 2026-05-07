-- +goose Up
ALTER TABLE product
  ADD COLUMN unit                varchar(16) NOT NULL DEFAULT '件',
  ADD COLUMN is_virtual          boolean NOT NULL DEFAULT false,
  ADD COLUMN freight_template_id bigint REFERENCES freight_template(id) ON DELETE SET NULL,
  ADD COLUMN virtual_sales       int NOT NULL DEFAULT 0;

CREATE INDEX idx_product_freight ON product(freight_template_id) WHERE freight_template_id IS NOT NULL;

ALTER TABLE sku
  ADD COLUMN barcode varchar(64);

-- +goose Down
ALTER TABLE product
  DROP COLUMN unit,
  DROP COLUMN is_virtual,
  DROP COLUMN freight_template_id,
  DROP COLUMN virtual_sales;

DROP INDEX IF EXISTS idx_product_freight;

ALTER TABLE sku
  DROP COLUMN barcode;
