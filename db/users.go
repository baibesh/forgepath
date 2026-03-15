package db

import (
	"context"
)

func (d *DB) CreateUser(id int64, username, firstName string) error {
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO users (id, username, first_name) VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET username = $2, first_name = $3`,
		id, username, firstName,
	)
	return err
}

func (d *DB) GetUser(id int64) (*User, error) {
	var u User
	err := d.Pool.QueryRow(context.Background(),
		`SELECT id, username, COALESCE(first_name, ''), tz_offset, level,
		        COALESCE(language, 'en'), active, COALESCE(onboarded, false),
		        COALESCE(skip_count, 0), COALESCE(current_grammar_week, 1),
		        COALESCE(word_hour, 7), COALESCE(word_min, 30),
		        COALESCE(writing_hour, 12), COALESCE(writing_min, 0),
		        COALESCE(media_hour, 18), COALESCE(media_min, 0),
		        COALESCE(review_hour, 21), COALESCE(review_min, 30),
		        created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.FirstName, &u.TzOffset, &u.Level,
		&u.Language, &u.Active, &u.Onboarded,
		&u.SkipCount, &u.CurrentGrammarWeek,
		&u.Schedule.WordHour, &u.Schedule.WordMin,
		&u.Schedule.WritingHour, &u.Schedule.WritingMin,
		&u.Schedule.MediaHour, &u.Schedule.MediaMin,
		&u.Schedule.ReviewHour, &u.Schedule.ReviewMin,
		&u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) GetActiveUsers() ([]User, error) {
	rows, err := d.Pool.Query(context.Background(),
		`SELECT id, username, COALESCE(first_name, ''), tz_offset, level,
		        COALESCE(language, 'en'), active, COALESCE(onboarded, false),
		        COALESCE(skip_count, 0), COALESCE(current_grammar_week, 1),
		        COALESCE(word_hour, 7), COALESCE(word_min, 30),
		        COALESCE(writing_hour, 12), COALESCE(writing_min, 0),
		        COALESCE(media_hour, 18), COALESCE(media_min, 0),
		        COALESCE(review_hour, 21), COALESCE(review_min, 30),
		        created_at
		 FROM users WHERE active = true AND COALESCE(onboarded, false) = true`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.TzOffset, &u.Level,
			&u.Language, &u.Active, &u.Onboarded,
			&u.SkipCount, &u.CurrentGrammarWeek,
			&u.Schedule.WordHour, &u.Schedule.WordMin,
			&u.Schedule.WritingHour, &u.Schedule.WritingMin,
			&u.Schedule.MediaHour, &u.Schedule.MediaMin,
			&u.Schedule.ReviewHour, &u.Schedule.ReviewMin,
			&u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (d *DB) UpdateUserTimezone(userID int64, offset int) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET tz_offset = $2 WHERE id = $1`, userID, offset)
	return err
}

func (d *DB) UpdateUserLevel(userID int64, level string) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET level = $2 WHERE id = $1`, userID, level)
	return err
}

func (d *DB) UpdateUserLanguage(userID int64, language string) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET language = $2 WHERE id = $1`, userID, language)
	return err
}

func (d *DB) IncrementSkipCount(userID int64) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET skip_count = skip_count + 1 WHERE id = $1`, userID)
	return err
}

func (d *DB) ResetWeeklySkips(userID int64) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET skip_count = 0 WHERE id = $1`, userID)
	return err
}

func (d *DB) SetOnboarded(userID int64) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET onboarded = TRUE WHERE id = $1`, userID)
	return err
}

func (d *DB) AdvanceGrammarWeek(userID int64) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE users SET current_grammar_week = (current_grammar_week % 7) + 1 WHERE id = $1`, userID)
	return err
}
