package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleUpdate(update *tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallbackQuery(update)
	} else if update.Message.IsCommand() {
		b.handleCommand(update)
	} else if update.Message.Text != "" {
		//b.handleTextMessage(update)
	}
}

func renderMainMenu() (string, *tgbotapi.InlineKeyboardMarkup) {
	text := "Главное меню:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Найти снаряжение", "menu:find"),
			tgbotapi.NewInlineKeyboardButtonData("👤 Мой профиль", "menu:profile"),
		),
	)

	return text, &keyboard
}

func renderCategoryKeyboard() *tgbotapi.InlineKeyboardMarkup {
	categories := []string{
		"Палатки",
		"Рюкзаки",
		"Спальники",
		"Обувь",
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, c := range categories {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			c,
			"search:category:"+c,
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "menu:main"),
	))

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (b *Bot) displayMessage(chatID int64, messageID *int, text string, kb *tgbotapi.InlineKeyboardMarkup) {
	if messageID == nil {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = kb
		b.api.Send(msg)
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID, *messageID, text)
	edit.ReplyMarkup = kb
	b.api.Send(edit)
}

func (b *Bot) removeInlineKeyboard(chatID int64, messageID int) {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, tgbotapi.InlineKeyboardMarkup{})
	b.api.Send(edit)
}
