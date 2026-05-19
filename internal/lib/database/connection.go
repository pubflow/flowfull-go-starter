package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
	libsqlgorm "github.com/ytsruh/gorm-libsql"
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
	case "libsql":
		safeURL, authToken, err := parseLibSQLConnection(cfg.DatabaseURL, cfg.DatabaseAuthToken)
		if err != nil {
			return nil, err
		}
		connector, err := libsql.NewConnector(safeURL, libsql.WithAuthToken(authToken))
		if err != nil {
			return nil, fmt.Errorf("failed to create libsql connector: %w", err)
		}
		conn := sql.OpenDB(connector)
		dialector = libsqlgorm.New(libsqlgorm.Config{
			Conn: conn,
		})
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

func parseLibSQLConnection(databaseURL string, fallbackAuthToken string) (string, string, error) {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid libsql database url: %w", err)
	}

	query := parsedURL.Query()
	authToken := firstQueryValue(query, "authToken", "token", "auth_token", "jwt")
	if authToken == "" {
		authToken = strings.TrimSpace(fallbackAuthToken)
	}
	if authToken == "" {
		return "", "", fmt.Errorf("libsql database requires authToken in DATABASE_URL or DATABASE_AUTH_TOKEN/TURSO_AUTH_TOKEN")
	}

	for _, key := range []string{"authToken", "token", "auth_token", "jwt"} {
		query.Del(key)
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), authToken, nil
}

func firstQueryValue(values url.Values, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(values.Get(key))
		if value != "" {
			return value
		}
	}
	return ""
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
	if strings.HasPrefix(url, "libsql://") || strings.HasPrefix(url, "wss://") || strings.HasPrefix(url, "https://") {
		return "libsql"
	}
	// Default to postgres if no prefix
	return "postgres"
}
