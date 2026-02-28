package bot

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
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
		b.handleStartCommand(update.Message)
	case "add":
		b.handleAddCommand(update)
	case "edit":
		b.handleEditCommand(update)
	case "getadmin":
		b.handleGetAdminCommand(update)
	default:
		return
	}
}

func (b *Bot) handleStartCommand(msg *tgbotapi.Message) {
	if msg == nil {
		return
	}

	userID := msg.Chat.ID
	if msg.From != nil {
		userID = msg.From.ID
	}

	if err := b.repo.EnsureUserState(b.ctx, userID); err != nil {
		log.Printf("ensure user state failed for user %d: %v", userID, err)
	}

	text, kb := renderMainMenu()
	b.displayMessage(msg.Chat.ID, nil, text, kb)
}

func (b *Bot) handleGetAdminCommand(update *tgbotapi.Update) {
	if update == nil || update.Message == nil {
		return
	}

	code := strings.TrimSpace(update.Message.CommandArguments())
	if code == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Укажите код: /getadmin <код>")
		b.api.Send(msg)
		return
	}

	adminCode := strings.TrimSpace(os.Getenv("ADMIN_CODE"))
	if adminCode == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ ADMIN_CODE не настроен на сервере")
		b.api.Send(msg)
		return
	}
	if code != adminCode {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Неверный код")
		b.api.Send(msg)
		return
	}

	userID := update.Message.Chat.ID
	if update.Message.From != nil {
		userID = update.Message.From.ID
	}

	if err := b.repo.SetUserRole(b.ctx, userID, models.ADMIN); err != nil {
		log.Printf("set admin role failed for user %d: %v", userID, err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Не удалось выдать роль администратора")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Роль администратора выдана")
	b.api.Send(msg)

}

func (b *Bot) handleAddCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

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

//TODO: команда для админа. вывести список категорий в которых можно добавлять предметы (LEAF категории)

func (b *Bot) handleEditCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text
	if strings.TrimSpace(text) == "" {
		text = update.Message.Caption
	}

	lines := strings.Split(text, "\n")
	if len(lines) < 4 {
		msg := tgbotapi.NewMessage(
			chatID,
			"❌ Неверный формат.\n\nИспользуй:\n/edit <id>\nНовое название\nНовая категория\nНовое описание",
		)
		b.api.Send(msg)
		return
	}

	headFields := strings.Fields(strings.TrimSpace(lines[0]))
	if len(headFields) < 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Укажите ID: /edit <id>")
		b.api.Send(msg)
		return
	}

	itemID := strings.TrimSpace(headFields[1])
	title := strings.TrimSpace(lines[1])
	categoryID := strings.TrimSpace(lines[2])
	description := strings.TrimSpace(lines[3])

	if itemID == "" || title == "" || categoryID == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ ID, название и категория обязательны")
		b.api.Send(msg)
		return
	}

	oldItem, err := b.repo.GetItemByID(b.ctx, itemID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Предмет не найден")
		b.api.Send(msg)
		return
	}

	photoURLs := oldItem.PhotoURLs
	replacePhoto := len(update.Message.Photo) > 0
	if replacePhoto {
		photoURLs, err = b.uploadMessagePhotos(update.Message)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при загрузке нового фото")
			b.api.Send(msg)
			return
		}
	}

	err = b.repo.UpdateItem(b.ctx, itemID, title, categoryID, description, photoURLs)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при обновлении предмета")
		b.api.Send(msg)
		return
	}

	if replacePhoto {
		if err := b.deleteItemPhotos(oldItem.PhotoURLs); err != nil {
			log.Printf("delete old photos for item %s failed: %v", itemID, err)
			msg := tgbotapi.NewMessage(chatID, "⚠️ Предмет обновлен, но старое фото удалить не удалось")
			b.api.Send(msg)
			return
		}
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Предмет успешно обновлен")
	b.api.Send(msg)
}

//TODO: команда для админа. удалить предмет (по id). формат команды: /delete <id>. При этом удалить фото из хранилища

func (b *Bot) requireAdmin(update *tgbotapi.Update) bool {
	if update == nil || update.Message == nil {
		return false
	}

	userID := update.Message.Chat.ID
	if update.Message.From != nil {
		userID = update.Message.From.ID
	}

	role, err := b.repo.GetUserRole(b.ctx, userID)
	if err != nil {
		log.Printf("get user role failed for user %d: %v", userID, err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Не удалось проверить права доступа")
		b.api.Send(msg)
		return false
	}
	if role != models.ADMIN {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⛔ Команда доступна только администратору")
		b.api.Send(msg)
		return false
	}

	return true
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

func (b *Bot) deleteItemPhotos(photoURLs []string) error {
	if b.gcs == nil {
		return fmt.Errorf("gcs is not configured")
	}

	var failed []string
	for _, photoURL := range photoURLs {
		objectPath := gcsObjectPathFromPublicURL(photoURL)
		if objectPath == "" {
			continue
		}
		if err := b.gcs.Delete(b.ctx, objectPath); err != nil {
			failed = append(failed, err.Error())
		}
	}

	if len(failed) > 0 {
		return errors.New(strings.Join(failed, "; "))
	}

	return nil
}

func gcsObjectPathFromPublicURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if u.Host == "storage.googleapis.com" {
		pathValue := strings.TrimPrefix(u.Path, "/")
		parts := strings.SplitN(pathValue, "/", 2)
		if len(parts) != 2 {
			return ""
		}
		return strings.TrimSpace(parts[1])
	}

	if strings.HasSuffix(u.Host, ".storage.googleapis.com") {
		return strings.TrimPrefix(strings.TrimSpace(u.Path), "/")
	}

	return ""
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
