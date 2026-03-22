package bot

// Файл содержит обработчики slash-команд (/start, /add, /edit, /rent и т.д.),
// включая проверки ролей, парсинг пользовательского ввода и вызовы repository/storage.

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"ekapterka/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand роутит slash-команды в соответствующие обработчики.
// Команда может прийти как в тексте, так и в caption (для сообщений с фото).
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
	case "grantadmin":
		b.handleGrantAdminCommand(update)
	case "revokeadmin":
		b.handleRevokeAdminCommand(update)
	default:
		return
	}
}

// handleStartCommand инициализирует пользователя и показывает главное меню.
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

// handleGetAdminCommand повышает роль пользователя до admin при корректном коде.
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

// handleGrantAdminCommand выдает роль admin указанному пользователю при корректном коде.
// Доступна только действующим администраторам.
func (b *Bot) handleGrantAdminCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	args := strings.Fields(strings.TrimSpace(update.Message.CommandArguments()))
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Формат: /grantadmin <user_id> <код>")
		b.api.Send(msg)
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
	if err != nil || userID <= 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ Некорректный user_id")
		b.api.Send(msg)
		return
	}

	code := strings.TrimSpace(args[1])
	adminCode := strings.TrimSpace(os.Getenv("ADMIN_CODE"))
	if adminCode == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ ADMIN_CODE не настроен на сервере")
		b.api.Send(msg)
		return
	}
	if code != adminCode {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный код")
		b.api.Send(msg)
		return
	}

	if err := b.repo.SetUserRole(b.ctx, userID, models.ADMIN); err != nil {
		log.Printf("set admin role failed for user %d: %v", userID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Не удалось выдать роль администратора")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Роль администратора выдана")
	b.api.Send(msg)
}

// handleRevokeAdminCommand снимает роль admin с указанного пользователя при корректном коде.
// Доступна только действующим администраторам.
func (b *Bot) handleRevokeAdminCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	args := strings.Fields(strings.TrimSpace(update.Message.CommandArguments()))
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Формат: /revokeadmin <user_id> <код>")
		b.api.Send(msg)
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
	if err != nil || userID <= 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ Некорректный user_id")
		b.api.Send(msg)
		return
	}

	code := strings.TrimSpace(args[1])
	adminCode := strings.TrimSpace(os.Getenv("ADMIN_CODE"))
	if adminCode == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ ADMIN_CODE не настроен на сервере")
		b.api.Send(msg)
		return
	}
	if code != adminCode {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный код")
		b.api.Send(msg)
		return
	}

	if err := b.repo.SetUserRole(b.ctx, userID, models.USER); err != nil {
		log.Printf("revoke admin role failed for user %d: %v", userID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Не удалось снять роль администратора")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Роль администратора снята")
	b.api.Send(msg)
}

// handleAddCommand создает новый предмет.
// Формат поддерживает многострочное описание и опциональное фото.
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
	if len(lines) < 3 {
		msg := tgbotapi.NewMessage(
			chatID,
			"❌ Неверный формат.\n\nИспользуй:\n/add\nКатегорияID\nНазвание\n[Описание - опционально]",
		)
		b.api.Send(msg)
		return
	}

	categoryID := strings.TrimSpace(lines[1])
	title := strings.TrimSpace(lines[2])
	description := ""
	if len(lines) > 3 {
		description = strings.TrimSpace(strings.Join(lines[3:], "\n"))
	}

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

// handleLeafCategoriesCommand выводит список листовых категорий для админских команд.
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
		if strings.TrimSpace(title) == "" {
			title = cat.ID
		}
		sb.WriteString(fmt.Sprintf("<code>%s</code> %s\n", html.EscapeString(cat.ID), html.EscapeString(title)))
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = tgbotapi.ModeHTML
	b.api.Send(msg)
}

// handleEditCommand редактирует существующий предмет.
// При наличии нового фото старые ссылки заменяются, старые объекты удаляются.
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
	if len(lines) < 3 {
		msg := tgbotapi.NewMessage(
			chatID,
			"❌ Неверный формат.\n\nИспользуй:\n/edit <id>\nНовая категория\nНовое название\n[Новое описание - опционально]",
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
	description := ""
	if len(lines) > 3 {
		description = strings.TrimSpace(strings.Join(lines[3:], "\n"))
	}

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

// handleDeleteCommand удаляет предмет и связанные фото из GCS.
// ID может быть передан аргументом команды или в первой строке caption/text.
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

// handleRentCommand добавляет период аренды к предмету.
// Даты принимаются в формате DD.MM.YYYY.
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
	if len(lines) < 4 {
		msg := tgbotapi.NewMessage(
			chatID,
			"❌ Неверный формат.\n\nИспользуй:\n/rent <id>\n01.01.2025\n10.02.2025\ntg_id\n[описание для админа - опционально]",
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
	userIDRaw := strings.TrimSpace(lines[3])
	description := ""
	if len(lines) > 4 {
		description = strings.TrimSpace(strings.Join(lines[4:], "\n"))
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

	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil || userID <= 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ Некорректный tg_id. Укажите числовой Telegram ID пользователя.")
		b.api.Send(msg)
		return
	}

	period := models.Rental{
		Start:       startDate,
		End:         endDate,
		Description: description,
		UserID:      userID,
	}

	if err = b.repo.AddRental(b.ctx, itemID, period); err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при добавлении аренды")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Период аренды добавлен")
	b.api.Send(msg)
}

// handleUnrentCommand удаляет аренду по ID.
func (b *Bot) handleUnrentCommand(update *tgbotapi.Update) {
	if !b.requireAdmin(update) {
		return
	}

	chatID := update.Message.Chat.ID
	args := strings.Fields(strings.TrimSpace(update.Message.CommandArguments()))
	if len(args) < 1 {
		msg := tgbotapi.NewMessage(chatID, "❌ Формат: /unr <rental_id>")
		b.api.Send(msg)
		return
	}

	rentalID := strings.TrimSpace(args[0])
	if rentalID == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ Укажите rental_id")
		b.api.Send(msg)
		return
	}

	rental, err := b.repo.GetRentalByID(b.ctx, rentalID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Аренда не найдена")
		b.api.Send(msg)
		return
	}

	if err := b.repo.DeleteRental(b.ctx, rental); err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при отмене аренды")
		b.api.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Аренда отменена")
	b.api.Send(msg)
}

// handleCommandsCommand показывает список доступных команд для текущей роли.
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
			"/grantadmin",
			"/revokeadmin",
		)
	}

	msg := tgbotapi.NewMessage(chatID, strings.Join(lines, "\n"))
	b.api.Send(msg)
}

// requireAdmin проверяет роль пользователя и отправляет сообщение об ошибке,
// если команда недоступна по правам.
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

// isAdminUser возвращает true, если роль пользователя — admin.
func (b *Bot) isAdminUser(userID int64) (bool, error) {
	role, err := b.repo.GetUserRole(b.ctx, userID)
	if err != nil {
		return false, err
	}
	return role == models.ADMIN, nil
}

// uploadMessagePhotos загружает самое большое фото из Telegram в GCS.
// Возвращает публичный URL загруженного объекта.
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

// deleteItemPhotos удаляет набор фото по их публичным URL.
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

// gcsObjectPathFromPublicURL извлекает путь объекта из поддерживаемых форматов GCS URL.
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

// extractCommandFromCaption достает slash-команду из caption entities.
// Нужно для сценариев, где команда отправляется вместе с фото.
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

// resolveImageExt определяет расширение файла по Content-Type ответа Telegram.
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
