package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func RegisterCallbacks(b *tele.Bot, database *db.DB, openaiClient *ai.OpenAIClient) {

	// Language selection (onboarding step 1)
	b.Handle(&tele.Btn{Unique: "lang"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback lang=%s", userID, data)

		validLangs := map[string]bool{"en": true, "de": true}
		if !validLangs[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid language"})
		}

		database.UpdateUserLanguage(userID, data)
		database.SetState(userID, "onboarding_level", map[string]string{"language": data})

		c.Respond(&tele.CallbackResponse{Text: "Language set!"})
		langName := "English"
		if data == "de" {
			langName = "Deutsch"
		}
		return c.Edit(fmt.Sprintf("\u2705 Language: *%s*\n\nNow select your level:", langName),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: LevelSelectKeyboard()})
	})

	// Level selection (onboarding step 2)
	b.Handle(&tele.Btn{Unique: "level"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback level=%s", userID, data)

		validLevels := map[string]bool{"A1": true, "A2": true, "B1": true, "B2": true, "C1": true}
		if !validLevels[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid level"})
		}

		database.UpdateUserLevel(userID, data)
		database.SetState(userID, "onboarding_tz", map[string]string{"level": data})

		c.Respond(&tele.CallbackResponse{Text: "Level set to " + data})
		return c.Edit(fmt.Sprintf("\u2705 Level: *%s*\n\nNow select your timezone:", data),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: TimezoneKeyboard()})
	})

	// Timezone selection (onboarding step 3)
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
		c.Edit(fmt.Sprintf("\u2705 Setup complete!\n\nTimezone: UTC+%d\nYou're all set! \U0001F680\n\nUse the menu or /help to see commands.", offset))
		return c.Send("Let's go! \U0001F4AA")
	})

	// Quiz answer
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

		msg := c.Callback().Message
		if msg == nil || msg.ReplyMarkup == nil {
			return c.Respond(&tele.CallbackResponse{Text: "Quiz expired"})
		}

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

		user, _ := database.GetUser(userID)
		tzOffset := 5
		if user != nil {
			tzOffset = user.TzOffset
		}

		// Get real SRS values
		reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)

		isCorrect := strings.Contains(selectedText, word.Definition)

		if isCorrect {
			result := srs.Calculate(reps, interval, ease, 4)
			database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
			database.MarkReviewDone(userID, tzOffset)
			c.Respond(&tele.CallbackResponse{Text: "\u2705 Correct!"})
			return c.Edit(fmt.Sprintf("\u2705 *Correct!*\n\n*%s* — %s",
				escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
				&tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		result := srs.Calculate(reps, interval, ease, 1)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		c.Respond(&tele.CallbackResponse{Text: "\u274C Wrong!"})
		return c.Edit(fmt.Sprintf("\u274C *Wrong!*\n\n*%s* — %s\n\nYou'll see this again soon!",
			escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	// Skip confirmation
	b.Handle(&tele.Btn{Unique: "skip"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback skip=%s", userID, data)

		if data == "cancel" {
			c.Respond(&tele.CallbackResponse{Text: "Cancelled"})
			return c.Edit("\u2705 Skip cancelled. Keep going! \U0001F4AA")
		}

		user, err := database.GetUser(userID)
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Error"})
		}

		if user.SkipCount >= 2 {
			c.Respond(&tele.CallbackResponse{Text: "No skips left!"})
			return c.Edit("\u274C You've already used both skips this week.")
		}

		database.IncrementSkipCount(userID)
		database.MarkWordDone(userID, user.TzOffset)
		database.MarkWritingDone(userID, user.TzOffset)
		database.MarkReviewDone(userID, user.TzOffset)

		c.Respond(&tele.CallbackResponse{Text: "Day skipped"})
		return c.Edit(fmt.Sprintf("\u23ED Day skipped.\n\nSkips used: %d/2 this week.", user.SkipCount+1))
	})

	// Settings
	b.Handle(&tele.Btn{Unique: "settings"}, func(c tele.Context) error {
		data := c.Data()
		log.Printf("[user=%d] callback settings=%s", c.Sender().ID, data)
		switch data {
		case "timezone":
			return c.Edit("\U0001F550 Select your timezone:", &tele.SendOptions{ReplyMarkup: TimezoneKeyboard()})
		case "level":
			return c.Edit("\U0001F4DA Select your level:", &tele.SendOptions{ReplyMarkup: LevelSelectKeyboard()})
		case "language":
			return c.Edit("\U0001F310 Select your language:", &tele.SendOptions{ReplyMarkup: LanguageSelectKeyboard()})
		}
		return nil
	})

	// Media watched
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
			return c.Edit("\u2705 Marked as watched!")
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

		c.Edit("\u2705 Marked as watched!")
		return c.Send(FormatMediaTask(media, grammarFocus), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	// Listen — TTS pronunciation
	b.Handle(&tele.Btn{Unique: "listen"}, func(c tele.Context) error {
		wordID, err := strconv.Atoi(c.Data())
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid data"})
		}

		userID := c.Sender().ID
		log.Printf("[user=%d] callback listen word=%d", userID, wordID)

		word, err := database.GetWordByID(wordID)
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Word not found"})
		}

		if openaiClient == nil {
			return c.Respond(&tele.CallbackResponse{Text: "Audio not available"})
		}

		c.Respond(&tele.CallbackResponse{Text: "Generating audio..."})

		// Generate TTS: word + example
		ttsText := fmt.Sprintf("%s. %s", word.Word, word.Example)
		audioPath, err := openaiClient.TextToSpeech(ttsText)
		if err != nil {
			log.Printf("[user=%d] TTS error: %v", userID, err)
			return c.Send("Audio generation failed. Try again later.")
		}
		defer os.Remove(audioPath)

		voice := &tele.Voice{File: tele.FromDisk(audioPath)}
		return c.Send(voice)
	})

	// Settings language change (outside onboarding)
	b.Handle(&tele.Btn{Unique: "setlang"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] callback setlang=%s", userID, data)

		validLangs := map[string]bool{"en": true, "de": true}
		if !validLangs[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid language"})
		}

		database.UpdateUserLanguage(userID, data)
		c.Respond(&tele.CallbackResponse{Text: "Language updated!"})
		langName := "English"
		if data == "de" {
			langName = "Deutsch"
		}
		return c.Edit(fmt.Sprintf("\u2705 Language changed to *%s*!", langName),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})
}
