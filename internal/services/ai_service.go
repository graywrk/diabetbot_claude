package services

import "diabetbot/internal/models"

// AIService интерфейс для различных AI провайдеров
type AIService interface {
	GetGlucoseRecommendation(user *models.User, record *models.GlucoseRecord) string
	GetFoodRecommendation(user *models.User, foodDescription string) string
	GetGeneralRecommendation(user *models.User, question string) string
}

// Убеждаемся, что оба сервиса реализуют интерфейс
var _ AIService = (*GigaChatService)(nil)
var _ AIService = (*YandexGPTService)(nil)