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

Основные команды:
🩸 /glucose - записать уровень сахара
🍽 /food - записать прием пищи
📊 /stats - показать статистику
📱 /webapp - открыть веб-приложение
❓ /help - помощь

Вы также можете просто отправить мне число (уровень глюкозы) или описание еды.`, user.FirstName)

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	text := `📋 Список команд:

🩸 /glucose - Записать показания глюкометра
🍽 /food - Записать информацию о приеме пищи
📊 /stats - Показать статистику за период
📱 /webapp - Открыть веб-приложение с подробной статистикой

💡 Быстрые действия:
• Отправьте число (например, 5.6) чтобы записать уровень сахара
• Опишите что ели и я помогу записать это в дневник
• Задайте вопрос о диабете - получите рекомендации от ИИ

Для детальной работы с данными используйте веб-приложение /webapp`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleGlucoseCommand(message *tgbotapi.Message) {
	text := "🩸 Отправьте показания глюкометра (например: 5.6)"
	b.sendMessage(message.Chat.ID, text)
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
	stats, err := b.glucoseService.GetUserStats(user.ID, 7) // последние 7 дней
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка получения статистики")
		return
	}

	text := fmt.Sprintf(`📊 Статистика за 7 дней:

📈 Средний уровень: %.1f ммоль/л
📉 Минимум: %.1f ммоль/л  
📊 Максимум: %.1f ммоль/л
🔢 Всего измерений: %d

Для подробной статистики используйте /webapp`, 
		stats.Average, stats.Min, stats.Max, stats.Count)

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleWebAppCommand(message *tgbotapi.Message, user *models.User) {
	webAppURL := fmt.Sprintf("https://yourdomain.com/webapp?user_id=%d", user.TelegramID)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text: "📱 Открыть приложение",
				URL:  &webAppURL,
			},
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "📱 Нажмите на кнопку ниже, чтобы открыть веб-приложение:")
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

	switch {
	case data[:4] == "food":
		b.handleFoodTypeSelection(chatID, data[5:])
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