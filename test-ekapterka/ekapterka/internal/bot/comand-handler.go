package bot

import (
	"strings"

	"ekapterka/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCommand(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		b.handleStartCommand(update.Message.Chat.ID)
	case "add":
		b.handleAddCommand(update)
	default:
		return
	}
}

func (b *Bot) handleStartCommand(chatID int64) {
	text, kb := renderMainMenu()
	b.displayMessage(chatID, nil, text, kb)
}

func (b *Bot) handleAddCommand(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	lines := strings.Split(text, "\n")
	if len(lines) < 4 {
		msg := tgbotapi.NewMessage(
			chatID,
			"❌ Неверный формат.\n\nИспользуй:\n/add\nНазвание\nКатегорияID\nОписание",
		)
		b.api.Send(msg)
		return
	}

	title := strings.TrimSpace(lines[1])
	categoryID := strings.TrimSpace(lines[2])
	description := strings.TrimSpace(lines[3])

	if title == "" || categoryID == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ Название и категория обязательны")
		b.api.Send(msg)
		return
	}

	item := models.Item{
		Title:       title,
		Description: description,
		CategoryID:  categoryID,
		Tags:        []string{},
	}

	err := b.repo.AddItem(b.ctx, item)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при сохранении предмета")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Предмет успешно добавлен")
	b.api.Send(msg)
}
