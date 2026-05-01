package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InternalLog struct {
	UUID      string    `gorm:"primaryKey;uniqueIndex;not null" json:"uuid"`
	CreatedAt time.Time `json:"created_at"`

	// Who did it
	ActorUUID string `gorm:"index;not null" json:"actor_uuid"`
	ActorType string `gorm:"size:20;not null" json:"actor_type"` // "user" or "system"

	// What happened
	Action   string `gorm:"not null" json:"action"` // e.g., "update_bot_token", "delete_user"
	TargetID string `gorm:"index" json:"target_id"` // ID of the bot or user being modified

	// Details
	OldValue  string `gorm:"type:text" json:"old_value"`
	NewValue  string `gorm:"type:text" json:"new_value"`
	IPAddress string `gorm:"size:45" json:"ip_address"`
}

func (l *InternalLog) BeforeCreate(tx *gorm.DB) (err error) {
	l.UUID = uuid.NewString()
	return
}

func RecordInternalLog(db *gorm.DB, log InternalLog) error {
	return db.Create(&log).Error
}

func GetLogsByTarget(db *gorm.DB, targetID string, page int, pageSize int) ([]InternalLog, int64, error) {
	var logs []InternalLog
	var total int64

	query := db.Model(&InternalLog{}).Where("target_id = ?", targetID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Order("created_at desc").Find(&logs).Error

	return logs, total, err
}

func GetLogsByActor(db *gorm.DB, actorUUID string, page int, pageSize int) ([]InternalLog, int64, error) {
	var logs []InternalLog
	var total int64

	baseQuery := db.Model(&InternalLog{}).Where("actor_uuid = ?", actorUUID)

	baseQuery.Count(&total)

	offset := (page - 1) * pageSize
	err := baseQuery.Limit(pageSize).Offset(offset).Order("created_at desc").Find(&logs).Error

	return logs, total, err
}

func GetSystemGlobalLogs(db *gorm.DB, page int, pageSize int) ([]InternalLog, int64, error) {
	var logs []InternalLog
	var total int64

	db.Model(&InternalLog{}).Count(&total)

	offset := (page - 1) * pageSize
	err := db.Limit(pageSize).Offset(offset).Order("created_at desc").Find(&logs).Error

	return logs, total, err
}
