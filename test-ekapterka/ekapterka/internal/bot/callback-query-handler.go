package bot

import (
	"ekapterka/internal/models"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCallbackQuery(update *tgbotapi.Update) {
	cb := update.CallbackQuery
	if cb == nil {
		return
	}

	b.api.Request(tgbotapi.NewCallback(cb.ID, ""))

	//b.removeInlineKeyboard(cb.Message.Chat.ID, cb.Message.MessageID)

	switch {
	case cb.Data == "menu:main":
		b.handleMenuMainCallback(cb) // или handleMenuMainCallback (отдельная функция)
	case cb.Data == "menu:find":
		b.handleMenuFindCallback(cb)
	case cb.Data == "menu:profile":
		b.handleMenuProfileCallback(cb)
	case cb.Data == "search:root":
		b.handleMenuFindCallback(cb)
	case strings.HasPrefix(cb.Data, "search:category:"):
		id := strings.TrimPrefix(cb.Data, "search:category:")
		b.handleCategorySelect(cb, id)
	}
}

func (b *Bot) handleMenuMainCallback(cb *tgbotapi.CallbackQuery) {
	text, kb := renderMainMenu()
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, text, kb)
}

func strPtr(s string) *string {
	return &s
}

func (b *Bot) handleMenuFindCallback(cb *tgbotapi.CallbackQuery) {
	categories, err := b.repo.GetChildCategories(b.ctx, strPtr(models.RootParentID))
	if err != nil {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Ошибка загрузки категорий",
			nil,
		)
		return
	}

	kb := renderCategoriesKeyboard(categories, nil)

	edit := tgbotapi.NewEditMessageText(
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		"Выберите категорию:",
	)
	edit.ReplyMarkup = kb
	b.api.Send(edit)
}

func (b *Bot) handleCategorySelect(cb *tgbotapi.CallbackQuery, categoryID string) {
	ctx := b.ctx

	// 1. Загружаем категорию
	cat, err := b.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Категория не найдена",
			nil,
		)
		return
	}

	// 2. Если лист — дальше будут товары
	if cat.IsLeaf {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Здесь будет список товаров",
			nil,
		)
		return
	}

	// 3. Загружаем подкатегории
	children, err := b.repo.GetChildCategories(ctx, &cat.ID)
	if err != nil {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Ошибка загрузки подкатегорий",
			nil,
		)
		return
	}

	// 4. Рендерим клавиатуру
	kb := renderCategoriesKeyboard(children, cat.ParentID)

	edit := tgbotapi.NewEditMessageText(
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		"Выберите подкатегорию:",
	)
	edit.ReplyMarkup = kb
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
