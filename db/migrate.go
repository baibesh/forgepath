package db

import (
	"database/sql"
	"embed"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (d *DB) Migrate(databaseURL string) {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Migration: cannot open DB: %v", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Migration: cannot set dialect: %v", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("Migration: %v", err)
	}

	log.Println("Goose migrations applied")
}
