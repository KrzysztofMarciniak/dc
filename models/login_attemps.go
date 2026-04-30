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
	CreatedAt time.Time `gorm:"index;autoCreateTime"`
}

type LoginAttemptView struct {
	UUID            string
	UserUUID        string
	Username        string
	Success         bool
	IP              string
	UserAgent       string
	UserAgentPretty string
	CreatedAt       time.Time
	CreatedAtPretty string
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

func GetLoginAttemptsPaginated(db *gorm.DB, limit int, offset int) ([]LoginAttempt, int64, error) {
	var attempts []LoginAttempt
	var total int64

	if db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	if err := db.Model(&LoginAttempt{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&attempts).Error

	return attempts, total, err
}

func GetLoginAttemptsWithUser(db *gorm.DB, limit int, offset int) ([]LoginAttemptView, int64, error) {
	var attempts []LoginAttemptView
	var total int64

	if db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	if err := db.Model(&LoginAttempt{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Table("login_attempts").
		Select(`
			login_attempts.uuid,
			login_attempts.user_uuid,
			users.username,
			login_attempts.success,
			login_attempts.ip,
			login_attempts.user_agent,
			login_attempts.created_at
		`).
		Joins("LEFT JOIN users ON users.uuid = login_attempts.user_uuid").
		Order("login_attempts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&attempts).Error

	return attempts, total, err
}
