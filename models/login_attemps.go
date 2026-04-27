package models

import (
	"time"

	"gorm.io/gorm"
)

type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Success   bool      `gorm:"not null"`
	IP        string    `gorm:"index"`
	UserAgent string    `gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func RecordLoginAttempt(db *gorm.DB, userID uint, success bool, ip string, ua string) error {
	attempt := LoginAttempt{
		UserID:    userID,
		UserAgent: ua,
		Success:   success,
		IP:        ip,
	}

	return db.Create(&attempt).Error
}
