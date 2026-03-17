package bot

import (
	"fmt"
	"strings"

	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
)

func FormatWordOfDay(word *db.Word, grammar *db.GrammarWeek, lang string) string {
	m := content.GetMessages(lang)
	var sb strings.Builder
	sb.WriteString(m.LabelNewWordDay + "\n\n")
	sb.WriteString(fmt.Sprintf("*%s* — %s\n\n", escapeMarkdown(word.Word), escapeMarkdown(word.Definition)))
	sb.WriteString(fmt.Sprintf("\U0001F4AC _%s_\n", escapeMarkdown(word.Example)))

	if word.Construction != "" {
		sb.WriteString(fmt.Sprintf("\n%s: %s", m.LabelHowToUse, escapeMarkdown(word.Construction)))
	}
	if word.Collocations != "" {
		sb.WriteString(fmt.Sprintf("\n%s: %s", m.LabelGoesWith, escapeMarkdown(word.Collocations)))
	}

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\n\n%s %s\n", m.LabelGrammar, escapeMarkdown(grammar.TenseName)))
		sb.WriteString(fmt.Sprintf("\U0001F6AA %s", escapeMarkdown(grammar.Anchor)))
	}

	return sb.String()
}

func FormatQuizFillBlank(word *db.Word, options []string, correctIdx int, lang string) string {
	m := content.GetMessages(lang)
	example := strings.Replace(word.Example, word.Word, "\\_\\_\\_\\_\\_", 1)
	if example == word.Example {
		example = fmt.Sprintf("%s *%s*?", m.LabelWhatsWord, escapeMarkdown(word.Word))
	}

	var sb strings.Builder
	sb.WriteString(m.LabelQuickQuiz + "\n\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", example))

	letters := []string{"A", "B", "C", "D"}
	for i, opt := range options {
		if i < len(letters) {
			sb.WriteString(fmt.Sprintf("%s) %s\n", letters[i], escapeMarkdown(opt)))
		}
	}

	return sb.String()
}

func FormatQuizCollocation(word *db.Word, lang string) string {
	m := content.GetMessages(lang)
	return fmt.Sprintf("%s\n\n"+
		"%s: *%s*\n\n"+
		"Hint: %s",
		m.LabelQuickQuiz,
		m.LabelWord, escapeMarkdown(word.Word), escapeMarkdown(word.Construction))
}

func FormatQuizTypeWord(word *db.Word, lang string) string {
	m := content.GetMessages(lang)
	return fmt.Sprintf("%s\n\n"+
		"%s\n\n"+
		"\U0001F4AD %s\n\n"+
		"%s",
		m.LabelRemember,
		m.LabelWhatsWord,
		escapeMarkdown(word.Definition),
		m.LabelTypeAnswer)
}

func FormatQuizMakeSentence(word *db.Word, lang string) string {
	m := content.GetMessages(lang)
	return fmt.Sprintf("%s\n\n"+
		"%s *%s*\n"+
		"(%s)\n\n"+
		"%s",
		m.LabelSentence,
		m.LabelMakeSentence, escapeMarkdown(word.Word),
		escapeMarkdown(word.Definition),
		m.LabelTypeSentence)
}

func FormatStats(streak int, wordCount int, writingCount int, grammar *db.GrammarWeek, weekly *db.WeeklyStats, lang string) string {
	m := content.GetMessages(lang)
	var sb strings.Builder
	sb.WriteString("\U0001F4CA *" + m.LabelStreak + "*\n\n")
	sb.WriteString(fmt.Sprintf("\U0001F525 %s: *%d*\n", m.LabelStreak, streak))
	sb.WriteString(fmt.Sprintf("\U0001F4D6 %s: *%d*\n", m.LabelWords, wordCount))
	sb.WriteString(fmt.Sprintf("\u270D\uFE0F %s: *%d*\n", m.LabelWritings, writingCount))

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\n\U0001F3AF %s: *%s*\n", m.LabelThisWeek, escapeMarkdown(grammar.TenseName)))
	}

	if weekly != nil {
		sb.WriteString(fmt.Sprintf("\n%s\n", m.LabelLast7Days))
		sb.WriteString(fmt.Sprintf("  %s: %d \u2022 %s: %d \u2022 %s: %d\n",
			m.LabelWords, weekly.WordsDone,
			m.LabelWritings, weekly.WritingsDone,
			m.LabelQuizzes, weekly.ReviewsDone))
	}

	return sb.String()
}

