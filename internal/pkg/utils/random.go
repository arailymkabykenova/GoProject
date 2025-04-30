package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomString(length int) (string, error) {
	numBytes := (length*6)/8 + 1
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	randomString := base64.RawURLEncoding.EncodeToString(b)
	if len(randomString) > length {
		return randomString[:length], nil
	}
	return randomString, nil
}
