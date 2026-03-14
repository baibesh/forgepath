package bot

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/config"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

func RegisterHandlers(b *tele.Bot, database *db.DB, cfg *config.Config) {
	openaiClient := ai.NewOpenAIClient(cfg.OpenAIKey)

	// /start — onboarding or welcome back
	b.Handle("/start", func(c tele.Context) error {
		user := c.Sender()
		if err := database.CreateUser(user.ID, user.Username, user.FirstName); err != nil {
			log.Printf("Error creating user %d: %v", user.ID, err)
			return c.Send("Something went wrong. Try again later.")
		}

		existing, err := database.GetUser(user.ID)
		if err == nil && existing.Level != "" && existing.TzOffset != 0 {
			welcome := fmt.Sprintf(
				"Welcome back, %s! 👋\n\n"+
					"Your level: *%s*\n"+
					"Timezone: UTC+%d\n\n"+
					"Use the menu below to continue learning!",
				user.FirstName, existing.Level, existing.TzOffset,
			)
			return c.Send(welcome, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: MainMenu()})
		}

		// New user — start onboarding
		database.SetState(user.ID, "onboarding_level", map[string]string{})

		welcome := fmt.Sprintf(
			"Hey, %s! 👋\n\n"+
				"Welcome to *ForgePath* — your daily English learning companion.\n\n"+
				"📖 Morning — Word of the Day + Quiz\n"+
				"✍️ Afternoon — Free Writing (5 min)\n"+
				"🎬 Evening — Media Recommendation\n"+
				"📊 Night — Daily Review\n\n"+
				"Let's set you up! What's your English level?",
			user.FirstName,
		)
		return c.Send(welcome, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: LevelSelectKeyboard()})
	})

	// /today — show current/next task
	b.Handle("/today", func(c tele.Context) error {
		return handleToday(c, database)
	})

	// /word — manual word of day
	b.Handle("/word", func(c tele.Context) error {
		return handleWord(c, database, openaiClient)
	})

	// /quiz — manual quiz
	b.Handle("/quiz", func(c tele.Context) error {
		return handleQuiz(c, database, openaiClient)
	})

	// /write — manual writing prompt
	b.Handle("/write", func(c tele.Context) error {
		return handleWrite(c, database)
	})

	// /stats
	b.Handle("/stats", func(c tele.Context) error {
		return handleStats(c, database)
	})

	// /skip
	b.Handle("/skip", func(c tele.Context) error {
		return handleSkip(c, database)
	})

	// /settings
	b.Handle("/settings", func(c tele.Context) error {
		return c.Send("⚙️ *Settings*\n\nWhat would you like to change?",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard()})
	})

	// /help
	b.Handle("/help", func(c tele.Context) error {
		return c.Send(
			"📚 *ForgePath Help*\n\n"+
				"*Daily Schedule:*\n"+
				"07:30 — 📖 Word of the Day + Quiz\n"+
				"12:00 — ✍️ Free Writing (5 min)\n"+
				"18:00 — 🎬 Media Recommendation\n"+
				"21:30 — 📊 Daily Review\n\n"+
				"*Commands:*\n"+
				"/today — see current task\n"+
				"/word — get word of the day\n"+
				"/quiz — start a quiz\n"+
				"/write — start free writing\n"+
				"/stats — view your progress\n"+
				"/skip — skip today (max 2/week)\n"+
				"/words — your learned words\n"+
				"/settings — change timezone/level\n"+
				"/help — this message\n\n"+
				"*How it works:*\n"+
				"Each week focuses on one grammar tense.\n"+
				"Words come with constructions + collocations.\n"+
				"Quiz adapts to your recall level (SRS).\n"+
				"Writing gets AI feedback.\n"+
				"Stay consistent, build your streak! 🔥",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	})

	// /words — learned words list
	b.Handle("/words", func(c tele.Context) error {
		return handleWordsList(c, database)
	})

	// Reply keyboard buttons
	b.Handle(&tele.Btn{Text: "📖 Today"}, func(c tele.Context) error {
		return handleToday(c, database)
	})
	b.Handle(&tele.Btn{Text: "📊 Stats"}, func(c tele.Context) error {
		return handleStats(c, database)
	})
	b.Handle(&tele.Btn{Text: "⚙️ Settings"}, func(c tele.Context) error {
		return c.Send("⚙️ *Settings*",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SettingsKeyboard()})
	})

	// Callbacks
	RegisterCallbacks(b, database, openaiClient)

	// Text handler — FSM router
	b.Handle(tele.OnText, func(c tele.Context) error {
		return handleText(c, database, openaiClient)
	})
}

