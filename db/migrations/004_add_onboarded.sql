-- +goose Up

ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarded BOOL DEFAULT FALSE;

-- Mark existing users with language+level+timezone as onboarded
UPDATE users SET onboarded = TRUE WHERE language IS NOT NULL AND level IS NOT NULL AND tz_offset IS NOT NULL;

-- +goose Down

ALTER TABLE users DROP COLUMN IF EXISTS onboarded;
