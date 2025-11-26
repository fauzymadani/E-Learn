package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server           ServerConfig
	Database         DatabaseConfig
	JWT              JWTConfig
	NotificationGRPC string
	GCS              GCSConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type GCSConfig struct {
	BucketName string
	Enabled    bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARNING: Error loading .env file: %v", err)
	} else {
		log.Println("SUCCESS: .env file loaded successfully")
	}

	expHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION_HOURS: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "elearning"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: time.Duration(expHours) * time.Hour,
		},
		NotificationGRPC: getEnv("NOTIFICATION_GRPC_ADDR", "localhost:50051"),
		GCS: GCSConfig{
			BucketName: getEnv("GCS_BUCKET_NAME", ""),
			Enabled:    getEnv("GCS_ENABLED", "false") == "true",
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	log.Println("DEBUG: Loaded Database Config:")
	log.Printf("  DB_HOST: %s", cfg.Database.Host)
	log.Printf("  DB_PORT: %s", cfg.Database.Port)
	log.Printf("  DB_USER: %s", cfg.Database.User)
	log.Printf("  DB_NAME: %s", cfg.Database.DBName)
	log.Printf("  DB_SSLMODE: %s", cfg.Database.SSLMode)

	if cfg.GCS.Enabled {
		log.Printf("  GCS_ENABLED: true")
		log.Printf("  GCS_BUCKET: %s", cfg.GCS.BucketName)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
