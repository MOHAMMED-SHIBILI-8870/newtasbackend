package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		if strings.Contains(err.Error(), "cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {
			return nil
		}
		return err
	}
	return nil
}

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func MustGetEnv(key string) string {
	value := GetEnv(key, "")
	if value == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return value
}
