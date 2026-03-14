package cron

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/ai"
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
		case localHour == 7 && localMinute < 30:
			j.safeSend(user.ID, j.sendWordOfDay)
		case localHour == 12 && localMinute < 30:
			j.safeSend(user.ID, j.sendWritingPrompt)
		case localHour == 18 && localMinute < 30:
			j.safeSend(user.ID, j.sendMediaRecommendation)
		case localHour == 20 && localMinute < 30:
			j.safeSend(user.ID, j.sendMediaTask)
		case localHour == 21 && localMinute >= 30 && localMinute < 60:
			j.safeSend(user.ID, j.sendDailyReview)
		}

		if now.Weekday() == time.Sunday && localHour == 9 && localMinute < 30 {
			j.safeSend(user.ID, j.sendWeeklyReport)
		}
	}
}

func (j *Jobs) safeSend(userID int64, fn func(int64)) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[cron][user=%d] panic: %v", userID, r)
		}
	}()
	fn(userID)
}

func (j *Jobs) checkStateTimeout(userID int64) {
	// Check updated_at in DB — clear if >2 hours stale
	var updatedAt time.Time
	err := j.db.Pool.QueryRow(context.Background(),
		`SELECT updated_at FROM user_state WHERE user_id = $1 AND state != 'idle'`, userID,
	).Scan(&updatedAt)
	if err != nil {
		return
	}
	if time.Since(updatedAt) > 2*time.Hour {
		log.Printf("[cron][user=%d] clearing stale state (age=%s)", userID, time.Since(updatedAt).Round(time.Minute))
		j.db.ClearState(userID)
	}
}

