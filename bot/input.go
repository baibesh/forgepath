package bot

import (
	"fmt"
	"log"
	"os"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func buttonCommandsForLang(lang string) map[string]string {
	m := content.GetMessages(lang)
	return map[string]string{
		m.BtnNewWord:  "/word",
		m.BtnWrite:    "/write",
		m.BtnQuiz:     "/quiz",
		m.BtnToday:    "/today",
		m.BtnProgress: "/stats",
		m.BtnSettings: "/settings",
	}
}

func handleText(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())
	if text == "" {
		return nil
	}

	// Try matching button text for all supported languages
	for _, lang := range []string{"en", "ru", "kk"} {
		cmds := buttonCommandsForLang(lang)
		if cmd, ok := cmds[text]; ok {
			log.Printf("[user=%d] button -> %s", userID, cmd)
			c.Message().Text = cmd
			return routeButtonCommand(c, database, openaiClient, cmd)
		}
	}

	state, err := database.GetState(userID)
	if err != nil {
		return nil
	}

	log.Printf("[user=%d] text in state=%s len=%d", userID, state.State, len(text))
	return processTextInput(c, database, openaiClient, state, text)
}

func routeButtonCommand(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, cmd string) error {
	user, _ := database.GetUser(c.Sender().ID)
	m := userMessages(user)
	lang := userLang(user)
	state, _ := database.GetState(c.Sender().ID)
	if state.State != "idle" && state.State != "" {
		database.ClearState(c.Sender().ID)
		c.Send(m.PrevTaskCancelled)
	}

	switch cmd {
	case "/word":
		return handleWord(c, database, openaiClient)
	case "/write":
		return handleWrite(c, database)
	case "/quiz":
		return handleQuiz(c, database, openaiClient)
	case "/today":
		return handleToday(c, database)
	case "/stats":
		return handleStats(c, database)
	case "/settings":
		return c.Send(m.SettingsTitle,
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard(lang)})
	}
	return nil
}

