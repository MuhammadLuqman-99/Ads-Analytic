package postgres

import (
	"context"
	"fmt"

	"github.com/ads-aggregator/ads-aggregator/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM database connection
type Database struct {
	*gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &Database{db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks the database connection
func (db *Database) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// WithContext returns a new database with context
func (db *Database) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

// Transaction runs a function within a database transaction
func (db *Database) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}
