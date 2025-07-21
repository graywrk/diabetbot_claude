package services

import (
	"diabetbot/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetOrCreateUser(telegramID int64, username, firstName, lastName, languageCode string) (*models.User, error) {
	var user models.User
	
	// Пытаемся найти существующего пользователя
	err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error
	
	if err == gorm.ErrRecordNotFound {
		// Создаем нового пользователя
		user = models.User{
			TelegramID:   telegramID,
			Username:     username,
			FirstName:    firstName,
			LastName:     lastName,
			LanguageCode: languageCode,
			IsActive:     true,
		}
		
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}
		
		return &user, nil
	} else if err != nil {
		return nil, err
	}
	
	// Обновляем информацию существующего пользователя
	updates := map[string]interface{}{
		"username":      username,
		"first_name":    firstName,
		"last_name":     lastName,
		"language_code": languageCode,
		"is_active":     true,
	}
	
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (s *UserService) GetByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateDiabetesInfo(userID uint, diabetesType int, targetGlucose float64) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"diabetes_type":   diabetesType,
		"target_glucose":  targetGlucose,
	}).Error
}