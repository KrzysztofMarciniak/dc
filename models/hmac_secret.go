package models

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type HMACSecret struct {
	UUID      string    `gorm:"primaryKey;type:char(36)"`
	Secret    string    `gorm:"type:char(64);not null;uniqueIndex"` // 32-byte hex key
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

/* -------------------- generation -------------------- */

func GenerateHMACSecret() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func NewHMACSecret() (*HMACSecret, error) {
	secret, err := GenerateHMACSecret()
	if err != nil {
		return nil, err
	}

	return &HMACSecret{
		UUID:   uuid.NewString(),
		Secret: secret,
	}, nil
}

/* -------------------- HMAC -------------------- */

func Sign(value string, secret string) (string, error) {
	key, err := hex.DecodeString(secret)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(value))

	return hex.EncodeToString(h.Sum(nil)), nil
}

func Verify(value, sig, secret string) (bool, error) {
	expected, err := Sign(value, secret)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(expected), []byte(sig)), nil
}
