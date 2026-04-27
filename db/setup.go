package db

import (
	"_discord_bot/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, err
	}
	cfg := models.LoadConfig()

	if err := ensureAdminUser(db, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureAdminUser(db *gorm.DB, username, password string) error {
	var count int64

	db.Model(&models.User{}).
		Where("username = ?", username).
		Count(&count)

	if count > 0 {
		return nil
	}

	_, err := models.CreateUser(db, username, password)
	return err
}
