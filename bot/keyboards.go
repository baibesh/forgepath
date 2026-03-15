package bot

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func LanguageSelectKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F1EC\U0001F1E7 English", "lang", "en"),
			menu.Data("\U0001F1E9\U0001F1EA Deutsch", "lang", "de"),
		),
	)
	return menu
}

func LevelSelectKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("A1", "level", "A1"),
			menu.Data("A2", "level", "A2"),
			menu.Data("B1", "level", "B1"),
		),
		menu.Row(
			menu.Data("B2", "level", "B2"),
			menu.Data("C1", "level", "C1"),
		),
	)
	return menu
}

func TimezoneKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("UTC+2 \U0001F1EA\U0001F1FA", "tz", "2"),
			menu.Data("UTC+3 \U0001F1F7\U0001F1FA", "tz", "3"),
			menu.Data("UTC+4", "tz", "4"),
		),
		menu.Row(
			menu.Data("UTC+5 \U0001F1F0\U0001F1FF", "tz", "5"),
			menu.Data("UTC+6 \U0001F1F0\U0001F1FF", "tz", "6"),
			menu.Data("UTC+7", "tz", "7"),
		),
		menu.Row(
			menu.Data("UTC+8", "tz", "8"),
			menu.Data("UTC+9", "tz", "9"),
			menu.Data("UTC+10", "tz", "10"),
		),
		menu.Row(
			menu.Data("Other (type number)", "tz", "custom"),
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

func SettingsKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F550 Timezone", "settings", "timezone"),
			menu.Data("\U0001F4DA Level", "settings", "level"),
		),
		menu.Row(
			menu.Data("\U0001F310 Language", "settings", "language"),
		),
	)
	return menu
}

func SkipConfirmKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\u2705 Yes, skip", "skip", "confirm"),
			menu.Data("\u274C Cancel", "skip", "cancel"),
		),
	)
	return menu
}

func ListenKeyboard(wordID int) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001F50A Listen", "listen", fmt.Sprintf("%d", wordID)),
		),
	)
	return menu
}

func MediaDoneKeyboard(mediaID int) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\u2705 I watched it!", "media", fmt.Sprintf("done|%d", mediaID)),
		),
	)
	return menu
}
