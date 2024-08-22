-- +goose Up
CREATE TABLE skus (
    unique_identifier VARCHAR(50) PRIMARY KEY,
    prefix VARCHAR(50) NOT NULL,
    product_category VARCHAR(50) NOT NULL,
    manufacturer VARCHAR(100) NOT NULL,
    product_line VARCHAR(100) NOT NULL,
    attributes JSONB NOT NULL
);

ALTER TABLE items ADD COLUMN sku_id VARCHAR(50) REFERENCES skus(unique_identifier);

-- +goose Down
ALTER TABLE items DROP COLUMN sku_id;
DROP TABLE skus;
