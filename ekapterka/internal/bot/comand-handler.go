package bot

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
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
	case "cat":
		b.handleLeafCategoriesCommand(update)
	case "rm":
		b.handleDeleteCommand(update)
	case "rent":
		b.handleRentCommand(update)
	case "unr":
		b.handleUnrentCommand(update)
	case "cmd":
		b.handleCommandsCommand(update)
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
			"❌ Неверный формат.\n\nИспользуй:\n/add\nКатегорияID\nНазвание\nОписание",
		)
		b.api.Send(msg)
		return
	}

	categoryID := strings.TrimSpace(lines[1])
	title := strings.TrimSpace(lines[2])
	description := strings.TrimSpace(strings.Join(lines[3:], "\n"))

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
		Title:       title,
		Description: description,
		CategoryID:  categoryID,
		Tags:        []string{},
		PhotoURLs:   photoURLs,
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

func (b *Bot) handleLeafCategoriesCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	categories, err := b.repo.GetLeafCategories(b.ctx)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Ошибка загрузки категорий")
		b.api.Send(msg)
		return
	}
	if len(categories) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "LEAF категории не найдены")
		b.api.Send(msg)
		return
	}

	var sb strings.Builder
	for _, cat := range categories {
		title := cat.Title
		if strings.TrimSpace(title) == "" && len(cat.Path) > 0 {
			title = cat.Path[len(cat.Path)-1]
		}
		sb.WriteString(fmt.Sprintf("<code>%s</code> %s\n", html.EscapeString(cat.ID), html.EscapeString(title)))
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = tgbotapi.ModeHTML
	b.api.Send(msg)
}

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
			"❌ Неверный формат.\n\nИспользуй:\n/edit <id>\nНовая категория\nНовое название\nНовое описание",
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
	categoryID := strings.TrimSpace(lines[1])
	title := strings.TrimSpace(lines[2])
	description := strings.TrimSpace(strings.Join(lines[3:], "\n"))

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

func (b *Bot) handleDeleteCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	itemID := strings.TrimSpace(update.Message.CommandArguments())
	if itemID == "" {
		text := strings.TrimSpace(update.Message.Text)
		if text == "" {
			text = strings.TrimSpace(update.Message.Caption)
		}
		firstLine := strings.Split(text, "\n")[0]
		fields := strings.Fields(firstLine)
		if len(fields) > 1 {
			itemID = strings.TrimSpace(fields[1])
		}
	}

	if itemID == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ Укажите ID: /delete <id>")
		b.api.Send(msg)
		return
	}

	item, err := b.repo.GetItemByID(b.ctx, itemID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Предмет не найден")
		b.api.Send(msg)
		return
	}

	if err := b.repo.DeleteItemByID(b.ctx, itemID); err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при удалении предмета")
		b.api.Send(msg)
		return
	}

	if err := b.deleteItemPhotos(item.PhotoURLs); err != nil {
		log.Printf("delete photos for item %s failed: %v", itemID, err)
		msg := tgbotapi.NewMessage(chatID, "⚠️ Предмет удален, но фото удалить не удалось")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Предмет успешно удален")
	b.api.Send(msg)
}

func (b *Bot) handleRentCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text
	if strings.TrimSpace(text) == "" {
		text = update.Message.Caption
	}

	lines := strings.Split(text, "\n")
	if len(lines) < 3 {
		msg := tgbotapi.NewMessage(
			chatID,
			"❌ Неверный формат.\n\nИспользуй:\n/rent <id>\n01.01.2025\n10.02.2025\n[описание для админа - опционально]",
		)
		b.api.Send(msg)
		return
	}

	headFields := strings.Fields(strings.TrimSpace(lines[0]))
	if len(headFields) < 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Укажите ID: /rent <id>")
		b.api.Send(msg)
		return
	}

	itemID := strings.TrimSpace(headFields[1])
	startRaw := strings.TrimSpace(lines[1])
	endRaw := strings.TrimSpace(lines[2])
	description := ""
	if len(lines) > 3 {
		description = strings.TrimSpace(strings.Join(lines[3:], "\n"))
	}

	const rentalDateLayout = "02.01.2006"
	startDate, err := time.Parse(rentalDateLayout, startRaw)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Некорректная дата начала. Формат: 01.01.2025")
		b.api.Send(msg)
		return
	}

	endDate, err := time.Parse(rentalDateLayout, endRaw)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Некорректная дата окончания. Формат: 10.02.2025")
		b.api.Send(msg)
		return
	}

	if endDate.Before(startDate) {
		msg := tgbotapi.NewMessage(chatID, "❌ Дата окончания не может быть раньше даты начала")
		b.api.Send(msg)
		return
	}

	if _, err = b.repo.GetItemByID(b.ctx, itemID); err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Предмет не найден")
		b.api.Send(msg)
		return
	}

	period := models.Rental{
		Start:       startDate,
		End:         endDate,
		Description: description,
	}

	if err = b.repo.AddRentalPeriodToItem(b.ctx, itemID, period); err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при добавлении аренды")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Период аренды добавлен")
	b.api.Send(msg)
}

