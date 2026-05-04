package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginAttempt struct {
	UUID      string    `gorm:"primaryKey;type:char(36);not null;uniqueIndex"`
	UserUUID  string    `gorm:"index"`
	Success   bool      `gorm:"index;not null"`
	UserAgent string    `gorm:"index"`
	IP        string    `gorm:"index:idx_ip_created,priority:1"`
	CreatedAt time.Time `gorm:"index:idx_ip_created,priority:2;autoCreateTime"`
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

type LoginAttemptChartPoint struct {
	Date    string `json:"date"`
	IP      string `json:"ip"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
}

type LoginAttemptChartData struct {
	Month  string                   `json:"month"`
	Points []LoginAttemptChartPoint `json:"points"`
}

type loginAttemptChartRow struct {
	Date    string
	IP      string
	Success int64
	Failed  int64
}

func GetLoginAttemptChartData(db *gorm.DB, month string) (*LoginAttemptChartData, error) {
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	if month == "" {
		month = time.Now().Format("2006-01")
	}

	start, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, errors.New("invalid month format, expected YYYY-MM")
	}
	end := start.AddDate(0, 1, 0)

	var rows []loginAttemptChartRow

	err = db.Table("login_attempts").
		Select(`
			strftime('%Y-%m-%d', created_at) AS date,
			ip,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) AS success,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) AS failed
		`).
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("date, ip").
		Order("date ASC, ip ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	points := make([]LoginAttemptChartPoint, 0, len(rows))
	for _, r := range rows {
		points = append(points, LoginAttemptChartPoint{
			Date:    r.Date,
			IP:      r.IP,
			Success: r.Success,
			Failed:  r.Failed,
		})
	}

	return &LoginAttemptChartData{
		Month:  month,
		Points: points,
	}, nil
}
func IsIPBlocked(db *gorm.DB, ip string) (bool, error) {
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	err := db.Model(&LoginAttempt{}).
		Where("ip = ? AND success = ? AND created_at > ?", ip, false, oneHourAgo).
		Count(&count).Error

	return count >= 6, err
}
