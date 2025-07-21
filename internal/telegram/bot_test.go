package telegram

import (
	"testing"

	"diabetbot/internal/config"
	"diabetbot/internal/models"
	"diabetbot/internal/services"
	"diabetbot/internal/testutils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock BotAPI для тестирования
type MockBotAPI struct {
	sentMessages []tgbotapi.Chattable
	self         tgbotapi.User
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	return tgbotapi.Message{MessageID: len(m.sentMessages)}, nil
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return make(tgbotapi.UpdatesChannel)
}

func (m *MockBotAPI) GetLastSentMessage() tgbotapi.Chattable {
	if len(m.sentMessages) == 0 {
		return nil
	}
	return m.sentMessages[len(m.sentMessages)-1]
}

func (m *MockBotAPI) GetAllSentMessages() []tgbotapi.Chattable {
	return m.sentMessages
}

func (m *MockBotAPI) ClearMessages() {
	m.sentMessages = []tgbotapi.Chattable{}
}

func createTestBot() (*Bot, *MockBotAPI, *testutils.TestDB) {
	// Создаем тестовую БД
	db := testutils.SetupTestDB(&testing.T{})
	
	// Создаем mock GigaChat service
	cfg := &config.GigaChatConfig{APIKey: ""}
	gigachatService := services.NewGigaChatService(cfg)
	
	// Создаем бот с моками
	mockAPI := &MockBotAPI{
		self: tgbotapi.User{UserName: "testbot"},
	}
	
	bot := &Bot{
		api:             mockAPI,
		userService:     services.NewUserService(db),
		glucoseService:  services.NewGlucoseService(db),
		foodService:     services.NewFoodService(db),
		gigachatService: gigachatService,
	}
	
	return bot, mockAPI, &testutils.TestDB{DB: db}
}

func TestBot_HandleStartCommand(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	// Создаем тестового пользователя
	user := testutils.CreateTestUser(testDB.DB, 123456789)
	
	// Создаем сообщение /start
	message := &tgbotapi.Message{
		MessageID: 1,
		From: &tgbotapi.User{
			ID:           123456789,
			FirstName:    "Test",
			LastName:     "User",
			UserName:     "testuser",
			LanguageCode: "ru",
		},
		Chat: &tgbotapi.Chat{
			ID: 123456789,
		},
		Text: "/start",
	}

	bot.handleStartCommand(message, user)

	// Проверяем отправку сообщения
	require.Len(t, mockAPI.GetAllSentMessages(), 1)
	
	sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
	require.True(t, ok)
	
	assert.Contains(t, sentMsg.Text, "Привет, Test!")
	assert.Contains(t, sentMsg.Text, "/glucose")
	assert.Contains(t, sentMsg.Text, "/food")
	assert.Contains(t, sentMsg.Text, "/stats")
	assert.Contains(t, sentMsg.Text, "/webapp")
	assert.Contains(t, sentMsg.Text, "/help")
}

func TestBot_HandleHelpCommand(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat: &tgbotapi.Chat{ID: 123456789},
		Text: "/help",
	}

	bot.handleHelpCommand(message)

	require.Len(t, mockAPI.GetAllSentMessages(), 1)
	
	sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
	require.True(t, ok)
	
	assert.Contains(t, sentMsg.Text, "Список команд")
	assert.Contains(t, sentMsg.Text, "/glucose")
	assert.Contains(t, sentMsg.Text, "/food")
	assert.Contains(t, sentMsg.Text, "/stats")
	assert.Contains(t, sentMsg.Text, "/webapp")
}

func TestBot_HandleGlucoseInput(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	user := testutils.CreateTestUser(testDB.DB, 123456789)
	
	t.Run("ValidGlucoseValue", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 1,
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "6.5",
		}

		bot.handleGlucoseInput(message, user)

		// Проверяем, что запись создана
		records, err := bot.glucoseService.GetUserRecords(user.ID, 1)
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, 6.5, records[0].Value)

		// Проверяем отправку сообщения
		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		
		assert.Contains(t, sentMsg.Text, "Записал: 6.5 ммоль/л")
		assert.Contains(t, sentMsg.Text, "🤖") // должна быть рекомендация от ИИ
	})

	t.Run("InvalidGlucoseValue", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 2,
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "50.0", // слишком высокое значение
		}

		bot.handleGlucoseInput(message, user)

		// Проверяем отправку сообщения об ошибке
		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		
		assert.Contains(t, sentMsg.Text, "корректное значение")
		assert.Contains(t, sentMsg.Text, "1.0-30.0")
	})

	t.Run("NonNumericValue", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 3,
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "abc",
		}

		bot.handleGlucoseInput(message, user)

		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		
		assert.Contains(t, sentMsg.Text, "корректное значение")
	})
}

