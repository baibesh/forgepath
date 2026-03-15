package db

import (
	"context"
)

func (d *DB) SaveWriting(userID int64, topic, grammarFocus, text string, wordCount int) (int, error) {
	var id int
	err := d.Pool.QueryRow(context.Background(),
		`INSERT INTO writings (user_id, topic, grammar_focus, text, word_count)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, topic, grammarFocus, text, wordCount).Scan(&id)
	return id, err
}

func (d *DB) UpdateWritingFeedback(writingID int, feedback string) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE writings SET feedback = $2 WHERE id = $1`, writingID, feedback)
	return err
}

func (d *DB) GetUserWritingCount(userID int64) (int, error) {
	var count int
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM writings WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (d *DB) GetRecentWritings(userID int64, limit int) ([]Writing, error) {
	rows, err := d.Pool.Query(context.Background(),
		`SELECT id, user_id, COALESCE(topic,''), COALESCE(grammar_focus,''), COALESCE(text,''),
		        COALESCE(feedback,''), COALESCE(word_count,0), created_at
		 FROM writings WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var writings []Writing
	for rows.Next() {
		var w Writing
		if err := rows.Scan(&w.ID, &w.UserID, &w.Topic, &w.GrammarFocus, &w.Text,
			&w.Feedback, &w.WordCount, &w.CreatedAt); err != nil {
			return nil, err
		}
		writings = append(writings, w)
	}
	return writings, nil
}
