package models

import (
	"time"
	"gorm.io/gorm"
)

// AIUsage отслеживает использование AI запросов пользователями
type AIUsage struct {
	gorm.Model
	UserID      uint      `gorm:"not null;index:idx_user_date,unique:true"`
	Date        time.Time `gorm:"type:date;not null;index:idx_user_date,unique:true"`
	RequestCount int      `gorm:"default:0"`
	User        User      `gorm:"foreignKey:UserID"`
}