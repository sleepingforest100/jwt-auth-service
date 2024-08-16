package utils

import (
	"crypto/rand"
	"encoding/base64"
	"os"
)

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func GenerateRandomBase64String(size int) string {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}
