package bot

import tele "gopkg.in/telebot.v3"

func MainMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Reply(
		menu.Row(menu.Text("📖 Today"), menu.Text("📊 Stats")),
		menu.Row(menu.Text("⚙️ Settings")),
	)
	return menu
}
