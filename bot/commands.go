package bot

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
	"github.com/baibesh/forgepath/srs"
)

// PickQuizType selects a quiz format based on repetition count with some randomness.
func PickQuizType(reps int) string {
	switch {
	case reps >= 4:
		return "sentence"
	case reps == 3:
		choices := []string{"typing", "sentence"}
		return choices[rand.Intn(len(choices))]
	case reps == 2:
		choices := []string{"definition", "cloze", "reverse", "collocation"}
		return choices[rand.Intn(len(choices))]
	case reps == 1:
		choices := []string{"definition", "cloze", "reverse", "truefalse"}
		return choices[rand.Intn(len(choices))]
	default: // reps == 0
		choices := []string{"definition", "cloze", "truefalse"}
		return choices[rand.Intn(len(choices))]
	}
}

func requireOnboarded(c tele.Context, database *db.DB) (*db.User, error) {
	user, err := database.GetUser(c.Sender().ID)
	if err != nil {
		c.Send(content.GetMessages("en").NotStarted)
		return nil, err
	}
	if !user.Onboarded {
		m := userMessages(user)
		c.Send(m.FinishSetup)
		return nil, fmt.Errorf("not onboarded")
	}
	return user, nil
}

func warnActiveState(c tele.Context, database *db.DB, user *db.User) bool {
	state, _ := database.GetState(c.Sender().ID)
	if state.State != "idle" && state.State != "" {
		m := userMessages(user)
		c.Send(m.ActiveTask)
		return true
	}
	return false
}

