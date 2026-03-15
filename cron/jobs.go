package cron

import (
	"fmt"
	"log"
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

	for _, user := range users {
		localHour := (now.Hour() + user.TzOffset) % 24
		if localHour < 0 {
			localHour += 24
		}
		localMinute := now.Minute()

		j.checkStateTimeout(user.ID)

		switch {
		case localHour == 7 && localMinute >= 30:
			j.safeSend(user, j.sendWordOfDay)
		case localHour == 12 && localMinute < 30:
			j.safeSend(user, j.sendWritingPrompt)
		case localHour == 18 && localMinute < 30:
			j.safeSend(user, j.sendMediaRecommendation)
		case localHour == 20 && localMinute < 30:
			j.safeSend(user, j.sendMediaTask)
		case localHour == 21 && localMinute >= 30:
			j.safeSend(user, j.sendDailyReview)
		}

		if now.Weekday() == time.Sunday && localHour == 9 && localMinute < 30 {
			j.safeSend(user, j.sendWeeklyReport)
		}
	}
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

func (j *Jobs) sendWordOfDay(user db.User) {
	streak, _ := j.db.GetTodayStreak(user.ID, user.TzOffset)
	if streak.WordDone {
		return
	}

	word, err := j.db.GetRandomUnseen(user.ID, user.Level, user.Language)
	if err != nil {
		log.Printf("[cron][user=%d] no unseen words: %v", user.ID, err)
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)
	j.db.MarkWordSeen(user.ID, word.ID)
	j.db.MarkWordDone(user.ID, user.TzOffset)

	log.Printf("[cron][user=%d] word of day: %s", user.ID, word.Word)
	j.sendMessage(user.ID, bot.FormatWordOfDay(word, grammar))

	reviewWords, _ := j.db.GetWordsForReview(user.ID, 2)
	for _, rw := range reviewWords {
		j.sendQuizPoll(user.ID, &rw)
	}
}

func (j *Jobs) sendWritingPrompt(user db.User) {
	streak, _ := j.db.GetTodayStreak(user.ID, user.TzOffset)
	if streak.WritingDone {
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)
	grammar = bot.GrammarOrDefault(grammar, user.Language)

	topic := content.RandomTopic(user.Language)

	j.db.SetState(user.ID, "waiting_writing", map[string]string{
		"topic":         topic,
		"grammar_focus": grammar.TenseName,
	})

	log.Printf("[cron][user=%d] writing prompt: %s", user.ID, topic)
	j.sendMessage(user.ID, fmt.Sprintf("\u270D\uFE0F *Free Writing — 5 min*\n\n"+
		"\U0001F3AF Grammar: %s\n\U0001F6AA %s\n\n*Topic:* \"%s\"\n\n\U0001F4CD Formula: %s\n\U0001F4CD Markers: %s\n\n%s",
		grammar.TenseName, grammar.Anchor, topic, grammar.Formula, grammar.Markers,
		content.WritingHint(user.Language)))
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

	media, err := j.db.SearchMedia(user.ID, user.Level, user.Language, keywords)
	if err != nil {
		log.Printf("[cron][user=%d] no media: %v", user.ID, err)
		return
	}

	j.db.MarkMediaSent(user.ID, media.ID)

	log.Printf("[cron][user=%d] media: %s", user.ID, media.Title)

	recipient := &tele.User{ID: user.ID}
	_, sendErr := j.bot.Send(recipient,
		bot.FormatMediaRecommendation(media),
		&tele.SendOptions{ParseMode: tele.ModeMarkdown, ReplyMarkup: bot.MediaDoneKeyboard(media.ID)})
	if sendErr != nil {
		log.Printf("[cron][user=%d] send media error: %v", user.ID, sendErr)
	}
}

func (j *Jobs) sendMediaTask(user db.User) {
	um, err := j.db.GetTodayUnsentMedia(user.ID, user.TzOffset)
	if err != nil || um == nil {
		return
	}
	if um.TaskSent {
		return
	}

	j.db.MarkMediaTaskSent(user.ID, um.MediaID)

	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)
	grammarFocus := bot.GrammarTenseName(grammar)

	mediaTitle := j.db.GetMediaTitle(user.ID, um.MediaID)

	j.db.SetState(user.ID, "waiting_media_task", map[string]string{
		"media_id":    fmt.Sprintf("%d", um.MediaID),
		"media_title": mediaTitle,
	})

	log.Printf("[cron][user=%d] media task for: %s", user.ID, mediaTitle)
	j.sendMessage(user.ID, fmt.Sprintf("\U0001F4DD *Post-Media Task*\n\n"+
		"Write 3 sentences about what you watched:\nUse %s\n\n"+
		"1. What happened in the video?\n2. One new word or phrase you noticed\n"+
		"3. \"I think...\" (your opinion)\n\n_(type your sentences)_", grammarFocus))
}

func (j *Jobs) sendDailyReview(user db.User) {
	streak, _ := j.db.GetTodayStreak(user.ID, user.TzOffset)
	if streak.ReviewDone {
		return
	}

	streakDays, _ := j.db.GetCurrentStreak(user.ID, user.TzOffset)

	check := func(done bool) string {
		if done {
			return "\u2705"
		}
		return "\u274C"
	}

	log.Printf("[cron][user=%d] daily review (streak=%d)", user.ID, streakDays)
	j.sendMessage(user.ID, fmt.Sprintf("\U0001F4CA *Daily Review*\n\n"+
		"%s Word of the Day\n%s Free Writing\n%s Daily Review\n\n\U0001F525 Streak: *%d days*\n\n"+
		"Complete a /quiz to finish your day! \U0001F4AA",
		check(streak.WordDone), check(streak.WritingDone), check(streak.ReviewDone), streakDays))
}

func (j *Jobs) sendWeeklyReport(user db.User) {
	weekly, _ := j.db.GetWeeklyStats(user.ID, user.TzOffset)
	streakDays, _ := j.db.GetCurrentStreak(user.ID, user.TzOffset)
	grammar, _ := j.db.GetCurrentGrammarFocus(user.ID)

	grammarFocus := bot.GrammarTenseName(grammar)

	report, err := j.openai.GenerateWeeklyReport(weekly.WordsDone, weekly.WritingsDone, streakDays, grammarFocus)
	if err != nil {
		report = fmt.Sprintf("Great week! %d words, %d writings, %d day streak!", weekly.WordsDone, weekly.WritingsDone, streakDays)
	}

	log.Printf("[cron][user=%d] weekly report", user.ID)
	j.sendMessage(user.ID, fmt.Sprintf("\U0001F4CA *Weekly Report*\n\n"+
		"\U0001F4D6 Words: %d\n\u270D\uFE0F Writings: %d\n\U0001F4DD Reviews: %d\n\U0001F525 Streak: %d days\n\n%s\n\nNew grammar week starts now! \U0001F4DA",
		weekly.WordsDone, weekly.WritingsDone, weekly.ReviewsDone, streakDays, report))

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

func (j *Jobs) sendQuizPoll(userID int64, word *db.Word) {
	recipient := &tele.User{ID: userID}
	if err := bot.SendQuizPoll(j.bot, recipient, userID, word, j.openai); err != nil {
		log.Printf("[cron][user=%d] send quiz poll error: %v", userID, err)
	}
}
