package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	JWTSecret string
	TaxRate   float64
}

func Load() Config {
	return Config{
		Port:      getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "127.0.0.1"),
		DBPort:    getEnv("DB_PORT", "3306"),
		DBUser:    getEnv("DB_USER", "root"),
		DBPass:    getEnv("DB_PASS", "root"),
		DBName:    getEnv("DB_NAME", "mangea"),
		JWTSecret: getEnv("JWT_SECRET", "mangea-dev-secret-change-in-production"),
		TaxRate:   getEnvFloat("TAX_RATE", 0.10),
	}
}

func getEnvFloat(key string, fallback float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
	}
	return fallback
}
