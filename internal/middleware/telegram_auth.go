package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TelegramAuthMiddleware проверяет авторизацию пользователя через Telegram WebApp
func TelegramAuthMiddleware(botToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем Telegram ID из URL параметра
		telegramIDStr := c.Param("telegram_id")
		if telegramIDStr == "" {
			telegramIDStr = c.Param("user_id")
		}
		
		if telegramIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing telegram_id or user_id"})
			c.Abort()
			return
		}

		telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram_id format"})
			c.Abort()
			return
		}

		// Получаем данные инициализации из заголовка
		initData := c.GetHeader("X-Telegram-Init-Data")
		if initData == "" {
			log.Printf("No X-Telegram-Init-Data header for user %d", telegramID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Приложение должно запускаться только через Telegram бот"})
			c.Abort()
			return
		}

		// Проверяем подпись Telegram
		if !verifyTelegramInitData(initData, botToken) {
			log.Printf("Invalid Telegram signature for user %d", telegramID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительная подпись Telegram"})
			c.Abort()
			return
		}

		// Извлекаем данные пользователя из initData
		userData, err := extractUserData(initData)
		if err != nil {
			log.Printf("Failed to extract user data from initData for user %d: %v", telegramID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Некорректные данные пользователя"})
			c.Abort()
			return
		}

		// Проверяем, что ID в URL соответствует ID в данных Telegram
		if userData.ID != telegramID {
			log.Printf("Telegram ID mismatch: URL=%d, initData=%d", telegramID, userData.ID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен: несоответствие ID пользователя"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("telegram_user", userData)
		c.Next()
	}
}

// TelegramUser представляет данные пользователя Telegram
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// verifyTelegramInitData проверяет подпись данных инициализации Telegram WebApp
func verifyTelegramInitData(initData, botToken string) bool {
	// Парсим данные
	values, err := url.ParseQuery(initData)
	if err != nil {
		log.Printf("Failed to parse initData: %v", err)
		return false
	}

	// Получаем hash из данных
	hash := values.Get("hash")
	if hash == "" {
		log.Println("No hash in initData")
		return false
	}

	// Удаляем hash из данных для проверки
	values.Del("hash")

	// Создаем отсортированную строку данных
	var pairs []string
	for key, vals := range values {
		for _, val := range vals {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, val))
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Создаем секретный ключ
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	// Вычисляем HMAC-SHA256
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Сравниваем хеши
	return hmac.Equal([]byte(hash), []byte(expectedHash))
}

// extractUserData извлекает данные пользователя из initData
func extractUserData(initData string) (*TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %v", err)
	}

	// Получаем данные пользователя
	userStr := values.Get("user")
	if userStr == "" {
		return nil, fmt.Errorf("no user data in initData")
	}

	// Декодируем JSON данные пользователя (они в URL-encoded формате)
	userDecoded, err := url.QueryUnescape(userStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user data: %v", err)
	}

	// Простой парсинг JSON (можно заменить на json.Unmarshal)
	userData := &TelegramUser{}
	
	// Извлекаем ID
	if idStr := extractJSONField(userDecoded, "id"); idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			userData.ID = id
		}
	}
	
	// Извлекаем остальные поля
	userData.FirstName = extractJSONField(userDecoded, "first_name")
	userData.LastName = extractJSONField(userDecoded, "last_name")
	userData.Username = extractJSONField(userDecoded, "username")
	userData.LanguageCode = extractJSONField(userDecoded, "language_code")

	if userData.ID == 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Проверяем временную метку (не старше 1 дня)
	if authDateStr := values.Get("auth_date"); authDateStr != "" {
		if authDate, err := strconv.ParseInt(authDateStr, 10, 64); err == nil {
			if time.Now().Unix()-authDate > 86400 { // 24 часа
				return nil, fmt.Errorf("initData too old")
			}
		}
	}

	return userData, nil
}

// extractJSONField извлекает значение поля из JSON строки (простая реализация)
func extractJSONField(jsonStr, field string) string {
	if strings.Contains(jsonStr, fmt.Sprintf(`"%s":`, field)) {
		start := strings.Index(jsonStr, fmt.Sprintf(`"%s":`, field))
		if start == -1 {
			return ""
		}
		start += len(field) + 3 // длина поля + ":"
		
		// Пропускаем пробелы
		for start < len(jsonStr) && jsonStr[start] == ' ' {
			start++
		}
		
		if start >= len(jsonStr) || jsonStr[start] != '"' {
			// Возможно, это число
			end := start
			for end < len(jsonStr) && (jsonStr[end] >= '0' && jsonStr[end] <= '9') {
				end++
			}
			if end > start {
				return jsonStr[start:end]
			}
			return ""
		}
		
		start++ // пропускаем открывающую кавычку
		end := start
		for end < len(jsonStr) && jsonStr[end] != '"' {
			end++
		}
		
		if end < len(jsonStr) {
			return jsonStr[start:end]
		}
	}
	return ""
}