-- +goose Up

ALTER TABLE writings ADD COLUMN IF NOT EXISTS writing_type VARCHAR(20) DEFAULT 'free';

-- +goose Down

ALTER TABLE writings DROP COLUMN IF EXISTS writing_type;
