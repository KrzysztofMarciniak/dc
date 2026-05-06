package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// OnMessageBehaviour is our simplified "X -> Y" logic
type OnMessageBehaviour struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	BotUUID string `gorm:"index;not null" json:"bot_uuid"`

	// If empty, behavior applies to all channels
	ChannelID string `gorm:"size:512;index" json:"channel_id"`

	// Trigger: "!" or "/"
	Trigger string `gorm:"size:10;not null" json:"trigger"`
	// Command: "hello"
	Command string `gorm:"size:100;not null" json:"command"`
	// Reply: "Hello world!"
	Reply string `gorm:"type:text;not null" json:"reply"`

	Enabled   bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DiscordBotEvent remains for passing data from the service to the logic
type DiscordBotEvent struct {
	Name      string `json:"name"`
	BotUUID   string `json:"bot_uuid"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt time.Time
}

// BehaviourAction tells the manager what to actually do
type BehaviourAction struct {
	Type    string // e.g., "SEND_MESSAGE"
	Target  string // e.g., ChannelID
	Payload string // e.g., The Reply string
}

type BotRuntime struct {
	DB *gorm.DB
}

// -----------------------------
// EXECUTION LOGIC
// -----------------------------

func (r *BotRuntime) HandleEvent(evt DiscordBotEvent) ([]BehaviourAction, error) {
	if evt.Name != "on_message" {
		return nil, nil
	}

	var match OnMessageBehaviour
	// Logic: Match Trigger+Command AND ensure the channel is either a specific match or global ("")
	err := r.DB.Where("bot_uuid = ? AND enabled = ?", evt.BotUUID, true).
		Where("(channel_id = ? OR channel_id = '' OR channel_id IS NULL)", evt.ChannelID).
		Where("? = Trigger || Command", evt.Content).
		First(&match).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return []BehaviourAction{
		{
			Type:    "SEND_MESSAGE",
			Target:  evt.ChannelID,
			Payload: match.Reply,
		},
	}, nil
}

// -----------------------------
// CRUD FUNCTIONS
// -----------------------------

func CreateOnMessage(db *gorm.DB, botUUID, trigger, command, reply, channelID string) (OnMessageBehaviour, error) {
	b := OnMessageBehaviour{
		BotUUID:   botUUID,
		Trigger:   trigger,
		Command:   command,
		Reply:     reply,
		ChannelID: channelID, // Pass empty string for "All Channels"
		Enabled:   true,
	}
	err := db.Create(&b).Error
	return b, err
}

func ListOnMessage(db *gorm.DB, botUUID string) ([]OnMessageBehaviour, error) {
	var list []OnMessageBehaviour
	err := db.Where("bot_uuid = ?", botUUID).Order("created_at desc").Find(&list).Error
	return list, err
}

func UpdateOnMessage(db *gorm.DB, id uint, trigger, command, reply, channelID string, enabled bool) error {
	return db.Model(&OnMessageBehaviour{}).Where("id = ?", id).Updates(map[string]any{
		"trigger":    trigger,
		"command":    command,
		"reply":      reply,
		"channel_id": channelID,
		"enabled":    enabled,
	}).Error
}

func DeleteOnMessage(db *gorm.DB, id uint) error {
	return db.Delete(&OnMessageBehaviour{}, id).Error
}
