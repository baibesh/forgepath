package bot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/config"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
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

	b.Handle("/today", func(c tele.Context) error {
		log.Printf("[user=%d] /today", c.Sender().ID)
		return handleToday(c, database)
	})

	b.Handle("/word", func(c tele.Context) error {
		log.Printf("[user=%d] /word", c.Sender().ID)
		return handleWord(c, database, openaiClient)
	})

	b.Handle("/quiz", func(c tele.Context) error {
		log.Printf("[user=%d] /quiz", c.Sender().ID)
		return handleQuiz(c, database, openaiClient)
	})

	b.Handle("/write", func(c tele.Context) error {
		log.Printf("[user=%d] /write", c.Sender().ID)
		return handleWrite(c, database)
	})

	b.Handle("/stats", func(c tele.Context) error {
		log.Printf("[user=%d] /stats", c.Sender().ID)
		return handleStats(c, database)
	})

	b.Handle("/skip", func(c tele.Context) error {
		log.Printf("[user=%d] /skip", c.Sender().ID)
		return handleSkip(c, database)
	})

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
		log.Printf("[user=%d] /settings", c.Sender().ID)
		return c.Send("\u2699\uFE0F *Settings*\n\nWhat would you like to change?",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard()})
	})

	b.Handle("/help", func(c tele.Context) error {
		log.Printf("[user=%d] /help", c.Sender().ID)
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

	b.Handle("/words", func(c tele.Context) error {
		log.Printf("[user=%d] /words", c.Sender().ID)
		return handleWordsList(c, database)
	})

	RegisterCallbacks(b, database, openaiClient)

	b.Handle(tele.OnText, func(c tele.Context) error {
		return handleText(c, database, openaiClient)
	})

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		return handleVoice(c, b, database, openaiClient)
	})
}

func handleToday(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}
	streak, _ := database.GetTodayStreak(userID, user.TzOffset)

	var tasks []string
	if !streak.WordDone {
		tasks = append(tasks, "\U0001F4D6 Word of the Day — /word")
	}
	if !streak.WritingDone {
		tasks = append(tasks, "\u270D\uFE0F Free Writing — /write")
	}
	if !streak.ReviewDone {
		tasks = append(tasks, "\U0001F4DD Quiz Review — /quiz")
	}

	if len(tasks) == 0 {
		return c.Send("\u2705 *All done for today!*\n\nGreat job! See you tomorrow \U0001F4AA",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	return c.Send("\U0001F4CB *Today's Tasks:*\n\n"+strings.Join(tasks, "\n"),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleWord(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}

	word, err := database.GetRandomUnseen(userID, user.Level, user.Language)
	if err != nil {
		return c.Send("No new words available right now. You've learned them all! \U0001F389")
	}

	grammar, _ := database.GetCurrentGrammarFocus(userID)
	database.MarkWordSeen(userID, word.ID)
	database.MarkWordDone(userID, user.TzOffset)

	if err := c.Send(FormatWordOfDay(word, grammar), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdown,
		ReplyMarkup: ListenKeyboard(word.ID),
	}); err != nil {
		log.Printf("[user=%d] send word error: %v", userID, err)
		return err
	}

	return sendQuizForWord(c, database, word, openaiClient)
}

func sendQuizForWord(c tele.Context, database *db.DB, word *db.Word, openaiClient *ai.OpenAIClient) error {
	reps, _ := database.GetUserWordRepetitions(c.Sender().ID, word.ID)

	if reps >= 4 {
		database.SetState(c.Sender().ID, "waiting_quiz_sentence", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"word":    word.Word,
		})
		return c.Send(FormatQuizMakeSentence(word), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	if reps >= 3 {
		database.SetState(c.Sender().ID, "waiting_quiz_typing", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"answer":  strings.ToLower(word.Word),
		})
		return c.Send(FormatQuizTypeWord(word), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	options := []string{word.Definition}
	wrongOptions, _ := openaiClient.GenerateQuizOptions(word.Word, word.Definition, 3)
	options = append(options, wrongOptions...)

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	var correctIdx int
	for i, opt := range options {
		if opt == word.Definition {
			correctIdx = i
			break
		}
	}

	return c.Send(FormatQuizFillBlank(word, options, correctIdx),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: QuizKeyboard(word.ID, options, correctIdx)})
}

func handleQuiz(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	words, err := database.GetWordsForReview(userID, 3)
	if err != nil || len(words) == 0 {
		return c.Send("No words to review right now! Learn some with /word first.")
	}

	for _, w := range words {
		word := w
		if err := sendQuizForWord(c, database, &word, openaiClient); err != nil {
			log.Printf("[user=%d] quiz error word=%d: %v", userID, word.ID, err)
		}
	}
	return nil
}

func handleWrite(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}

	grammar, _ := database.GetCurrentGrammarFocus(userID)
	if grammar == nil {
		grammar = db.DefaultGrammar(user.Language)
	}

	topic := content.RandomTopic(user.Language)

	database.SetState(userID, "waiting_writing", map[string]string{
		"topic":         topic,
		"grammar_focus": grammar.TenseName,
	})

	return c.Send(FormatWritingPrompt(topic, grammar.TenseName, grammar, user.Language), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleStats(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}

	streak, _ := database.GetCurrentStreak(userID, user.TzOffset)
	wordCount, _ := database.GetUserWordCount(userID)
	writingCount, _ := database.GetUserWritingCount(userID)
	grammar, _ := database.GetCurrentGrammarFocus(userID)
	weekly, _ := database.GetWeeklyStats(userID, user.TzOffset)

	return c.Send(FormatStats(streak, wordCount, writingCount, grammar, weekly),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleSkip(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}

	if user.SkipCount >= 2 {
		return c.Send("\u274C You've already used both skips this week.\nKeep going! \U0001F4AA")
	}

	return c.Send(fmt.Sprintf("\u23ED *Skip Today?*\n\nYou have *%d/2* skips left this week.", 2-user.SkipCount),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SkipConfirmKeyboard()})
}

func handleWordsList(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	words, err := database.GetUserWords(userID, 0, 20)
	if err != nil || len(words) == 0 {
		return c.Send("You haven't learned any words yet. Start with /word!")
	}

	var sb strings.Builder
	sb.WriteString("\U0001F4DA *Your Words:*\n\n")
	for i, w := range words {
		sb.WriteString(fmt.Sprintf("%d. *%s* — %s\n", i+1, escapeMarkdown(w.Word), escapeMarkdown(w.Definition)))
	}

	count, _ := database.GetUserWordCount(userID)
	if count > 20 {
		sb.WriteString(fmt.Sprintf("\n_...and %d more_", count-20))
	}

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

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
