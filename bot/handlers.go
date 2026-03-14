package bot

import (
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/db"
)

func RegisterHandlers(b *tele.Bot, database *db.DB) {
	b.Handle("/start", func(c tele.Context) error {
		user := c.Sender()

		if err := database.CreateUser(user.ID, user.Username); err != nil {
			log.Printf("Error creating user %d: %v", user.ID, err)
			return c.Send("Something went wrong. Try again later.")
		}

		welcome := fmt.Sprintf(
			"Hey, %s! 👋\n\n"+
				"Welcome to *ForgePath* — your daily English learning companion.\n\n"+
				"Here's how it works:\n"+
				"📖 Morning — Word of the Day\n"+
				"✍️ Afternoon — Free Writing (5 min)\n"+
				"🎬 Evening — Media Recommendation\n"+
				"📊 Night — Daily Review\n\n"+
				"Stay consistent, build your streak, and watch your English grow.\n\n"+
				"Your current level: *%s*\n"+
				"Let's get started! 🚀",
			user.FirstName, "A2",
		)

		return c.Send(welcome, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})
}