func handleToday(c tele.Context, database *db.DB) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}
	m := userMessages(user)
	streak, _ := database.GetTodayStreak(user.ID, user.TzOffset)

	var tasks []string
	if !streak.WordDone {
		tasks = append(tasks, m.TodayWord)
	}
	if !streak.WritingDone {
		tasks = append(tasks, m.TodayWriting)
	}
	if !streak.ReviewDone {
		tasks = append(tasks, m.TodayQuiz)
	}

	if len(tasks) == 0 {
		return c.Send(m.TodayAllDone, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	return c.Send(m.TodayLeft+strings.Join(tasks, "\n"),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleWord(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	if warnActiveState(c, database, user) {
		return nil
	}

	m := userMessages(user)
	lang := userLang(user)

	count := user.WordsPerDay
	if count <= 0 {
		count = 3
	}

	grammar, _ := database.GetCurrentGrammarFocus(user.ID)
	sentCount := 0

	var lastWord *db.Word
	for i := 0; i < count; i++ {
		word, err := database.GetRandomUnseen(user.ID, user.Level, user.TargetLanguage)
		if err != nil {
			if i == 0 {
				return c.Send(m.AllWordsLearned)
			}
			break
		}

		database.MarkWordSeen(user.ID, word.ID)

		if err := c.Send(FormatWordOfDay(word, grammar, lang), &tele.SendOptions{
			ParseMode:   tele.ModeMarkdown,
			ReplyMarkup: ListenKeyboard(word.ID, lang),
		}); err != nil {
			log.Printf("[user=%d] send word error: %v", user.ID, err)
		}

		lastWord = word
		sentCount++
	}

	if sentCount > 0 {
		database.MarkWordDone(user.ID, user.TzOffset)
	}

	// Send quiz for the last word
	if lastWord != nil {
		return sendQuizForWord(c, database, lastWord, openaiClient)
	}
	return nil
}

func sendQuizForWord(c tele.Context, database *db.DB, word *db.Word, openaiClient *ai.OpenAIClient) error {
	user, _ := database.GetUser(c.Sender().ID)
	lang := userLang(user)
	reps, _ := database.GetUserWordRepetitions(c.Sender().ID, word.ID)

	quizType := PickQuizType(reps)
	return DispatchQuiz(c.Bot(), c.Recipient(), c.Sender().ID, database, word, openaiClient, lang, quizType)
}

// DispatchQuiz sends the appropriate quiz type, falling back to definition poll on error.
func DispatchQuiz(b *tele.Bot, recipient tele.Recipient, userID int64, database *db.DB, word *db.Word, openaiClient *ai.OpenAIClient, lang, quizType string) error {
	switch quizType {
	case "sentence":
		database.SetState(userID, "waiting_quiz_sentence", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"word":    word.Word,
		})
		_, err := b.Send(recipient, FormatQuizMakeSentence(word, lang), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		return err

	case "typing":
		database.SetState(userID, "waiting_quiz_typing", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"answer":  strings.ToLower(word.Word),
		})
		_, err := b.Send(recipient, FormatQuizTypeWord(word, lang), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		return err

	case "cloze":
		if err := SendClozeQuizPoll(b, recipient, userID, word, openaiClient, lang); err != nil {
			log.Printf("[user=%d] cloze quiz failed, fallback to definition: %v", userID, err)
			return SendQuizPoll(b, recipient, userID, word, openaiClient, lang)
		}
		return nil

	case "reverse":
		if err := SendReverseQuizPoll(b, recipient, userID, word, openaiClient, lang); err != nil {
			log.Printf("[user=%d] reverse quiz failed, fallback to definition: %v", userID, err)
			return SendQuizPoll(b, recipient, userID, word, openaiClient, lang)
		}
		return nil

	case "collocation":
		if word.Collocations == "" {
			return SendQuizPoll(b, recipient, userID, word, openaiClient, lang)
		}
		if err := SendCollocationQuizPoll(b, recipient, userID, word, openaiClient, lang); err != nil {
			log.Printf("[user=%d] collocation quiz failed, fallback to definition: %v", userID, err)
			return SendQuizPoll(b, recipient, userID, word, openaiClient, lang)
		}
		return nil

	case "truefalse":
		return SendTrueFalseQuiz(b, recipient, userID, word, database, lang)

	default:
		return SendQuizPoll(b, recipient, userID, word, openaiClient, lang)
	}
}

func SendQuizPoll(b *tele.Bot, recipient tele.Recipient, userID int64, word *db.Word, openaiClient *ai.OpenAIClient, lang string) error {
	m := content.GetMessages(lang)
	options := []string{word.Definition}
	wrongOptions, _ := openaiClient.GenerateQuizOptions(word.Word, word.Definition, word.Language, 3)
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

	poll := &tele.Poll{
		Type:          tele.PollQuiz,
		Question:      m.QuizPollQuestion(word.Word),
		CorrectOption: correctIdx,
		Anonymous:     false,
		Explanation:   fmt.Sprintf("%s — %s\n%s", word.Word, word.Definition, word.Example),
	}
	poll.AddOptions(options...)

	msg, err := poll.Send(b, recipient, nil)
	if err != nil {
		return err
	}

	if msg != nil && msg.Poll != nil {
		RegisterQuizPoll(msg.Poll.ID, userID, word.ID, correctIdx)
	}
	return nil
}

// SendClozeQuizPoll sends a fill-in-the-blank quiz using Telegram PollQuiz.
func SendClozeQuizPoll(b *tele.Bot, recipient tele.Recipient, userID int64, word *db.Word, openaiClient *ai.OpenAIClient, lang string) error {
	m := content.GetMessages(lang)
	sentence, wrongWords, err := openaiClient.GenerateClozeOptions(word.Word, word.Definition, word.Language)
	if err != nil {
		return err
	}

	options := []string{word.Word}
	options = append(options, wrongWords...)

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	var correctIdx int
	for i, opt := range options {
		if opt == word.Word {
			correctIdx = i
			break
		}
	}

	poll := &tele.Poll{
		Type:          tele.PollQuiz,
		Question:      m.QuizClozeQuestion(sentence),
		CorrectOption: correctIdx,
		Anonymous:     false,
		Explanation:   fmt.Sprintf("%s — %s", word.Word, word.Definition),
	}
	poll.AddOptions(options...)

	msg, err := poll.Send(b, recipient, nil)
	if err != nil {
		return err
	}
	if msg != nil && msg.Poll != nil {
		RegisterQuizPoll(msg.Poll.ID, userID, word.ID, correctIdx)
	}
	return nil
}

// SendReverseQuizPoll shows definition and asks to pick the correct word.
func SendReverseQuizPoll(b *tele.Bot, recipient tele.Recipient, userID int64, word *db.Word, openaiClient *ai.OpenAIClient, lang string) error {
	m := content.GetMessages(lang)

	// Generate 3 wrong word options (similar-level words, not definitions)
	wrongWords, err := openaiClient.GenerateQuizOptions(word.Word, word.Word, word.Language, 3)
	if err != nil {
		return err
	}

	options := []string{word.Word}
	options = append(options, wrongWords[:3]...)

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	var correctIdx int
	for i, opt := range options {
		if opt == word.Word {
			correctIdx = i
			break
		}
	}

	poll := &tele.Poll{
		Type:          tele.PollQuiz,
		Question:      m.QuizReverseQuestion(word.Definition),
		CorrectOption: correctIdx,
		Anonymous:     false,
		Explanation:   fmt.Sprintf("%s — %s", word.Word, word.Definition),
	}
	poll.AddOptions(options...)

	msg, sendErr := poll.Send(b, recipient, nil)
	if sendErr != nil {
		return sendErr
	}
	if msg != nil && msg.Poll != nil {
		RegisterQuizPoll(msg.Poll.ID, userID, word.ID, correctIdx)
	}
	return nil
}

// SendCollocationQuizPoll sends a collocation quiz using PollQuiz.
func SendCollocationQuizPoll(b *tele.Bot, recipient tele.Recipient, userID int64, word *db.Word, openaiClient *ai.OpenAIClient, lang string) error {
	m := content.GetMessages(lang)

	_, options, _, err := openaiClient.GenerateCollocationQuiz(word.Word, word.Collocations, word.Language)
	if err != nil {
		return err
	}

	correctAnswer := options[0] // first is always correct from GenerateCollocationQuiz

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	var correctIdx int
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	poll := &tele.Poll{
		Type:          tele.PollQuiz,
		Question:      m.QuizCollocationQuestion(word.Word),
		CorrectOption: correctIdx,
		Anonymous:     false,
		Explanation:   fmt.Sprintf("%s: %s", word.Word, word.Collocations),
	}
	poll.AddOptions(options...)

	msg, sendErr := poll.Send(b, recipient, nil)
	if sendErr != nil {
		return sendErr
	}
	if msg != nil && msg.Poll != nil {
		RegisterQuizPoll(msg.Poll.ID, userID, word.ID, correctIdx)
	}
	return nil
}

// SendTrueFalseQuiz sends a True/False poll (regular, not quiz type).
func SendTrueFalseQuiz(b *tele.Bot, recipient tele.Recipient, userID int64, word *db.Word, database *db.DB, lang string) error {
	m := content.GetMessages(lang)

	// 50% chance show correct definition, 50% show wrong one
	showCorrect := rand.Intn(2) == 0
	shownDef := word.Definition

	if !showCorrect {
		// Get a random word to use its definition as the wrong one
		wrongWords, err := database.GetRandomWordsExcluding(word.ID, word.Level, 1)
		if err == nil && len(wrongWords) > 0 {
			shownDef = wrongWords[0].Definition
		} else {
			showCorrect = true
			shownDef = word.Definition
		}
	}

	var correctIdx int
	if showCorrect {
		correctIdx = 0 // True
	} else {
		correctIdx = 1 // False
	}

	poll := &tele.Poll{
		Type:          tele.PollQuiz,
		Question:      m.QuizTrueFalseQuestion(word.Word, shownDef),
		CorrectOption: correctIdx,
		Anonymous:     false,
		Explanation:   fmt.Sprintf("%s — %s", word.Word, word.Definition),
	}
	poll.AddOptions(m.LabelTrue, m.LabelFalse)

	msg, err := poll.Send(b, recipient, nil)
	if err != nil {
		return err
	}
	if msg != nil && msg.Poll != nil {
		RegisterQuizPoll(msg.Poll.ID, userID, word.ID, correctIdx)
	}
	return nil
}

func handleQuiz(c tele.Context, database *db.DB, openaiClient *ai.OpenAIClient) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	if warnActiveState(c, database, user) {
		return nil
	}

	m := userMessages(user)
	lang := userLang(user)
	words, err := database.GetWordsForReview(user.ID, 5)
	if err != nil || len(words) == 0 {
		return c.Send(m.NothingToReview)
	}

	sentTextQuiz := false
	for _, w := range words {
		word := w
		reps, _ := database.GetUserWordRepetitions(user.ID, word.ID)
		quizType := PickQuizType(reps)

		// Only one text-input quiz per session (typing or sentence)
		if (quizType == "typing" || quizType == "sentence") && sentTextQuiz {
			quizType = "definition"
		}

		if quizType == "typing" || quizType == "sentence" {
			sentTextQuiz = true
		}

		if err := DispatchQuiz(c.Bot(), c.Recipient(), c.Sender().ID, database, &word, openaiClient, lang, quizType); err != nil {
			log.Printf("[user=%d] quiz error word=%d type=%s: %v", user.ID, word.ID, quizType, err)
		}

		if sentTextQuiz && (quizType == "typing" || quizType == "sentence") {
			break // text quiz needs user response, stop sending more
		}
	}
	return nil
}

func handleWrite(c tele.Context, database *db.DB) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	if warnActiveState(c, database, user) {
		return nil
	}

	grammar, _ := database.GetCurrentGrammarFocus(user.ID)
	grammar = GrammarOrDefault(grammar, user.TargetLanguage)

	topic := content.RandomTopic(user.TargetLanguage)

	database.SetState(user.ID, "waiting_writing", map[string]string{
		"topic":         topic,
		"grammar_focus": grammar.TenseName,
	})

	return c.Send(FormatWritingPrompt(topic, grammar.TenseName, grammar, user.Language), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleStats(c tele.Context, database *db.DB) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	lang := userLang(user)
	streak, _ := database.GetCurrentStreak(user.ID, user.TzOffset)
	wordCount, _ := database.GetUserWordCount(user.ID)
	writingCount, _ := database.GetUserWritingCount(user.ID)
	grammar, _ := database.GetCurrentGrammarFocus(user.ID)
	weekly, _ := database.GetWeeklyStats(user.ID, user.TzOffset)

	return c.Send(FormatStats(streak, wordCount, writingCount, grammar, weekly, lang),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleSkip(c tele.Context, database *db.DB) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	m := userMessages(user)
	lang := userLang(user)
	if user.SkipCount >= 2 {
		return c.Send(m.SkipMaxReached)
	}

	return c.Send(m.SkipConfirm(2-user.SkipCount),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: SkipConfirmKeyboard(lang)})
}

func handleWordsList(c tele.Context, database *db.DB) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	m := userMessages(user)
	words, err := database.GetUserWords(user.ID, 0, 20)
	if err != nil || len(words) == 0 {
		return c.Send(m.NoWordsYet)
	}

	var sb strings.Builder
	sb.WriteString(m.WordsYouKnow)
	for i, w := range words {
		sb.WriteString(fmt.Sprintf("%d. *%s* — %s\n", i+1, escapeMarkdown(w.Word), escapeMarkdown(w.Definition)))
	}

	count, _ := database.GetUserWordCount(user.ID)
	if count > 20 {
		sb.WriteString(m.AndMore(count - 20))
	}

	return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

func handleAddWord(c tele.Context, database *db.DB) error {
	user, err := requireOnboarded(c, database)
	if err != nil {
		return nil
	}

	if warnActiveState(c, database, user) {
		return nil
	}

	m := userMessages(user)
	database.SetState(user.ID, "waiting_addword", map[string]string{})
	return c.Send(m.AddWordPrompt)
}

func handlePollAnswer(c tele.Context, database *db.DB) error {
	answer := c.PollAnswer()
	if answer == nil {
		return nil
	}

	entry, ok := GetQuizPoll(answer.PollID)
	if !ok {
		return nil
	}

	userID := entry.UserID
	wordID := entry.WordID

	user, _ := database.GetUser(userID)
	tzOffset := userTzOffset(user)

	reps, interval, ease, _ := database.GetUserWordSRS(userID, wordID)

	isCorrect := len(answer.Options) > 0 && answer.Options[0] == entry.CorrectIdx
	if isCorrect {
		result := srs.Calculate(reps, interval, ease, 4)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		database.MarkReviewDone(userID, tzOffset)
		log.Printf("[user=%d] quiz poll correct word=%d", userID, wordID)
	} else {
		result := srs.Calculate(reps, interval, ease, 1)
		database.UpdateWordReview(userID, wordID, result.IntervalDays, result.EaseFactor, result.Repetitions)
		log.Printf("[user=%d] quiz poll wrong word=%d", userID, wordID)
	}

	return nil
}