func TestBot_HandleFoodDescription(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	user := testutils.CreateTestUser(testDB.DB, 123456789)
	
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat: &tgbotapi.Chat{ID: 123456789},
		Text: "съел овсянку с ягодами",
	}

	bot.handleFoodDescription(message, user)

	// Проверяем создание записи о еде
	records, err := bot.foodService.GetUserRecords(user.ID, 1)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "съел овсянку с ягодами", records[0].FoodName)
	assert.Equal(t, "неопределено", records[0].FoodType)

	// Проверяем отправку сообщения
	require.Len(t, mockAPI.GetAllSentMessages(), 1)
	sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
	require.True(t, ok)
	
	assert.Contains(t, sentMsg.Text, "Записал в дневник питания")
	assert.Contains(t, sentMsg.Text, "съел овсянку с ягодами")
	assert.Contains(t, sentMsg.Text, "🤖")
}

func TestBot_HandleStatsCommand(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	user := testutils.CreateTestUser(testDB.DB, 123456789)
	
	// Создаем тестовые записи
	testutils.CreateTestGlucoseRecord(testDB.DB, user.ID, 5.5)
	testutils.CreateTestGlucoseRecord(testDB.DB, user.ID, 6.0)
	testutils.CreateTestGlucoseRecord(testDB.DB, user.ID, 6.5)
	testutils.CreateTestGlucoseRecord(testDB.DB, user.ID, 7.0)
	
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat: &tgbotapi.Chat{ID: 123456789},
		Text: "/stats",
	}

	bot.handleStatsCommand(message, user)

	require.Len(t, mockAPI.GetAllSentMessages(), 1)
	sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
	require.True(t, ok)
	
	assert.Contains(t, sentMsg.Text, "Статистика за 7 дней")
	assert.Contains(t, sentMsg.Text, "Средний уровень: 6.2") // (5.5+6.0+6.5+7.0)/4
	assert.Contains(t, sentMsg.Text, "Минимум: 5.5")
	assert.Contains(t, sentMsg.Text, "Максимум: 7.0")
	assert.Contains(t, sentMsg.Text, "Всего измерений: 4")
	assert.Contains(t, sentMsg.Text, "/webapp")
}

func TestBot_HandleFoodCommand(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat: &tgbotapi.Chat{ID: 123456789},
		Text: "/food",
	}

	bot.handleFoodCommand(message)

	require.Len(t, mockAPI.GetAllSentMessages(), 1)
	
	// Проверяем, что отправлено сообщение с inline клавиатурой
	sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
	require.True(t, ok)
	
	assert.Contains(t, sentMsg.Text, "Выберите тип приема пищи")
	assert.NotNil(t, sentMsg.ReplyMarkup)
}

func TestBot_HandleWebAppCommand(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	user := testutils.CreateTestUser(testDB.DB, 123456789)
	
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat: &tgbotapi.Chat{ID: 123456789},
		Text: "/webapp",
	}

	bot.handleWebAppCommand(message, user)

	require.Len(t, mockAPI.GetAllSentMessages(), 1)
	
	sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
	require.True(t, ok)
	
	assert.Contains(t, sentMsg.Text, "веб-приложение")
	assert.NotNil(t, sentMsg.ReplyMarkup)
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"6.5", true},
		{"5", true},
		{"10.25", true},
		{"0.5", true},
		{"abc", false},
		{"6.5abc", false},
		{"", false},
		{".", false},
		{"-5.0", false}, // отрицательные не допускаются
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := isNumeric(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsFoodDescription(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"съел овсянку", true},
		{"поел хлеб", true},
		{"завтрак был вкусный", true},
		{"выпил кофе", true},
		{"ел мясо с овощами", true},
		{"как дела?", false},
		{"погода хорошая", false},
		{"123", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := isFoodDescription(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestBot_HandleMessage_Classification(t *testing.T) {
	bot, mockAPI, testDB := createTestBot()
	defer testutils.CleanupTestDB(testDB.DB)
	
	user := testutils.CreateTestUser(testDB.DB, 123456789)
	
	t.Run("CommandMessage", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 1,
			From: &tgbotapi.User{ID: 123456789, FirstName: "Test"},
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "/help",
		}

		bot.handleMessage(message)

		// Должно обработаться как команда
		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Contains(t, sentMsg.Text, "Список команд")
	})

	t.Run("NumericMessage", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 2,
			From: &tgbotapi.User{ID: 123456789, FirstName: "Test"},
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "6.5",
		}

		bot.handleMessage(message)

		// Должно обработаться как глюкоза
		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Contains(t, sentMsg.Text, "Записал: 6.5 ммоль/л")
	})

	t.Run("FoodMessage", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 3,
			From: &tgbotapi.User{ID: 123456789, FirstName: "Test"},
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "съел яблоко",
		}

		bot.handleMessage(message)

		// Должно обработаться как еда
		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Contains(t, sentMsg.Text, "Записал в дневник питания")
	})

	t.Run("QuestionMessage", func(t *testing.T) {
		mockAPI.ClearMessages()
		
		message := &tgbotapi.Message{
			MessageID: 4,
			From: &tgbotapi.User{ID: 123456789, FirstName: "Test"},
			Chat: &tgbotapi.Chat{ID: 123456789},
			Text: "Какие упражнения полезны при диабете?",
		}

		bot.handleMessage(message)

		// Должно обработаться как вопрос для ИИ
		require.Len(t, mockAPI.GetAllSentMessages(), 1)
		sentMsg, ok := mockAPI.GetLastSentMessage().(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Contains(t, sentMsg.Text, "🤖")
	})
}