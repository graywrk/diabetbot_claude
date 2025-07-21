package services

import (
	"diabetbot/internal/models"
	"time"

	"gorm.io/gorm"
)

type FoodService struct {
	db *gorm.DB
}

func NewFoodService(db *gorm.DB) *FoodService {
	return &FoodService{db: db}
}

func (s *FoodService) CreateRecord(userID uint, foodName, foodType string, carbs *float64, calories *int, quantity, notes string) (*models.FoodRecord, error) {
	record := models.FoodRecord{
		UserID:     userID,
		FoodName:   foodName,
		FoodType:   foodType,
		Carbs:      carbs,
		Calories:   calories,
		Quantity:   quantity,
		ConsumedAt: time.Now(),
		Notes:      notes,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func (s *FoodService) GetUserRecords(userID uint, days int) ([]models.FoodRecord, error) {
	var records []models.FoodRecord
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Where("user_id = ? AND consumed_at >= ?", userID, startDate).
		Order("consumed_at DESC").
		Find(&records).Error
	
	return records, err
}

func (s *FoodService) GetRecordsByType(userID uint, foodType string, days int) ([]models.FoodRecord, error) {
	var records []models.FoodRecord
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Where("user_id = ? AND food_type = ? AND consumed_at >= ?", userID, foodType, startDate).
		Order("consumed_at DESC").
		Find(&records).Error
	
	return records, err
}

func (s *FoodService) DeleteRecord(userID, recordID uint) error {
	return s.db.Where("user_id = ? AND id = ?", userID, recordID).
		Delete(&models.FoodRecord{}).Error
}

func (s *FoodService) UpdateRecord(userID, recordID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.FoodRecord{}).
		Where("user_id = ? AND id = ?", userID, recordID).
		Updates(updates).Error
}

func (s *FoodService) GetTodayCalories(userID uint) (int, error) {
	var totalCalories int
	
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	
	err := s.db.Model(&models.FoodRecord{}).
		Where("user_id = ? AND consumed_at >= ? AND consumed_at < ?", userID, today, tomorrow).
		Select("COALESCE(SUM(calories), 0)").
		Scan(&totalCalories).Error
	
	return totalCalories, err
}

func (s *FoodService) GetTodayCarbs(userID uint) (float64, error) {
	var totalCarbs float64
	
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	
	err := s.db.Model(&models.FoodRecord{}).
		Where("user_id = ? AND consumed_at >= ? AND consumed_at < ?", userID, today, tomorrow).
		Select("COALESCE(SUM(carbs), 0)").
		Scan(&totalCarbs).Error
	
	return totalCarbs, err
}