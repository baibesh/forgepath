package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func RegisterCallbacks(b *tele.Bot, database *db.DB, openaiClient *ai.OpenAIClient) {
	b.Handle(&tele.Btn{Unique: "level"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback level=%s", userID, data)

		// Edge case: validate level
		validLevels := map[string]bool{"A1": true, "A2": true, "B1": true, "B2": true, "C1": true}
		if !validLevels[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid level"})
		}

		database.UpdateUserLevel(userID, data)
		database.SetState(userID, "onboarding_tz", map[string]string{"level": data})

		c.Respond(&tele.CallbackResponse{Text: "Level set to " + data})
		return c.Edit(fmt.Sprintf("✅ Level: *%s*\n\nNow select your timezone:", data),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: TimezoneKeyboard()})
	})

	b.Handle(&tele.Btn{Unique: "tz"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback tz=%s", userID, data)

		if data == "custom" {
			database.SetState(userID, "onboarding_tz_custom", map[string]string{})
			c.Respond(&tele.CallbackResponse{Text: "Type your UTC offset"})
			return c.Edit("Type your UTC offset (e.g. 5 for UTC+5, -3 for UTC-3):")
		}

		offset, err := strconv.Atoi(data)
		if err != nil || offset < -12 || offset > 14 {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid timezone"})
		}

		database.UpdateUserTimezone(userID, offset)
		database.ClearState(userID)

		c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("Timezone: UTC+%d", offset)})
		c.Edit(fmt.Sprintf("✅ Setup complete!\n\nTimezone: UTC+%d\nYou're all set! 🚀\n\nUse the menu or /help to see commands.", offset))
		return c.Send("Let's go! 💪", &tele.SendOptions{ReplyMarkup: MainMenu()})
	})

	b.Handle(&tele.Btn{Unique: "quiz"}, func(c tele.Context) error {
		parts := strings.Split(c.Data(), "|")
		if len(parts) != 2 {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid quiz data"})
		}

		wordID, err := strconv.Atoi(parts[0])
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid quiz data"})
		}
		selectedIdx, err := strconv.Atoi(parts[1])
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid quiz data"})
		}

		userID := c.Sender().ID
		log.Printf("[user=%d] callback quiz word=%d idx=%d", userID, wordID, selectedIdx)

		word, err := database.GetWordByID(wordID)
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Word not found"})
		}

		// Edge case: message or keyboard gone
		msg := c.Callback().Message
		if msg == nil || msg.ReplyMarkup == nil {
			return c.Respond(&tele.CallbackResponse{Text: "Quiz expired"})
		}

		// Edge case: index out of range
		var selectedText string
		if selectedIdx >= 0 && selectedIdx < len(msg.ReplyMarkup.InlineKeyboard) {
			row := msg.ReplyMarkup.InlineKeyboard[selectedIdx]
			if len(row) > 0 {
				selectedText = row[0].Text
			}
		}
		if selectedText == "" {
			return c.Respond(&tele.CallbackResponse{Text: "Quiz expired"})
		}

		isCorrect := strings.Contains(selectedText, word.Definition)

		if isCorrect {
			result := srs.Calculate(0, 1, 2.5, 4)
			database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
			database.MarkReviewDone(userID)
			c.Respond(&tele.CallbackResponse{Text: "✅ Correct!"})
			return c.Edit(fmt.Sprintf("✅ *Correct!*\n\n*%s* — %s",
				escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
				&tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		result := srs.Calculate(0, 1, 2.5, 1)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		c.Respond(&tele.CallbackResponse{Text: "❌ Wrong!"})
		return c.Edit(fmt.Sprintf("❌ *Wrong!*\n\n*%s* — %s\n\nYou'll see this again soon!",
			escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	b.Handle(&tele.Btn{Unique: "skip"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback skip=%s", userID, data)

		if data == "cancel" {
			c.Respond(&tele.CallbackResponse{Text: "Cancelled"})
			return c.Edit("✅ Skip cancelled. Keep going! 💪")
		}

		user, err := database.GetUser(userID)
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Error"})
		}

		if user.SkipCount >= 2 {
			c.Respond(&tele.CallbackResponse{Text: "No skips left!"})
			return c.Edit("❌ You've already used both skips this week.")
		}

		database.IncrementSkipCount(userID)
		database.MarkWordDone(userID)
		database.MarkWritingDone(userID)
		database.MarkReviewDone(userID)

		c.Respond(&tele.CallbackResponse{Text: "Day skipped"})
		return c.Edit(fmt.Sprintf("⏭ Day skipped.\n\nSkips used: %d/2 this week.", user.SkipCount+1))
	})

	b.Handle(&tele.Btn{Unique: "settings"}, func(c tele.Context) error {
		data := c.Data()
		log.Printf("[user=%d] callback settings=%s", c.Sender().ID, data)
		switch data {
		case "timezone":
			return c.Edit("🕐 Select your timezone:", &tele.SendOptions{ReplyMarkup: TimezoneKeyboard()})
		case "level":
			return c.Edit("📚 Select your level:", &tele.SendOptions{ReplyMarkup: LevelSelectKeyboard()})
		}
		return nil
	})

	b.Handle(&tele.Btn{Unique: "media"}, func(c tele.Context) error {
		parts := strings.Split(c.Data(), "|")
		if len(parts) != 2 || parts[0] != "done" {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid data"})
		}

		mediaID, err := strconv.Atoi(parts[1])
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid data"})
		}

		userID := c.Sender().ID
		log.Printf("[user=%d] callback media done=%d", userID, mediaID)

		database.MarkMediaTaskSent(userID, mediaID)
		c.Respond(&tele.CallbackResponse{Text: "Great! Task incoming..."})

		_, media, err := database.GetPendingMediaTask(userID)
		if err != nil {
			log.Printf("[user=%d] media task error: %v", userID, err)
			return c.Edit("✅ Marked as watched!")
		}

		grammar, _ := database.GetCurrentGrammarFocus(userID)
		grammarFocus := "Past Simple"
		if grammar != nil {
			grammarFocus = grammar.TenseName
		}

		database.SetState(userID, "waiting_media_task", map[string]string{
			"media_id":    fmt.Sprintf("%d", mediaID),
			"media_title": media.Title,
		})

		c.Edit("✅ Marked as watched!")
		return c.Send(FormatMediaTask(media, grammarFocus), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})
}
