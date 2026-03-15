package bot

import (
	"fmt"
	"log"
	"os"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func handleText(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())
	if text == "" {
		return nil
	}

	state, err := database.GetState(userID)
	if err != nil {
		return nil
	}

	log.Printf("[user=%d] text in state=%s len=%d", userID, state.State, len(text))
	return processTextInput(c, database, openaiClient, state, text)
}

func handleVoice(c tele.Context, b *tele.Bot, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	voice := c.Message().Voice
	if voice == nil {
		return nil
	}

	log.Printf("[user=%d] voice message, duration=%ds", userID, voice.Duration)

	if openaiClient == nil {
		return c.Send("Voice recognition is not available right now.")
	}

	file, err := b.FileByID(voice.FileID)
	if err != nil {
		log.Printf("[user=%d] voice download error: %v", userID, err)
		return c.Send("Could not process your voice message. Try again.")
	}

	tmpPath := fmt.Sprintf("/tmp/forgepath-voice-%d.ogg", userID)
	if err := b.Download(&file, tmpPath); err != nil {
		log.Printf("[user=%d] voice save error: %v", userID, err)
		return c.Send("Could not process your voice message. Try again.")
	}
	defer os.Remove(tmpPath)

	text, err := openaiClient.SpeechToText(tmpPath)
	if err != nil {
		log.Printf("[user=%d] transcription error: %v", userID, err)
		return c.Send("Could not recognize your voice. Try sending text instead.")
	}
	if text == "" {
		return c.Send("Could not hear anything. Try again or send text.")
	}

	log.Printf("[user=%d] transcribed: %s", userID, text)
	c.Send(fmt.Sprintf("\U0001F399 _Heard:_ \"%s\"", escapeMarkdown(text)), &tele.SendOptions{ParseMode: tele.ModeMarkdown})

	state, _ := database.GetState(userID)
	if state.State == "idle" {
		return c.Send("You can send voice messages when the bot is waiting for your text (writing, quiz, etc.)")
	}

	return processTextInput(c, database, openaiClient, state, text)
}

func processTextInput(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	switch state.State {
	case "onboarding_tz_custom":
		return handleOnboardingTzCustom(c, database, text)
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
	default:
		return nil
	}
}

func handleOnboardingTzCustom(c tele.Context, database *db.DB, text string) error {
	userID := c.Sender().ID

	var offset int
	_, err := fmt.Sscanf(text, "%d", &offset)
	if err != nil || offset < -12 || offset > 14 {
		return c.Send("Please enter a number between -12 and 14:")
	}

	database.UpdateUserTimezone(userID, offset)
	database.SetOnboarded(userID)
	database.ClearState(userID)

	return c.Send(
		fmt.Sprintf("\u2705 Setup complete!\n\nTimezone: UTC+%d\nYou're all set! \U0001F680\n\nUse /help to see commands.", offset))
}

func handleSettingsTzCustom(c tele.Context, database *db.DB, text string) error {
	userID := c.Sender().ID

	var offset int
	_, err := fmt.Sscanf(text, "%d", &offset)
	if err != nil || offset < -12 || offset > 14 {
		return c.Send("Please enter a number between -12 and 14:")
	}

	database.UpdateUserTimezone(userID, offset)
	database.ClearState(userID)

	return c.Send(fmt.Sprintf("\u2705 Timezone changed to UTC+%d!", offset))
}

func processWriting(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	userID := c.Sender().ID

	if len(text) < 10 {
		return c.Send("Your text is too short. Write at least a few sentences!")
	}
	if len(text) > 3000 {
		text = text[:3000]
	}

	wordCount := len(strings.Fields(text))
	topic := state.Context["topic"]
	grammarFocus := state.Context["grammar_focus"]

	writingID, err := database.SaveWriting(userID, topic, grammarFocus, text, wordCount)
	if err != nil {
		log.Printf("[user=%d] save writing error: %v", userID, err)
		return c.Send("Error saving your writing. Try again.")
	}

	user, _ := database.GetUser(userID)
	tzOffset := 5
	level := "A2"
	language := "en"
	if user != nil {
		tzOffset = user.TzOffset
		level = user.Level
		language = user.Language
	}

	database.MarkWritingDone(userID, tzOffset)
	database.ClearState(userID)

	c.Send(fmt.Sprintf("\u2705 Saved! (%d words)\n\nAnalyzing...", wordCount))

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
	tzOffset := 5
	if user != nil {
		tzOffset = user.TzOffset
	}

	reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)
	database.ClearState(userID)

	if text == answer {
		result := srs.Calculate(reps, interval, ease, 5)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		database.MarkReviewDone(userID, tzOffset)
		return c.Send("\u2705 Correct! Great recall! \U0001F389")
	}

	result := srs.Calculate(reps, interval, ease, 1)
	database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)

	word, _ := database.GetWordByID(wordID)
	if word != nil {
		return c.Send(fmt.Sprintf("\u274C Not quite.\n\nCorrect answer: *%s*\n(%s)\n\nYou'll see this word again soon!",
			escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}
	return c.Send("\u274C Not quite. You'll see this word again soon!")
}

func processQuizSentence(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	userID := c.Sender().ID
	targetWord := state.Context["word"]

	var wordID int
	fmt.Sscanf(state.Context["word_id"], "%d", &wordID)

	if len(text) < 5 {
		return c.Send("Write a full sentence, please!")
	}

	if !strings.Contains(strings.ToLower(text), strings.ToLower(targetWord)) {
		return c.Send(fmt.Sprintf("\u274C Please use *%s* in your sentence!", escapeMarkdown(targetWord)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	database.ClearState(userID)

	user, _ := database.GetUser(userID)
	tzOffset := 5
	if user != nil {
		tzOffset = user.TzOffset
	}

	reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)
	result := srs.Calculate(reps, interval, ease, 4)
	database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
	database.MarkReviewDone(userID, tzOffset)

	if openaiClient != nil {
		feedback, err := openaiClient.CheckSentences(text, targetWord)
		if err == nil {
			return c.Send("\u2705 Great sentence!\n\n"+feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
	}

	return c.Send("\u2705 Great sentence! Keep practicing! \U0001F4AA")
}

func processMediaTask(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState, text string) error {
	userID := c.Sender().ID

	var mediaID int
	fmt.Sscanf(state.Context["media_id"], "%d", &mediaID)
	mediaTitle := state.Context["media_title"]

	if len(text) < 10 {
		return c.Send("Write at least a few sentences!")
	}

	database.ClearState(userID)

	user, _ := database.GetUser(userID)
	tzOffset := 5
	if user != nil {
		tzOffset = user.TzOffset
	}

	database.SaveMediaTaskResponse(userID, mediaID, text)
	database.MarkReviewDone(userID, tzOffset)

	c.Send("\u2705 Saved! Checking your sentences...")

	feedback, err := openaiClient.CheckSentences(text, mediaTitle)
	if err != nil {
		log.Printf("[user=%d] AI media feedback error: %v", userID, err)
		return c.Send("Good job! Keep watching and writing! \U0001F4AA")
	}

	return c.Send(feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
