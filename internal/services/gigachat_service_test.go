package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"diabetbot/internal/config"
	"diabetbot/internal/models"
	"diabetbot/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGigaChatService_Authenticate(t *testing.T) {
	// Mock server for GigaChat API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/oauth" {
			if r.Header.Get("Authorization") == "Basic test_api_key" {
				response := AuthResponse{
					AccessToken: "test_access_token",
					ExpiresIn:   3600,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	}))
	defer server.Close()

	cfg := &config.GigaChatConfig{
		APIKey:  "test_api_key",
		BaseURL: server.URL,
	}
	
	service := NewGigaChatService(cfg)

	t.Run("SuccessfulAuthentication", func(t *testing.T) {
		err := service.authenticate()
		require.NoError(t, err)
		assert.NotEmpty(t, service.authToken)
		assert.True(t, service.tokenExp.After(time.Now()))
	})

	t.Run("TokenReuseWhenValid", func(t *testing.T) {
		// Устанавливаем валидный токен
		service.authToken = "existing_token"
		service.tokenExp = time.Now().Add(10 * time.Minute)

		oldToken := service.authToken
		err := service.authenticate()
		
		require.NoError(t, err)
		assert.Equal(t, oldToken, service.authToken) // токен не должен обновляться
	})
}

func TestGigaChatService_SendChatRequest(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth":
			response := AuthResponse{
				AccessToken: "test_token",
				ExpiresIn:   3600,
			}
			json.NewEncoder(w).Encode(response)
		case "/api/v2/chat/completions":
			if r.Header.Get("Authorization") != "Bearer test_token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			
			var req ChatRequest
			json.NewDecoder(r.Body).Decode(&req)
			
			response := ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Test response from GigaChat",
						},
						FinishReason: "stop",
					},
				},
				Usage: Usage{
					TotalTokens: 50,
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := &config.GigaChatConfig{
		APIKey:  "test_api_key",
		BaseURL: server.URL,
	}
	
	service := NewGigaChatService(cfg)

	t.Run("SuccessfulChatRequest", func(t *testing.T) {
		messages := []Message{
			{Role: "user", Content: "Test question"},
		}

		response, err := service.sendChatRequest(messages)
		
		require.NoError(t, err)
		assert.Equal(t, "Test response from GigaChat", response)
	})

	t.Run("EmptyMessages", func(t *testing.T) {
		response, err := service.sendChatRequest([]Message{})
		
		require.NoError(t, err)
		assert.NotEmpty(t, response)
	})
}

func TestGigaChatService_GetGlucoseRecommendation(t *testing.T) {
	// Mock server with realistic response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth":
			response := AuthResponse{
				AccessToken: "test_token",
				ExpiresIn:   3600,
			}
			json.NewEncoder(w).Encode(response)
		case "/api/v2/chat/completions":
			response := ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Ваш уровень глюкозы 6.5 ммоль/л находится в нормальном диапазоне. Продолжайте контролировать питание и физическую активность.",
						},
						FinishReason: "stop",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := &config.GigaChatConfig{
		APIKey:  "test_api_key",
		BaseURL: server.URL,
	}
	
	service := NewGigaChatService(cfg)

	t.Run("WithValidAPIKey", func(t *testing.T) {
		user := &models.User{
			TelegramID:    123,
			FirstName:     "Test",
			DiabetesType:  testutils.IntPtr(2),
			TargetGlucose: testutils.FloatPtr(6.0),
		}
		
		record := &models.GlucoseRecord{
			Value:      6.5,
			MeasuredAt: time.Now(),
		}

		recommendation := service.GetGlucoseRecommendation(user, record)
		
		assert.Contains(t, recommendation, "6.5 ммоль/л")
		assert.NotContains(t, recommendation, "временно недоступны")
	})

	t.Run("WithoutAPIKey", func(t *testing.T) {
		emptyCfg := &config.GigaChatConfig{
			APIKey:  "",
			BaseURL: server.URL,
		}
		emptyService := NewGigaChatService(emptyCfg)

		user := &models.User{
			TelegramID: 123,
			FirstName:  "Test",
		}
		
		record := &models.GlucoseRecord{
			Value:      6.5,
			MeasuredAt: time.Now(),
		}

		recommendation := emptyService.GetGlucoseRecommendation(user, record)
		
		assert.Contains(t, recommendation, "временно недоступны")
	})

	t.Run("WithUserWithoutDiabetesInfo", func(t *testing.T) {
		user := &models.User{
			TelegramID: 123,
			FirstName:  "Test",
			// DiabetesType и TargetGlucose не заданы
		}
		
		record := &models.GlucoseRecord{
			Value:      6.5,
			MeasuredAt: time.Now(),
		}

		recommendation := service.GetGlucoseRecommendation(user, record)
		
		assert.NotEmpty(t, recommendation)
		assert.Contains(t, recommendation, "6.5 ммоль/л")
	})
}

