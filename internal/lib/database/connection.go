package database

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pubflow/flowfull-go-starter/internal/config"
)

// Connection holds database connection
type Connection struct {
	DB     *gorm.DB
	logger *zap.Logger
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config, zapLogger *zap.Logger) (*Connection, error) {
	// Determine database type from URL
	dbType := getDatabaseType(cfg.DatabaseURL)

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Create dialector based on database type
	var dialector gorm.Dialector
	switch dbType {
	case "postgres", "postgresql":
		dialector = postgres.Open(cfg.DatabaseURL)
	case "mysql":
		dialector = mysql.Open(cfg.DatabaseURL)
	case "sqlite":
		// Extract path from sqlite://path
		path := strings.TrimPrefix(cfg.DatabaseURL, "sqlite://")
		dialector = sqlite.Open(path)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Open connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DatabaseMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DatabaseConnMaxLifetime) * time.Second)

	zapLogger.Info("database connected",
		zap.String("type", dbType),
		zap.Int("max_idle_conns", cfg.DatabaseMaxIdleConns),
		zap.Int("max_open_conns", cfg.DatabaseMaxOpenConns),
	)

	return &Connection{
		DB:     db,
		logger: zapLogger,
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate runs auto-migration for models
func (c *Connection) AutoMigrate(models ...interface{}) error {
	c.logger.Info("running auto-migration")
	if err := c.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}
	c.logger.Info("auto-migration completed")
	return nil
}

// getDatabaseType extracts database type from connection URL
func getDatabaseType(url string) string {
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		return "postgres"
	}
	if strings.HasPrefix(url, "mysql://") {
		return "mysql"
	}
	if strings.HasPrefix(url, "sqlite://") {
		return "sqlite"
	}
	// Default to postgres if no prefix
	return "postgres"
}

