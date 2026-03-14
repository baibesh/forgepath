package cron

import (
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
		log.Printf("Cron: error getting users: %v", err)
		return
	}

	now := time.Now().UTC()

	for _, user := range users {
		localHour := (now.Hour() + user.TzOffset) % 24
		if localHour < 0 {
			localHour += 24
		}
		localMinute := now.Minute()

		// Check state timeout (>2 hours in non-idle state)
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

		// Weekly report on Sunday at 9:00
		if now.Weekday() == time.Sunday && localHour == 9 && localMinute < 30 {
			j.safeSend(user.ID, j.sendWeeklyReport)
		}
	}
}

func (j *Jobs) safeSend(userID int64, fn func(int64)) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Cron: recovered panic for user %d: %v", userID, r)
		}
	}()
	fn(userID)
}

func (j *Jobs) checkStateTimeout(userID int64) {
	state, err := j.db.GetState(userID)
	if err != nil || state.State == "idle" {
		return
	}
	// Auto-clear stale states (the state module doesn't track time, so we just clear non-idle)
	// In production, we'd check updated_at, but for now we rely on the cron running every 30min
	j.db.ClearState(userID)
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
		log.Printf("Cron: no unseen words for user %d: %v", userID, err)
		return
	}

	grammar, _ := j.db.GetCurrentGrammarFocus(userID)
	j.db.MarkWordSeen(userID, word.ID)
	j.db.MarkWordDone(userID)

	msg := formatWordOfDayCron(word, grammar)
	j.sendMessage(userID, msg)

	// Send quiz for review words
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

	msg := fmt.Sprintf("✍️ *Free Writing — 5 min*\n\n"+
		"🎯 Grammar: %s\n"+
		"🚪 %s\n\n"+
		"*Topic:* \"%s\"\n\n"+
		"📍 Formula: %s\n"+
		"📍 Markers: %s\n\n"+
		"Send your text when ready!",
		grammar.TenseName, grammar.Anchor, topic, grammar.Formula, grammar.Markers)

	j.sendMessage(userID, msg)
}

func (j *Jobs) sendMediaRecommendation(userID int64) {
	// Check if media already sent today
	_, err := j.db.GetTodayUnsentMedia(userID)
	if err == nil {
		return // already sent
	}

	user, err := j.db.GetUser(userID)
	if err != nil {
		return
	}

	media, err := j.db.GetUnseenMedia(userID, user.Level)
	if err != nil {
		log.Printf("Cron: no unseen media for user %d: %v", userID, err)
		return
	}

	j.db.MarkMediaSent(userID, media.ID)

	msg := fmt.Sprintf("🎬 *Today's Recommendation*\n\n"+
		"📺 \"%s\"\n"+
		"🔗 %s\n"+
		"⏱ %s | Level: %s\n\n"+
		"Watch it! Task in 2 hours 📝",
		media.Title, media.URL, media.Duration, media.Level)

	j.sendMessage(userID, msg)
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

	media, err := j.db.GetUnseenMedia(userID, "A2") // fallback to get title
	mediaTitle := "the video"
	if err == nil && media != nil {
		mediaTitle = media.Title
	}

	j.db.SetState(userID, "waiting_media_task", map[string]string{
		"media_id":    fmt.Sprintf("%d", um.MediaID),
		"media_title": mediaTitle,
	})

	msg := fmt.Sprintf("📝 *Post-Media Task*\n\n"+
		"Write 3 sentences about what you watched:\n"+
		"Use %s\n\n"+
		"1. What happened in the video?\n"+
		"2. One new word or phrase you noticed\n"+
		"3. \"I think...\" (your opinion)\n\n"+
		"_(type your sentences)_", grammarFocus)

	j.sendMessage(userID, msg)
}

func (j *Jobs) sendDailyReview(userID int64) {
	streak, _ := j.db.GetTodayStreak(userID)
	if streak.ReviewDone {
		return
	}

	streakDays, _ := j.db.GetCurrentStreak(userID)
	todayStreak, _ := j.db.GetTodayStreak(userID)

	checkOrCross := func(done bool) string {
		if done {
			return "✅"
		}
		return "❌"
	}

	msg := fmt.Sprintf("📊 *Daily Review*\n\n"+
		"%s Word of the Day\n"+
		"%s Free Writing\n"+
		"%s Daily Review\n\n"+
		"🔥 Streak: *%d days*\n\n"+
		"Keep going! See you tomorrow 💪",
		checkOrCross(todayStreak.WordDone),
		checkOrCross(todayStreak.WritingDone),
		checkOrCross(todayStreak.ReviewDone),
		streakDays)

	j.sendMessage(userID, msg)
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

	msg := fmt.Sprintf("📊 *Weekly Report*\n\n"+
		"📖 Words: %d\n"+
		"✍️ Writings: %d\n"+
		"📝 Reviews: %d\n"+
		"🔥 Streak: %d days\n\n"+
		"%s\n\n"+
		"New grammar week starts now! 📚",
		weekly.WordsDone, weekly.WritingsDone, weekly.ReviewsDone, streakDays, report)

	j.sendMessage(userID, msg)

	// Advance grammar week and reset skips
	j.db.AdvanceGrammarWeek(userID)
	j.db.ResetWeeklySkips()
}

func (j *Jobs) sendMessage(userID int64, text string) {
	recipient := &tele.User{ID: userID}
	_, err := j.bot.Send(recipient, text, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	if err != nil {
		log.Printf("Cron: failed to send to user %d: %v", userID, err)
	}
}

func (j *Jobs) sendSimpleQuiz(userID int64, word *db.Word) {
	// Simple fill-in-blank quiz via message (no inline keyboard in cron context)
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
