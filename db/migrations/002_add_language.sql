-- +goose Up

-- Users: add language column
ALTER TABLE users ADD COLUMN IF NOT EXISTS language TEXT DEFAULT 'en';

-- Words: add language column + update unique index
ALTER TABLE words ADD COLUMN IF NOT EXISTS language TEXT DEFAULT 'en';
DROP INDEX IF EXISTS words_word_unique;
CREATE UNIQUE INDEX IF NOT EXISTS words_word_lang_unique ON words (word, language);

-- Grammar weeks: add language + unique index
ALTER TABLE grammar_weeks ADD COLUMN IF NOT EXISTS language TEXT DEFAULT 'en';
ALTER TABLE grammar_weeks DROP CONSTRAINT IF EXISTS grammar_weeks_pkey;
CREATE UNIQUE INDEX IF NOT EXISTS grammar_weeks_week_lang_unique ON grammar_weeks (week_num, language);

-- Media resources: add language column
ALTER TABLE media_resources ADD COLUMN IF NOT EXISTS language TEXT DEFAULT 'en';

-- Backfill: existing data is English
UPDATE users SET language = 'en' WHERE language IS NULL;
UPDATE words SET language = 'en' WHERE language IS NULL;
UPDATE media_resources SET language = 'en' WHERE language IS NULL;

-- +goose Down

ALTER TABLE users DROP COLUMN IF EXISTS language;
ALTER TABLE words DROP COLUMN IF EXISTS language;
DROP INDEX IF EXISTS words_word_lang_unique;
CREATE UNIQUE INDEX IF NOT EXISTS words_word_unique ON words (word);
ALTER TABLE grammar_weeks DROP COLUMN IF EXISTS language;
ALTER TABLE media_resources DROP COLUMN IF EXISTS language;
