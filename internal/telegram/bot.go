package telegram

import (
	"fmt"
	"log"
	"strconv"

	"diabetbot/internal/config"
	"diabetbot/internal/models"
	"diabetbot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	userService     *services.UserService
	glucoseService  *services.GlucoseService
	foodService     *services.FoodService
	gigachatService *services.GigaChatService
	config          *config.TelegramConfig
}

func NewBot(cfg *config.TelegramConfig, db *gorm.DB, gigachatService *services.GigaChatService) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Bot{
		api:             bot,
		userService:     services.NewUserService(db),
		glucoseService:  services.NewGlucoseService(db),
		foodService:     services.NewFoodService(db),
		gigachatService: gigachatService,
		config:          cfg,
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	user, err := b.userService.GetOrCreateUser(message.From.ID, message.From.UserName, 
		message.From.FirstName, message.From.LastName, message.From.LanguageCode)
	if err != nil {
		log.Printf("Error getting/creating user: %v", err)
		return
	}

	switch {
	case message.IsCommand():
		b.handleCommand(message, user)
	case isNumeric(message.Text):
		b.handleGlucoseInput(message, user)
	case b.isKeyboardButton(message.Text):
		b.handleKeyboardButton(message, user)
	default:
		b.handleTextMessage(message, user)
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message, user *models.User) {
	switch message.Command() {
	case "start":
		b.handleStartCommand(message, user)
	case "help":
		b.handleHelpCommand(message)
	case "glucose":
		b.handleGlucoseCommand(message)
	case "food":
		b.handleFoodCommand(message)
	case "stats":
		b.handleStatsCommand(message, user)
	case "webapp":
		b.handleWebAppCommand(message, user)
	default:
		b.sendMessage(message.Chat.ID, "Неизвестная команда. Используйте /help для списка команд.")
	}
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message, user *models.User) {
	text := fmt.Sprintf(`Привет, %s! 👋

Я помогу вам контролировать уровень сахара в крови и питание.

Используйте кнопки ниже для быстрого доступа к функциям или просто отправьте мне:
• Число (уровень глюкозы, например: 5.6)
• Описание еды (что съели)
• Вопрос о диабете

Для подробной работы откройте веб-приложение! 📱`, user.FirstName)

	keyboard := b.getMainKeyboard()
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	text := `📋 Как пользоваться ботом:

🔘 Используйте кнопки ниже для быстрого доступа
🔘 Или просто напишите мне:
  • Число (например, 5.6) - записать уровень сахара
  • Описание еды - записать в дневник питания
  • Вопрос о диабете - получить рекомендацию от ИИ

📱 Веб-приложение:
Для подробной статистики, графиков и управления данными используйте веб-приложение - нажмите кнопку "📱 Веб-приложение"

🤖 ИИ помощник:
Бот анализирует ваши данные и дает персональные рекомендации на основе уровня сахара и питания.`

	keyboard := b.getMainKeyboard()
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleGlucoseCommand(message *tgbotapi.Message) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌅 До еды", "glucose_before"),
			tgbotapi.NewInlineKeyboardButtonData("🍽 После еды", "glucose_after"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌅 Утром натощак", "glucose_morning"),
			tgbotapi.NewInlineKeyboardButtonData("🌙 Перед сном", "glucose_night"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "🩸 Выберите время измерения или просто отправьте число (например: 5.6):")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleFoodCommand(message *tgbotapi.Message) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌅 Завтрак", "food_breakfast"),
			tgbotapi.NewInlineKeyboardButtonData("🌞 Обед", "food_lunch"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌙 Ужин", "food_dinner"),
			tgbotapi.NewInlineKeyboardButtonData("🍎 Перекус", "food_snack"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "🍽 Выберите тип приема пищи:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleStatsCommand(message *tgbotapi.Message, user *models.User) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 1 день", "stats_1"),
			tgbotapi.NewInlineKeyboardButtonData("📅 7 дней", "stats_7"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 30 дней", "stats_30"),
			tgbotapi.NewInlineKeyboardButtonData("📅 90 дней", "stats_90"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "📊 Выберите период для статистики:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleWebAppCommand(message *tgbotapi.Message, user *models.User) {
	if b.config.WebAppURL == "" {
		b.sendMessage(message.Chat.ID, "📱 Веб-приложение временно недоступно. Обратитесь к администратору для настройки.")
		return
	}

	webAppURL := fmt.Sprintf("%s/webapp", b.config.WebAppURL)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text: "📱 Открыть веб-приложение",
				URL:  &webAppURL,
			},
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, `📱 Веб-приложение DiabetBot

🔹 Подробная статистика и графики
🔹 Управление записями глюкозы и питания  
🔹 Персонализированные рекомендации
🔹 Экспорт данных

Нажмите кнопку ниже для открытия:`)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleGlucoseInput(message *tgbotapi.Message, user *models.User) {
	value, err := strconv.ParseFloat(message.Text, 64)
	if err != nil || value < 1.0 || value > 30.0 {
		b.sendMessage(message.Chat.ID, "Пожалуйста, введите корректное значение глюкозы (1.0-30.0 ммоль/л)")
		return
	}

	record, err := b.glucoseService.CreateRecord(user.ID, value, "")
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка сохранения данных")
		return
	}

	// Получаем рекомендации от ИИ
	recommendation := b.gigachatService.GetGlucoseRecommendation(user, record)
	
	response := fmt.Sprintf("✅ Записал: %.1f ммоль/л\n\n🤖 %s", value, recommendation)
	b.sendMessage(message.Chat.ID, response)
}

func (b *Bot) handleTextMessage(message *tgbotapi.Message, user *models.User) {
	// Обработка текстового сообщения как возможного описания еды или вопроса
	if len(message.Text) < 3 {
		b.sendMessage(message.Chat.ID, "Не понял вас. Используйте /help для получения помощи.")
		return
	}

	// Пытаемся определить, это еда или вопрос
	if isFoodDescription(message.Text) {
		b.handleFoodDescription(message, user)
	} else {
		b.handleQuestion(message, user)
	}
}

func (b *Bot) handleFoodDescription(message *tgbotapi.Message, user *models.User) {
	// Сохраняем запись о еде
	_, err := b.foodService.CreateRecord(user.ID, message.Text, "неопределено", nil, nil, "", "")
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка сохранения записи о питании")
		return
	}

	// Получаем рекомендации от ИИ
	recommendation := b.gigachatService.GetFoodRecommendation(user, message.Text)
	
	response := fmt.Sprintf("✅ Записал в дневник питания: %s\n\n🤖 %s", message.Text, recommendation)
	b.sendMessage(message.Chat.ID, response)
}

func (b *Bot) handleQuestion(message *tgbotapi.Message, user *models.User) {
	response := b.gigachatService.GetGeneralRecommendation(user, message.Text)
	b.sendMessage(message.Chat.ID, "🤖 "+response)
}

func (b *Bot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	// Обработка inline кнопок
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	b.api.Request(callback)

	data := callbackQuery.Data
	chatID := callbackQuery.Message.Chat.ID
	
	// Получаем пользователя для callback query
	user, err := b.userService.GetByTelegramID(callbackQuery.From.ID)
	if err != nil {
		log.Printf("Error getting user for callback: %v", err)
		return
	}

	switch {
	case len(data) >= 4 && data[:4] == "food":
		b.handleFoodTypeSelection(chatID, data[5:])
	case len(data) >= 7 && data[:7] == "glucose":
		b.handleGlucosePeriodSelection(chatID, data[8:], user)
	case len(data) >= 5 && data[:5] == "stats":
		b.handleStatsSelection(chatID, data[6:], user)
	}
}

func (b *Bot) handleFoodTypeSelection(chatID int64, foodType string) {
	var typeText string
	switch foodType {
	case "breakfast":
		typeText = "завтрак"
	case "lunch":
		typeText = "обед"
	case "dinner":
		typeText = "ужин"
	case "snack":
		typeText = "перекус"
	}

	text := fmt.Sprintf("🍽 Опишите ваш %s (например: овсянка с ягодами, 200г)", typeText)
	b.sendMessage(chatID, text)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isFoodDescription(text string) bool {
	foodKeywords := []string{"съел", "поел", "ел", "завтрак", "обед", "ужин", "перекус", 
		"каша", "хлеб", "мясо", "рыба", "овощи", "фрукты", "молоко", "кофе", "чай"}
	
	for _, keyword := range foodKeywords {
		if contains(text, keyword) {
			return true
		}
	}
	return false
}

func contains(text, substr string) bool {
	return len(text) >= len(substr) && 
		(text == substr || 
		 (len(text) > len(substr) && 
		  (text[:len(substr)] == substr || 
		   text[len(text)-len(substr):] == substr ||
		   containsMiddle(text, substr))))
}

func containsMiddle(text, substr string) bool {
	for i := 1; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// getMainKeyboard возвращает основную клавиатуру
func (b *Bot) getMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🩸 Записать глюкозу"),
			tgbotapi.NewKeyboardButton("🍽 Записать еду"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Статистика"),
			tgbotapi.NewKeyboardButton("📱 Веб-приложение"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Помощь"),
			tgbotapi.NewKeyboardButton("🔄 Главное меню"),
		),
	)
}

// isKeyboardButton проверяет, является ли текст кнопкой клавиатуры
func (b *Bot) isKeyboardButton(text string) bool {
	buttons := []string{
		"🩸 Записать глюкозу",
		"🍽 Записать еду", 
		"📊 Статистика",
		"📱 Веб-приложение",
		"❓ Помощь",
		"🔄 Главное меню",
	}
	
	for _, button := range buttons {
		if text == button {
			return true
		}
	}
	return false
}

// handleKeyboardButton обрабатывает нажатие кнопки клавиатуры
func (b *Bot) handleKeyboardButton(message *tgbotapi.Message, user *models.User) {
	switch message.Text {
	case "🩸 Записать глюкозу":
		b.handleGlucoseCommand(message)
	case "🍽 Записать еду":
		b.handleFoodCommand(message)
	case "📊 Статистика":
		b.handleStatsCommand(message, user)
	case "📱 Веб-приложение":
		b.handleWebAppCommand(message, user)
	case "❓ Помощь":
		b.handleHelpCommand(message)
	case "🔄 Главное меню":
		b.handleStartCommand(message, user)
	}
}

// handleGlucosePeriodSelection обрабатывает выбор периода измерения глюкозы
func (b *Bot) handleGlucosePeriodSelection(chatID int64, period string, user *models.User) {
	var periodText string
	switch period {
	case "before":
		periodText = "до еды"
	case "after":
		periodText = "после еды"
	case "morning":
		periodText = "утром натощак"
	case "night":
		periodText = "перед сном"
	default:
		periodText = "общее измерение"
	}

	text := fmt.Sprintf("🩸 Отправьте показания глюкометра (%s)\nНапример: 5.6", periodText)
	
	// Сохраняем контекст периода для следующего сообщения пользователя
	// В реальном приложении это можно сохранить в базе или кэше
	b.sendMessage(chatID, text)
}

// handleStatsSelection обрабатывает выбор периода статистики
func (b *Bot) handleStatsSelection(chatID int64, period string, user *models.User) {
	days, err := strconv.Atoi(period)
	if err != nil {
		b.sendMessage(chatID, "Ошибка обработки периода")
		return
	}

	stats, err := b.glucoseService.GetUserStats(user.ID, days)
	if err != nil {
		b.sendMessage(chatID, "Ошибка получения статистики")
		return
	}

	if stats.Count == 0 {
		text := fmt.Sprintf("📊 Статистика за %d дней:\n\n❌ Нет данных за выбранный период\n\nНачните записывать показания глюкозы!", days)
		b.sendMessage(chatID, text)
		return
	}

	var periodText string
	switch days {
	case 1:
		periodText = "сегодня"
	case 7:
		periodText = "за неделю"
	case 30:
		periodText = "за месяц"
	case 90:
		periodText = "за 3 месяца"
	default:
		periodText = fmt.Sprintf("за %d дней", days)
	}

	// Определяем статус по среднему уровню
	var statusEmoji, statusText string
	if stats.Average <= 5.5 {
		statusEmoji = "✅"
		statusText = "Отличный контроль!"
	} else if stats.Average <= 7.0 {
		statusEmoji = "⚠️"
		statusText = "Хороший контроль, но можно улучшить"
	} else {
		statusEmoji = "❗"
		statusText = "Требуется внимание"
	}

	text := fmt.Sprintf(`📊 Статистика %s:

📈 Средний уровень: %.1f ммоль/л %s
📉 Минимум: %.1f ммоль/л  
📊 Максимум: %.1f ммоль/л
🔢 Всего измерений: %d

%s %s

💡 Для подробных графиков и трендов используйте веб-приложение`, 
		periodText, stats.Average, statusEmoji, stats.Min, stats.Max, stats.Count,
		statusEmoji, statusText)

	b.sendMessage(chatID, text)
}