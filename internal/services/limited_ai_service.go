package services

import (
	"diabetbot/internal/models"
	"fmt"
	"gorm.io/gorm"
)

// LimitedAIService оборачивает AI сервис и добавляет проверку лимитов
type LimitedAIService struct {
	aiService      AIService
	aiUsageService *AIUsageService
}

func NewLimitedAIService(aiService AIService, db *gorm.DB) *LimitedAIService {
	return &LimitedAIService{
		aiService:      aiService,
		aiUsageService: NewAIUsageService(db),
	}
}

func (s *LimitedAIService) GetGlucoseRecommendation(user *models.User, record *models.GlucoseRecord) string {
	// Проверяем лимит
	allowed, remaining, err := s.aiUsageService.CheckAndIncrementUsage(user.ID)
	if err != nil {
		return "Ошибка проверки лимита запросов. Обратитесь к врачу для консультации."
	}
	
	if !allowed {
		return fmt.Sprintf("🚫 Достигнут дневной лимит AI запросов (%d в день). Лимит обновится завтра. Обратитесь к врачу для консультации.", DailyAIRequestLimit)
	}
	
	// Получаем рекомендацию от AI
	recommendation := s.aiService.GetGlucoseRecommendation(user, record)
	
	// Добавляем информацию об оставшихся запросах
	if remaining > 0 {
		recommendation += fmt.Sprintf("\n\n📊 Осталось AI запросов на сегодня: %d", remaining)
	} else {
		recommendation += fmt.Sprintf("\n\n⚠️ Это был последний AI запрос на сегодня")
	}
	
	return recommendation
}

func (s *LimitedAIService) GetFoodRecommendation(user *models.User, foodDescription string) string {
	// Проверяем лимит
	allowed, remaining, err := s.aiUsageService.CheckAndIncrementUsage(user.ID)
	if err != nil {
		return "Ошибка проверки лимита запросов. Следите за углеводами в рационе."
	}
	
	if !allowed {
		return fmt.Sprintf("🚫 Достигнут дневной лимит AI запросов (%d в день). Лимит обновится завтра. Контролируйте количество углеводов в рационе.", DailyAIRequestLimit)
	}
	
	// Получаем рекомендацию от AI
	recommendation := s.aiService.GetFoodRecommendation(user, foodDescription)
	
	// Добавляем информацию об оставшихся запросах
	if remaining > 0 {
		recommendation += fmt.Sprintf("\n\n📊 Осталось AI запросов на сегодня: %d", remaining)
	} else {
		recommendation += fmt.Sprintf("\n\n⚠️ Это был последний AI запрос на сегодня")
	}
	
	return recommendation
}

func (s *LimitedAIService) GetGeneralRecommendation(user *models.User, question string) string {
	// Проверяем лимит
	allowed, remaining, err := s.aiUsageService.CheckAndIncrementUsage(user.ID)
	if err != nil {
		return "Ошибка проверки лимита запросов. Обратитесь к лечащему врачу за консультацией."
	}
	
	if !allowed {
		return fmt.Sprintf("🚫 Достигнут дневной лимит AI запросов (%d в день). Лимит обновится завтра. Рекомендую обратиться к лечащему врачу за консультацией.", DailyAIRequestLimit)
	}
	
	// Получаем рекомендацию от AI
	recommendation := s.aiService.GetGeneralRecommendation(user, question)
	
	// Добавляем информацию об оставшихся запросах
	if remaining > 0 {
		recommendation += fmt.Sprintf("\n\n📊 Осталось AI запросов на сегодня: %d", remaining)
	} else {
		recommendation += fmt.Sprintf("\n\n⚠️ Это был последний AI запрос на сегодня")
	}
	
	return recommendation
}

// Убеждаемся, что LimitedAIService реализует интерфейс AIService
var _ AIService = (*LimitedAIService)(nil)