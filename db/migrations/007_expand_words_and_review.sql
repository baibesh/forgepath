-- Expand words table with richer content
ALTER TABLE words ADD COLUMN IF NOT EXISTS synonyms TEXT DEFAULT '';
ALTER TABLE words ADD COLUMN IF NOT EXISTS antonyms TEXT DEFAULT '';
ALTER TABLE words ADD COLUMN IF NOT EXISTS examples TEXT DEFAULT '';

-- User preference for words per day
ALTER TABLE users ADD COLUMN IF NOT EXISTS words_per_day INT DEFAULT 3;

-- Separate review session schedule
ALTER TABLE users ADD COLUMN IF NOT EXISTS review_session_hour INT DEFAULT 14;
ALTER TABLE users ADD COLUMN IF NOT EXISTS review_session_min INT DEFAULT 0;