func (b *Bot) handleUnrentCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	args := strings.Fields(strings.TrimSpace(update.Message.CommandArguments()))
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Формат: /unr <id> <номер аренды>")
		b.api.Send(msg)
		return
	}

	itemID := strings.TrimSpace(args[0])
	rentalNumber, err := strconv.Atoi(strings.TrimSpace(args[1]))
	if err != nil || rentalNumber <= 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ Номер аренды должен быть положительным числом")
		b.api.Send(msg)
		return
	}

	item, err := b.repo.GetItemByID(b.ctx, itemID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Предмет не найден")
		b.api.Send(msg)
		return
	}

	if len(item.Rentals) == 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ У предмета нет аренд")
		b.api.Send(msg)
		return
	}

	type indexedRental struct {
		Index  int
		Rental models.Rental
	}
	sorted := make([]indexedRental, 0, len(item.Rentals))
	for i, rental := range item.Rentals {
		sorted = append(sorted, indexedRental{Index: i, Rental: rental})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Rental.Start.Equal(sorted[j].Rental.Start) {
			return sorted[i].Rental.End.Before(sorted[j].Rental.End)
		}
		return sorted[i].Rental.Start.Before(sorted[j].Rental.Start)
	})

	if rentalNumber > len(sorted) {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Номер аренды вне диапазона: 1..%d", len(sorted)))
		b.api.Send(msg)
		return
	}

	removeIndex := sorted[rentalNumber-1].Index
	updated := make([]models.Rental, 0, len(item.Rentals)-1)
	updated = append(updated, item.Rentals[:removeIndex]...)
	updated = append(updated, item.Rentals[removeIndex+1:]...)

	if err := b.repo.UpdateItemRentals(b.ctx, itemID, updated); err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при отмене аренды")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Аренда отменена")
	b.api.Send(msg)
}

func (b *Bot) handleCommandsCommand(update *tgbotapi.Update) {
	if update == nil || update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	lines := []string{
		"Команды:",
		"/start",
	}

	userID := chatID
	if update.Message.From != nil {
		userID = update.Message.From.ID
	}

	isAdmin, err := b.isAdminUser(userID)
	if err != nil {
		log.Printf("resolve role for user %d failed: %v", userID, err)
	}
	if isAdmin {
		lines = append(lines,
			"/add",
			"/edit",
			"/rm",
			"/cat",
			"/rent",
			"/unr",
		)
	}

	msg := tgbotapi.NewMessage(chatID, strings.Join(lines, "\n"))
	b.api.Send(msg)
}

func (b *Bot) requireAdmin(update *tgbotapi.Update) bool {
	if update == nil || update.Message == nil {
		return false
	}

	userID := update.Message.Chat.ID
	if update.Message.From != nil {
		userID = update.Message.From.ID
	}

	isAdmin, err := b.isAdminUser(userID)
	if err != nil {
		log.Printf("get user role failed for user %d: %v", userID, err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Не удалось проверить права доступа")
		b.api.Send(msg)
		return false
	}
	if !isAdmin {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⛔ Команда доступна только администратору")
		b.api.Send(msg)
		return false
	}

	return true
}

func (b *Bot) isAdminUser(userID int64) (bool, error) {
	role, err := b.repo.GetUserRole(b.ctx, userID)
	if err != nil {
		return false, err
	}
	return role == models.ADMIN, nil
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
