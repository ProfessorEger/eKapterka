package bot

// Файл реализует обработку callback_query от inline-кнопок:
// навигацию по категориям, пагинацию предметов и рендер карточек товаров.

import (
	"ekapterka/internal/models"
	"encoding/json"
	"html"
	"log"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCallbackQuery роутит inline callback-события меню и каталога.
func (b *Bot) handleCallbackQuery(update *tgbotapi.Update) {
	cb := update.CallbackQuery
	if cb == nil {
		return
	}

	b.api.Request(tgbotapi.NewCallback(cb.ID, ""))

	switch {
	case cb.Data == "menu:main":
		b.handleMenuMainCallback(cb)
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

// handleMenuMainCallback возвращает пользователя в главное меню.
func (b *Bot) handleMenuMainCallback(cb *tgbotapi.CallbackQuery) {
	text, kb := renderMainMenu()
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, text, kb)
}

// strPtr — локальный helper для передачи строковых указателей.
func strPtr(s string) *string {
	return &s
}

// handleMenuFindCallback открывает корневые категории каталога.
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
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, "Выберите категорию:", kb)
}

// handleCategorySelect обрабатывает выбор категории:
// лист -> список предметов, не лист -> подкатегории.
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
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, "Выберите подкатегорию:", kb)
}

// handleItemsPageSelect парсит callback пагинации и открывает нужную страницу.
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

// showItemsPage отображает список предметов категории с пагинацией и кнопкой "Назад".
func (b *Bot) showItemsPage(cb *tgbotapi.CallbackQuery, categoryID string, page int) {
	backCallback := b.getBackCallbackForCategory(categoryID)

	items, hasNext, err := b.repo.GetItemsByCategoryPage(b.ctx, categoryID, page, 10)
	if err != nil {
		b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, "Ошибка загрузки товаров", &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", backCallback),
				},
			},
		})
		return
	}

	if len(items) == 0 {
		b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, "В этой категории пока нет товаров", &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", backCallback),
				},
			},
		})
		return
	}

	kb := renderItemsKeyboard(items, categoryID, page, hasNext, backCallback)
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, "Выберите товар:", kb)
}

// getBackCallbackForCategory вычисляет, куда должна вести кнопка "Назад" из текущей категории.
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

// handleItemSelect открывает карточку конкретного предмета.
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

	isAdmin := false
	if cb.From != nil {
		isAdmin, err = b.isAdminUser(cb.From.ID)
		if err != nil {
			log.Printf("resolve role for user %d failed: %v", cb.From.ID, err)
		}
	}

	text, parseMode := renderItemCardText(item, isAdmin, b.getQuartermasterContact())

	kb := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData(
					"⬅ Назад",
					"search:items:"+item.CategoryID+":"+strconv.Itoa(page),
				),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "menu:main"),
			},
		},
	}

	if photoURL := firstPhotoURL(item.PhotoURLs); photoURL != "" {
		b.displayPhotoMessageWithParseMode(
			cb.Message.Chat.ID,
			&cb.Message.MessageID,
			photoURL,
			text,
			kb,
			parseMode,
		)
		return
	}

	b.displayMessageWithParseMode(cb.Message.Chat.ID, &cb.Message.MessageID, text, kb, parseMode)
}

// renderItemCardText собирает HTML-текст карточки предмета.
// Для админа дополнительно показываются CategoryID и описания аренд.
func renderItemCardText(item *models.Item, isAdmin bool, quartermasterContact string) (string, string) {
	const rentalDateLayout = "02.01.2006"

	lines := []string{
		"<code>" + html.EscapeString(item.ID) + "</code>",
	}
	if isAdmin {
		lines = append(lines, "<code>"+html.EscapeString(item.CategoryID)+"</code>")
	}

	lines = append(lines, "<b>"+html.EscapeString(item.Title)+"</b>")
	if desc := strings.TrimSpace(item.Description); desc != "" {
		lines = append(lines, html.EscapeString(desc))
	}

	if len(item.Rentals) == 0 {
		lines = append(lines, "")
		lines = append(lines, "✅ Свободно")
	} else {
		lines = append(lines, "")
		lines = append(lines, "Арендовано:")
		periods := make([]models.Rental, len(item.Rentals))
		copy(periods, item.Rentals)
		sort.Slice(periods, func(i, j int) bool {
			if periods[i].Start.Equal(periods[j].Start) {
				return periods[i].End.Before(periods[j].End)
			}
			return periods[i].Start.Before(periods[j].Start)
		})

		for i, period := range periods {
			lines = append(lines, strconv.Itoa(i+1)+". "+period.Start.Format(rentalDateLayout)+"-"+period.End.Format(rentalDateLayout))
			if isAdmin {
				if desc := strings.TrimSpace(period.Description); desc != "" {
					lines = append(lines, "Описание: "+html.EscapeString(desc))
				}
			}
		}
	}

	lines = append(lines, "\nДля того чтобы арендовать, пишите каптерщику "+html.EscapeString(quartermasterContact))

	return strings.Join(lines, "\n"), tgbotapi.ModeHTML
}

// getQuartermasterContact получает контакт каптерщика из описания бота
// (первое слово) и кэширует результат в памяти процесса.
func (b *Bot) getQuartermasterContact() string {
	const fallbackContact = ""

	b.quartermasterContactOnce.Do(func() {
		b.quartermasterContact = fallbackContact

		resp, err := b.api.MakeRequest("getMyDescription", nil)
		if err != nil {
			log.Printf("get bot description failed: %v", err)
			return
		}

		var descriptionResp struct {
			Description string `json:"description"`
		}
		if err := json.Unmarshal(resp.Result, &descriptionResp); err != nil {
			log.Printf("decode bot description failed: %v", err)
			return
		}

		words := strings.Fields(strings.TrimSpace(descriptionResp.Description))
		if len(words) == 0 {
			return
		}

		contact := strings.Trim(words[0], ".,;:!?()[]{}")
		if contact == "" {
			return
		}

		b.quartermasterContact = contact
	})

	return b.quartermasterContact
}

// firstPhotoURL возвращает первый непустой URL фото.
func firstPhotoURL(photoURLs []string) string {
	for _, photoURL := range photoURLs {
		photoURL = strings.TrimSpace(photoURL)
		if photoURL != "" {
			return photoURL
		}
	}
	return ""
}

// handleMenuProfileCallback — заглушка профиля пользователя.
func (b *Bot) handleMenuProfileCallback(cb *tgbotapi.CallbackQuery) {
	b.displayMessage(cb.Message.Chat.ID, &cb.Message.MessageID, "Пустой профиль", &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "menu:main"),
			},
		},
	})
}
