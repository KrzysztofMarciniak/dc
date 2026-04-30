package models

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

/*
CSRF token format:

	token = randomValue.signature

Where:
- randomValue = 32-byte random hex string
- signature   = HMAC(randomValue, HMACSecret.Secret)
*/

/* -------------------- CONFIG -------------------- */

const CSRFTokenBytes = 32

/* -------------------- GENERATION -------------------- */

func NewCSRFToken(secret string) (string, error) {
	b := make([]byte, CSRFTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	value := hex.EncodeToString(b)

	sig, err := Sign(value, secret)
	if err != nil {
		return "", err
	}

	return value + "." + sig, nil
}

/* -------------------- VALIDATION -------------------- */

func ValidateCSRFToken(token string, secret string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, nil
	}

	value := parts[0]
	sig := parts[1]

	return Verify(value, sig, secret)
}
