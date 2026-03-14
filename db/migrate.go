package db

import (
	"context"
	"log"
)

func (d *DB) Migrate() {
	ctx := context.Background()

	queries := []string{
		// Base tables (create if first deploy)
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY,
			username TEXT,
			tz_offset INT DEFAULT 5,
			level TEXT DEFAULT 'A2',
			active BOOL DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS streaks (
			user_id BIGINT REFERENCES users(id),
			date DATE,
			completed BOOL DEFAULT FALSE,
			PRIMARY KEY (user_id, date)
		)`,

		`CREATE TABLE IF NOT EXISTS words (
			id SERIAL PRIMARY KEY,
			word TEXT NOT NULL UNIQUE,
			definition TEXT,
			example TEXT,
			level TEXT DEFAULT 'A2',
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS user_words (
			user_id BIGINT REFERENCES users(id),
			word_id INT REFERENCES words(id),
			seen_at TIMESTAMP DEFAULT NOW(),
			next_review TIMESTAMP,
			score INT DEFAULT 0,
			PRIMARY KEY (user_id, word_id)
		)`,

		// Ensure unique constraint on words for ON CONFLICT
		`CREATE UNIQUE INDEX IF NOT EXISTS words_word_unique ON words (word)`,

		// Extend users table
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS skip_count INT DEFAULT 0`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS current_grammar_week INT DEFAULT 1`,

		// Extend streaks table
		`ALTER TABLE streaks ADD COLUMN IF NOT EXISTS word_done BOOL DEFAULT FALSE`,
		`ALTER TABLE streaks ADD COLUMN IF NOT EXISTS writing_done BOOL DEFAULT FALSE`,
		`ALTER TABLE streaks ADD COLUMN IF NOT EXISTS review_done BOOL DEFAULT FALSE`,

		// Extend words table
		`ALTER TABLE words ADD COLUMN IF NOT EXISTS collocations TEXT DEFAULT ''`,
		`ALTER TABLE words ADD COLUMN IF NOT EXISTS construction TEXT DEFAULT ''`,

		// Extend user_words table
		`ALTER TABLE user_words ADD COLUMN IF NOT EXISTS interval_days INT DEFAULT 1`,
		`ALTER TABLE user_words ADD COLUMN IF NOT EXISTS ease_factor REAL DEFAULT 2.5`,
		`ALTER TABLE user_words ADD COLUMN IF NOT EXISTS repetitions INT DEFAULT 0`,

		// FSM state table
		`CREATE TABLE IF NOT EXISTS user_state (
			user_id BIGINT PRIMARY KEY REFERENCES users(id),
			state TEXT DEFAULT 'idle',
			context JSONB DEFAULT '{}',
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Writings table
		`CREATE TABLE IF NOT EXISTS writings (
			id SERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			topic TEXT,
			grammar_focus TEXT,
			text TEXT,
			feedback TEXT,
			word_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Curated media resources
		`CREATE TABLE IF NOT EXISTS media_resources (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			media_type TEXT DEFAULT 'video',
			level TEXT DEFAULT 'A2',
			topic TEXT DEFAULT '',
			duration TEXT DEFAULT '',
			active BOOL DEFAULT TRUE
		)`,

		// Ensure unique on url for pre-existing media_resources
		`CREATE UNIQUE INDEX IF NOT EXISTS media_resources_url_unique ON media_resources (url)`,

		// Extend media_resources for smart search
		`ALTER TABLE media_resources ADD COLUMN IF NOT EXISTS tags TEXT DEFAULT ''`,
		`ALTER TABLE media_resources ADD COLUMN IF NOT EXISTS view_count INT DEFAULT 0`,
		`ALTER TABLE media_resources ADD COLUMN IF NOT EXISTS has_subtitles BOOL DEFAULT FALSE`,
		`ALTER TABLE media_resources ADD COLUMN IF NOT EXISTS description TEXT DEFAULT ''`,

		// User media tracking
		`CREATE TABLE IF NOT EXISTS user_media (
			user_id BIGINT REFERENCES users(id),
			media_id INT REFERENCES media_resources(id),
			sent_at TIMESTAMP DEFAULT NOW(),
			task_sent BOOL DEFAULT FALSE,
			task_response TEXT,
			completed BOOL DEFAULT FALSE,
			PRIMARY KEY (user_id, media_id)
		)`,

		// Grammar weeks
		`CREATE TABLE IF NOT EXISTS grammar_weeks (
			week_num INT PRIMARY KEY,
			family TEXT NOT NULL,
			focus TEXT NOT NULL,
			tense_name TEXT NOT NULL,
			anchor TEXT NOT NULL,
			markers TEXT NOT NULL,
			formula TEXT NOT NULL,
			example TEXT NOT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := d.Pool.Exec(ctx, q); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}

	d.seedGrammarWeeks(ctx)
	d.seedWords(ctx)
	d.seedMedia(ctx)

	log.Println("Database migration completed")
}

func (d *DB) seedGrammarWeeks(ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Simple", "Past Simple", "Past Simple",
			"🚪 Закрытая дверь — действие завершено и всё",
			"yesterday, last week, ago, in 2020",
			"S + V2 (ed / irregular)",
			"I watched a movie yesterday."},
		{2, "Simple", "Present Simple", "Present Simple",
			"🔄 Карусель — повторяется снова и снова",
			"always, usually, every day, sometimes",
			"S + V1 (he/she +s)",
			"I usually wake up at 7."},
		{3, "Simple", "Future Simple", "Future Simple",
			"🔮 Хрустальный шар — решение прямо сейчас",
			"tomorrow, next week, I think, probably",
			"S + will + V1",
			"I will call you tomorrow."},
		{4, "Continuous", "Present Continuous", "Present Continuous",
			"📸 Фотография — прямо сейчас в процессе",
			"now, right now, at the moment, look!",
			"S + am/is/are + Ving",
			"I am reading a book right now."},
		{5, "Continuous", "Past Continuous", "Past Continuous",
			"🎬 Кадр из фильма — фон, процесс в прошлом",
			"while, when, at that moment, all day yesterday",
			"S + was/were + Ving",
			"I was cooking when you called."},
		{6, "Perfect", "Present Perfect", "Present Perfect",
			"🌉 Мост — из прошлого в настоящее, результат важен",
			"already, yet, just, ever, never, since, for",
			"S + have/has + V3",
			"I have already finished my homework."},
		{7, "Perfect", "Past Perfect", "Past Perfect",
			"⏪ Перемотка — действие ДО другого прошлого",
			"before, after, by the time, already (past context)",
			"S + had + V3",
			"I had eaten before she arrived."},
	}

	for _, w := range weeks {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (week_num) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
		if err != nil {
			log.Printf("Seed grammar week %d warning: %v", w.num, err)
		}
	}
}

func (d *DB) seedWords(ctx context.Context) {
	words := []struct {
		word, definition, example, level, collocations, construction string
	}{
		{"figure out", "понять, разобраться", "I finally figured out how to use this app.", "A2",
			"figure out a problem, figure out the answer, figure out how", "figure out + how/what/why"},
		{"give up", "сдаться, бросить", "Don't give up! You can do it.", "A2",
			"give up hope, give up trying, give up smoking", "give up + Ving / noun"},
		{"look forward to", "ждать с нетерпением", "I'm looking forward to the weekend.", "A2",
			"look forward to meeting, look forward to seeing", "look forward to + Ving / noun"},
		{"turn out", "оказаться", "It turned out to be a great movie.", "A2",
			"turn out to be, turn out well, turn out that", "turn out + to be / that"},
		{"come up with", "придумать", "She came up with a brilliant idea.", "A2",
			"come up with an idea, come up with a plan, come up with a solution", "come up with + noun"},
		{"run out of", "закончиться, иссякнуть", "We ran out of milk.", "A2",
			"run out of time, run out of money, run out of patience", "run out of + noun"},
		{"get along with", "ладить с", "I get along with my coworkers.", "A2",
			"get along with people, get along well, get along with someone", "get along with + person"},
		{"pick up", "подобрать, выучить", "I picked up some new words from the movie.", "A2",
			"pick up a language, pick up the phone, pick up a skill", "pick up + noun"},
		{"put off", "откладывать", "Stop putting off your homework!", "A2",
			"put off a meeting, put off doing something", "put off + Ving / noun"},
		{"carry on", "продолжать", "Carry on with your work.", "A2",
			"carry on working, carry on with something", "carry on + Ving / with + noun"},
		{"break down", "сломаться, расплакаться", "My car broke down on the highway.", "A2",
			"break down in tears, break down a problem", "break down + (no object) / break down + noun"},
		{"set up", "организовать, настроить", "Let's set up a meeting for Monday.", "A2",
			"set up a business, set up a meeting, set up an account", "set up + noun"},
		{"find out", "узнать, выяснить", "I found out the truth yesterday.", "A2",
			"find out the truth, find out about, find out that", "find out + about / that / noun"},
		{"go through", "пройти через, пережить", "She went through a difficult time.", "A2",
			"go through a phase, go through changes, go through a process", "go through + noun"},
		{"bring up", "поднять тему, воспитать", "Don't bring up that topic again.", "A2",
			"bring up a topic, bring up children, bring up an issue", "bring up + noun"},
		{"deal with", "справляться с", "I need to deal with this problem.", "A2",
			"deal with a problem, deal with stress, deal with people", "deal with + noun"},
		{"point out", "указать на", "She pointed out my mistake.", "A2",
			"point out a mistake, point out that, point out a problem", "point out + noun / that"},
		{"end up", "в итоге оказаться", "We ended up staying home.", "A2",
			"end up doing, end up in a place, end up being", "end up + Ving / in + place"},
		{"take off", "взлететь, снять", "The plane took off on time.", "A2",
			"take off clothes, take off from work, plane takes off", "take off + noun / (no object)"},
		{"show up", "появиться, прийти", "He didn't show up to the meeting.", "A2",
			"show up late, show up on time, show up unexpectedly", "show up + (no object) / at + place"},
		{"meanwhile", "тем временем", "I cooked dinner. Meanwhile, she set the table.", "A2",
			"meanwhile in, meanwhile back", "meanwhile + clause"},
		{"although", "хотя", "Although it was raining, we went for a walk.", "A2",
			"although it seems, although I know", "although + clause"},
		{"actually", "на самом деле", "I actually enjoyed the movie.", "A2",
			"actually quite, actually think, actually happened", "actually + verb / adjective"},
		{"definitely", "определённо", "I will definitely come to your party.", "A2",
			"definitely agree, definitely need, definitely want", "definitely + verb"},
		{"probably", "вероятно", "She will probably be late.", "A2",
			"probably not, probably the best, probably should", "probably + verb / adjective"},
		{"recently", "недавно", "I recently started learning English.", "A2",
			"recently discovered, recently started, recently moved", "recently + Past Simple / Present Perfect"},
		{"especially", "особенно", "I love fruits, especially mangoes.", "A2",
			"especially when, especially important, especially for", "especially + noun / when / adjective"},
		{"instead", "вместо этого", "I didn't go out. Instead, I stayed home.", "A2",
			"instead of doing, instead of that", "instead + clause / instead of + Ving"},
		{"however", "однако", "The test was hard. However, I passed it.", "A2",
			"however much, however difficult", "however + clause"},
		{"manage to", "суметь, удаться", "I managed to finish the project on time.", "A2",
			"manage to do, manage to find, manage to get", "manage to + V1"},
		{"be supposed to", "должен (по плану/ожиданию)", "You are supposed to be here at 9.", "A2",
			"supposed to do, supposed to be, supposed to know", "be supposed to + V1"},
		{"used to", "раньше (привычка в прошлом)", "I used to play football every day.", "A2",
			"used to live, used to be, used to think", "used to + V1"},
		{"be about to", "вот-вот, собираться", "The movie is about to start.", "A2",
			"about to leave, about to start, about to happen", "be about to + V1"},
		{"afford", "позволить себе", "I can't afford a new phone.", "A2",
			"afford to buy, can't afford, afford the time", "can/can't afford + to V1 / noun"},
		{"improve", "улучшить", "I want to improve my English.", "A2",
			"improve skills, improve performance, improve quality", "improve + noun"},
		{"appreciate", "ценить", "I really appreciate your help.", "A2",
			"appreciate help, appreciate the effort, appreciate it", "appreciate + noun / Ving"},
		{"avoid", "избегать", "Try to avoid making the same mistake.", "A2",
			"avoid doing, avoid mistakes, avoid problems", "avoid + Ving / noun"},
		{"recommend", "рекомендовать", "I recommend watching this movie.", "A2",
			"recommend doing, recommend a book, highly recommend", "recommend + Ving / noun"},
		{"require", "требовать", "This job requires experience.", "A2",
			"require experience, require attention, require effort", "require + noun / Ving"},
		{"consider", "рассмотреть, считать", "I consider him a good friend.", "A2",
			"consider doing, consider the options, consider important", "consider + Ving / noun + adjective"},
		{"suggest", "предложить", "I suggest taking a break.", "A2",
			"suggest doing, suggest an idea, suggest that", "suggest + Ving / that + clause"},
		{"depend on", "зависеть от", "It depends on the weather.", "A2",
			"depend on someone, depend on the situation", "depend on + noun"},
		{"belong to", "принадлежать", "This book belongs to me.", "A2",
			"belong to someone, belong to a group", "belong to + noun"},
		{"consist of", "состоять из", "The team consists of five people.", "A2",
			"consist of parts, consist of members", "consist of + noun"},
		{"respond to", "отвечать на", "She didn't respond to my message.", "A2",
			"respond to a question, respond to a message, respond quickly", "respond to + noun"},
		{"ordinary", "обычный", "It was just an ordinary day.", "A2",
			"ordinary people, ordinary life, ordinary day", "ordinary + noun"},
		{"essential", "необходимый", "Sleep is essential for health.", "A2",
			"essential for, essential part, essential information", "essential + for / noun"},
		{"obvious", "очевидный", "The answer was obvious.", "A2",
			"obvious reason, obvious choice, obviously wrong", "obvious + noun / that"},
		{"entire", "весь, целый", "I spent the entire day reading.", "A2",
			"entire day, entire life, entire team", "entire + noun"},
		{"convenient", "удобный", "This time is convenient for me.", "A2",
			"convenient time, convenient location, convenient for", "convenient + for + noun"},
	}

	for _, w := range words {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO words (word, definition, example, level, collocations, construction)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (word) DO NOTHING`,
			w.word, w.definition, w.example, w.level, w.collocations, w.construction)
		if err != nil {
			log.Printf("Seed word '%s' warning: %v", w.word, err)
		}
	}
}

func (d *DB) seedMedia(ctx context.Context) {
	media := []struct {
		title, url, mediaType, level, topic, duration string
	}{
		{"Morning Routine — Easy English", "https://www.youtube.com/watch?v=GGp25fn25Cs", "video", "A2", "daily life", "5 min"},
		{"At the Restaurant — Easy English", "https://www.youtube.com/watch?v=BGHxLfRGk3I", "video", "A2", "food", "6 min"},
		{"My Daily Routine — Bob the Canadian", "https://www.youtube.com/watch?v=MIuoBGFMEAo", "video", "A2", "daily life", "8 min"},
		{"Shopping Vocabulary — English with Lucy", "https://www.youtube.com/watch?v=h4X-Oyl91sE", "video", "A2", "shopping", "10 min"},
		{"Travel English — Easy Conversations", "https://www.youtube.com/watch?v=tfJRwNo2SJI", "video", "A2", "travel", "7 min"},
		{"English Listening Practice — Slow Easy", "https://www.youtube.com/watch?v=MqR0GbVfIqk", "video", "A2", "listening", "10 min"},
		{"Past Simple Stories — Easy English", "https://www.youtube.com/watch?v=aBq4MJuxI2c", "video", "A2", "grammar", "6 min"},
		{"Present Perfect Explained — BBC Learning", "https://www.youtube.com/watch?v=WjpCNe_JwBs", "video", "A2", "grammar", "5 min"},
		{"Everyday Phrasal Verbs — Rachel's English", "https://www.youtube.com/watch?v=wLgS3t_EXak", "video", "A2", "vocabulary", "8 min"},
		{"How to Talk About Your Weekend", "https://www.youtube.com/watch?v=wBHLJGHxCgQ", "video", "A2", "speaking", "6 min"},
		{"English Weather Vocabulary", "https://www.youtube.com/watch?v=N4TBw9Y1hS0", "video", "A2", "vocabulary", "5 min"},
		{"Telling Time in English", "https://www.youtube.com/watch?v=IBBQXBhSNUs", "video", "A2", "basics", "7 min"},
		{"Describing People in English", "https://www.youtube.com/watch?v=A5fNZnpXBzQ", "video", "A2", "speaking", "6 min"},
		{"Job Interview English — Easy Level", "https://www.youtube.com/watch?v=naIkpQ_cIt0", "video", "A2", "work", "8 min"},
		{"English at the Doctor's Office", "https://www.youtube.com/watch?v=xdDbp6RnUfU", "video", "A2", "health", "6 min"},
		{"Cooking Vocabulary in English", "https://www.youtube.com/watch?v=ZjXwOIFbyoo", "video", "A2", "food", "7 min"},
		{"Giving Directions in English", "https://www.youtube.com/watch?v=bBMmEL5Fzno", "video", "A2", "travel", "5 min"},
		{"English Phone Conversations", "https://www.youtube.com/watch?v=VGMWQeEQPKQ", "video", "A2", "speaking", "6 min"},
		{"Feelings and Emotions Vocabulary", "https://www.youtube.com/watch?v=dEL3xPNpJJE", "video", "A2", "vocabulary", "5 min"},
		{"Weekend Plans — Easy Listening", "https://www.youtube.com/watch?v=CrHMjC99jas", "video", "A2", "listening", "8 min"},
	}

	for _, m := range media {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO media_resources (title, url, media_type, level, topic, duration)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (url) DO NOTHING`,
			m.title, m.url, m.mediaType, m.level, m.topic, m.duration)
		if err != nil {
			log.Printf("Seed media '%s' warning: %v", m.title, err)
		}
	}
}
