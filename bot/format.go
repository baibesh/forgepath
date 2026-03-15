package bot

import (
	"fmt"
	"strings"

	"github.com/baibesh/forgepath/db"
)

func FormatWordOfDay(word *db.Word, grammar *db.GrammarWeek) string {
	var sb strings.Builder
	sb.WriteString("\U0001F31F *New word for you!*\n\n")
	sb.WriteString(fmt.Sprintf("*%s* — %s\n\n", escapeMarkdown(word.Word), escapeMarkdown(word.Definition)))
	sb.WriteString(fmt.Sprintf("\U0001F4AC _%s_\n", escapeMarkdown(word.Example)))

	if word.Construction != "" {
		sb.WriteString(fmt.Sprintf("\n\U0001F527 How to use: %s", escapeMarkdown(word.Construction)))
	}
	if word.Collocations != "" {
		sb.WriteString(fmt.Sprintf("\n\U0001F517 Goes with: %s", escapeMarkdown(word.Collocations)))
	}

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\n\n\U0001F4A1 *This week's grammar:* %s\n", escapeMarkdown(grammar.TenseName)))
		sb.WriteString(fmt.Sprintf("\U0001F6AA %s", escapeMarkdown(grammar.Anchor)))
	}

	return sb.String()
}

func FormatQuizFillBlank(word *db.Word, options []string, correctIdx int) string {
	example := strings.Replace(word.Example, word.Word, "\\_\\_\\_\\_\\_", 1)
	if example == word.Example {
		example = fmt.Sprintf("What does *%s* mean?", escapeMarkdown(word.Word))
	}

	var sb strings.Builder
	sb.WriteString("\U0001F9E9 *Quick quiz!*\n\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", example))

	letters := []string{"A", "B", "C", "D"}
	for i, opt := range options {
		if i < len(letters) {
			sb.WriteString(fmt.Sprintf("%s) %s\n", letters[i], escapeMarkdown(opt)))
		}
	}

	return sb.String()
}

func FormatQuizCollocation(word *db.Word) string {
	return fmt.Sprintf("\U0001F9E9 *What goes with this word?*\n\n"+
		"Word: *%s*\n\n"+
		"Hint: %s",
		escapeMarkdown(word.Word), escapeMarkdown(word.Construction))
}

func FormatQuizTypeWord(word *db.Word) string {
	return fmt.Sprintf("\U0001F9E9 *Can you remember?*\n\n"+
		"What's the word for:\n\n"+
		"\U0001F4AD %s\n\n"+
		"Type your answer:",
		escapeMarkdown(word.Definition))
}

func FormatQuizMakeSentence(word *db.Word) string {
	return fmt.Sprintf("\U0001F9E9 *Use it in a sentence!*\n\n"+
		"Make a sentence with: *%s*\n"+
		"(%s)\n\n"+
		"Type your sentence:",
		escapeMarkdown(word.Word), escapeMarkdown(word.Definition))
}

func FormatStats(streak int, wordCount int, writingCount int, grammar *db.GrammarWeek, weekly *db.WeeklyStats) string {
	var sb strings.Builder
	sb.WriteString("\U0001F4CA *Your progress*\n\n")
	sb.WriteString(fmt.Sprintf("\U0001F525 Streak: *%d days*\n", streak))
	sb.WriteString(fmt.Sprintf("\U0001F4D6 Words: *%d*\n", wordCount))
	sb.WriteString(fmt.Sprintf("\u270D\uFE0F Writings: *%d*\n", writingCount))

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\n\U0001F3AF This week: *%s*\n", escapeMarkdown(grammar.TenseName)))
	}

	if weekly != nil {
		sb.WriteString(fmt.Sprintf("\n*Last 7 days:*\n"))
		sb.WriteString(fmt.Sprintf("  Words: %d \u2022 Writings: %d \u2022 Quizzes: %d\n", weekly.WordsDone, weekly.WritingsDone, weekly.ReviewsDone))
	}

	return sb.String()
}

func FormatWritingPrompt(topic, grammarFocus string, grammar *db.GrammarWeek, language string) string {
	var sb strings.Builder
	sb.WriteString("\u270D\uFE0F *Time to write!*\n\n")
	sb.WriteString(fmt.Sprintf("*Topic:* %s\n\n", escapeMarkdown(topic)))

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\U0001F4A1 Try to use *%s*\n", escapeMarkdown(grammar.TenseName)))
		sb.WriteString(fmt.Sprintf("Example: _%s_\n", escapeMarkdown(grammar.Example)))
		sb.WriteString(fmt.Sprintf("Markers: _%s_\n\n", escapeMarkdown(grammar.Markers)))
	}

	hint := "Write a few sentences and send them. Don't worry about mistakes — I'll help!"
	if language == "de" {
		hint = "Schreib ein paar Satze und schick sie ab. Keine Angst vor Fehlern!"
	}
	sb.WriteString(hint)

	return sb.String()
}

func FormatMediaRecommendation(media *db.MediaResource) string {
	return fmt.Sprintf("\U0001F3AC *Watch this!*\n\n"+
		"\U0001F4FA %s\n"+
		"\U0001F517 %s\n"+
		"\u23F1 %s\n\n"+
		"After watching, press the button and I'll give you a small task!",
		escapeMarkdown(media.Title), media.URL, media.Duration)
}

func FormatMediaTask(media *db.MediaResource, grammarFocus string) string {
	return fmt.Sprintf("\U0001F4DD *What did you think?*\n\n"+
		"Write a few sentences about what you watched.\n\n"+
		"For example:\n"+
		"\u2022 What was it about?\n"+
		"\u2022 What new word did you hear?\n"+
		"\u2022 What do you think about it?\n\n"+
		"Try to use *%s*!",
		escapeMarkdown(grammarFocus))
}

func FormatDailyReview(todayWord *db.Word, streak *db.TodayStreak, streakDays int) string {
	var sb strings.Builder
	sb.WriteString("\U0001F31B *End of day!*\n\n")

	check := func(done bool) string {
		if done {
			return "\u2705"
		}
		return "\u2B1C"
	}

	sb.WriteString(fmt.Sprintf("%s New word\n", check(streak.WordDone)))
	sb.WriteString(fmt.Sprintf("%s Writing\n", check(streak.WritingDone)))
	sb.WriteString(fmt.Sprintf("%s Quiz\n\n", check(streak.ReviewDone)))

	if todayWord != nil {
		sb.WriteString(fmt.Sprintf("Today's word: *%s* — %s\n\n", escapeMarkdown(todayWord.Word), escapeMarkdown(todayWord.Definition)))
	}

	sb.WriteString(fmt.Sprintf("\U0001F525 *%d days* in a row!\n\n", streakDays))
	sb.WriteString("See you tomorrow!")

	return sb.String()
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
