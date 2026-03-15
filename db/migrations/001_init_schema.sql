-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username TEXT,
    first_name TEXT DEFAULT '',
    tz_offset INT DEFAULT 5,
    level TEXT DEFAULT 'A2',
    active BOOL DEFAULT TRUE,
    skip_count INT DEFAULT 0,
    current_grammar_week INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS streaks (
    user_id BIGINT REFERENCES users(id),
    date DATE,
    completed BOOL DEFAULT FALSE,
    word_done BOOL DEFAULT FALSE,
    writing_done BOOL DEFAULT FALSE,
    review_done BOOL DEFAULT FALSE,
    PRIMARY KEY (user_id, date)
);

CREATE TABLE IF NOT EXISTS words (
    id SERIAL PRIMARY KEY,
    word TEXT NOT NULL,
    definition TEXT,
    example TEXT,
    collocations TEXT DEFAULT '',
    construction TEXT DEFAULT '',
    level TEXT DEFAULT 'A2',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS words_word_unique ON words (word);

CREATE TABLE IF NOT EXISTS user_words (
    user_id BIGINT REFERENCES users(id),
    word_id INT REFERENCES words(id),
    seen_at TIMESTAMP DEFAULT NOW(),
    next_review TIMESTAMP,
    interval_days INT DEFAULT 1,
    ease_factor REAL DEFAULT 2.5,
    repetitions INT DEFAULT 0,
    score INT DEFAULT 0,
    PRIMARY KEY (user_id, word_id)
);

CREATE TABLE IF NOT EXISTS user_state (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    state TEXT DEFAULT 'idle',
    context JSONB DEFAULT '{}',
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS writings (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    topic TEXT,
    grammar_focus TEXT,
    text TEXT,
    feedback TEXT,
    word_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS media_resources (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    media_type TEXT DEFAULT 'video',
    level TEXT DEFAULT 'A2',
    topic TEXT DEFAULT '',
    duration TEXT DEFAULT '',
    tags TEXT DEFAULT '',
    view_count INT DEFAULT 0,
    has_subtitles BOOL DEFAULT FALSE,
    description TEXT DEFAULT '',
    active BOOL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS user_media (
    user_id BIGINT REFERENCES users(id),
    media_id INT REFERENCES media_resources(id),
    sent_at TIMESTAMP DEFAULT NOW(),
    task_sent BOOL DEFAULT FALSE,
    task_response TEXT,
    completed BOOL DEFAULT FALSE,
    PRIMARY KEY (user_id, media_id)
);

CREATE TABLE IF NOT EXISTS grammar_weeks (
    week_num INT,
    family TEXT NOT NULL,
    focus TEXT NOT NULL,
    tense_name TEXT NOT NULL,
    anchor TEXT NOT NULL,
    markers TEXT NOT NULL,
    formula TEXT NOT NULL,
    example TEXT NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS grammar_weeks;
DROP TABLE IF EXISTS user_media;
DROP TABLE IF EXISTS media_resources;
DROP TABLE IF EXISTS writings;
DROP TABLE IF EXISTS user_state;
DROP TABLE IF EXISTS user_words;
DROP TABLE IF EXISTS words;
DROP TABLE IF EXISTS streaks;
DROP TABLE IF EXISTS users;
