package bot

import (
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/config"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func SetBotCommands(b *tele.Bot) {
	commands := []tele.Command{
		{Text: "today", Description: "What's left for today"},
		{Text: "word", Description: "Learn a new word"},
		{Text: "quiz", Description: "Practice words"},
		{Text: "write", Description: "Write something (5 min)"},
		{Text: "stats", Description: "Your progress"},
		{Text: "words", Description: "Words you've learned"},
		{Text: "skip", Description: "Take a day off"},
		{Text: "cancel", Description: "Stop current task"},
		{Text: "settings", Description: "Change settings"},
		{Text: "help", Description: "How this works"},
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
				"Hey, %s! %s\n\n"+
					"You're learning %s, level *%s*\n\n"+
					"Type /today to see what's next!",
				user.FirstName, content.LanguageFlag(existing.Language),
				content.LanguageName(existing.Language), existing.Level,
			), &tele.SendOptions{
				ParseMode:   tele.ModeMarkdown,
				ReplyMarkup: &tele.ReplyMarkup{RemoveKeyboard: true},
			})
		}

		database.SetState(user.ID, "onboarding_language", map[string]string{})
		return c.Send(fmt.Sprintf(
			"Hey, %s! \U0001F44B\n\n"+
				"I'm *ForgePath* — I'll help you learn a language every day.\n\n"+
				"Here's how it works:\n"+
				"\U0001F31F Morning — a new word + quiz\n"+
				"\u270D\uFE0F Afternoon — write a few sentences\n"+
				"\U0001F3AC Evening — watch something fun\n"+
				"\U0001F31B Night — see how your day went\n\n"+
				"Let's start! What language?",
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
			return c.Send("Nothing to cancel right now.")
		}

		if state.State == "waiting_quiz_typing" || state.State == "waiting_quiz_sentence" {
			var wordID int
			fmt.Sscanf(state.Context["word_id"], "%d", &wordID)
			if wordID > 0 {
				reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)
				result := srs.Calculate(reps, interval, ease, 2)
				database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
			}
		}

		database.ClearState(userID)
		return c.Send("\u2705 Done! You can start something new anytime.")
	})

	b.Handle("/settings", func(c tele.Context) error {
		return c.Send("\u2699\uFE0F *Settings*\n\nWhat do you want to change?",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard()})
	})

	b.Handle("/help", func(c tele.Context) error {
		return c.Send(
			"\U0001F4DA *How ForgePath works*\n\n"+
				"Every day you get:\n"+
				"\U0001F31F *New word* at 7:30 — learn it and take a quiz\n"+
				"\u270D\uFE0F *Writing* at 12:00 — write a few sentences on a topic\n"+
				"\U0001F3AC *Video* at 18:00 — watch something and write about it\n"+
				"\U0001F31B *Review* at 21:30 — see how your day went\n\n"+
				"*Main commands:*\n"+
				"/word — learn a new word\n"+
				"/write — write something\n"+
				"/quiz — practice your words\n"+
				"/today — what's left for today\n"+
				"/stats — your progress\n"+
				"/skip — take a day off\n\n"+
				"Each week focuses on one grammar topic.\n"+
				"Don't worry about mistakes — that's how you learn! \U0001F4AA",
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
