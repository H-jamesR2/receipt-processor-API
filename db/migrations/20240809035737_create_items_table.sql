-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    short_description TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price_paid NUMERIC(10, 2) NOT NULL,
    receipt_id UUID NOT NULL REFERENCES receipts(id)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS items;

