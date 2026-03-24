package bot

import (
	"fmt"

	tele "gopkg.in/telebot.v3"

	"github.com/baibesh/forgepath/content"
)

func MainKeyboard(lang string) *tele.ReplyMarkup {
	m := content.GetMessages(lang)
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Reply(
		menu.Row(
			menu.Text(m.BtnNewWord),
			menu.Text(m.BtnWrite),
			menu.Text(m.BtnQuiz),
		),
		menu.Row(
			menu.Text(m.BtnToday),
			menu.Text(m.BtnProgress),
			menu.Text(m.BtnSettings),
		),
	)
	return menu
}

func LanguageSelectKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F1EC\U0001F1E7 English", "lang", "en"),
			menu.Data("\U0001F1F7\U0001F1FA Русский", "lang", "ru"),
		),
		menu.Row(
			menu.Data("\U0001F1F0\U0001F1FF Қазақша", "lang", "kk"),
		),
	)
	return menu
}

func LevelSelectKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F331 Beginner (A1)", "level", "A1"),
			menu.Data("\U0001F33F A bit (A2)", "level", "A2"),
		),
		menu.Row(
			menu.Data("\U0001F333 Middle (B1)", "level", "B1"),
			menu.Data("\U0001F4AA Strong (B2)", "level", "B2"),
		),
		menu.Row(
			menu.Data("\U0001F31F Advanced (C1)", "level", "C1"),
		),
	)
	return menu
}

func TimezoneKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F1EA\U0001F1FA Europe +2", "tz", "2"),
			menu.Data("\U0001F1F7\U0001F1FA Moscow +3", "tz", "3"),
		),
		menu.Row(
			menu.Data("\U0001F1F0\U0001F1FF Astana +5", "tz", "5"),
			menu.Data("\U0001F1F0\U0001F1FF Almaty +6", "tz", "6"),
		),
		menu.Row(
			menu.Data("+4", "tz", "4"),
			menu.Data("+7", "tz", "7"),
			menu.Data("+8", "tz", "8"),
		),
		menu.Row(
			menu.Data("Other", "tz", "custom"),
		),
	)
	return menu
}

func SettingsKeyboard(lang string) *tele.ReplyMarkup {
	m := content.GetMessages(lang)
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data(m.BtnTimezone, "settings", "timezone"),
			menu.Data(m.BtnLevel, "settings", "level"),
		),
		menu.Row(
			menu.Data(m.BtnLanguage, "settings", "language"),
			menu.Data(m.BtnSchedule, "settings", "schedule"),
		),
		menu.Row(
			menu.Data(m.BtnWordsPerDay, "settings", "wordsperday"),
		),
	)
	return menu
}

func WordsPerDayKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("1", "setwpd", "1"),
			menu.Data("3", "setwpd", "3"),
			menu.Data("5", "setwpd", "5"),
			menu.Data("7", "setwpd", "7"),
		),
	)
	return menu
}

var webAppURL string

func SetWebAppURL(url string) {
	webAppURL = url
}

func SettingsLanguageKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F1EC\U0001F1E7 English", "setlang", "en"),
			menu.Data("\U0001F1F7\U0001F1FA Русский", "setlang", "ru"),
		),
		menu.Row(
			menu.Data("\U0001F1F0\U0001F1FF Қазақша", "setlang", "kk"),
		),
	)
	return menu
}

func SettingsLevelKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F331 Beginner (A1)", "setlevel", "A1"),
			menu.Data("\U0001F33F A bit (A2)", "setlevel", "A2"),
		),
		menu.Row(
			menu.Data("\U0001F333 Middle (B1)", "setlevel", "B1"),
			menu.Data("\U0001F4AA Strong (B2)", "setlevel", "B2"),
		),
		menu.Row(
			menu.Data("\U0001F31F Advanced (C1)", "setlevel", "C1"),
		),
	)
	return menu
}

func SettingsTimezoneKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F1EA\U0001F1FA Europe +2", "settz", "2"),
			menu.Data("\U0001F1F7\U0001F1FA Moscow +3", "settz", "3"),
		),
		menu.Row(
			menu.Data("\U0001F1F0\U0001F1FF Astana +5", "settz", "5"),
			menu.Data("\U0001F1F0\U0001F1FF Almaty +6", "settz", "6"),
		),
		menu.Row(
			menu.Data("+4", "settz", "4"),
			menu.Data("+7", "settz", "7"),
			menu.Data("+8", "settz", "8"),
		),
		menu.Row(
			menu.Data("Other", "settz", "custom"),
		),
	)
	return menu
}

func QuizKeyboard(wordID int, options []string, correctIdx int) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	letters := []string{"A", "B", "C", "D"}
	var rows []tele.Row
	for i, opt := range options {
		if i >= 4 {
			break
		}
		label := fmt.Sprintf("%s) %s", letters[i], opt)
		if len(label) > 40 {
			label = label[:37] + "..."
		}
		rows = append(rows, menu.Row(
			menu.Data(label, "quiz", fmt.Sprintf("%d|%d", wordID, i)),
		))
	}
	menu.Inline(rows...)
	return menu
}

func SkipConfirmKeyboard(lang string) *tele.ReplyMarkup {
	m := content.GetMessages(lang)
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data(m.BtnYesSkip, "skip", "confirm"),
			menu.Data(m.BtnNoSkip, "skip", "cancel"),
		),
	)
	return menu
}

func ListenKeyboard(wordID int, lang string) *tele.ReplyMarkup {
	m := content.GetMessages(lang)
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data(m.BtnListen, "listen", fmt.Sprintf("%d", wordID)),
		),
	)
	return menu
}

func MediaDoneKeyboard(mediaID int, lang string) *tele.ReplyMarkup {
	m := content.GetMessages(lang)
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data(m.BtnDoneWatching, "media", fmt.Sprintf("done|%d", mediaID)),
		),
	)
	return menu
}

func ScheduleKeyboard(webAppURL, lang string) *tele.ReplyMarkup {
	m := content.GetMessages(lang)
	menu := &tele.ReplyMarkup{}
	rows := []tele.Row{}
	if webAppURL != "" {
		rows = append(rows, menu.Row(
			menu.WebApp(m.BtnCustomize, &tele.WebApp{URL: webAppURL + "/schedule"}),
		))
	}
	menu.Inline(rows...)
	return menu
}
