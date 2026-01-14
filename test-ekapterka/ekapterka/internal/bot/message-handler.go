package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleTextMessage(update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Echo: "+update.Message.Text)
	b.api.Send(msg)
}
