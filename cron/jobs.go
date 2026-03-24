package cron

import (
	"fmt"
	"log"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
	"github.com/baibesh/forgepath/bot"
	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
)

type Jobs struct {
	bot    *tele.Bot
	db     *db.DB
	openai *ai.OpenAIClient
}

func NewJobs(b *tele.Bot, database *db.DB, openai *ai.OpenAIClient) *Jobs {
	return &Jobs{bot: b, db: database, openai: openai}
}

func (j *Jobs) DispatchTasks() {
	users, err := j.db.GetActiveUsers()
	if err != nil {
		log.Printf("[cron] error getting users: %v", err)
		return
	}

	now := time.Now().UTC()
	log.Printf("[cron] dispatch: %d active users, UTC %s", len(users), now.Format("15:04"))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, user := range users {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			localHour := ((now.Hour() + user.TzOffset) % 24 + 24) % 24
			localMinute := now.Minute()

			j.checkStateTimeout(user.ID)

			inWindow := func(targetHour, targetMinute int) bool {
				totalLocal := localHour*60 + localMinute
				totalTarget := targetHour*60 + targetMinute
				return totalLocal >= totalTarget && totalLocal < totalTarget+30
			}

			s := user.Schedule

			// Use if instead of switch so multiple tasks can fire in the same cycle
			if inWindow(s.WordHour, s.WordMin) {
				j.safeSend(user, j.sendWordsOfDay)
			}
			if inWindow(s.WritingHour, s.WritingMin) {
				j.safeSend(user, j.sendWritingPrompt)
			}
			if inWindow(s.MediaHour, s.MediaMin) {
				j.safeSend(user, j.sendMediaRecommendation)
			}
			if inWindow(s.ReviewSessionHour, s.ReviewSessionMin) {
				j.safeSend(user, j.sendReviewSession)
			}
			if inWindow(s.ReviewHour, s.ReviewMin) {
				j.safeSend(user, j.sendDailyReview)
			}

			// Use local weekday instead of UTC weekday
			localDay := now.Add(time.Duration(user.TzOffset) * time.Hour).Weekday()
			if localDay == time.Sunday && inWindow(9, 0) {
				j.safeSend(user, j.sendWeeklyReport)
			}
		}()
	}
	wg.Wait()
}

func (j *Jobs) safeSend(user db.User, fn func(db.User)) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[cron][user=%d] panic: %v", user.ID, r)
		}
	}()
	fn(user)
}

func (j *Jobs) checkStateTimeout(userID int64) {
	age, found := j.db.GetStaleStateAge(userID)
	if !found {
		return
	}
	if age > 2*time.Hour {
		log.Printf("[cron][user=%d] clearing stale state (age=%s)", userID, age.Round(time.Minute))
		j.db.ClearState(userID)
	}
}

func (j *Jobs) sendWordsOfDay(user db.User) {
	streak, _ := j.db.GetTodayStreak(user.ID, user.TzOffset)
	if streak.WordDone {
		return
	}

	count := user.WordsPerDay
	if count <= 0 {
		count = 3
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)
	sentCount := 0

	for i := 0; i < count; i++ {
		word, err := j.db.GetRandomUnseen(user.ID, user.Level, user.TargetLanguage)
		if err != nil {
			if i == 0 {
				log.Printf("[cron][user=%d] no unseen words: %v", user.ID, err)
			}
			break
		}

		// Enrich word if missing synonyms/examples
		if word.Synonyms == "" && word.Examples == "" && j.openai != nil {
			synonyms, antonyms, examples, err := j.openai.EnrichWord(word.Word, word.Definition, user.Language)
			if err == nil {
				j.db.UpdateWordEnrichment(word.ID, synonyms, antonyms, examples)
				word.Synonyms = synonyms
				word.Antonyms = antonyms
				word.Examples = examples
			}
		}

		j.db.MarkWordSeen(user.ID, word.ID)
		log.Printf("[cron][user=%d] word of day %d/%d: %s", user.ID, i+1, count, word.Word)
		j.sendMessage(user.ID, bot.FormatWordOfDay(word, grammar, user.Language))
		sentCount++
	}

	if sentCount > 0 {
		j.db.MarkWordDone(user.ID, user.TzOffset)
	}
}

func (j *Jobs) sendReviewSession(user db.User) {
	reviewWords, _ := j.db.GetWordsForReview(user.ID, 10)
	if len(reviewWords) == 0 {
		return
	}

	log.Printf("[cron][user=%d] review session: %d words due", user.ID, len(reviewWords))

	m := content.GetMessages(user.Language)
	j.sendMessage(user.ID, m.LabelReviewTime)

	for _, rw := range reviewWords {
		word := rw
		reps, _ := j.db.GetUserWordRepetitions(user.ID, word.ID)
		quizType := bot.PickQuizType(reps)

		// In cron context, skip text-input quizzes (typing/sentence) since there's no interactive session
		if quizType == "typing" || quizType == "sentence" {
			quizType = "definition"
		}

		recipient := &tele.User{ID: user.ID}
		if err := bot.DispatchQuiz(j.bot, recipient, user.ID, j.db, &word, j.openai, user.Language, quizType); err != nil {
			log.Printf("[cron][user=%d] review quiz error word=%d: %v", user.ID, word.ID, err)
		}
	}
}