func TestGigaChatService_GetFoodRecommendation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth":
			response := AuthResponse{
				AccessToken: "test_token",
				ExpiresIn:   3600,
			}
			json.NewEncoder(w).Encode(response)
		case "/api/v2/chat/completions":
			response := ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Овсянка с ягодами - отличный выбор для завтрака. Содержит медленные углеводы, которые обеспечат стабильный уровень глюкозы.",
						},
						FinishReason: "stop",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := &config.GigaChatConfig{
		APIKey:  "test_api_key",
		BaseURL: server.URL,
	}
	
	service := NewGigaChatService(cfg)

	t.Run("ValidFoodDescription", func(t *testing.T) {
		user := &models.User{
			TelegramID:   123,
			FirstName:    "Test",
			DiabetesType: testutils.IntPtr(1),
		}
		
		foodDescription := "овсянка с ягодами"

		recommendation := service.GetFoodRecommendation(user, foodDescription)
		
		assert.Contains(t, strings.ToLower(recommendation), "овсянка")
		assert.NotContains(t, recommendation, "временно недоступны")
	})

	t.Run("WithoutAPIKey", func(t *testing.T) {
		emptyCfg := &config.GigaChatConfig{
			APIKey: "",
		}
		emptyService := NewGigaChatService(emptyCfg)

		user := &models.User{
			TelegramID: 123,
			FirstName:  "Test",
		}

		recommendation := emptyService.GetFoodRecommendation(user, "тест")
		
		assert.Contains(t, recommendation, "временно недоступны")
	})
}

func TestGigaChatService_GetGeneralRecommendation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth":
			response := AuthResponse{
				AccessToken: "test_token",
				ExpiresIn:   3600,
			}
			json.NewEncoder(w).Encode(response)
		case "/api/v2/chat/completions":
			response := ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role:    "assistant",
							Content: "Физические упражнения помогают улучшить чувствительность к инсулину. Рекомендую начать с ходьбы 30 минут в день.",
						},
						FinishReason: "stop",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	cfg := &config.GigaChatConfig{
		APIKey:  "test_api_key",
		BaseURL: server.URL,
	}
	
	service := NewGigaChatService(cfg)

	t.Run("ValidQuestion", func(t *testing.T) {
		user := &models.User{
			TelegramID:   123,
			FirstName:    "Test",
			DiabetesType: testutils.IntPtr(2),
		}
		
		question := "Какие упражнения полезны при диабете?"

		recommendation := service.GetGeneralRecommendation(user, question)
		
		assert.Contains(t, recommendation, "упражнения")
		assert.NotContains(t, recommendation, "временно недоступны")
	})

	t.Run("APIError", func(t *testing.T) {
		// Сервер, который всегда возвращает ошибку
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorServer.Close()

		errorCfg := &config.GigaChatConfig{
			APIKey:  "test_key",
			BaseURL: errorServer.URL,
		}
		errorService := NewGigaChatService(errorCfg)

		user := &models.User{
			TelegramID: 123,
			FirstName:  "Test",
		}

		recommendation := errorService.GetGeneralRecommendation(user, "тест")
		
		assert.Contains(t, recommendation, "Не удалось получить ответ")
		assert.Contains(t, recommendation, "врачу")
	})
}