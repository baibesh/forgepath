-- +goose Up

ALTER TABLE users ADD COLUMN IF NOT EXISTS target_language TEXT DEFAULT 'en';

-- +goose Down

ALTER TABLE users DROP COLUMN IF EXISTS target_language;
