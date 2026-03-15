-- +goose Up

ALTER TABLE media_resources ALTER COLUMN view_count TYPE BIGINT;

-- +goose Down

ALTER TABLE media_resources ALTER COLUMN view_count TYPE INT;
