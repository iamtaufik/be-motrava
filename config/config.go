package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config stores application runtime configuration values.
type Config struct {
	AppPort            string
	DatabaseURL        string
	LogFile            string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GoogleFrontendURL  string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppPort:            getEnv("APP_PORT", "3000"),
		DatabaseURL:        getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=motrava port=5432 sslmode=disable TimeZone=UTC"),
		LogFile:            getEnv("LOG_FILE", "log.json"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/api/auth/google/callback"),
		GoogleFrontendURL:  getEnv("GOOGLE_FRONTEND_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
