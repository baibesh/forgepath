package db

import (
	"context"
	"time"
)

type TodayStreak struct {
	WordDone    bool
	WritingDone bool
	ReviewDone  bool
}

type WeeklyStats struct {
	Days         int
	WordsDone    int
	WritingsDone int
	ReviewsDone  int
}

func (d *DB) MarkWordDone(userID int64) error {
	today := time.Now().UTC().Format("2006-01-02")
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO streaks (user_id, date, word_done) VALUES ($1, $2, true)
		 ON CONFLICT (user_id, date) DO UPDATE SET word_done = true`,
		userID, today)
	return err
}

func (d *DB) MarkWritingDone(userID int64) error {
	today := time.Now().UTC().Format("2006-01-02")
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO streaks (user_id, date, writing_done) VALUES ($1, $2, true)
		 ON CONFLICT (user_id, date) DO UPDATE SET writing_done = true`,
		userID, today)
	return err
}

func (d *DB) MarkReviewDone(userID int64) error {
	today := time.Now().UTC().Format("2006-01-02")
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO streaks (user_id, date, review_done) VALUES ($1, $2, true)
		 ON CONFLICT (user_id, date) DO UPDATE SET review_done = true`,
		userID, today)
	return err
}

func (d *DB) GetTodayStreak(userID int64) (*TodayStreak, error) {
	today := time.Now().UTC().Format("2006-01-02")
	var s TodayStreak
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(word_done, false), COALESCE(writing_done, false), COALESCE(review_done, false)
		 FROM streaks WHERE user_id = $1 AND date = $2`,
		userID, today).Scan(&s.WordDone, &s.WritingDone, &s.ReviewDone)
	if err != nil {
		return &TodayStreak{}, nil
	}
	return &s, nil
}

func (d *DB) GetCurrentStreak(userID int64) (int, error) {
	var streak int
	err := d.Pool.QueryRow(context.Background(), `
		WITH ordered AS (
			SELECT date, ROW_NUMBER() OVER (ORDER BY date DESC) as rn
			FROM streaks
			WHERE user_id = $1
			  AND (COALESCE(word_done, false) OR COALESCE(writing_done, false) OR COALESCE(review_done, false))
		)
		SELECT COUNT(*) FROM ordered
		WHERE date = CURRENT_DATE - (rn - 1)::int
	`, userID).Scan(&streak)
	if err != nil {
		return 0, nil
	}
	return streak, nil
}

func (d *DB) GetWeeklyStats(userID int64) (*WeeklyStats, error) {
	var s WeeklyStats
	err := d.Pool.QueryRow(context.Background(), `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE COALESCE(word_done, false)),
			COUNT(*) FILTER (WHERE COALESCE(writing_done, false)),
			COUNT(*) FILTER (WHERE COALESCE(review_done, false))
		FROM streaks
		WHERE user_id = $1 AND date >= CURRENT_DATE - INTERVAL '7 days'
	`, userID).Scan(&s.Days, &s.WordsDone, &s.WritingsDone, &s.ReviewsDone)
	if err != nil {
		return &WeeklyStats{}, nil
	}
	return &s, nil
}
