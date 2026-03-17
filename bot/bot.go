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
		{Text: "addword", Description: "Add a custom word"},
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
	SetWebAppURL(cfg.WebAppURL)

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
			m := userMessages(existing)
			lang := userLang(existing)
			scheduleText := FormatSchedule(existing.Schedule, lang)
			msg := m.StartReturning(
				user.FirstName, content.LanguageFlag(existing.Language),
				content.LanguageName(existing.Language), existing.Level,
				scheduleText,
			)
			c.Send(msg, &tele.SendOptions{
				ParseMode:   tele.ModeMarkdown,
				ReplyMarkup: ScheduleKeyboard(cfg.WebAppURL, lang),
			})
			return c.Send(m.ChooseAction, &tele.SendOptions{
				ReplyMarkup: MainKeyboard(lang),
			})
		}

		database.SetState(user.ID, "onboarding_language", map[string]string{})
		return c.Send(fmt.Sprintf(
			"Hey, %s! \U0001F44B\n\n"+
				"What language do you want to learn?\n"+
				"Какой язык хочешь учить?\n"+
				"Қай тілді үйренгің келеді?",
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

	b.Handle("/addword", func(c tele.Context) error { return handleAddWord(c, database) })

	b.Handle("/cancel", func(c tele.Context) error {
		userID := c.Sender().ID
		log.Printf("[user=%d] /cancel", userID)
		user, _ := database.GetUser(userID)
		m := userMessages(user)

		state, _ := database.GetState(userID)
		if state.State == "idle" {
			return c.Send(m.CancelNothing)
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
		return c.Send(m.CancelDone)
	})

	b.Handle("/settings", func(c tele.Context) error {
		user, _ := database.GetUser(c.Sender().ID)
		m := userMessages(user)
		lang := userLang(user)
		return c.Send(m.SettingsTitle,
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard(lang)})
	})

	b.Handle("/help", func(c tele.Context) error {
		user, _ := database.GetUser(c.Sender().ID)
		m := userMessages(user)
		return c.Send(m.Help, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
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
