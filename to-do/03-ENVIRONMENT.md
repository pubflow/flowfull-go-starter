# 03 - Environment Configuration

## 🔧 Configuration with Viper

Flowfull-Go uses **Viper** for configuration management with validation and type safety.

---

## 📝 internal/config/environment.go

```go
package config

import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
    // Server
    Port        int    `mapstructure:"PORT" validate:"required,min=1,max=65535"`
    Host        string `mapstructure:"HOST" validate:"required"`
    Environment string `mapstructure:"ENVIRONMENT" validate:"required,oneof=development production test"`
    BaseURL     string `mapstructure:"BASE_URL" validate:"required,url"`

    // Database
    DatabaseURL            string `mapstructure:"DATABASE_URL" validate:"required"`
    DatabaseMaxIdleConns   int    `mapstructure:"DATABASE_MAX_IDLE_CONNS" validate:"min=1"`
    DatabaseMaxOpenConns   int    `mapstructure:"DATABASE_MAX_OPEN_CONNS" validate:"min=1"`
    DatabaseConnMaxLifetime int   `mapstructure:"DATABASE_CONN_MAX_LIFETIME" validate:"min=0"`

    // Flowless Integration
    FlowlessAPIURL           string `mapstructure:"FLOWLESS_API_URL" validate:"required,url"`
    BridgeValidationSecret   string `mapstructure:"BRIDGE_VALIDATION_SECRET" validate:"required,min=32"`
    BridgeValidationTimeout  int    `mapstructure:"BRIDGE_VALIDATION_TIMEOUT" validate:"min=1000"`
    BridgeRetryAttempts      int    `mapstructure:"BRIDGE_RETRY_ATTEMPTS" validate:"min=0,max=10"`

    // Session Management
    SessionValidationCacheTTL int    `mapstructure:"SESSION_VALIDATION_CACHE_TTL" validate:"min=60"`
    SessionHeaderName         string `mapstructure:"SESSION_HEADER_NAME" validate:"required"`
    SessionCookieName         string `mapstructure:"SESSION_COOKIE_NAME" validate:"required"`

    // Authentication & Validation Mode
    AuthValidationMode         string `mapstructure:"AUTH_VALIDATION_MODE" validate:"required,oneof=DISABLED STANDARD ADVANCED STRICT"`
    AuthEnableValidationMode   bool   `mapstructure:"AUTH_ENABLE_VALIDATION_MODE"`
    AuthIPValidation           bool   `mapstructure:"AUTH_IP_VALIDATION"`
    AuthUserAgentValidation    bool   `mapstructure:"AUTH_USER_AGENT_VALIDATION"`
    AuthDeviceValidation       bool   `mapstructure:"AUTH_DEVICE_VALIDATION"`
    AuthAutoInvalidate         bool   `mapstructure:"AUTH_AUTO_INVALIDATE"`
    AuthLogViolations          bool   `mapstructure:"AUTH_LOG_VIOLATIONS"`

    // HybridCache
    CacheEnabled  bool   `mapstructure:"CACHE_ENABLED"`
    CacheMaxSize  int64  `mapstructure:"CACHE_MAX_SIZE" validate:"min=1000"`
    RedisURL      string `mapstructure:"REDIS_URL"`

    // Trust Tokens (PASETO)
    PasetoPrivateKey                string `mapstructure:"PASETO_PRIVATE_KEY"`
    TokenTTLHours                   int    `mapstructure:"TOKEN_TTL_HOURS" validate:"min=1"`
    TokenEmailVerificationTTLHours  int    `mapstructure:"TOKEN_EMAIL_VERIFICATION_TTL_HOURS" validate:"min=1"`
    TokenPasswordResetTTLHours      int    `mapstructure:"TOKEN_PASSWORD_RESET_TTL_HOURS" validate:"min=1"`
    TokenInvitationTTLHours         int    `mapstructure:"TOKEN_INVITATION_TTL_HOURS" validate:"min=1"`

    // Security & CORS
    CORSOrigins     string `mapstructure:"CORS_ORIGINS" validate:"required"`
    CORSMethods     string `mapstructure:"CORS_METHODS" validate:"required"`
    CORSHeaders     string `mapstructure:"CORS_HEADERS" validate:"required"`
    CORSCredentials bool   `mapstructure:"CORS_CREDENTIALS"`
    CORSMaxAge      int    `mapstructure:"CORS_MAX_AGE" validate:"min=0"`

    // Rate Limiting
    RateLimitEnabled  bool `mapstructure:"RATE_LIMIT_ENABLED"`
    RateLimitRequests int  `mapstructure:"RATE_LIMIT_REQUESTS" validate:"min=1"`
    RateLimitWindow   int  `mapstructure:"RATE_LIMIT_WINDOW" validate:"min=1"`

    // Logging & Monitoring
    LogLevel  string `mapstructure:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
    LogFormat string `mapstructure:"LOG_FORMAT" validate:"required,oneof=json console"`
    LogMode   bool   `mapstructure:"LOG_MODE"`

    // Development Settings
    DevMode         bool `mapstructure:"DEV_MODE"`
    DevCORSRelaxed  bool `mapstructure:"DEV_CORS_RELAXED"`
    DevLogRequests  bool `mapstructure:"DEV_LOG_REQUESTS"`
    Reload          bool `mapstructure:"RELOAD"`
}

