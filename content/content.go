package content

import "math/rand"

var topicsEN = []string{
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

var topicsDE = []string{
	"Was hast du letztes Wochenende gemacht?",
	"Beschreibe deinen Morgen.",
	"Erzähle von deinem Lieblingsfilm.",
	"Was möchtest du gerne lernen?",
	"Beschreibe eine Person, die du bewunderst.",
	"Was hast du gestern gegessen?",
	"Erzähle von deiner besten Reise.",
	"Was macht dich glücklich?",
	"Beschreibe deinen Arbeitsplatz.",
	"Was sind deine Pläne für diese Woche?",
}

func RandomTopic(language string) string {
	topics := GetTopics(language)
	return topics[rand.Intn(len(topics))]
}

func GetTopics(language string) []string {
	switch language {
	case "de":
		return topicsDE
	default:
		return topicsEN
	}
}

func LanguageName(code string) string {
	switch code {
	case "de":
		return "Deutsch"
	default:
		return "English"
	}
}

func LanguageFlag(code string) string {
	switch code {
	case "de":
		return "\U0001F1E9\U0001F1EA"
	default:
		return "\U0001F1EC\U0001F1E7"
	}
}

func WritingHint(language string) string {
	switch language {
	case "de":
		return "Schreib ein paar Satze und schick sie ab. Keine Angst vor Fehlern!"
	default:
		return "Write a few sentences and send them. Don't worry about mistakes — I'll help!"
	}
}
