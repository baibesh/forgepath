package db

import (
	"context"
	"encoding/json"
	"time"
)

func (d *DB) GetState(userID int64) (*UserState, error) {
	var state string
	var ctxJSON []byte

	err := d.Pool.QueryRow(context.Background(),
		`SELECT state, context FROM user_state WHERE user_id = $1`, userID,
	).Scan(&state, &ctxJSON)

	if err != nil {
		return &UserState{UserID: userID, State: "idle", Context: map[string]string{}}, nil
	}

	var ctxMap map[string]string
	if err := json.Unmarshal(ctxJSON, &ctxMap); err != nil {
		ctxMap = map[string]string{}
	}

	return &UserState{UserID: userID, State: state, Context: ctxMap}, nil
}

func (d *DB) SetState(userID int64, state string, ctx map[string]string) error {
	ctxJSON, err := json.Marshal(ctx)
	if err != nil {
		ctxJSON = []byte("{}")
	}

	_, err = d.Pool.Exec(context.Background(),
		`INSERT INTO user_state (user_id, state, context, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET state = $2, context = $3, updated_at = NOW()`,
		userID, state, ctxJSON)
	return err
}

func (d *DB) ClearState(userID int64) error {
	return d.SetState(userID, "idle", map[string]string{})
}

func (d *DB) GetStaleStateAge(userID int64) (time.Duration, bool) {
	var updatedAt time.Time
	err := d.Pool.QueryRow(context.Background(),
		`SELECT updated_at FROM user_state WHERE user_id = $1 AND state != 'idle'`, userID,
	).Scan(&updatedAt)
	if err != nil {
		return 0, false
	}
	return time.Since(updatedAt), true
}

func (d *DB) GetMediaTitle(userID int64, mediaID int) string {
	var title string
	err := d.Pool.QueryRow(context.Background(),
		`SELECT mr.title FROM user_media um JOIN media_resources mr ON mr.id = um.media_id
		 WHERE um.user_id = $1 AND um.media_id = $2`, userID, mediaID).Scan(&title)
	if err != nil {
		return "the video"
	}
	return title
}
