package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	// Server configuration
	ServerPort string

	// JWT configuration
	JWTSecret string

	// File upload configuration
	UploadDir     string
	MaxUploadSize int64

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Log levels
	LogInfo  string
	LogWarn  string
	LogError string
)

// LoadConfig loads environment variables from .env file
func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values or environment variables")
	}

	// Server configuration
	ServerPort = getEnv("SERVER_PORT", ":8080")

	// JWT configuration
	JWTSecret = getEnv("JWT_SECRET", "your-secret-key-change-in-production")

	// File upload configuration
	UploadDir = getEnv("UPLOAD_DIR", "./uploads")
	MaxUploadSize = getEnvAsInt64("MAX_UPLOAD_SIZE", 10<<20) // 10MB default

	// Database configuration
	DBHost = getEnv("DB_HOST", "localhost")
	DBPort = getEnv("DB_PORT", "5432")
	DBUser = getEnv("DB_USER", "postgres")
	DBPassword = getEnv("DB_PASSWORD", "postgres")
	DBName = getEnv("DB_NAME", "petclinic")
	DBSSLMode = getEnv("DB_SSLMODE", "disable")

	// Log levels
	LogInfo = getEnv("LOG_INFO", "INFO")
	LogWarn = getEnv("LOG_WARN", "WARN")
	LogError = getEnv("LOG_ERROR", "ERROR")
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt64 reads an environment variable as int64 or returns a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

// GetDBConnectionString returns the PostgreSQL connection string
func GetDBConnectionString() string {
	return "host=" + DBHost +
		" port=" + DBPort +
		" user=" + DBUser +
		" password=" + DBPassword +
		" dbname=" + DBName +
		" sslmode=" + DBSSLMode
}