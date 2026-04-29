package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginAttempt struct {
	UUID      string    `gorm:"primaryKey;type:char(36)"`
	UserUUID  string    `gorm:"index"`
	Success   bool      `gorm:"index;not null"`
	IP        string    `gorm:"index"`
	UserAgent string    `gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func RecordLoginAttempt(db *gorm.DB, userUUID string, success bool, ip string, ua string) error {
	if db == nil {
		return errors.New("db is nil")
	}

	attempt := LoginAttempt{
		UUID:      uuid.NewString(),
		UserUUID:  userUUID,
		UserAgent: ua,
		Success:   success,
		IP:        ip,
	}

	return db.Create(&attempt).Error
}