// LoadConfig loads configuration from environment
func LoadConfig() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    // Set defaults
    setDefaults()

    // Read config file (optional)
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("error reading config file: %w", err)
        }
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("unable to decode config: %w", err)
    }

    // Validate configuration
    validate := validator.New()
    if err := validate.Struct(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return &config, nil
}

// setDefaults sets default values
func setDefaults() {
    viper.SetDefault("PORT", 3001)
    viper.SetDefault("HOST", "0.0.0.0")
    viper.SetDefault("ENVIRONMENT", "development")
    viper.SetDefault("BASE_URL", "http://localhost:3001")

    viper.SetDefault("DATABASE_MAX_IDLE_CONNS", 10)
    viper.SetDefault("DATABASE_MAX_OPEN_CONNS", 100)
    viper.SetDefault("DATABASE_CONN_MAX_LIFETIME", 3600)

    viper.SetDefault("BRIDGE_VALIDATION_TIMEOUT", 5000)
    viper.SetDefault("BRIDGE_RETRY_ATTEMPTS", 3)

    viper.SetDefault("SESSION_VALIDATION_CACHE_TTL", 300)
    viper.SetDefault("SESSION_HEADER_NAME", "X-Session-Id")
    viper.SetDefault("SESSION_COOKIE_NAME", "session_id")

    viper.SetDefault("AUTH_VALIDATION_MODE", "STANDARD")
    viper.SetDefault("AUTH_ENABLE_VALIDATION_MODE", true)
    viper.SetDefault("AUTH_IP_VALIDATION", true)
    viper.SetDefault("AUTH_USER_AGENT_VALIDATION", true)
    viper.SetDefault("AUTH_DEVICE_VALIDATION", false)
    viper.SetDefault("AUTH_AUTO_INVALIDATE", false)
    viper.SetDefault("AUTH_LOG_VIOLATIONS", true)

    viper.SetDefault("CACHE_ENABLED", true)
    viper.SetDefault("CACHE_MAX_SIZE", 50000)

    viper.SetDefault("TOKEN_TTL_HOURS", 168)
    viper.SetDefault("TOKEN_EMAIL_VERIFICATION_TTL_HOURS", 24)
    viper.SetDefault("TOKEN_PASSWORD_RESET_TTL_HOURS", 1)
    viper.SetDefault("TOKEN_INVITATION_TTL_HOURS", 168)

    viper.SetDefault("CORS_ORIGINS", "http://localhost:3000")
    viper.SetDefault("CORS_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
    viper.SetDefault("CORS_HEADERS", "Content-Type,Authorization,X-Session-Id")
    viper.SetDefault("CORS_CREDENTIALS", true)
    viper.SetDefault("CORS_MAX_AGE", 86400)

    viper.SetDefault("RATE_LIMIT_ENABLED", true)
    viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
    viper.SetDefault("RATE_LIMIT_WINDOW", 60)

    viper.SetDefault("LOG_LEVEL", "info")
    viper.SetDefault("LOG_FORMAT", "json")
    viper.SetDefault("LOG_MODE", false)

    viper.SetDefault("DEV_MODE", true)
    viper.SetDefault("DEV_CORS_RELAXED", true)
    viper.SetDefault("DEV_LOG_REQUESTS", true)
    viper.SetDefault("RELOAD", true)
}

// GetCORSOrigins returns CORS origins as slice
func (c *Config) GetCORSOrigins() []string {
    return strings.Split(c.CORSOrigins, ",")
}

// IsDevelopment checks if environment is development
func (c *Config) IsDevelopment() bool {
    return c.Environment == "development"
}

// IsProduction checks if environment is production
func (c *Config) IsProduction() bool {
    return c.Environment == "production"
}
```

---

**Next**: [04-USAGE-GUIDE.md](./04-USAGE-GUIDE.md) - Usage guide with examples

