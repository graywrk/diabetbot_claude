package services

import (
	"diabetbot/internal/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

const (
	DailyAIRequestLimit = 10 // Лимит AI запросов на пользователя в день
)

type AIUsageService struct {
	db *gorm.DB
}

func NewAIUsageService(db *gorm.DB) *AIUsageService {
	return &AIUsageService{db: db}
}

// CheckAndIncrementUsage проверяет лимит и увеличивает счетчик использования
func (s *AIUsageService) CheckAndIncrementUsage(userID uint) (bool, int, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	
	var usage models.AIUsage
	err := s.db.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error
	
	if err == gorm.ErrRecordNotFound {
		// Создаем новую запись для сегодня
		usage = models.AIUsage{
			UserID:       userID,
			Date:         today,
			RequestCount: 1,
		}
		if err := s.db.Create(&usage).Error; err != nil {
			return false, 0, fmt.Errorf("failed to create AI usage record: %w", err)
		}
		log.Printf("AI usage: user %d made 1/%d requests today", userID, DailyAIRequestLimit)
		return true, DailyAIRequestLimit - 1, nil
	}
	
	if err != nil {
		return false, 0, fmt.Errorf("failed to get AI usage: %w", err)
	}
	
	// Проверяем лимит
	if usage.RequestCount >= DailyAIRequestLimit {
		log.Printf("AI usage limit reached: user %d has %d/%d requests today", userID, usage.RequestCount, DailyAIRequestLimit)
		return false, 0, nil
	}
	
	// Увеличиваем счетчик
	usage.RequestCount++
	if err := s.db.Save(&usage).Error; err != nil {
		return false, 0, fmt.Errorf("failed to update AI usage: %w", err)
	}
	
	remainingRequests := DailyAIRequestLimit - usage.RequestCount
	log.Printf("AI usage: user %d made %d/%d requests today, %d remaining", userID, usage.RequestCount, DailyAIRequestLimit, remainingRequests)
	
	return true, remainingRequests, nil
}

// GetUsageToday возвращает количество использованных запросов за сегодня
func (s *AIUsageService) GetUsageToday(userID uint) (int, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	
	var usage models.AIUsage
	err := s.db.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error
	
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	
	if err != nil {
		return 0, fmt.Errorf("failed to get AI usage: %w", err)
	}
	
	return usage.RequestCount, nil
}

// GetRemainingRequests возвращает количество оставшихся запросов на сегодня
func (s *AIUsageService) GetRemainingRequests(userID uint) (int, error) {
	used, err := s.GetUsageToday(userID)
	if err != nil {
		return 0, err
	}
	
	remaining := DailyAIRequestLimit - used
	if remaining < 0 {
		remaining = 0
	}
	
	return remaining, nil
}

// ResetDailyUsage сбрасывает счетчики для всех пользователей (для cron job)
func (s *AIUsageService) ResetDailyUsage() error {
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	
	// Удаляем записи старше вчерашнего дня
	result := s.db.Where("date < ?", yesterday).Delete(&models.AIUsage{})
	if result.Error != nil {
		return fmt.Errorf("failed to clean old AI usage records: %w", result.Error)
	}
	
	if result.RowsAffected > 0 {
		log.Printf("Cleaned %d old AI usage records", result.RowsAffected)
	}
	
	return nil
}