package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func GetUserByID(db *gorm.DB, id uint) (*User, error) {
	var u User
	if err := db.First(&u, id).Error; err != nil {
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

func UpdatePassword(db *gorm.DB, id uint, newPassword string) error {
	var u User
	if err := db.First(&u, id).Error; err != nil {
		return err
	}
	u.Password = newPassword
	if err := u.HashPassword(); err != nil {
		return err
	}
	return db.Save(&u).Error
}

func UpdateUsername(db *gorm.DB, id uint, username string) error {
	return db.Model(&User{}).
		Where("id = ?", id).
		Update("username", username).Error
}

func DeleteUser(db *gorm.DB, id uint) error {
	return db.Delete(&User{}, id).Error
}

func Authenticate(db *gorm.DB, username, password, ip string, ua string) (*User, error) {
	u, err := GetUserByUsername(db, username)
	if err != nil {
		// avoid leaking whether user exists
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyhashdummyhashdummyha"), []byte(password))
		return nil, errors.New("invalid credentials")
	}

	if !u.CheckPassword(password) {
		_ = RecordLoginAttempt(db, u.ID, false, ip, ua)
		return nil, errors.New("invalid credentials")
	}

	_ = RecordLoginAttempt(db, u.ID, true, ip, ua)
	return u, nil
}
