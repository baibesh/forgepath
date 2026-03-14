package bot

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func MainMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Reply(
		menu.Row(menu.Text("📖 Today"), menu.Text("📊 Stats")),
		menu.Row(menu.Text("⚙️ Settings")),
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
			menu.Data("UTC+2 🇪🇺", "tz", "2"),
			menu.Data("UTC+3 🇷🇺", "tz", "3"),
			menu.Data("UTC+4", "tz", "4"),
		),
		menu.Row(
			menu.Data("UTC+5 🇰🇿", "tz", "5"),
			menu.Data("UTC+6 🇰🇿", "tz", "6"),
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
			menu.Data("🕐 Timezone", "settings", "timezone"),
			menu.Data("📚 Level", "settings", "level"),
		),
	)
	return menu
}

func SkipConfirmKeyboard() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("✅ Yes, skip", "skip", "confirm"),
			menu.Data("❌ Cancel", "skip", "cancel"),
		),
	)
	return menu
}

func MediaDoneKeyboard(mediaID int) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("✅ I watched it!", "media", fmt.Sprintf("done|%d", mediaID)),
		),
	)
	return menu
}
