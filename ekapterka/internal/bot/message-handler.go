package bot

// Файл содержит обработку обычных текстовых сообщений (не команд).
// Сейчас это базовая заготовка/echo для потенциального расширения UX-логики.

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleTextMessage — заготовка для обработки обычного текста (не команд).
// Сейчас используется как simple echo и в основном потоке отключена.
func (b *Bot) handleTextMessage(update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Echo: "+update.Message.Text)
	b.api.Send(msg)
}
