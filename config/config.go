package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken    string
	DatabaseURL string
	OpenAIKey string
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
	}

	if cfg.BotToken == "" {
		log.Fatal("BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.OpenAIKey == "" {
		log.Println("WARNING: OPENAI_API_KEY not set, AI features will be disabled")
	}
	return cfg
}
