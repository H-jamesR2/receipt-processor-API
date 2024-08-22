-- +goose Up

CREATE INDEX IF NOT EXISTS idx_receipt_id ON items(receipt_id);

-- +goose Down

DROP INDEX IF EXISTS idx_receipt_id;


