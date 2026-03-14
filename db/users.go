package db

import (
	"context"
	"time"
)

type User struct {
	ID        int64
	Username  string
	TzOffset  int
	Level     string
	Active    bool
	CreatedAt time.Time
}

func (d *DB) CreateUser(id int64, username string) error {
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO users (id, username) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
		id, username,
	)
	return err
}

func (d *DB) GetUser(id int64) (*User, error) {
	var u User
	err := d.Pool.QueryRow(context.Background(),
		`SELECT id, username, tz_offset, level, active, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.TzOffset, &u.Level, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) GetActiveUsers() ([]User, error) {
	rows, err := d.Pool.Query(context.Background(),
		`SELECT id, username, tz_offset, level, active, created_at FROM users WHERE active = true`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.TzOffset, &u.Level, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