func handleToday(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	streak, _ := database.GetTodayStreak(userID)

	var tasks []string
	if !streak.WordDone {
		tasks = append(tasks, "📖 Word of the Day — /word")
	}
	if !streak.WritingDone {
		tasks = append(tasks, "✍️ Free Writing — /write")
	}
	if !streak.ReviewDone {
		tasks = append(tasks, "📝 Quiz Review — /quiz")
	}

	if len(tasks) == 0 {
		return c.Send("✅ *All done for today!*\n\nGreat job! See you tomorrow 💪",
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	msg := "📋 *Today's Tasks:*\n\n" + strings.Join(tasks, "\n")
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleWord(c tele.Context, database *db.DB, gemini *ai.OpenAIClient) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}

	word, err := database.GetRandomUnseen(userID, user.Level)
	if err != nil {
		return c.Send("No new words available right now. You've learned them all! 🎉")
	}

	grammar, _ := database.GetCurrentGrammarFocus(userID)
	database.MarkWordSeen(userID, word.ID)
	database.MarkWordDone(userID)

	msg := FormatWordOfDay(word, grammar)

	// Send word first
	if err := c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown}); err != nil {
		return err
	}

	// Then send a quiz for review words
	return sendQuizForWord(c, database, word, gemini)
}

func sendQuizForWord(c tele.Context, database *db.DB, word *db.Word, gemini *ai.OpenAIClient) error {
	reps, _ := database.GetUserWordRepetitions(c.Sender().ID, word.ID)

	switch {
	case reps >= 3:
		// Active recall: type the word
		database.SetState(c.Sender().ID, "waiting_quiz_typing", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"answer":  strings.ToLower(word.Word),
		})
		return c.Send(FormatQuizTypeWord(word), &tele.SendOptions{ParseMode: tele.ModeMarkdown})

	case reps >= 4:
		// Production: make a sentence
		database.SetState(c.Sender().ID, "waiting_quiz_sentence", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"word":    word.Word,
		})
		return c.Send(FormatQuizMakeSentence(word), &tele.SendOptions{ParseMode: tele.ModeMarkdown})

	default:
		// Fill-in-blank quiz
		options := []string{word.Definition}
		wrongOptions, _ := gemini.GenerateQuizOptions(word.Word, word.Definition, 3)
		options = append(options, wrongOptions...)

		// Shuffle
		correctIdx := 0
		rand.Shuffle(len(options), func(i, j int) {
			options[i], options[j] = options[j], options[i]
			if i == correctIdx {
				correctIdx = j
			} else if j == correctIdx {
				correctIdx = i
			}
		})

		// Find correct index after shuffle
		for i, opt := range options {
			if opt == word.Definition {
				correctIdx = i
				break
			}
		}

		msg := FormatQuizFillBlank(word, options, correctIdx)
		kb := QuizKeyboard(word.ID, options, correctIdx)
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: kb})
	}
}

func handleQuiz(c tele.Context, database *db.DB, gemini *ai.OpenAIClient) error {
	userID := c.Sender().ID
	words, err := database.GetWordsForReview(userID, 3)
	if err != nil || len(words) == 0 {
		return c.Send("No words to review right now! Learn some with /word first.")
	}

	for _, w := range words {
		word := w
		if err := sendQuizForWord(c, database, &word, gemini); err != nil {
			log.Printf("Quiz error for word %d: %v", word.ID, err)
		}
	}
	return nil
}

func handleWrite(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	grammar, _ := database.GetCurrentGrammarFocus(userID)
	if grammar == nil {
		grammar = &db.GrammarWeek{TenseName: "Past Simple", Anchor: "🚪 Закрытая дверь", Formula: "S + V2", Markers: "yesterday, last week"}
	}

	topics := []string{
		"What did you do last weekend?",
		"Describe your morning routine.",
		"Tell about your favorite movie.",
		"What would you like to learn?",
		"Describe a person you admire.",
		"What did you eat yesterday?",
		"Tell about your best trip.",
		"What makes you happy?",
		"Describe your workplace.",
		"What are your plans for this week?",
	}
	topic := topics[rand.Intn(len(topics))]

	database.SetState(userID, "waiting_writing", map[string]string{
		"topic":         topic,
		"grammar_focus": grammar.TenseName,
	})

	msg := FormatWritingPrompt(topic, grammar.TenseName, grammar)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleStats(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	streak, _ := database.GetCurrentStreak(userID)
	wordCount, _ := database.GetUserWordCount(userID)
	writingCount, _ := database.GetUserWritingCount(userID)
	grammar, _ := database.GetCurrentGrammarFocus(userID)
	weekly, _ := database.GetWeeklyStats(userID)

	msg := FormatStats(streak, wordCount, writingCount, grammar, weekly)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleSkip(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	user, err := database.GetUser(userID)
	if err != nil {
		return c.Send("Please /start first!")
	}

	if user.SkipCount >= 2 {
		return c.Send("❌ You've already used both skips this week.\nKeep going! 💪")
	}

	msg := fmt.Sprintf("⏭ *Skip Today?*\n\nYou have *%d/2* skips left this week.", 2-user.SkipCount)
	return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SkipConfirmKeyboard()})
}

