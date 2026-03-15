package bot

import (
	"fmt"
	"strings"

	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
)

func FormatWordOfDay(word *db.Word, grammar *db.GrammarWeek) string {
	var sb strings.Builder
	sb.WriteString("\U0001F4D6 *Word of the Day*\n\n")
	sb.WriteString(fmt.Sprintf("*%s* — %s\n\n", escapeMarkdown(word.Word), escapeMarkdown(word.Definition)))
	sb.WriteString(fmt.Sprintf("\U0001F4A1 \"%s\"\n\n", escapeMarkdown(word.Example)))

	if word.Construction != "" {
		sb.WriteString(fmt.Sprintf("\U0001F4CC Construction: %s\n", escapeMarkdown(word.Construction)))
	}
	if word.Collocations != "" {
		sb.WriteString(fmt.Sprintf("\U0001F517 Collocations: %s\n\n", escapeMarkdown(word.Collocations)))
	}

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\U0001F3AF Grammar: %s — %s\n", grammar.Family, grammar.TenseName))
		sb.WriteString(fmt.Sprintf("\U0001F4A1 Think of: %s\n", escapeMarkdown(grammar.Anchor)))
		sb.WriteString(fmt.Sprintf("\U0001F4CD Markers: %s\n", escapeMarkdown(grammar.Markers)))
	}

	return sb.String()
}

func FormatQuizFillBlank(word *db.Word, options []string, correctIdx int) string {
	example := strings.Replace(word.Example, word.Word, "\\_\\_\\_\\_\\_", 1)
	if example == word.Example {
		example = fmt.Sprintf("Choose the correct meaning of *%s*:", escapeMarkdown(word.Word))
	}

	var sb strings.Builder
	sb.WriteString("\U0001F9E0 *Quiz Time!*\n\n")
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
	return fmt.Sprintf("\U0001F9E0 *Collocation Quiz*\n\n"+
		"Which collocation goes with *%s*?\n\n"+
		"Hint: %s",
		escapeMarkdown(word.Word), escapeMarkdown(word.Construction))
}

func FormatQuizTypeWord(word *db.Word) string {
	return fmt.Sprintf("\U0001F9E0 *Active Recall*\n\n"+
		"Type the word/phrase:\n\n"+
		"\U0001F4DD %s\n\n"+
		"_(type your answer)_",
		escapeMarkdown(word.Definition))
}

func FormatQuizMakeSentence(word *db.Word) string {
	return fmt.Sprintf("\U0001F9E0 *Make a Sentence*\n\n"+
		"Write a sentence using: *%s*\n"+
		"(%s)\n\n"+
		"_(type your sentence)_",
		escapeMarkdown(word.Word), escapeMarkdown(word.Definition))
}

func FormatStats(streak int, wordCount int, writingCount int, grammar *db.GrammarWeek, weekly *db.WeeklyStats) string {
	var sb strings.Builder
	sb.WriteString("\U0001F4CA *Your Stats*\n\n")
	sb.WriteString(fmt.Sprintf("\U0001F525 Current streak: *%d days*\n", streak))
	sb.WriteString(fmt.Sprintf("\U0001F4D6 Words learned: *%d*\n", wordCount))
	sb.WriteString(fmt.Sprintf("\u270D\uFE0F Writings done: *%d*\n\n", writingCount))

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\U0001F3AF Grammar week: *%s*\n\n", escapeMarkdown(grammar.TenseName)))
	}

	if weekly != nil {
		sb.WriteString("*Last 7 days:*\n")
		sb.WriteString(fmt.Sprintf("  \U0001F4D6 Words: %d/7\n", weekly.WordsDone))
		sb.WriteString(fmt.Sprintf("  \u270D\uFE0F Writings: %d/7\n", weekly.WritingsDone))
		sb.WriteString(fmt.Sprintf("\U0001F4DD Reviews: %d/7\n", weekly.ReviewsDone))
	}

	return sb.String()
}

func FormatWritingPrompt(topic, grammarFocus string, grammar *db.GrammarWeek, language string) string {
	var sb strings.Builder
	sb.WriteString("\u270D\uFE0F *Free Writing — 5 min*\n\n")
	sb.WriteString(fmt.Sprintf("\U0001F3AF Grammar: %s\n", escapeMarkdown(grammarFocus)))

	if grammar != nil {
		sb.WriteString(fmt.Sprintf("\U0001F4A1 %s\n\n", escapeMarkdown(grammar.Anchor)))
		sb.WriteString(fmt.Sprintf("*Topic:* \"%s\"\n\n", escapeMarkdown(topic)))
		sb.WriteString(fmt.Sprintf("\U0001F4CD Formula: %s\n", escapeMarkdown(grammar.Formula)))
		sb.WriteString(fmt.Sprintf("\U0001F4CD Markers: %s\n\n", escapeMarkdown(grammar.Markers)))
	} else {
		sb.WriteString(fmt.Sprintf("\n*Topic:* \"%s\"\n\n", escapeMarkdown(topic)))
	}

	sb.WriteString(content.WritingHint(language))

	return sb.String()
}

func FormatMediaRecommendation(media *db.MediaResource) string {
	return fmt.Sprintf("\U0001F3AC *Today's Recommendation*\n\n"+
		"\U0001F4FA \"%s\"\n"+
		"\U0001F517 %s\n"+
		"\u23F1 %s \\| Level: %s\n\n"+
		"Watch it! Then press the button below \U0001F4DD",
		escapeMarkdown(media.Title), media.URL, media.Duration, media.Level)
}

func FormatMediaTask(media *db.MediaResource, grammarFocus string) string {
	return fmt.Sprintf("\U0001F4DD *Post-Media Task*\n\n"+
		"Write 3 sentences about what you watched:\n"+
		"Use %s\n\n"+
		"1. What happened in the video?\n"+
		"2. One new word or phrase you noticed\n"+
		"3. \"I think...\" (your opinion)\n\n"+
		"_(type your sentences)_",
		escapeMarkdown(grammarFocus))
}

func FormatDailyReview(todayWord *db.Word, streak *db.TodayStreak, streakDays int) string {
	var sb strings.Builder
	sb.WriteString("\U0001F4CA *Daily Review*\n\n")

	checkOrCross := func(done bool) string {
		if done {
			return "\u2705"
		}
		return "\u274C"
	}

	sb.WriteString(fmt.Sprintf("%s Word of the Day\n", checkOrCross(streak.WordDone)))
	sb.WriteString(fmt.Sprintf("%s Free Writing\n", checkOrCross(streak.WritingDone)))
	sb.WriteString(fmt.Sprintf("%s Daily Review\n\n", checkOrCross(streak.ReviewDone)))

	if todayWord != nil {
		sb.WriteString(fmt.Sprintf("\U0001F4D6 Today's word: *%s* — %s\n\n", escapeMarkdown(todayWord.Word), escapeMarkdown(todayWord.Definition)))
	}

	sb.WriteString(fmt.Sprintf("\U0001F525 Streak: *%d days*\n\n", streakDays))
	sb.WriteString("Keep going! See you tomorrow \U0001F4AA")

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
