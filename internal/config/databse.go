package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() error {
	if err := godotenv.Load(); err != nil {
		// .env is optional in production where env vars are injected.
		_ = err
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Kolkata",
		MustGetEnv("DB_HOST"),
		MustGetEnv("DB_USER"),
		MustGetEnv("DB_PASSWORD"),
		MustGetEnv("DB_NAME"),
		MustGetEnv("DB_PORT"),
		GetEnv("DB_SSLMODE", "disable"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = database
	return nil
}