func handleWordsList(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	words, err := database.GetUserWords(userID, 0, 20)
	if err != nil || len(words) == 0 {
		return c.Send("You haven't learned any words yet. Start with /word!")
	}

	var sb strings.Builder
	sb.WriteString("📚 *Your Words:*\n\n")
	for i, w := range words {
		sb.WriteString(fmt.Sprintf("%d. *%s* — %s\n", i+1, escapeMarkdown(w.Word), escapeMarkdown(w.Definition)))
	}

	count, _ := database.GetUserWordCount(userID)
	if count > 20 {
		sb.WriteString(fmt.Sprintf("\n_...and %d more_", count-20))
	}

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleText routes text messages based on FSM state
func handleText(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	userID := c.Sender().ID
	state, err := database.GetState(userID)
	if err != nil {
		return nil
	}

	switch state.State {
	case "onboarding_tz_custom":
		return handleOnboardingTzCustom(c, database)

	case "waiting_writing":
		return handleWritingSubmit(c, database, openaiClient, state)

	case "waiting_quiz_typing":
		return handleQuizTyping(c, database, state)

	case "waiting_quiz_sentence":
		return handleQuizSentence(c, database, openaiClient, state)

	case "waiting_media_task":
		return handleMediaTaskSubmit(c, database, openaiClient, state)

	default:
		return nil
	}
}

func handleOnboardingTzCustom(c tele.Context, database *db.DB) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())

	var offset int
	_, err := fmt.Sscanf(text, "%d", &offset)
	if err != nil || offset < -12 || offset > 14 {
		return c.Send("Please enter a number between -12 and 14:")
	}

	database.UpdateUserTimezone(userID, offset)
	database.ClearState(userID)

	return c.Send(
		fmt.Sprintf("✅ Setup complete!\n\nTimezone: UTC+%d\nYou're all set! 🚀\n\nUse the menu or /help to see commands.", offset),
		&tele.SendOptions{ReplyMarkup: MainMenu()})
}

func handleWritingSubmit(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())

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
		log.Printf("Error saving writing for user %d: %v", userID, err)
		return c.Send("Error saving your writing. Try again.")
	}

	database.MarkWritingDone(userID)
	database.ClearState(userID)

	c.Send(fmt.Sprintf("✅ Saved! (%d words)\n\nAnalyzing...", wordCount))

	user, _ := database.GetUser(userID)
	level := "A2"
	if user != nil {
		level = user.Level
	}

	feedback, err := openaiClient.CheckWriting(text, grammarFocus, level)
	if err != nil {
		log.Printf("AI feedback error for writing %d: %v", writingID, err)
		feedback = "AI feedback is not available right now. Keep writing!"
	}

	database.UpdateWritingFeedback(writingID, feedback)
	return c.Send(feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleQuizTyping(c tele.Context, database *db.DB, state *db.UserState) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(strings.ToLower(c.Text()))
	answer := state.Context["answer"]

	var wordID int
	fmt.Sscanf(state.Context["word_id"], "%d", &wordID)

	database.ClearState(userID)

	if text == answer {
		// Correct — score 5
		result := srs.Calculate(0, 1, 2.5, 5)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		database.MarkReviewDone(userID)
		return c.Send("✅ Correct! Great recall! 🎉")
	}

	// Wrong — score 1
	result := srs.Calculate(0, 1, 2.5, 1)
	database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)

	word, _ := database.GetWordByID(wordID)
	if word != nil {
		return c.Send(fmt.Sprintf("❌ Not quite.\n\nCorrect answer: *%s*\n(%s)\n\nYou'll see this word again soon!", escapeMarkdown(word.Word), escapeMarkdown(word.Definition)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}
	return c.Send("❌ Not quite. You'll see this word again soon!")
}

func handleQuizSentence(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())
	targetWord := state.Context["word"]

	var wordID int
	fmt.Sscanf(state.Context["word_id"], "%d", &wordID)

	database.ClearState(userID)

	if len(text) < 5 {
		return c.Send("Write a full sentence, please!")
	}

	// Check if they used the word
	if !strings.Contains(strings.ToLower(text), strings.ToLower(targetWord)) {
		return c.Send(fmt.Sprintf("❌ Please use *%s* in your sentence!", escapeMarkdown(targetWord)),
			&tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Score 4 for effort
	result := srs.Calculate(0, 1, 2.5, 4)
	database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
	database.MarkReviewDone(userID)

	// AI feedback if available
	if openaiClient != nil {
		feedback, err := openaiClient.CheckSentences(text, targetWord)
		if err == nil {
			return c.Send("✅ Great sentence!\n\n"+feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
	}

	return c.Send("✅ Great sentence! Keep practicing! 💪")
}

func handleMediaTaskSubmit(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient, state *db.UserState) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())

	var mediaID int
	fmt.Sscanf(state.Context["media_id"], "%d", &mediaID)
	mediaTitle := state.Context["media_title"]

	database.ClearState(userID)

	if len(text) < 10 {
		return c.Send("Write at least a few sentences!")
	}

	database.SaveMediaTaskResponse(userID, mediaID, text)
	database.MarkReviewDone(userID)

	c.Send("✅ Saved! Checking your sentences...")

	feedback, err := openaiClient.CheckSentences(text, mediaTitle)
	if err != nil {
		return c.Send("Good job! Keep watching and writing! 💪")
	}

	return c.Send(feedback, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
