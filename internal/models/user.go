package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `json:"id" gorm:"primarykey"`
	TelegramID     int64          `json:"telegram_id" gorm:"uniqueIndex;not null"`
	Username       string         `json:"username" gorm:"size:255"`
	FirstName      string         `json:"first_name" gorm:"size:255"`
	LastName       string         `json:"last_name" gorm:"size:255"`
	LanguageCode   string         `json:"language_code" gorm:"size:10;default:'ru'"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Medical information
	DiabetesType   *int           `json:"diabetes_type" gorm:"check:diabetes_type IN (1,2)"`
	TargetGlucose  *float64       `json:"target_glucose"` // mmol/L
	
	// Relations
	GlucoseRecords []GlucoseRecord `json:"glucose_records" gorm:"foreignKey:UserID"`
	FoodRecords    []FoodRecord    `json:"food_records" gorm:"foreignKey:UserID"`
}

type GlucoseRecord struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Value     float64        `json:"value" gorm:"not null"` // mmol/L
	MeasuredAt time.Time     `json:"measured_at" gorm:"not null"`
	Notes     string         `json:"notes" gorm:"size:500"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type FoodRecord struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	FoodName    string         `json:"food_name" gorm:"size:255;not null"`
	FoodType    string         `json:"food_type" gorm:"size:100"` // завтрак, обед, ужин, перекус
	Carbs       *float64       `json:"carbs"`                     // граммы углеводов
	Calories    *int           `json:"calories"`
	Quantity    string         `json:"quantity" gorm:"size:100"`  // порция, описание количества
	ConsumedAt  time.Time      `json:"consumed_at" gorm:"not null"`
	Notes       string         `json:"notes" gorm:"size:500"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type AIRecommendation struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Type      string         `json:"type" gorm:"size:50;not null"` // glucose, food, general
	Content   string         `json:"content" gorm:"type:text;not null"`
	Context   string         `json:"context" gorm:"type:json"`     // JSON с контекстными данными
	CreatedAt time.Time      `json:"created_at"`
	
	User User `json:"user" gorm:"foreignKey:UserID"`
}