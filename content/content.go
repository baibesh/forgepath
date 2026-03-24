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

var topicsRU = []string{
	"Что ты делал в прошлые выходные?",
	"Опиши своё утро.",
	"Расскажи о любимом фильме.",
	"Чему ты хотел бы научиться?",
	"Опиши человека которым восхищаешься.",
	"Что ты ел вчера?",
	"Расскажи о лучшей поездке.",
	"Что делает тебя счастливым?",
	"Опиши своё рабочее место.",
	"Какие планы на эту неделю?",
}

var topicsKK = []string{
	"Өткен демалыста не істедің?",
	"Таңертеңгі күн тәртібіңді сипатта.",
	"Сүйікті фильмің туралы айтып бер.",
	"Нені үйренгің келеді?",
	"Тамсанатын адамды сипатта.",
	"Кеше не жедің?",
	"Ең жақсы саяхатың туралы айтып бер.",
	"Сені не бақытты етеді?",
	"Жұмыс орныңды сипатта.",
	"Осы аптаға қандай жоспарларың бар?",
}

func RandomTopic(language string) string {
	topics := GetTopics(language)
	return topics[rand.Intn(len(topics))]
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

func GetTopics(language string) []string {
	switch language {
	case "ru":
		return topicsRU
	case "kk":
		return topicsKK
	case "de":
		return topicsDE
	default:
		return topicsEN
	}
}

func LanguageName(code string) string {
	switch code {
	case "ru":
		return "Русский"
	case "kk":
		return "Қазақша"
	case "de":
		return "Deutsch"
	default:
		return "English"
	}
}

func LanguageFlag(code string) string {
	switch code {
	case "ru":
		return "\U0001F1F7\U0001F1FA"
	case "kk":
		return "\U0001F1F0\U0001F1FF"
	case "de":
		return "\U0001F1E9\U0001F1EA"
	default:
		return "\U0001F1EC\U0001F1E7"
	}
}

func WritingHint(language string) string {
	switch language {
	case "ru":
		return "Напиши несколько предложений и отправь. Не бойся ошибок — я помогу!"
	case "kk":
		return "Бірнеше сөйлем жазып жібер. Қателіктен қорықпа — мен көмектесемін!"
	default:
		return "Write a few sentences and send them. Don't worry about mistakes — I'll help!"
	}
}

// LanguageNameInLanguage returns the language name as displayed in the target language's own script.
func LanguageNameInLanguage(code string) string {
	switch code {
	case "ru":
		return "Русский"
	case "kk":
		return "Қазақша"
	default:
		return "English"
	}
}
