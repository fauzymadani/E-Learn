package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection
func NewDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	// Use URL format instead of key=value format
	// URL format works, key=value format has issues on this system
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
		cfg.TimeZone,
	)

	log.Printf("DSN: postgresql://%s@%s:%s/%s", cfg.User, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		NowFunc:     func() time.Time { return time.Now().UTC() },
		PrepareStmt: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Second)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify connected database
	var currentDB string
	db.Raw("SELECT current_database()").Scan(&currentDB)
	log.Printf("VERIFY: Connected to database '%s'", currentDB)

	if currentDB != cfg.DBName {
		return nil, fmt.Errorf("failed to connect to correct database: expected '%s', got '%s'", cfg.DBName, currentDB)
	}

	log.Println("Database connected successfully")
	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