func FormatWritingPrompt(topic, grammarFocus string, grammar *db.GrammarWeek, language string) string {
	m := content.GetMessages(language)
	var sb strings.Builder
	sb.WriteString(m.LabelTimeToWrite + "\n\n")
	sb.WriteString(fmt.Sprintf("%s %s\n\n", m.LabelTopic, escapeMarkdown(topic)))

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\U0001F4A1 %s *%s*\n", m.LabelTryToUse, escapeMarkdown(grammar.TenseName)))
		sb.WriteString(fmt.Sprintf("%s: _%s_\n", m.LabelExample, escapeMarkdown(grammar.Example)))
		sb.WriteString(fmt.Sprintf("%s: _%s_\n\n", m.LabelMarkers, escapeMarkdown(grammar.Markers)))
	}

	sb.WriteString(content.WritingHint(language))

	return sb.String()
}

func FormatMediaRecommendation(media *db.MediaResource, lang string) string {
	m := content.GetMessages(lang)
	return fmt.Sprintf("%s\n\n"+
		"\U0001F4FA %s\n"+
		"\U0001F517 %s\n"+
		"%s %s\n\n"+
		"%s",
		m.LabelWatchThis,
		escapeMarkdown(media.Title), media.URL, m.LabelDuration, media.Duration,
		m.LabelAfterWatch)
}

func FormatMediaTask(media *db.MediaResource, grammarFocus, lang string) string {
	m := content.GetMessages(lang)
	return fmt.Sprintf("%s\n\n"+
		"%s\n\n"+
		"\u2022 %s\n"+
		"\u2022 %s\n"+
		"\u2022 %s\n\n"+
		"%s *%s*!",
		m.LabelWhatThink,
		m.LabelWriteAbout,
		m.LabelWhatAbout,
		m.LabelNewWordHeard,
		m.LabelWhatDoYouThink,
		m.LabelTryUseGrammar, escapeMarkdown(grammarFocus))
}

func FormatDailyReview(todayWord *db.Word, streak *db.TodayStreak, streakDays int, lang string) string {
	m := content.GetMessages(lang)
	var sb strings.Builder
	sb.WriteString(m.LabelEndOfDay + "\n\n")

	check := func(done bool) string {
		if done {
			return "\u2705"
		}
		return "\u2B1C"
	}

	sb.WriteString(fmt.Sprintf("%s %s\n", check(streak.WordDone), m.LabelWord))
	sb.WriteString(fmt.Sprintf("%s %s\n", check(streak.WritingDone), m.LabelWriting))
	sb.WriteString(fmt.Sprintf("%s %s\n\n", check(streak.ReviewDone), m.LabelQuiz))

	if todayWord != nil {
		sb.WriteString(fmt.Sprintf("%s: *%s* — %s\n\n", m.LabelWord, escapeMarkdown(todayWord.Word), escapeMarkdown(todayWord.Definition)))
	}

	sb.WriteString(fmt.Sprintf("\U0001F525 *%d* %s!\n\n", streakDays, m.LabelStreak))
	sb.WriteString(m.LabelSeeYouTmrw)

	return sb.String()
}

func FormatSchedule(s db.UserSchedule, lang string) string {
	m := content.GetMessages(lang)
	return fmt.Sprintf(
		"%s — *%02d:%02d*\n"+
			"%s — *%02d:%02d*\n"+
			"%s — *%02d:%02d*\n"+
			"%s — *%02d:%02d*",
		m.LabelScheduleWord, s.WordHour, s.WordMin,
		m.LabelScheduleWriting, s.WritingHour, s.WritingMin,
		m.LabelScheduleMedia, s.MediaHour, s.MediaMin,
		m.LabelScheduleReview, s.ReviewHour, s.ReviewMin,
	)
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
