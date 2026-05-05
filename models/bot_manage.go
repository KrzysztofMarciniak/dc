package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscordBotManagement struct {
	UUID      string `gorm:"primaryKey;uniqueIndex;not null" json:"uuid"`
	Name      string `gorm:"size:100;not null" json:"name"`
	Token     string `gorm:"type:text;not null" json:"-"` // never expose
	PublicKey string `gorm:"size:255;not null" json:"public_key"`

	Retired bool `gorm:"default:false;index"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *DiscordBotManagement) BeforeCreate(tx *gorm.DB) (err error) {
	b.UUID = uuid.NewString()
	return
}

type DiscordBotDTO struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func toDTO(b DiscordBotManagement) DiscordBotDTO {
	return DiscordBotDTO{
		UUID:      b.UUID,
		Name:      b.Name,
		PublicKey: b.PublicKey,
	}
}
func CreateDiscordBot(db *gorm.DB, name, token, publicKey string) (DiscordBotDTO, error) {
	bot := DiscordBotManagement{
		Name:      name,
		Token:     token,
		PublicKey: publicKey,
	}

	if err := db.Create(&bot).Error; err != nil {
		return DiscordBotDTO{}, err
	}

	return toDTO(bot), nil
}
func GetDiscordBot(db *gorm.DB, uuid string) (DiscordBotDTO, error) {
	var bot DiscordBotManagement

	err := db.
		Select("uuid", "name", "public_key").
		Where("uuid = ?", uuid).
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
		Select("uuid", "name", "public_key").
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

func DeleteDiscordBot(db *gorm.DB, uuid string) error {
	return db.
		Where("uuid = ?", uuid).
		Delete(&DiscordBotManagement{}).
		Error
}
