package models

import (
	"time"

	"gorm.io/gorm"
)

type OnTimeBehaviour struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	BotUUID   string `gorm:"index;not null" json:"bot_uuid"`
	ChannelID string `gorm:"size:512;not null" json:"channel_id"`
	Message   string `gorm:"type:text;not null" json:"message"`

	// "interval" (every X seconds) or "specific" (one-time date)
	TimeType string `gorm:"size:20;not null" json:"time_type"`

	// If interval: seconds (e.g., 3600). If specific: Unix timestamp.
	TimeValue int64 `json:"time_value"`

	// Unix timestamp of last execution to prevent double-firing
	LastRun int64 `gorm:"default:0" json:"last_run"`

	Enabled   bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// -----------------------------
// CRUD FUNCTIONS
// -----------------------------

func CreateOnTime(db *gorm.DB, botUUID, channelID, message, timeType string, timeValue int64) (OnTimeBehaviour, error) {
	b := OnTimeBehaviour{
		BotUUID:   botUUID,
		ChannelID: channelID,
		Message:   message,
		TimeType:  timeType,
		TimeValue: timeValue,
		Enabled:   true,
	}
	err := db.Create(&b).Error
	return b, err
}

func ListOnTime(db *gorm.DB, botUUID string) ([]OnTimeBehaviour, error) {
	var list []OnTimeBehaviour
	err := db.Where("bot_uuid = ?", botUUID).Order("created_at desc").Find(&list).Error
	return list, err
}

func UpdateOnTime(db *gorm.DB, id uint, channelID, message, timeType string, timeValue int64, enabled bool) error {
	return db.Model(&OnTimeBehaviour{}).Where("id = ?", id).Updates(map[string]any{
		"channel_id": channelID,
		"message":    message,
		"time_type":  timeType,
		"time_value": timeValue,
		"enabled":    enabled,
	}).Error
}

func DeleteOnTime(db *gorm.DB, id uint) error {
	return db.Delete(&OnTimeBehaviour{}, id).Error
}

// ResetLastRun allows re-triggering a "specific" type event or restarting an interval
func ResetLastRun(db *gorm.DB, id uint) error {
	return db.Model(&OnTimeBehaviour{}).Where("id = ?", id).Update("last_run", 0).Error
}
