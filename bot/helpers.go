package bot

import (
	"fmt"

	"github.com/baibesh/forgepath/content"
	"github.com/baibesh/forgepath/db"
)

func userTzOffset(user *db.User) int {
	if user != nil {
		return user.TzOffset
	}
	return 0
}

func FormatUTCOffset(offset int) string {
	if offset > 0 {
		return fmt.Sprintf("UTC+%d", offset)
	}
	if offset < 0 {
		return fmt.Sprintf("UTC%d", offset)
	}
	return "UTC"
}

func GrammarOrDefault(grammar *db.GrammarWeek, language string) *db.GrammarWeek {
	if grammar != nil {
		return grammar
	}
	return db.DefaultGrammar(language)
}

func GrammarTenseName(grammar *db.GrammarWeek) string {
	if grammar != nil {
		return grammar.TenseName
	}
	return "Past Simple"
}

func userLang(user *db.User) string {
	if user != nil {
		return user.Language
	}
	return "en"
}

func userMessages(user *db.User) *content.Messages {
	return content.GetMessages(userLang(user))
}

func messagesForLang(lang string) *content.Messages {
	return content.GetMessages(lang)
}
