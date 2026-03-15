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
		Question:      fmt.Sprintf("What does \"%s\" mean?", word.Word),
		CorrectOption: correctIdx,
		Anonymous:     false,
		Explanation:   fmt.Sprintf("%s — %s\n%s", word.Word, word.Definition, word.Example),
	}
	poll.AddOptions(options...)

	msg, err := poll.Send(c.Bot(), c.Recipient(), nil)
	if err != nil {
		return err
	}

	if msg != nil && msg.Poll != nil {
		RegisterQuizPoll(msg.Poll.ID, c.Sender().ID, word.ID, correctIdx)
	}
	return nil
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
	tzOffset := 5
	if user != nil {
		tzOffset = user.TzOffset
	}

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
