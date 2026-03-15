package db

import (
	"context"
	"fmt"
	"time"
)

type MediaResource struct {
	ID           int
	Title        string
	URL          string
	MediaType    string
	Level        string
	Topic        string
	Duration     string
	Active       bool
	Tags         string
	ViewCount    int
	HasSubtitles bool
	Description  string
	Language     string
}

type UserMedia struct {
	UserID       int64
	MediaID      int
	SentAt       time.Time
	TaskSent     bool
	TaskResponse string
	Completed    bool
}

func (d *DB) GetUnseenMedia(userID int64, level, language string) (*MediaResource, error) {
	var m MediaResource
	err := d.Pool.QueryRow(context.Background(),
		`SELECT mr.id, mr.title, mr.url, mr.media_type, mr.level,
		        COALESCE(mr.topic,''), COALESCE(mr.duration,''), mr.active,
		        COALESCE(mr.tags,''), COALESCE(mr.view_count,0),
		        COALESCE(mr.has_subtitles,false), COALESCE(mr.description,''),
		        COALESCE(mr.language,'en')
		 FROM media_resources mr
		 WHERE mr.active = true AND mr.level = $2 AND COALESCE(mr.language,'en') = $3
		   AND mr.id NOT IN (SELECT media_id FROM user_media WHERE user_id = $1)
		 ORDER BY mr.view_count DESC, RANDOM() LIMIT 1`, userID, level, language,
	).Scan(&m.ID, &m.Title, &m.URL, &m.MediaType, &m.Level, &m.Topic, &m.Duration, &m.Active,
		&m.Tags, &m.ViewCount, &m.HasSubtitles, &m.Description, &m.Language)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// SearchMedia finds media matching tags/topic keywords, prioritizing popular unseen content.
func (d *DB) SearchMedia(userID int64, level, language string, keywords []string) (*MediaResource, error) {
	query := `SELECT mr.id, mr.title, mr.url, mr.media_type, mr.level,
		        COALESCE(mr.topic,''), COALESCE(mr.duration,''), mr.active,
		        COALESCE(mr.tags,''), COALESCE(mr.view_count,0),
		        COALESCE(mr.has_subtitles,false), COALESCE(mr.description,''),
		        COALESCE(mr.language,'en')
		 FROM media_resources mr
		 WHERE mr.active = true AND mr.level = $1 AND COALESCE(mr.language,'en') = $3
		   AND mr.id NOT IN (SELECT media_id FROM user_media WHERE user_id = $2)
		   AND (`
	args := []interface{}{level, userID, language}

	for i, kw := range keywords {
		if i > 0 {
			query += " OR "
		}
		args = append(args, "%"+kw+"%")
		idx := len(args)
		query += fmt.Sprintf("mr.tags ILIKE $%d OR mr.topic ILIKE $%d OR mr.title ILIKE $%d OR mr.description ILIKE $%d", idx, idx, idx, idx)
	}
	query += `) ORDER BY mr.view_count DESC LIMIT 1`

	var m MediaResource
	err := d.Pool.QueryRow(context.Background(), query, args...).Scan(
		&m.ID, &m.Title, &m.URL, &m.MediaType, &m.Level, &m.Topic, &m.Duration, &m.Active,
		&m.Tags, &m.ViewCount, &m.HasSubtitles, &m.Description, &m.Language)
	if err != nil {
		// Fallback to any unseen media
		return d.GetUnseenMedia(userID, level, language)
	}
	return &m, nil
}

func (d *DB) MarkMediaSent(userID int64, mediaID int) error {
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO user_media (user_id, media_id, sent_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id, media_id) DO NOTHING`, userID, mediaID)
	return err
}

func (d *DB) MarkMediaTaskSent(userID int64, mediaID int) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE user_media SET task_sent = true WHERE user_id = $1 AND media_id = $2`, userID, mediaID)
	return err
}

func (d *DB) GetPendingMediaTask(userID int64) (*UserMedia, *MediaResource, error) {
	var um UserMedia
	var mr MediaResource
	err := d.Pool.QueryRow(context.Background(),
		`SELECT um.user_id, um.media_id, um.sent_at, um.task_sent, COALESCE(um.task_response,''), um.completed,
		        mr.id, mr.title, mr.url, mr.media_type, mr.level, COALESCE(mr.topic,''), COALESCE(mr.duration,''), mr.active
		 FROM user_media um
		 JOIN media_resources mr ON mr.id = um.media_id
		 WHERE um.user_id = $1 AND um.task_sent = true AND um.completed = false
		 ORDER BY um.sent_at DESC LIMIT 1`, userID,
	).Scan(&um.UserID, &um.MediaID, &um.SentAt, &um.TaskSent, &um.TaskResponse, &um.Completed,
		&mr.ID, &mr.Title, &mr.URL, &mr.MediaType, &mr.Level, &mr.Topic, &mr.Duration, &mr.Active)
	if err != nil {
		return nil, nil, err
	}
	return &um, &mr, nil
}

func (d *DB) SaveMediaTaskResponse(userID int64, mediaID int, response string) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE user_media SET task_response = $3, completed = true
		 WHERE user_id = $1 AND media_id = $2`, userID, mediaID, response)
	return err
}

func (d *DB) GetTodayUnsentMedia(userID int64, tzOffset int) (*UserMedia, error) {
	today := UserLocalDate(tzOffset)
	var um UserMedia
	err := d.Pool.QueryRow(context.Background(),
		`SELECT user_id, media_id, sent_at, task_sent, COALESCE(task_response,''), completed
		 FROM user_media
		 WHERE user_id = $1 AND DATE(sent_at) = $2::DATE AND task_sent = false
		 LIMIT 1`, userID, today,
	).Scan(&um.UserID, &um.MediaID, &um.SentAt, &um.TaskSent, &um.TaskResponse, &um.Completed)
	if err != nil {
		return nil, err
	}
	return &um, nil
}
