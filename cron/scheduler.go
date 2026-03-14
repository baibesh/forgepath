package cron

import (
	"log"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/config"
	"github.com/baibesh/forgepath/db"
	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
)

func StartScheduler(b *tele.Bot, database *db.DB, cfg *config.Config) *cron.Cron {
	openaiClient := ai.NewOpenAIClient(cfg.OpenAIKey)

	jobs := NewJobs(b, database, openaiClient)

	c := cron.New()
	c.AddFunc("0,30 * * * *", jobs.DispatchTasks)

	c.Start()
	log.Println("Scheduler started with 30-minute dispatch cycle")
	return c
}
