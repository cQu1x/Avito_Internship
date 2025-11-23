package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	POSTGRES_DB       string
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	DB_HOST_PORT      string
	DB_DSN            string
	APP_PORT          string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load("deploy/.env"); err != nil {
		return nil, err
	}
	return &Config{
		POSTGRES_DB:       os.Getenv("POSTGRES_DB"),
		POSTGRES_USER:     os.Getenv("POSTGRES_USER"),
		POSTGRES_PASSWORD: os.Getenv("POSTGRES_PASSWORD"),
		DB_HOST_PORT:      os.Getenv("DB_HOST_PORT"),
		APP_PORT:          os.Getenv("APP_PORT"),
	}, nil
}
