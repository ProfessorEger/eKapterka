package bot

import (
	"ekapterka/internal/models"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func renderCategoriesKeyboard(categories []models.Category, parentID *string) *tgbotapi.InlineKeyboardMarkup {

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, c := range categories {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				c.Title,
				"search:category:"+c.ID,
			),
		))
	}

	rows = append(rows, backButton(parentID))

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func renderItemsKeyboard(items []models.Item, categoryID string, page int, hasNext bool, backCallback string) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, item := range items {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				item.Title,
				"search:item:"+item.ID+":p:"+strconv.Itoa(page),
			),
		))
	}

	pagination := []tgbotapi.InlineKeyboardButton{}
	if page > 0 {
		pagination = append(pagination, tgbotapi.NewInlineKeyboardButtonData(
			"⬅",
			"search:items:"+categoryID+":"+strconv.Itoa(page-1),
		))
	}
	if hasNext {
		pagination = append(pagination, tgbotapi.NewInlineKeyboardButtonData(
			"➡",
			"search:items:"+categoryID+":"+strconv.Itoa(page+1),
		))
	}
	if len(pagination) > 0 {
		rows = append(rows, pagination)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"⬅ Назад",
			backCallback,
		),
	))

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (b *Bot) displayMessage(chatID int64, messageID *int, text string, kb *tgbotapi.InlineKeyboardMarkup) {
	b.displayMessageWithParseMode(chatID, messageID, text, kb, "")
}

func (b *Bot) displayMessageWithParseMode(chatID int64, messageID *int, text string, kb *tgbotapi.InlineKeyboardMarkup, parseMode string) {
	if messageID == nil {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = kb
		msg.ParseMode = parseMode
		b.api.Send(msg)
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID, *messageID, text)
	edit.ReplyMarkup = kb
	edit.ParseMode = parseMode
	if _, err := b.api.Send(edit); err == nil {
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	msg.ParseMode = parseMode
	if _, err := b.api.Send(msg); err == nil {
		b.deleteMessage(chatID, *messageID)
	}
}

func (b *Bot) displayPhotoMessage(chatID int64, messageID *int, photoURL, caption string, kb *tgbotapi.InlineKeyboardMarkup) {
	b.displayPhotoMessageWithParseMode(chatID, messageID, photoURL, caption, kb, "")
}

func (b *Bot) displayPhotoMessageWithParseMode(chatID int64, messageID *int, photoURL, caption string, kb *tgbotapi.InlineKeyboardMarkup, parseMode string) {
	if messageID != nil {
		media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(photoURL))
		media.Caption = truncateTelegramCaption(caption)
		media.ParseMode = parseMode

		edit := tgbotapi.EditMessageMediaConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      chatID,
				MessageID:   *messageID,
				ReplyMarkup: kb,
			},
			Media: media,
		}
		if _, err := b.api.Send(edit); err == nil {
			return
		}
	}

	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))
	msg.Caption = truncateTelegramCaption(caption)
	msg.ParseMode = parseMode
	msg.ReplyMarkup = kb
	if _, err := b.api.Send(msg); err == nil && messageID != nil {
		b.deleteMessage(chatID, *messageID)
	}
}

func (b *Bot) removeInlineKeyboard(chatID int64, messageID int) {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, tgbotapi.InlineKeyboardMarkup{})
	b.api.Send(edit)
}

func (b *Bot) deleteMessage(chatID int64, messageID int) {
	del := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.api.Send(del)
}

func truncateTelegramCaption(text string) string {
	const maxCaptionRunes = 1024
	runes := []rune(text)
	if len(runes) <= maxCaptionRunes {
		return text
	}
	return string(runes[:maxCaptionRunes-1]) + "…"
}

func backButton(parentID *string) []tgbotapi.InlineKeyboardButton {
	if parentID == nil {
		return []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				"⬅ Назад",
				"menu:main",
			),
		}
	}

	if *parentID == models.RootParentID {
		return []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				"⬅ Назад",
				"search:root",
			),
		}
	}

	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"⬅ Назад",
			"search:category:"+*parentID,
		),
	}
}
