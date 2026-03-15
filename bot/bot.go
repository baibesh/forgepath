package bot

import (
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/config"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
)

func SetBotCommands(b *tele.Bot) {
	commands := []tele.Command{
		{Text: "today", Description: "Current tasks for today"},
		{Text: "word", Description: "Get word of the day"},
		{Text: "quiz", Description: "Start a review quiz"},
		{Text: "write", Description: "Start free writing (5 min)"},
		{Text: "stats", Description: "View your progress"},
		{Text: "words", Description: "Your learned words"},
		{Text: "skip", Description: "Skip today (max 2/week)"},
		{Text: "cancel", Description: "Cancel current task"},
		{Text: "settings", Description: "Change language/level/timezone"},
		{Text: "help", Description: "How the bot works"},
	}
	if err := b.SetCommands(commands); err != nil {
		log.Printf("SetCommands error: %v", err)
	}
}

func RegisterHandlers(b *tele.Bot, database *db.DB, cfg *config.Config) {
	openaiClient := ai.NewOpenAIClient(cfg.OpenAIKey)

	b.Handle("/start", func(c tele.Context) error {
		user := c.Sender()
		log.Printf("[user=%d] /start", user.ID)

		if err := database.CreateUser(user.ID, user.Username, user.FirstName); err != nil {
			log.Printf("[user=%d] create error: %v", user.ID, err)
			return c.Send("Something went wrong. Try again later.")
		}

		existing, err := database.GetUser(user.ID)

		if err == nil && !existing.Onboarded {
			state, _ := database.GetState(user.ID)
			if state.State != "idle" && state.State != "" {
				database.ClearState(user.ID)
			}
		}

		if err == nil && existing.Onboarded {
			return c.Send(fmt.Sprintf(
				"Welcome back, %s! %s\n\n%s %s | Level: *%s*\nTimezone: UTC+%d\n\nUse /help to see commands.",
				user.FirstName, content.LanguageFlag(existing.Language),
				content.LanguageFlag(existing.Language), content.LanguageName(existing.Language),
				existing.Level, existing.TzOffset,
			), &tele.SendOptions{
				ParseMode:   tele.ModeMarkdown,
				ReplyMarkup: &tele.ReplyMarkup{RemoveKeyboard: true},
			})
		}

		database.SetState(user.ID, "onboarding_language", map[string]string{})
		return c.Send(fmt.Sprintf(
			"Hey, %s! \U0001F44B\n\n"+
				"Welcome to *ForgePath* — your daily language learning companion.\n\n"+
				"\U0001F4D6 Morning — Word of the Day + Quiz\n\u270D\uFE0F Afternoon — Free Writing (5 min)\n"+
				"\U0001F3AC Evening — Media Recommendation\n\U0001F4CA Night — Daily Review\n\n"+
				"Let's set you up! Choose your language:",
			user.FirstName,
		), &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: LanguageSelectKeyboard()})
	})

	b.Handle("/today", func(c tele.Context) error { return handleToday(c, database) })
	b.Handle("/word", func(c tele.Context) error { return handleWord(c, database, openaiClient) })
	b.Handle("/quiz", func(c tele.Context) error { return handleQuiz(c, database, openaiClient) })
	b.Handle("/write", func(c tele.Context) error { return handleWrite(c, database) })
	b.Handle("/stats", func(c tele.Context) error { return handleStats(c, database) })
	b.Handle("/skip", func(c tele.Context) error { return handleSkip(c, database) })
	b.Handle("/words", func(c tele.Context) error { return handleWordsList(c, database) })

	b.Handle("/cancel", func(c tele.Context) error {
		userID := c.Sender().ID
		log.Printf("[user=%d] /cancel", userID)
		state, _ := database.GetState(userID)
		if state.State == "idle" {
			return c.Send("Nothing to cancel.")
		}
		database.ClearState(userID)
		return c.Send("\u2705 Cancelled. You can start a new task anytime.")
	})

	b.Handle("/settings", func(c tele.Context) error {
		return c.Send("\u2699\uFE0F *Settings*\n\nWhat would you like to change?",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard()})
	})

	b.Handle("/help", func(c tele.Context) error {
		return c.Send(
			"\U0001F4DA *ForgePath Help*\n\n"+
				"*Daily Schedule:*\n"+
				"07:30 — \U0001F4D6 Word of the Day + Quiz\n"+
				"12:00 — \u270D\uFE0F Free Writing (5 min)\n"+
				"18:00 — \U0001F3AC Media Recommendation\n"+
				"21:30 — \U0001F4CA Daily Review\n\n"+
				"*Commands:*\n"+
				"/today — see current task\n"+
				"/word — get word of the day\n"+
				"/quiz — start a quiz\n"+
				"/write — start free writing\n"+
				"/stats — view your progress\n"+
				"/skip — skip today (max 2/week)\n"+
				"/cancel — cancel current task\n"+
				"/words — your learned words\n"+
				"/settings — change timezone/level/language\n"+
				"/help — this message\n\n"+
				"*How it works:*\n"+
				"Each week focuses on one grammar tense.\n"+
				"Words come with constructions + collocations.\n"+
				"Quiz adapts to your recall level (SRS).\n"+
				"Writing gets AI feedback.\n"+
				"Stay consistent, build your streak! \U0001F525",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	RegisterCallbacks(b, database, openaiClient)

	b.Handle(tele.OnPollAnswer, func(c tele.Context) error {
		return handlePollAnswer(c, database)
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		return handleText(c, database, openaiClient)
	})

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		return handleVoice(c, b, database, openaiClient)
	})
}
