package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

func (b *Bot) handleUpdate(update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	b.sendText(update.Message.Chat.ID, "HELLO, HELLO")
}
