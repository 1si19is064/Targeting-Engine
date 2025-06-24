package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
	Environment string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:Tejas%402001@db.wywkucanulrrkqgexwcp.supabase.co:5432/postgres?sslmode=require"),
		RedisURL:    getEnv("REDIS_URL", "redis://default:hDbjDRpv9yi892LytkwuAs1yrKSw8cjL@redis-14159.c206.ap-south-1-1.ec2.redns.redis-cloud.com:14159"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
