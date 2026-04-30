package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UUID               string    `gorm:"primaryKey;uniqueIndex;not null" json:"uuid"`
	Username           string    `gorm:"uniqueIndex;not null" json:"username"`
	Password           string    `gorm:"not null" json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	LastPasswordChange time.Time `gorm:"index"`
}

func (u *User) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(password),
	) == nil
}

func CreateUser(db *gorm.DB, username, password string) (*User, error) {
	u := &User{
		UUID:     uuid.NewString(),
		Username: username,
		Password: password,
	}

	if err := u.HashPassword(); err != nil {
		return nil, err
	}

	if err := db.Create(u).Error; err != nil {
		return nil, err
	}

	return u, nil
}

func GetUserByUUID(db *gorm.DB, uuid string) (*User, error) {
	var u User
	if err := db.Where("uuid = ?", uuid).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var u User
	if err := db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdatePassword(db *gorm.DB, uuid string, newPassword string) error {
	var u User
	if err := db.Where("uuid = ?", uuid).First(&u).Error; err != nil {
		return err
	}

	u.Password = newPassword
	if err := u.HashPassword(); err != nil {
		return err
	}

	return db.Save(&u).Error
}

func UpdateUsername(db *gorm.DB, uuid string, username string) error {
	return db.Model(&User{}).
		Where("uuid = ?", uuid).
		Update("username", username).Error
}

func DeleteUser(db *gorm.DB, uuid string) error {
	return db.Where("uuid = ?", uuid).Delete(&User{}).Error
}

func Authenticate(db *gorm.DB, username, password, ip, ua string) (*User, error) {
	u, err := GetUserByUsername(db, username)

	if err != nil {
		_ = RecordLoginAttempt(db, "", false, ip, ua)
		return nil, errors.New("invalid credentials")
	}

	if !u.CheckPassword(password) {
		_ = RecordLoginAttempt(db, u.UUID, false, ip, ua)
		return nil, errors.New("invalid credentials")
	}

	if err := RecordLoginAttempt(db, u.UUID, true, ip, ua); err != nil {
		return nil, err
	}

	return u, nil
}

func PasswordExpired(u *User) bool {
	return time.Since(u.LastPasswordChange) > 30*24*time.Hour
}

func PasswordTimeLeft(u *User) time.Duration {
	expiry := u.LastPasswordChange.Add(30 * 24 * time.Hour)
	return time.Until(expiry)
}
