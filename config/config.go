package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config stores application runtime configuration values.
type Config struct {
	AppPort               string
	DatabaseURL           string
	LogFile               string
	JWTSecret             string
	JWTIssuer             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	GoogleFrontendURL     string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppPort:               getEnv("APP_PORT", "3000"),
		DatabaseURL:           getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=motrava port=5432 sslmode=disable TimeZone=UTC"),
		LogFile:               getEnv("LOG_FILE", "log.json"),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		JWTIssuer:             getEnv("JWT_ISSUER", "motrava"),
		AccessTokenTTLMinutes: getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15),
		RefreshTokenTTLHours:  getEnvInt("REFRESH_TOKEN_TTL_HOURS", 720),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:     getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/api/auth/google/callback"),
		GoogleFrontendURL:     getEnv("GOOGLE_FRONTEND_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
