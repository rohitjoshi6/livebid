package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	HTTPAddr           string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	JWTSecret          string
	JWTIssuer          string
	AccessTokenTTL     time.Duration
	CORSAllowedOrigins []string
	MigrationsDir      string
	ExpirationInterval time.Duration
	ExpirationBatch    int
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://livebid:livebid@localhost:5432/livebid?sslmode=disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		JWTSecret:          getEnv("JWT_SECRET", "replace-with-a-local-dev-secret"),
		JWTIssuer:          getEnv("JWT_ISSUER", "livebid"),
		AccessTokenTTL:     time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 60)) * time.Minute,
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		MigrationsDir:      getEnv("MIGRATIONS_DIR", "migrations"),
		ExpirationInterval: time.Duration(getEnvInt("AUCTION_EXPIRATION_INTERVAL_SECONDS", 2)) * time.Second,
		ExpirationBatch:    getEnvInt("AUCTION_EXPIRATION_BATCH_SIZE", 50),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
