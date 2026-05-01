package db

import (
	"_discord_bot/models"
	"errors"
	"os/exec"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var Secret string

const DBPath = "database.db"

func Init() error {
	if err := checkSQLite3Exists(); err != nil {
		return err
	}

	database, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = database

	if err := DB.AutoMigrate(
		&models.User{},
		&models.LoginAttempt{},
		&models.HMACSecret{},
		&models.InternalLog{},
	); err != nil {
		return err
	}

	cfg := models.LoadConfig()

	if err := ensureAdminUser(DB, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return err
	}

	sec, err := ensureHMACSecret(DB)
	if err != nil {
		return err
	}

	Secret = sec
	return nil
}

/* -------------------- ensure functions -------------------- */

func ensureAdminUser(db *gorm.DB, username, password string) error {
	var count int64

	if err := db.Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err := models.CreateUser(db, username, password)
	return err
}

func checkSQLite3Exists() error {
	_, err := exec.LookPath("sqlite3")
	return err
}

func ensureHMACSecret(db *gorm.DB) (string, error) {
	if db == nil {
		return "", errors.New("db is nil")
	}

	var s models.HMACSecret

	err := db.Order("created_at DESC").First(&s).Error
	if err == nil {
		return s.Secret, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	secret, err := models.GenerateHMACSecret()
	if err != nil {
		return "", err
	}

	newSecret := models.HMACSecret{
		UUID:   uuid.NewString(),
		Secret: secret,
	}

	if err := db.Create(&newSecret).Error; err != nil {
		return "", err
	}

	return newSecret.Secret, nil
}
