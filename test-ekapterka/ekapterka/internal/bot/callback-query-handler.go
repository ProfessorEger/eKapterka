package bot

import (
	"ekapterka/internal/models"
	"strconv"
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
	case strings.HasPrefix(cb.Data, "search:items:"):
		b.handleItemsPageSelect(cb, strings.TrimPrefix(cb.Data, "search:items:"))
	case strings.HasPrefix(cb.Data, "search:item:"):
		b.handleItemSelect(cb, strings.TrimPrefix(cb.Data, "search:item:"))
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

	// 2. Если лист — показываем товары
	if cat.IsLeaf {
		b.showItemsPage(cb, cat.ID, 0)
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

func (b *Bot) handleItemsPageSelect(cb *tgbotapi.CallbackQuery, payload string) {
	parts := strings.Split(payload, ":")
	if len(parts) != 2 {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Некорректный запрос списка товаров",
			nil,
		)
		return
	}

	categoryID := parts[0]
	page, err := strconv.Atoi(parts[1])
	if err != nil || page < 0 {
		page = 0
	}

	b.showItemsPage(cb, categoryID, page)
}

func (b *Bot) showItemsPage(cb *tgbotapi.CallbackQuery, categoryID string, page int) {
	backCallback := b.getBackCallbackForCategory(categoryID)

	items, hasNext, err := b.repo.GetItemsByCategoryPage(b.ctx, categoryID, page, 10)
	if err != nil {
		edit := tgbotapi.NewEditMessageText(
			cb.Message.Chat.ID,
			cb.Message.MessageID,
			"Ошибка загрузки товаров",
		)
		edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", backCallback),
				},
			},
		}
		b.api.Send(edit)
		return
	}

	if len(items) == 0 {
		edit := tgbotapi.NewEditMessageText(
			cb.Message.Chat.ID,
			cb.Message.MessageID,
			"В этой категории пока нет товаров",
		)
		edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", backCallback),
				},
			},
		}
		b.api.Send(edit)
		return
	}

	kb := renderItemsKeyboard(items, categoryID, page, hasNext, backCallback)
	edit := tgbotapi.NewEditMessageText(
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		"Выберите товар:",
	)
	edit.ReplyMarkup = kb
	b.api.Send(edit)
}

func (b *Bot) getBackCallbackForCategory(categoryID string) string {
	cat, err := b.repo.GetCategoryByID(b.ctx, categoryID)
	if err != nil || cat.ParentID == nil {
		return "search:root"
	}

	if *cat.ParentID == models.RootParentID {
		return "search:root"
	}

	return "search:category:" + *cat.ParentID
}

func (b *Bot) handleItemSelect(cb *tgbotapi.CallbackQuery, payload string) {
	parts := strings.Split(payload, ":p:")
	if len(parts) != 2 {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Некорректный запрос товара",
			nil,
		)
		return
	}

	itemID := parts[0]
	page, err := strconv.Atoi(parts[1])
	if err != nil || page < 0 {
		page = 0
	}

	item, err := b.repo.GetItemByID(b.ctx, itemID)
	if err != nil {
		b.displayMessage(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			"Товар не найден",
			nil,
		)
		return
	}

	text := item.Title + "/n" + item.Description

	edit := tgbotapi.NewEditMessageText(
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		text,
	)
	edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData(
					"⬅ Назад",
					"search:items:"+item.CategoryID+":"+strconv.Itoa(page),
				),
			},
		},
	}
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
