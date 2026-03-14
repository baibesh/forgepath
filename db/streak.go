package db

import (
	"context"
	"time"
)

func (d *DB) UpdateStreak(userID int64) error {
	today := time.Now().UTC().Format("2006-01-02")
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO streaks (user_id, date, completed) VALUES ($1, $2, true)
		 ON CONFLICT (user_id, date) DO UPDATE SET completed = true`,
		userID, today,
	)
	return err
}

func (d *DB) GetCurrentStreak(userID int64) (int, error) {
	var streak int
	err := d.Pool.QueryRow(context.Background(), `
		WITH dates AS (
			SELECT date FROM streaks
			WHERE user_id = $1 AND completed = true
			ORDER BY date DESC
		)
		SELECT COUNT(*) FROM dates
		WHERE date >= CURRENT_DATE - (
			SELECT COUNT(*) - 1 FROM dates
			WHERE date >= CURRENT_DATE - INTERVAL '1 day' * (
				ROW_NUMBER() OVER (ORDER BY date DESC) - 1
			)
		)
	`, userID).Scan(&streak)

	if err != nil {
		// Simpler fallback query
		err = d.Pool.QueryRow(context.Background(), `
			SELECT COUNT(*) FROM streaks
			WHERE user_id = $1 AND completed = true
			AND date >= CURRENT_DATE - INTERVAL '30 days'
		`, userID).Scan(&streak)
	}
	return streak, err
}
