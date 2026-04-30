package models

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

/*
CSRF token format:

	token = randomValue|timestamp.signature

Where:
- randomValue = 32-byte random hex string
- timestamp   = unix seconds
- signature   = HMAC(randomValue|timestamp, HMACSecret.Secret)
*/

const CSRFTokenBytes = 32
const CSRFTokenTTL = time.Hour

func NewCSRFToken(secret string) (string, error) {
	b := make([]byte, CSRFTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	value := hex.EncodeToString(b)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	payload := value + "|" + ts

	sig, err := Sign(payload, secret)
	if err != nil {
		return "", err
	}

	return payload + "." + sig, nil
}

func ValidateCSRFToken(token string, secret string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, nil
	}

	payload := parts[0]
	sig := parts[1]

	ok, err := Verify(payload, sig, secret)
	if err != nil || !ok {
		return false, err
	}

	parts = strings.Split(payload, "|")
	if len(parts) != 2 {
		return false, nil
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, nil
	}

	if time.Since(time.Unix(ts, 0)) > CSRFTokenTTL {
		return false, nil
	}

	return true, nil
}