func (j *Jobs) sendWritingPrompt(user db.User) {
	streak, _ := j.db.GetTodayStreak(user.ID, user.TzOffset)
	if streak.WritingDone {
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)
	grammar = bot.GrammarOrDefault(grammar, user.TargetLanguage)

	topic := content.RandomTopic(user.TargetLanguage)

	j.db.SetState(user.ID, "waiting_writing", map[string]string{
		"topic":         topic,
		"grammar_focus": grammar.TenseName,
	})

	log.Printf("[cron][user=%d] writing prompt: %s", user.ID, topic)
	j.sendMessage(user.ID, bot.FormatWritingPrompt(topic, grammar.TenseName, grammar, user.Language))
}

func (j *Jobs) sendMediaRecommendation(user db.User) {
	_, err := j.db.GetTodayUnsentMedia(user.ID, user.TzOffset)
	if err == nil {
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)
	grammarFocus := bot.GrammarTenseName(grammar)
	todayWord := ""

	words, _ := j.db.GetUserWords(user.ID, 0, 1)
	if len(words) > 0 {
		todayWord = words[0].Word
	}

	keywords, _ := j.openai.SuggestMediaKeywords(grammarFocus, todayWord, user.Level)
	log.Printf("[cron][user=%d] media keywords: %v", user.ID, keywords)

	media, err := j.db.SearchMedia(user.ID, user.Level, user.TargetLanguage, keywords)
	if err != nil {
		log.Printf("[cron][user=%d] no media: %v", user.ID, err)
		return
	}

	j.db.MarkMediaSent(user.ID, media.ID)

	log.Printf("[cron][user=%d] media: %s", user.ID, media.Title)

	recipient := &tele.User{ID: user.ID}
	_, sendErr := j.bot.Send(recipient,
		bot.FormatMediaRecommendation(media, user.Language),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: bot.MediaDoneKeyboard(media.ID, user.Language)})
	if sendErr != nil {
		log.Printf("[cron][user=%d] send media error, trying fallback: %v", user.ID, sendErr)
		fallback, err := j.db.GetUnseenMedia(user.ID, user.Level, user.TargetLanguage)
		if err == nil && fallback.ID != media.ID {
			j.db.MarkMediaSent(user.ID, fallback.ID)
			_, retryErr := j.bot.Send(recipient,
				bot.FormatMediaRecommendation(fallback, user.Language),
				&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: bot.MediaDoneKeyboard(fallback.ID, user.Language)})
			if retryErr != nil {
				log.Printf("[cron][user=%d] fallback media also failed: %v", user.ID, retryErr)
			}
		}
	}
}

func (j *Jobs) sendDailyReview(user db.User) {
	streak, _ := j.db.GetTodayStreak(user.ID, user.TzOffset)
	if streak.ReviewDone {
		return
	}

	m := content.GetMessages(user.Language)
	streakDays, _ := j.db.GetCurrentStreak(user.ID, user.TzOffset)

	check := func(done bool) string {
		if done {
			return "\u2705"
		}
		return "\u274C"
	}

	log.Printf("[cron][user=%d] daily review (streak=%d)", user.ID, streakDays)
	j.sendMessage(user.ID, fmt.Sprintf("%s\n\n"+
		"%s %s\n%s %s\n%s %s\n\n\U0001F525 *%d* %s!\n\n"+
		"%s",
		m.LabelEndOfDay,
		check(streak.WordDone), m.LabelWord,
		check(streak.WritingDone), m.LabelWriting,
		check(streak.ReviewDone), m.LabelQuiz,
		streakDays, m.LabelStreak,
		m.LabelTakeQuiz))
}

func (j *Jobs) sendWeeklyReport(user db.User) {
	m := content.GetMessages(user.Language)
	weekly, _ := j.db.GetWeeklyStats(user.ID, user.TzOffset)
	streakDays, _ := j.db.GetCurrentStreak(user.ID, user.TzOffset)
	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)

	grammarFocus := bot.GrammarTenseName(grammar)

	report, err := j.openai.GenerateWeeklyReport(weekly.WordsDone, weekly.WritingsDone, streakDays, grammarFocus, user.Level, user.Language)
	if err != nil {
		report = fmt.Sprintf("Nice work this week! %d words learned, %d texts written, %d day streak!", weekly.WordsDone, weekly.WritingsDone, streakDays)
	}

	log.Printf("[cron][user=%d] weekly report", user.ID)
	j.sendMessage(user.ID, fmt.Sprintf("%s\n\n"+
		"\U0001F4D6 %s: %d\n\u270D\uFE0F %s: %d\n\U0001F9E9 %s: %d\n\U0001F525 %s: %d\n\n%s\n\n%s",
		m.LabelYourWeek,
		m.LabelWords, weekly.WordsDone,
		m.LabelWritings, weekly.WritingsDone,
		m.LabelQuizzes, weekly.ReviewsDone,
		m.LabelStreak, streakDays,
		report,
		m.LabelNewGrammar))

	j.db.AdvanceGrammarWeek(user.ID)
	j.db.ResetWeeklySkips(user.ID)
}

func (j *Jobs) sendMessage(userID int64, text string) {
	recipient := &tele.User{ID: userID}
	_, err := j.bot.Send(recipient, text, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	if err != nil {
		log.Printf("[cron][user=%d] send error: %v", userID, err)
	}
}

func (j *Jobs) sendQuizPoll(user db.User, word *db.Word) {
	recipient := &tele.User{ID: user.ID}
	if err := bot.SendQuizPoll(j.bot, recipient, user.ID, word, j.openai, user.Language); err != nil {
		log.Printf("[cron][user=%d] send quiz poll error: %v", user.ID, err)
	}
}
