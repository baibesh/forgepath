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
	word, err := database.GetRandomUnseen(user.ID, user.Level, user.Language)
	if err != nil {
		return c.Send(m.AllWordsLearned)
	}

	grammar, _ := database.GetCurrentGrammarFocus(user.ID)
	database.MarkWordSeen(user.ID, word.ID)
	database.MarkWordDone(user.ID, user.TzOffset)

	if err := c.Send(FormatWordOfDay(word, grammar, lang), &tele.SendOptions{
		ParseMode:   tele.ModeMarkdown,
		ReplyMarkup: ListenKeyboard(word.ID, lang),
	}); err != nil {
		log.Printf("[user=%d] send word error: %v", user.ID, err)
		return err
	}

	return sendQuizForWord(c, database, word, openaiClient)
}

func sendQuizForWord(c tele.Context, database *db.DB, word *db.Word, openaiClient *ai.OpenAIClient) error {
	user, _ := database.GetUser(c.Sender().ID)
	lang := userLang(user)
	reps, _ := database.GetUserWordRepetitions(c.Sender().ID, word.ID)

	if reps >= 4 {
		database.SetState(c.Sender().ID, "waiting_quiz_sentence", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"word":    word.Word,
		})
		return c.Send(FormatQuizMakeSentence(word, lang), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	if reps >= 3 {
		database.SetState(c.Sender().ID, "waiting_quiz_typing", map[string]string{
			"word_id": fmt.Sprintf("%d", word.ID),
			"answer":  strings.ToLower(word.Word),
		})
		return c.Send(FormatQuizTypeWord(word, lang), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	return SendQuizPoll(c.Bot(), c.Recipient(), c.Sender().ID, word, openaiClient, userLang(user))
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
	words, err := database.GetWordsForReview(user.ID, 3)
	if err != nil || len(words) == 0 {
		return c.Send(m.NothingToReview)
	}

	sentTypingQuiz := false
	for _, w := range words {
		word := w
		reps, _ := database.GetUserWordRepetitions(user.ID, word.ID)
		if reps >= 3 && !sentTypingQuiz {
			if err := sendQuizForWord(c, database, &word, openaiClient); err != nil {
				log.Printf("[user=%d] quiz error word=%d: %v", user.ID, word.ID, err)
			}
			sentTypingQuiz = true
			continue
		}
		if reps < 3 {
			if err := SendQuizPoll(c.Bot(), c.Recipient(), user.ID, &word, openaiClient, lang); err != nil {
				log.Printf("[user=%d] quiz error word=%d: %v", user.ID, word.ID, err)
			}
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
	grammar = GrammarOrDefault(grammar, user.Language)

	topic := content.RandomTopic(user.Language)

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
