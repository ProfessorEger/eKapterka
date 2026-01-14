package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
