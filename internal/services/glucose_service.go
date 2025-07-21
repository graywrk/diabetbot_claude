package services

import (
	"diabetbot/internal/models"
	"time"

	"gorm.io/gorm"
)

type GlucoseService struct {
	db *gorm.DB
}

type GlucoseStats struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Count   int64   `json:"count"`
}

func NewGlucoseService(db *gorm.DB) *GlucoseService {
	return &GlucoseService{db: db}
}

func (s *GlucoseService) CreateRecord(userID uint, value float64, notes string) (*models.GlucoseRecord, error) {
	record := models.GlucoseRecord{
		UserID:     userID,
		Value:      value,
		MeasuredAt: time.Now(),
		Notes:      notes,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func (s *GlucoseService) GetUserRecords(userID uint, days int) ([]models.GlucoseRecord, error) {
	var records []models.GlucoseRecord
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Where("user_id = ? AND measured_at >= ?", userID, startDate).
		Order("measured_at DESC").
		Find(&records).Error
	
	return records, err
}

func (s *GlucoseService) GetUserStats(userID uint, days int) (*GlucoseStats, error) {
	var stats GlucoseStats
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.GlucoseRecord{}).
		Where("user_id = ? AND measured_at >= ?", userID, startDate).
		Select("AVG(value) as average, MIN(value) as min, MAX(value) as max, COUNT(*) as count").
		Scan(&stats).Error
	
	return &stats, err
}

func (s *GlucoseService) GetRecentRecord(userID uint) (*models.GlucoseRecord, error) {
	var record models.GlucoseRecord
	
	err := s.db.Where("user_id = ?", userID).
		Order("measured_at DESC").
		First(&record).Error
	
	if err != nil {
		return nil, err
	}
	
	return &record, nil
}

func (s *GlucoseService) DeleteRecord(userID, recordID uint) error {
	return s.db.Where("user_id = ? AND id = ?", userID, recordID).
		Delete(&models.GlucoseRecord{}).Error
}

func (s *GlucoseService) UpdateRecord(userID, recordID uint, value float64, notes string) error {
	return s.db.Model(&models.GlucoseRecord{}).
		Where("user_id = ? AND id = ?", userID, recordID).
		Updates(map[string]interface{}{
			"value": value,
			"notes": notes,
		}).Error
}

func (s *GlucoseService) DeleteAllUserRecords(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.GlucoseRecord{}).Error
}