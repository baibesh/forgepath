package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func RegisterCallbacks(b *tele.Bot, database *db.DB, openaiClient *ai.OpenAIClient) {

	b.Handle(&tele.Btn{Unique: "lang"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] onboarding lang=%s", userID, data)

		validLangs := map[string]bool{"en": true, "ru": true, "kk": true}
		if !validLangs[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid language"})
		}

		m := messagesForLang(data)
		database.UpdateUserLanguage(userID, data)
		database.SetState(userID, "onboarding_level", map[string]string{"language": data})

		c.Respond(&tele.CallbackResponse{Text: "OK!"})
		return c.Edit(m.Welcome(c.Sender().FirstName),
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

		user, _ := database.GetUser(userID)
		m := userMessages(user)

		database.UpdateUserLevel(userID, data)
		database.SetState(userID, "onboarding_tz", map[string]string{"level": data})

		c.Respond(&tele.CallbackResponse{Text: data})
		return c.Edit(fmt.Sprintf("\u2705 Level: *%s*\n\n%s", data, m.TimezonePrompt),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: TimezoneKeyboard()})
	})

	b.Handle(&tele.Btn{Unique: "tz"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] onboarding tz=%s", userID, data)

		user, _ := database.GetUser(userID)
		m := userMessages(user)
		lang := userLang(user)

		if data == "custom" {
			database.SetState(userID, "onboarding_tz_custom", map[string]string{})
			c.Respond(&tele.CallbackResponse{})
			return c.Edit(m.TzCustomPrompt)
		}

		offset, err := strconv.Atoi(data)
		if err != nil || offset < -12 || offset > 14 {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid timezone"})
		}

		database.UpdateUserTimezone(userID, offset)
		database.SetOnboarded(userID)
		database.ClearState(userID)

		c.Respond(&tele.CallbackResponse{Text: FormatUTCOffset(offset)})
		c.Edit(m.AllSet)

		user, _ = database.GetUser(userID)
		if user != nil {
			word, err := database.GetRandomUnseen(userID, user.Level, user.Language)
			if err == nil {
				grammar, _ := database.GetCurrentGrammarFocus(userID)
				database.MarkWordSeen(userID, word.ID)
				database.MarkWordDone(userID, user.TzOffset)
				c.Send(FormatWordOfDay(word, grammar, lang), &tele.SendOptions{
					ParseMode:   tele.ModeMarkdown,
					ReplyMarkup: MainKeyboard(lang),
				})
				return sendQuizForWord(c, database, word, openaiClient)
			}
		}
		return c.Send(m.ChooseAction, &tele.SendOptions{ReplyMarkup: MainKeyboard(lang)})
	})

	b.Handle(&tele.Btn{Unique: "settings"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] settings=%s", userID, data)

		user, _ := database.GetUser(userID)
		m := userMessages(user)
		lang := userLang(user)

		switch data {
		case "timezone":
			return c.Edit(m.BtnTimezone+":", &tele.SendOptions{ReplyMarkup: SettingsTimezoneKeyboard()})
		case "level":
			return c.Edit(m.BtnLevel+":", &tele.SendOptions{ReplyMarkup: SettingsLevelKeyboard()})
		case "language":
			return c.Edit(m.BtnLanguage+":", &tele.SendOptions{ReplyMarkup: SettingsLanguageKeyboard()})
		case "schedule":
			if webAppURL != "" {
				return c.Edit(m.BtnSchedule+":", &tele.SendOptions{ReplyMarkup: ScheduleKeyboard(webAppURL, lang)})
			}
			return c.Respond(&tele.CallbackResponse{Text: "Schedule settings not available"})
		}
		return nil
	})

	b.Handle(&tele.Btn{Unique: "setlang"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] setlang=%s", userID, data)

		validLangs := map[string]bool{"en": true, "ru": true, "kk": true}
		if !validLangs[data] {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid language"})
		}

		database.UpdateUserLanguage(userID, data)
		c.Respond(&tele.CallbackResponse{Text: "Language updated!"})
		langName := content.LanguageName(data)
		return c.Edit(fmt.Sprintf("\u2705 %s: *%s*!", content.GetMessages(data).BtnLanguage, langName),
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
		return c.Edit(fmt.Sprintf("\u2705 Level: *%s*!", data),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	b.Handle(&tele.Btn{Unique: "settz"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] settz=%s", userID, data)

		user, _ := database.GetUser(userID)
		m := userMessages(user)

		if data == "custom" {
			database.SetState(userID, "settings_tz_custom", map[string]string{})
			c.Respond(&tele.CallbackResponse{Text: "Type your UTC offset"})
			return c.Edit(m.TzCustomPrompt)
		}

		offset, err := strconv.Atoi(data)
		if err != nil || offset < -12 || offset > 14 {
			return c.Respond(&tele.CallbackResponse{Text: "Invalid timezone"})
		}

		database.UpdateUserTimezone(userID, offset)
		c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("Timezone: %s", FormatUTCOffset(offset))})
		return c.Edit(fmt.Sprintf("\u2705 Timezone: %s!", FormatUTCOffset(offset)))
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
			c.Respond(&tele.CallbackResponse{Text: "\u2705"})
			return c.Edit(fmt.Sprintf("\u2705 *%s* — %s",
				escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
				&tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		result := srs.Calculate(reps, interval, ease, 1)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		c.Respond(&tele.CallbackResponse{Text: "\u274C"})
		return c.Edit(fmt.Sprintf("\u274C *%s* — %s",
			escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	b.Handle(&tele.Btn{Unique: "skip"}, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID
		log.Printf("[user=%d] skip=%s", userID, data)

		user, err := database.GetUser(userID)
		m := userMessages(user)

		if data == "cancel" {
			c.Respond(&tele.CallbackResponse{})
			return c.Edit(m.SkipCancelled)
		}

		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: "Error"})
		}

		if user.SkipCount >= 2 {
			c.Respond(&tele.CallbackResponse{})
			return c.Edit(m.SkipMaxReached)
		}

		database.IncrementSkipCount(userID)
		database.MarkWordDone(userID, user.TzOffset)
		database.MarkWritingDone(userID, user.TzOffset)
		database.MarkReviewDone(userID, user.TzOffset)
		database.ClearState(userID)

		c.Respond(&tele.CallbackResponse{})
		return c.Edit(m.SkipDone(2 - user.SkipCount - 1))
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

		user, _ := database.GetUser(userID)
		lang := userLang(user)

		database.MarkMediaTaskSent(userID, mediaID)
		c.Respond(&tele.CallbackResponse{Text: "OK"})

		_, media, err := database.GetPendingMediaTask(userID)
		if err != nil {
			log.Printf("[user=%d] media task error: %v", userID, err)
			return c.Edit("\u2705")
		}

		grammar, _ := database.GetCurrentGrammarFocus(userID)
		grammarFocus := GrammarTenseName(grammar)

		database.SetState(userID, "waiting_media_task", map[string]string{
			"media_id":    fmt.Sprintf("%d", mediaID),
			"media_title": media.Title,
		})

		c.Edit("\u2705")
		return c.Send(FormatMediaTask(media, grammarFocus, lang), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
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

		user, _ := database.GetUser(userID)
		m := userMessages(user)

		if openaiClient == nil {
			return c.Respond(&tele.CallbackResponse{Text: m.AudioNotAvail})
		}

		c.Respond(&tele.CallbackResponse{Text: m.AudioGenerating})

		ttsText := fmt.Sprintf("%s. %s", word.Word, word.Example)
		audioPath, err := openaiClient.TextToSpeech(ttsText)
		if err != nil {
			log.Printf("[user=%d] TTS error: %v", userID, err)
			return c.Send(m.AudioFailed)
		}
		defer os.Remove(audioPath)

		voice := &tele.Voice{File: tele.FromDisk(audioPath)}
		return c.Send(voice)
	})
}
