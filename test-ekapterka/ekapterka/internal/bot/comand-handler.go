package bot

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"ekapterka/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCommand(update *tgbotapi.Update) {
	command := update.Message.Command()
	if command == "" {
		command = extractCommandFromCaption(update.Message)
	}

	switch command {
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
	if strings.TrimSpace(text) == "" {
		text = update.Message.Caption
	}

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

	photoURLs, err := b.uploadMessagePhotos(update.Message)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при загрузке фото")
		b.api.Send(msg)
		return
	}

	item := models.Item{
		Title:        title,
		Description:  description,
		CategoryID:   categoryID,
		CategoryPath: []string{},
		Tags:         []string{},
		PhotoURLs:    photoURLs,
	}

	err = b.repo.AddItem(b.ctx, item)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при сохранении предмета")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Предмет успешно добавлен")
	b.api.Send(msg)
}

func (b *Bot) uploadMessagePhotos(msg *tgbotapi.Message) ([]string, error) {
	if msg == nil || len(msg.Photo) == 0 {
		return []string{}, nil
	}
	if b.gcs == nil {
		return nil, fmt.Errorf("gcs is not configured")
	}

	photos := msg.Photo
	largestPhoto := photos[len(photos)-1]

	fileURL, err := b.api.GetFileDirectURL(largestPhoto.FileID)
	if err != nil {
		return nil, fmt.Errorf("get telegram file url: %w", err)
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("download telegram file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("telegram file download status: %d", resp.StatusCode)
	}

	objectPath := fmt.Sprintf(
		"items/%d/%d_%s%s",
		msg.Chat.ID,
		time.Now().UnixNano(),
		largestPhoto.FileUniqueID,
		resolveImageExt(resp.Header.Get("Content-Type")),
	)

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	if err := b.gcs.Upload(b.ctx, objectPath, resp.Body, contentType); err != nil {
		return nil, fmt.Errorf("upload to gcs: %w", err)
	}

	return []string{b.gcs.PublicURL(objectPath)}, nil
}

func extractCommandFromCaption(msg *tgbotapi.Message) string {
	if msg == nil {
		return ""
	}

	for _, entity := range msg.CaptionEntities {
		if entity.Offset != 0 || !entity.IsCommand() {
			continue
		}

		runes := []rune(msg.Caption)
		if entity.Length > len(runes) {
			return ""
		}

		command := strings.TrimPrefix(string(runes[:entity.Length]), "/")
		if atIndex := strings.Index(command, "@"); atIndex >= 0 {
			command = command[:atIndex]
		}
		return command
	}

	return ""
}

func resolveImageExt(contentType string) string {
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
