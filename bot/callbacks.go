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

	b.Handle(&tele.Btn{Unique: "lang"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] onboarding lang=%s", userID, data)

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

	b.Handle(&tele.Btn{Unique: "level"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] onboarding level=%s", userID, data)

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

	b.Handle(&tele.Btn{Unique: "tz"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] onboarding tz=%s", userID, data)

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
		database.SetOnboarded(userID)
		database.ClearState(userID)

		c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("Timezone: %s", FormatUTCOffset(offset))})
		c.Edit(fmt.Sprintf("\u2705 All set! Your first word is coming! \U0001F680"))

		user, _ := database.GetUser(userID)
		if user != nil {
			word, err := database.GetRandomUnseen(userID, user.Level, user.Language)
			if err == nil {
				grammar, _ := database.GetCurrentGrammarFocus(userID)
				database.MarkWordSeen(userID, word.ID)
				database.MarkWordDone(userID, user.TzOffset)
				c.Send(FormatWordOfDay(word, grammar), &tele.SendOptions{
					ParseMode:   tele.ModeMarkdown,
					ReplyMarkup: MainKeyboard(),
				})
				return sendQuizForWord(c, database, word, openaiClient)
			}
		}
		return c.Send("You're all set! Use the buttons below.", &tele.SendOptions{ReplyMarkup: MainKeyboard()})
	})

	b.Handle(&tele.Btn{Unique: "settings"}, func(c tele.Context) error {
		data := c.Data()
		log.Printf("[user=%d] settings=%s", c.Sender().ID, data)
		switch data {
		case "timezone":
			return c.Edit("\U0001F550 Select your timezone:", &tele.SendOptions{ReplyMarkup: SettingsTimezoneKeyboard()})
		case "level":
			return c.Edit("\U0001F4DA Select your level:", &tele.SendOptions{ReplyMarkup: SettingsLevelKeyboard()})
		case "language":
			return c.Edit("\U0001F310 Select your language:", &tele.SendOptions{ReplyMarkup: SettingsLanguageKeyboard()})
		case "schedule":
			if webAppURL != "" {
				return c.Edit("\U0001F514 Set your notification times:", &tele.SendOptions{ReplyMarkup: ScheduleKeyboard(webAppURL)})
			}
			return c.Respond(&tele.CallbackResponse{Text: "Schedule settings not available"})
		}
		return nil
	})

	b.Handle(&tele.Btn{Unique: "setlang"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] setlang=%s", userID, data)

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

	b.Handle(&tele.Btn{Unique: "setlevel"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] setlevel=%s", userID, data)

		validLevels := map[string]bool{"A1": true, "A2": true, "B1": true, "B2": true, "C1": true}
		if !validLevels[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid level"})
		}

		database.UpdateUserLevel(userID, data)
		c.Respond(&tele.CallbackResponse{Text: "Level set to " + data})
		return c.Edit(fmt.Sprintf("\u2705 Level changed to *%s*!", data),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	b.Handle(&tele.Btn{Unique: "settz"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] settz=%s", userID, data)

		if data == "custom" {
			database.SetState(userID, "settings_tz_custom", map[string]string{})
			c.Respond(&tele.CallbackResponse{Text: "Type your UTC offset"})
			return c.Edit("Type your UTC offset (e.g. 5 for UTC+5, -3 for UTC-3):")
		}

		offset, err := strconv.Atoi(data)
		if err != nil || offset < -12 || offset > 14 {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid timezone"})
		}

		database.UpdateUserTimezone(userID, offset)
		c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("Timezone: %s", FormatUTCOffset(offset))})
		return c.Edit(fmt.Sprintf("\u2705 Timezone changed to %s!", FormatUTCOffset(offset)))
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
		log.Printf("[user=%d] quiz word=%d idx=%d", userID, wordID, selectedIdx)

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
		tzOffset := userTzOffset(user)

		reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)
		isCorrect := strings.Contains(selectedText, word.Definition)

		if isCorrect {
			result := srs.Calculate(reps, interval, ease, 4)
			database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
			database.MarkReviewDone(userID, tzOffset)
			c.Respond(&tele.CallbackResponse{Text: "\u2705 Yes!"})
			return c.Edit(fmt.Sprintf("\u2705 *You got it!*\n\n*%s* — %s",
				escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
				&tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		result := srs.Calculate(reps, interval, ease, 1)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		c.Respond(&tele.CallbackResponse{Text: "Not this time"})
		return c.Edit(fmt.Sprintf("\u274C The answer was: *%s* — %s\n\nNo worries, you'll see it again!",
			escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	b.Handle(&tele.Btn{Unique: "skip"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] skip=%s", userID, data)

		if data == "cancel" {
			c.Respond(&tele.CallbackResponse{Text: "Cancelled"})
			return c.Edit("\u2705 Good choice! Let's keep going!")
		}

		user, err := database.GetUser(userID)
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Error"})
		}

		if user.SkipCount >= 2 {
			c.Respond(&tele.CallbackResponse{Text: "No days off left"})
			return c.Edit("You've already taken 2 days off this week.")
		}

		database.IncrementSkipCount(userID)
		database.MarkWordDone(userID, user.TzOffset)
		database.MarkWritingDone(userID, user.TzOffset)
		database.MarkReviewDone(userID, user.TzOffset)
		database.ClearState(userID)

		c.Respond(&tele.CallbackResponse{Text: "Day off!"})
		return c.Edit(fmt.Sprintf("\U0001F634 Rest day! You have %d day(s) off left this week.", 2-user.SkipCount-1))
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
		log.Printf("[user=%d] media done=%d", userID, mediaID)

		database.MarkMediaTaskSent(userID, mediaID)
		c.Respond(&tele.CallbackResponse{Text: "Great! Task incoming..."})

		_, media, err := database.GetPendingMediaTask(userID)
		if err != nil {
			log.Printf("[user=%d] media task error: %v", userID, err)
			return c.Edit("\u2705 Marked as watched!")
		}

		grammar, _ := database.GetCurrentGrammarFocus(userID)
		grammarFocus := GrammarTenseName(grammar)

		database.SetState(userID, "waiting_media_task", map[string]string{
			"media_id":    fmt.Sprintf("%d", mediaID),
			"media_title": media.Title,
		})

		c.Edit("\u2705 Marked as watched!")
		return c.Send(FormatMediaTask(media, grammarFocus), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	b.Handle(&tele.Btn{Unique: "listen"}, func(c tele.Context) error {
		wordID, err := strconv.Atoi(c.Data())
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid data"})
		}

		userID := c.Sender().ID
		log.Printf("[user=%d] listen word=%d", userID, wordID)

		word, err := database.GetWordByID(wordID)
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Word not found"})
		}

		if openaiClient == nil {
			return c.Respond(&tele.CallbackResponse{Text: "Audio not available"})
		}

		c.Respond(&tele.CallbackResponse{Text: "Generating audio..."})

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
}
