package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) removeInlineKeyboard(chatID int64, messageID int) {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, tgbotapi.InlineKeyboardMarkup{})
	b.api.Send(edit)
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

func (b *Bot) handleUpdate(update *tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallbackQuery(update)
	} else if update.Message.IsCommand() {
		b.handleCommand(update)
	} else if update.Message.Text != "" {
		//b.handleTextMessage(update)
	}
}

func (b *Bot) handleTextMessage(update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Echo: "+update.Message.Text)
	b.api.Send(msg)
}

func (b *Bot) handleCallbackQuery(update *tgbotapi.Update) {
	cb := update.CallbackQuery
	if cb == nil {
		return
	}

	b.api.Request(tgbotapi.NewCallback(cb.ID, ""))

	//b.removeInlineKeyboard(cb.Message.Chat.ID, cb.Message.MessageID)

	switch cb.Data {
	case "menu:main":
		b.handleMenuMainCallback(cb) // или handleMenuMainCallback (отдельная функция)
	case "menu:find":
		b.handleMenuFindCallback(cb)
	case "menu:profile":
		b.handleMenuProfileCallback(cb)
	default:
		return
	}
}

func (b *Bot) handleCommand(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		b.handleStartCommand(update.Message.Chat.ID)
	default:
		return
	}
}

func (b *Bot) handleStartCommand(chatID int64) {
	text, kb := renderMainMenu()
	b.displayMessage(chatID, nil, text, kb)
}

func (b *Bot) handleMenuMainCallback(cb *tgbotapi.CallbackQuery) {
	text, kb := renderMainMenu()
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, text, kb)
}

func (b *Bot) handleMenuFindCallback(cb *tgbotapi.CallbackQuery) {
	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Добавьте категориюи или введите название")
	edit.ReplyMarkup = renderCategoryKeyboard()
	b.api.Send(edit)
}

func (b *Bot) handleMenuProfileCallback(cb *tgbotapi.CallbackQuery) {
	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, "Пустой профиль")
	edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "menu:main"),
			},
		},
	}

	b.api.Send(edit)
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
