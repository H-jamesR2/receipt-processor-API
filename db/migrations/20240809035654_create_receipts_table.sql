-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS receipts (
    id UUID PRIMARY KEY,
    retailer TEXT NOT NULL,
    purchase_date DATE NOT NULL,
    purchase_time TIME NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    points INTEGER NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS receipts;

