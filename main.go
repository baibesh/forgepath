package main

import (
	"log"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/bot"
	"github.com/baibesh/forgepath/config"
	"github.com/baibesh/forgepath/cron"
	"github.com/baibesh/forgepath/db"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	database.Migrate()

	b, err := tele.NewBot(tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	bot.RegisterHandlers(b, database, cfg)
	cron.StartScheduler(b, database, cfg)

	log.Println("ForgePath bot started")
	b.Start()
}