func (j *Jobs) sendWordOfDay(userID int64) {
	streak, _ := j.db.GetTodayStreak(userID)
	if streak.WordDone {
		return
	}

	user, err := j.db.GetUser(userID)
	if err != nil {
		return
	}

	word, err := j.db.GetRandomUnseen(userID, user.Level)
	if err != nil {
		log.Printf("[cron][user=%d] no unseen words: %v", userID, err)
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(userID)
	j.db.MarkWordSeen(userID, word.ID)
	j.db.MarkWordDone(userID)

	log.Printf("[cron][user=%d] word of day: %s", userID, word.Word)
	j.sendMessage(userID, formatWordOfDayCron(word, grammar))

	reviewWords, _ := j.db.GetWordsForReview(userID, 2)
	for _, rw := range reviewWords {
		j.sendSimpleQuiz(userID, &rw)
	}
}

func (j *Jobs) sendWritingPrompt(userID int64) {
	streak, _ := j.db.GetTodayStreak(userID)
	if streak.WritingDone {
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(userID)
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

	j.db.SetState(userID, "waiting_writing", map[string]string{
		"topic":         topic,
		"grammar_focus": grammar.TenseName,
	})

	log.Printf("[cron][user=%d] writing prompt: %s", userID, topic)
	j.sendMessage(userID, fmt.Sprintf("✍️ *Free Writing — 5 min*\n\n"+
		"🎯 Grammar: %s\n🚪 %s\n\n*Topic:* \"%s\"\n\n📍 Formula: %s\n📍 Markers: %s\n\nSend your text when ready!",
		grammar.TenseName, grammar.Anchor, topic, grammar.Formula, grammar.Markers))
}

func (j *Jobs) sendMediaRecommendation(userID int64) {
	_, err := j.db.GetTodayUnsentMedia(userID)
	if err == nil {
		return
	}

	user, err := j.db.GetUser(userID)
	if err != nil {
		return
	}

	// Smart media selection via AI keywords
	grammar, _ := j.db.GetCurrentGrammarFocus(userID)
	grammarFocus := "english"
	todayWord := ""
	if grammar != nil {
		grammarFocus = grammar.TenseName
	}

	words, _ := j.db.GetUserWords(userID, 0, 1)
	if len(words) > 0 {
		todayWord = words[0].Word
	}

	keywords, _ := j.openai.SuggestMediaKeywords(grammarFocus, todayWord, user.Level)
	log.Printf("[cron][user=%d] media keywords: %v", userID, keywords)

	media, err := j.db.SearchMedia(userID, user.Level, keywords)
	if err != nil {
		log.Printf("[cron][user=%d] no media: %v", userID, err)
		return
	}

	j.db.MarkMediaSent(userID, media.ID)

	log.Printf("[cron][user=%d] media: %s", userID, media.Title)
	j.sendMessage(userID, fmt.Sprintf("🎬 *Today's Recommendation*\n\n"+
		"📺 \"%s\"\n🔗 %s\n⏱ %s | Level: %s\n\nWatch it! Task in 2 hours 📝",
		media.Title, media.URL, media.Duration, media.Level))
}

func (j *Jobs) sendMediaTask(userID int64) {
	um, err := j.db.GetTodayUnsentMedia(userID)
	if err != nil || um == nil {
		return
	}
	if um.TaskSent {
		return
	}

	j.db.MarkMediaTaskSent(userID, um.MediaID)

	grammar, _ := j.db.GetCurrentGrammarFocus(userID)
	grammarFocus := "Past Simple"
	if grammar != nil {
		grammarFocus = grammar.TenseName
	}

	// Get the media title from the media we sent today
	var mediaTitle string
	err = j.db.Pool.QueryRow(context.Background(),
		`SELECT mr.title FROM user_media um JOIN media_resources mr ON mr.id = um.media_id
		 WHERE um.user_id = $1 AND um.media_id = $2`, userID, um.MediaID).Scan(&mediaTitle)
	if err != nil {
		mediaTitle = "the video"
	}

	j.db.SetState(userID, "waiting_media_task", map[string]string{
		"media_id":    fmt.Sprintf("%d", um.MediaID),
		"media_title": mediaTitle,
	})

	log.Printf("[cron][user=%d] media task for: %s", userID, mediaTitle)
	j.sendMessage(userID, fmt.Sprintf("📝 *Post-Media Task*\n\n"+
		"Write 3 sentences about what you watched:\nUse %s\n\n"+
		"1. What happened in the video?\n2. One new word or phrase you noticed\n"+
		"3. \"I think...\" (your opinion)\n\n_(type your sentences)_", grammarFocus))
}

func (j *Jobs) sendDailyReview(userID int64) {
	streak, _ := j.db.GetTodayStreak(userID)
	if streak.ReviewDone {
		return
	}

	streakDays, _ := j.db.GetCurrentStreak(userID)
	todayStreak, _ := j.db.GetTodayStreak(userID)

	check := func(done bool) string {
		if done {
			return "✅"
		}
		return "❌"
	}

	log.Printf("[cron][user=%d] daily review (streak=%d)", userID, streakDays)
	j.sendMessage(userID, fmt.Sprintf("📊 *Daily Review*\n\n"+
		"%s Word of the Day\n%s Free Writing\n%s Daily Review\n\n🔥 Streak: *%d days*\n\nKeep going! See you tomorrow 💪",
		check(todayStreak.WordDone), check(todayStreak.WritingDone), check(todayStreak.ReviewDone), streakDays))

	j.db.MarkReviewDone(userID)
}

func (j *Jobs) sendWeeklyReport(userID int64) {
	weekly, _ := j.db.GetWeeklyStats(userID)
	streakDays, _ := j.db.GetCurrentStreak(userID)
	grammar, _ := j.db.GetCurrentGrammarFocus(userID)

	grammarFocus := "grammar"
	if grammar != nil {
		grammarFocus = grammar.TenseName
	}

	report, err := j.openai.GenerateWeeklyReport(weekly.WordsDone, weekly.WritingsDone, streakDays, grammarFocus)
	if err != nil {
		report = fmt.Sprintf("Great week! %d words, %d writings, %d day streak!", weekly.WordsDone, weekly.WritingsDone, streakDays)
	}

	log.Printf("[cron][user=%d] weekly report", userID)
	j.sendMessage(userID, fmt.Sprintf("📊 *Weekly Report*\n\n"+
		"📖 Words: %d\n✍️ Writings: %d\n📝 Reviews: %d\n🔥 Streak: %d days\n\n%s\n\nNew grammar week starts now! 📚",
		weekly.WordsDone, weekly.WritingsDone, weekly.ReviewsDone, streakDays, report))

	j.db.AdvanceGrammarWeek(userID)
	j.db.ResetWeeklySkips()
}

func (j *Jobs) sendMessage(userID int64, text string) {
	recipient := &tele.User{ID: userID}
	_, err := j.bot.Send(recipient, text, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	if err != nil {
		log.Printf("[cron][user=%d] send error: %v", userID, err)
	}
}

func (j *Jobs) sendSimpleQuiz(userID int64, word *db.Word) {
	wrongOptions, _ := j.openai.GenerateQuizOptions(word.Word, word.Definition, 3)
	options := []string{word.Definition}
	options = append(options, wrongOptions...)

	rand.Shuffle(len(options), func(i, k int) {
		options[i], options[k] = options[k], options[i]
	})

	var sb strings.Builder
	sb.WriteString("🧠 *Quick Review*\n\n")
	sb.WriteString(fmt.Sprintf("What does *%s* mean?\n\n", word.Word))
	letters := []string{"A", "B", "C", "D"}
	for i, opt := range options {
		if i < 4 {
			sb.WriteString(fmt.Sprintf("%s) %s\n", letters[i], opt))
		}
	}

	j.sendMessage(userID, sb.String())
}

func formatWordOfDayCron(word *db.Word, grammar *db.GrammarWeek) string {
	var sb strings.Builder
	sb.WriteString("📖 *Word of the Day*\n\n")
	sb.WriteString(fmt.Sprintf("*%s* — %s\n\n", word.Word, word.Definition))
	sb.WriteString(fmt.Sprintf("💡 \"%s\"\n\n", word.Example))

	if word.Construction != "" {
		sb.WriteString(fmt.Sprintf("📌 Construction: %s\n", word.Construction))
	}
	if word.Collocations != "" {
		sb.WriteString(fmt.Sprintf("🔗 Collocations: %s\n\n", word.Collocations))
	}

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("🎯 Grammar: %s — %s\n", grammar.Family, grammar.TenseName))
		sb.WriteString(fmt.Sprintf("🚪 Anchor: %s\n", grammar.Anchor))
		sb.WriteString(fmt.Sprintf("📍 Markers: %s\n", grammar.Markers))
	}

	return sb.String()
}
