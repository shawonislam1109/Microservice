package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Shared
	PostgresURL string
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// RADIUS Server
	RadiusSecret   string
	AuthPort       int
	AcctPort       int
	WorkerPoolSize int

	// API Server
	APIPort string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file for local development
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		PostgresURL:    getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/radius"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvAsInt("REDIS_DB", 0),
		RadiusSecret:   getEnv("RADIUS_SECRET", "mikrotik_secret"),
		AuthPort:       getEnvAsInt("AUTH_PORT", 1812),
		AcctPort:       getEnvAsInt("ACCT_PORT", 1813),
		WorkerPoolSize: getEnvAsInt("WORKER_POOL_SIZE", 10),
		APIPort:        getEnv("API_PORT", ":8080"),
	}
}

// Helper function to get an environment variable or return a default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Helper function to get an environment variable as an integer or return a default value
func getEnvAsInt(key string, fallback int) int {
	if valueStr, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return fallback
}