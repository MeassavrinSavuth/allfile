package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateVerificationToken generates a random hex token of 32 bytes (64 hex chars)
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
