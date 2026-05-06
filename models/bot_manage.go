package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscordBotManagement struct {
	UUID     string `gorm:"primaryKey;uniqueIndex;not null" json:"uuid"`
	Name     string `gorm:"size:100;not null" json:"name"`
	Token    string `gorm:"type:text;not null" json:"-"` // never expose
	ClientID string `gorm:"size:255;not null;default:''" json:"client_id"`

	Retired bool `gorm:"default:false;index"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *DiscordBotManagement) BeforeCreate(tx *gorm.DB) (err error) {
	b.UUID = uuid.NewString()
	return
}

type DiscordBotDTO struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
}

func toDTO(b DiscordBotManagement) DiscordBotDTO {
	return DiscordBotDTO{
		UUID:     b.UUID,
		Name:     b.Name,
		ClientID: b.ClientID,
	}
}
func CreateDiscordBot(db *gorm.DB, name, token, clientId string) (DiscordBotDTO, error) {
	bot := DiscordBotManagement{
		Name:     name,
		Token:    token,
		Retired:  false,
		ClientID: clientId,
	}

	if err := db.Create(&bot).Error; err != nil {
		return DiscordBotDTO{}, err
	}

	return toDTO(bot), nil
}
func GetDiscordBot(db *gorm.DB, uuid string) (DiscordBotDTO, error) {
	var bot DiscordBotManagement
	err := db.
		Select("uuid", "name", "client_id").
		Where("uuid = ? AND retired = ?", uuid, false).
		First(&bot).Error
	if err != nil {
		return DiscordBotDTO{}, err
	}

	return toDTO(bot), nil
}
func ListDiscordBots(db *gorm.DB, page, pageSize int) ([]DiscordBotDTO, int64, error) {
	var bots []DiscordBotManagement
	var total int64

	db.Model(&DiscordBotManagement{}).Count(&total)

	offset := (page - 1) * pageSize

	err := db.
		Select("uuid", "name").
		Where("retired = ?", false).
		Limit(pageSize).
		Offset(offset).
		Order("created_at desc").
		Find(&bots).Error

	if err != nil {
		return nil, 0, err
	}

	out := make([]DiscordBotDTO, len(bots))
	for i, b := range bots {
		out[i] = toDTO(b)
	}

	return out, total, nil
}
func UpdateDiscordBotName(db *gorm.DB, uuid, newName string) error {
	return db.Model(&DiscordBotManagement{}).
		Where("uuid = ?", uuid).
		Update("name", newName).Error
}

// OLD, use `RetireDiscordBot`
func DeleteDiscordBot(db *gorm.DB, uuid string) error {
	return db.
		Where("uuid = ?", uuid).
		Delete(&DiscordBotManagement{}).
		Error
}
func RetireDiscordBot(db *gorm.DB, uuid string) error {
	return db.Model(&DiscordBotManagement{}).
		Where("uuid = ?", uuid).
		Update("retired", true).Error
}

func GetFullDiscordBot(db *gorm.DB, uuid string) (DiscordBotManagement, error) {
	var bot DiscordBotManagement
	err := db.Where("uuid = ? AND retired = ?", uuid, false).First(&bot).Error
	return bot, err
}

func ValidateBotCredentials(token string) (string, error) {
	cleanedToken := strings.TrimSpace(token)
	if cleanedToken == "" {
		return "", fmt.Errorf("token is empty")
	}

	dg, err := discordgo.New("Bot " + cleanedToken)
	if err != nil {
		return "", fmt.Errorf("failed to initialize session: %w", err)
	}

	user, err := dg.User("@me")
	if err != nil {
		return "", fmt.Errorf("discord auth failed: %w", err)
	}

	defer dg.Close()

	if user.Discriminator == "0" || user.Discriminator == "" {
		return user.Username, nil
	}
	return fmt.Sprintf("%s#%s", user.Username, user.Discriminator), nil
}
