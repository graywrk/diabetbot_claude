package testutils

import (
	"fmt"
	"testing"
	"time"

	"diabetbot/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB создает временную SQLite базу данных для тестов
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Автомиграция моделей
	err = db.AutoMigrate(
		&models.User{},
		&models.GlucoseRecord{},
		&models.FoodRecord{},
		&models.AIRecommendation{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB очищает тестовую базу данных
func CleanupTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// CreateTestUser создает тестового пользователя
func CreateTestUser(db *gorm.DB, telegramID int64) *models.User {
	user := &models.User{
		TelegramID:   telegramID,
		FirstName:    fmt.Sprintf("Test User %d", telegramID),
		IsActive:     true,
		DiabetesType: intPtr(1),
		TargetGlucose: floatPtr(6.0),
	}
	db.Create(user)
	return user
}

// CreateTestGlucoseRecord создает тестовую запись глюкозы
func CreateTestGlucoseRecord(db *gorm.DB, userID uint, value float64) *models.GlucoseRecord {
	record := &models.GlucoseRecord{
		UserID:     userID,
		Value:      value,
		MeasuredAt: time.Now(),
	}
	db.Create(record)
	return record
}

// CreateTestFoodRecord создает тестовую запись питания
func CreateTestFoodRecord(db *gorm.DB, userID uint, foodName, foodType string) *models.FoodRecord {
	record := &models.FoodRecord{
		UserID:     userID,
		FoodName:   foodName,
		FoodType:   foodType,
		ConsumedAt: time.Now(),
	}
	db.Create(record)
	return record
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

func FloatPtr(f float64) *float64 {
	return &f
}

// TestDB wrapper for compatibility
type TestDB struct {
	*gorm.DB
}