func handleVoice(c tele.Context, b *tele.Bot, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	voice := c.Message().Voice
	if voice == nil {
		return nil
	}

	log.Printf("[user=%d] voice message, duration=%ds", userID, voice.Duration)

	user, _ := database.GetUser(userID)
	m := userMessages(user)

	if openaiClient == nil {
		return c.Send(m.VoiceNotAvailable)
	}

	file, err := b.FileByID(voice.FileID)
	if err != nil {
		log.Printf("[user=%d] voice download error: %v", userID, err)
		return c.Send(m.VoiceError)
	}

	tmpFile, tmpErr := os.CreateTemp("", fmt.Sprintf("forgepath-voice-%d-*.ogg", userID))
	if tmpErr != nil {
		log.Printf("[user=%d] voice temp file error: %v", userID, tmpErr)
		return c.Send(m.VoiceError)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	if err := b.Download(&file, tmpPath); err != nil {
		log.Printf("[user=%d] voice save error: %v", userID, err)
		return c.Send(m.VoiceError)
	}
	defer os.Remove(tmpPath)

	text, err := openaiClient.SpeechToText(tmpPath)
	if err != nil {
		log.Printf("[user=%d] transcription error: %v", userID, err)
		return c.Send(m.VoiceError)
	}
	if text == "" {
		return c.Send(m.VoiceCantHear)
	}

	log.Printf("[user=%d] transcribed: %s", userID, text)
	c.Send(fmt.Sprintf("\U0001F399 _Heard:_ \"%s\"", escapeMarkdown(text)), &tele.SendOptions{ParseMode: tele.ModeMarkdown})

	state, _ := database.GetState(userID)
	if state.State == "idle" {
		return c.Send(m.VoiceIdleHint)
	}

	return processTextInput(c, database, openaiClient, state, text)
}

func processTextInput(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	switch state.State {
	case "onboarding_tz_custom":
		return handleOnboardingTzCustom(c, database, openaiClient, text)
	case "settings_tz_custom":
		return handleSettingsTzCustom(c, database, text)
	case "waiting_writing":
		return processWriting(c, database, openaiClient, state, text)
	case "waiting_quiz_typing":
		return processQuizTyping(c, database, state, text)
	case "waiting_quiz_sentence":
		return processQuizSentence(c, database, openaiClient, state, text)
	case "waiting_media_task":
		return processMediaTask(c, database, openaiClient, state, text)
	case "waiting_addword":
		return processAddWord(c, database, openaiClient, text)
	default:
		return nil
	}
}

func handleOnboardingTzCustom(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, text string) error {
	userID := c.Sender().ID
	user, _ := database.GetUser(userID)
	m := userMessages(user)
	lang := userLang(user)

	var offset int
	_, err := fmt.Sscanf(text, "%d", &offset)
	if err != nil || offset < -12 || offset > 14 {
		return c.Send(m.TzInvalid)
	}

	database.UpdateUserTimezone(userID, offset)
	database.SetOnboarded(userID)
	database.ClearState(userID)

	c.Send(m.AllSet)

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
}

func handleSettingsTzCustom(c tele.Context, database *db.DB, text string) error {
	userID := c.Sender().ID
	user, _ := database.GetUser(userID)
	m := userMessages(user)

	var offset int
	_, err := fmt.Sscanf(text, "%d", &offset)
	if err != nil || offset < -12 || offset > 14 {
		return c.Send(m.TzInvalid)
	}

	database.UpdateUserTimezone(userID, offset)
	database.ClearState(userID)

	return c.Send(fmt.Sprintf("\u2705 Timezone changed to %s!", FormatUTCOffset(offset)))
}

func processWriting(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	userID := c.Sender().ID
	user, _ := database.GetUser(userID)
	m := userMessages(user)

	wordCount := len(strings.Fields(text))
	if wordCount < 5 {
		return c.Send(m.WritingTooShort)
	}
	if len(text) > 3000 {
		text = text[:3000]
	}

	topic := state.Context["topic"]
	grammarFocus := state.Context["grammar_focus"]

	writingID, err := database.SaveWriting(userID, topic, grammarFocus, text, wordCount)
	if err != nil {
		log.Printf("[user=%d] save writing error: %v", userID, err)
		return c.Send(m.WritingSaveError)
	}

	tzOffset := userTzOffset(user)
	level := "A2"
	language := "en"
	if user != nil {
		level = user.Level
		language = user.Language
	}

	database.MarkWritingDone(userID, tzOffset)
	database.ClearState(userID)

	c.Send(m.WritingSaved(wordCount))

	feedback, err := openaiClient.CheckWriting(text, grammarFocus, level, language)
	if err != nil {
		log.Printf("[user=%d] AI feedback error writing=%d: %v", userID, writingID, err)
		feedback = "AI feedback is not available right now. Keep writing!"
	}

	database.UpdateWritingFeedback(writingID, feedback)
	return c.Send(feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func processQuizTyping(c tele.Context, database *db.DB, state *db.UserState, text string) error {
	userID := c.Sender().ID
	text = strings.TrimSpace(strings.ToLower(text))
	answer := state.Context["answer"]

	var wordID int
	fmt.Sscanf(state.Context["word_id"], "%d", &wordID)

	user, _ := database.GetUser(userID)
	m := userMessages(user)
	tzOffset := userTzOffset(user)

	reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)
	database.ClearState(userID)

	if text == answer {
		result := srs.Calculate(reps, interval, ease, 5)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		database.MarkReviewDone(userID, tzOffset)
		return c.Send(m.QuizCorrect)
	}

	result := srs.Calculate(reps, interval, ease, 1)
	database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)

	word, _ := database.GetWordByID(wordID)
	if word != nil {
		return c.Send(m.QuizWrong(escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}
	return c.Send(m.QuizWrongSimple)
}

func processQuizSentence(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	userID := c.Sender().ID
	targetWord := state.Context["word"]

	var wordID int
	fmt.Sscanf(state.Context["word_id"], "%d", &wordID)

	user, _ := database.GetUser(userID)
	m := userMessages(user)

	if len(text) < 5 {
		return c.Send(m.QuizTrySentence)
	}

	if !strings.Contains(strings.ToLower(text), strings.ToLower(targetWord)) {
		return c.Send(m.TryUseWord(escapeMarkdown(targetWord)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	database.ClearState(userID)

	tzOffset := userTzOffset(user)
	level := "A2"
	language := "en"
	if user != nil {
		level = user.Level
		language = user.Language
	}

	reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)
	result := srs.Calculate(reps, interval, ease, 4)
	database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
	database.MarkReviewDone(userID, tzOffset)

	if openaiClient != nil {
		feedback, err := openaiClient.CheckSentences(text, targetWord, level, language)
		if err == nil {
			return c.Send("\u2705 Nice sentence!\n\n"+feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
	}

	return c.Send(m.MediaGoodJob)
}

func processMediaTask(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	userID := c.Sender().ID
	user, _ := database.GetUser(userID)
	m := userMessages(user)

	var mediaID int
	fmt.Sscanf(state.Context["media_id"], "%d", &mediaID)
	mediaTitle := state.Context["media_title"]

	if len(text) < 10 {
		return c.Send(m.MediaTooShort)
	}

	database.ClearState(userID)

	tzOffset := userTzOffset(user)
	level := "A2"
	language := "en"
	if user != nil {
		level = user.Level
		language = user.Language
	}

	database.SaveMediaTaskResponse(userID, mediaID, text)
	database.MarkReviewDone(userID, tzOffset)

	wordCount := len(strings.Fields(text))
	topic := fmt.Sprintf("Video: %s", mediaTitle)
	writingID, _ := database.SaveWritingWithType(userID, topic, "", text, wordCount, "media")
	database.MarkWritingDone(userID, tzOffset)

	c.Send(m.MediaGotIt)

	if openaiClient == nil {
		return c.Send(m.MediaGoodJob)
	}

	feedback, err := openaiClient.CheckSentences(text, mediaTitle, level, language)
	if err != nil {
		log.Printf("[user=%d] AI media feedback error: %v", userID, err)
		return c.Send(m.MediaGoodJob)
	}

	if writingID > 0 {
		database.UpdateWritingFeedback(writingID, feedback)
	}

	return c.Send(feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func processAddWord(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, text string) error {
	userID := c.Sender().ID
	user, _ := database.GetUser(userID)
	m := userMessages(user)
	lang := userLang(user)

	word := strings.TrimSpace(strings.ToLower(text))
	if word == "" || len(word) > 100 {
		return c.Send(m.AddWordNotFound)
	}

	database.ClearState(userID)

	// Check if user already has this word
	existing, _ := database.GetWordByText(word, "en")
	if existing != nil {
		reps, _ := database.GetUserWordRepetitions(userID, existing.ID)
		if reps > 0 {
			return c.Send(m.AddWordExists(escapeMarkdown(existing.Word)),
				&tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
		database.MarkWordSeen(userID, existing.ID)
		return c.Send(m.AddWordAdded(escapeMarkdown(existing.Word), escapeMarkdown(existing.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	c.Send(m.AddWordSearching)

	info, err := openaiClient.LookupWord(word, lang)
	if err != nil {
		log.Printf("[user=%d] addword lookup error: %v", userID, err)
		return c.Send(m.AddWordNotFound)
	}

	level := "A2"
	if user != nil {
		level = user.Level
	}

	wordID, err := database.InsertCustomWord(word, info.Definition, info.Example, info.Collocations, info.Construction, level, "en")
	if err != nil {
		log.Printf("[user=%d] addword insert error: %v", userID, err)
		return c.Send(m.SomethingWrong)
	}

	database.MarkWordSeen(userID, wordID)

	return c.Send(m.AddWordAdded(escapeMarkdown(word), escapeMarkdown(info.Definition)),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
