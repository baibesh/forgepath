package cron

import (
	"log"

	"github.com/baibesh/forgepath/db"
	tele "gopkg.in/telebot.v3"
)

func StartScheduler(b *tele.Bot, database *db.DB) {
	// TODO: add cron jobs later
	log.Println("Scheduler initialized (no jobs yet)")
}
