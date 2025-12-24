# 02 - Core Concepts Implementation

## 🎯 The 7 Core Concepts

This document provides complete implementation examples for all 7 core concepts in Go.

---

## 1️⃣ Bridge Validation

### internal/lib/auth/bridge_validator.go

```go
package auth

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-resty/resty/v2"
    "go.uber.org/zap"
    
    "github.com/yourusername/flowfull-go/internal/config"
)

// SessionData represents validated session information
type SessionData struct {
    UserID         string    `json:"user_id"`
    Email          string    `json:"email"`
    Name           string    `json:"name"`
    UserType       string    `json:"user_type"`
    OrganizationID *string   `json:"organization_id"`
    Permissions    []string  `json:"permissions"`
    ExpiresAt      time.Time `json:"expires_at"`
    ValidatedAt    time.Time `json:"validated_at"`
}

// ValidationOptions contains options for session validation
type ValidationOptions struct {
    IP        string
    UserAgent string
    DeviceID  string
}

// BridgeValidator handles session validation with Flowless
type BridgeValidator struct {
    flowlessURL      string
    validationSecret string
    timeout          time.Duration
    retryAttempts    int
    client           *resty.Client
    logger           *zap.Logger
}

// NewBridgeValidator creates a new bridge validator instance
func NewBridgeValidator(cfg *config.Config, logger *zap.Logger) *BridgeValidator {
    client := resty.New().
        SetTimeout(time.Duration(cfg.BridgeValidationTimeout) * time.Millisecond).
        SetRetryCount(cfg.BridgeRetryAttempts).
        SetRetryWaitTime(100 * time.Millisecond).
        SetRetryMaxWaitTime(2 * time.Second)

    return &BridgeValidator{
        flowlessURL:      cfg.FlowlessAPIURL,
        validationSecret: cfg.BridgeValidationSecret,
        timeout:          time.Duration(cfg.BridgeValidationTimeout) * time.Millisecond,
        retryAttempts:    cfg.BridgeRetryAttempts,
        client:           client,
        logger:           logger,
    }
}

// ValidateSession validates a session with Flowless
func (bv *BridgeValidator) ValidateSession(
    ctx context.Context,
    sessionID string,
    opts *ValidationOptions,
) (*SessionData, error) {
    if sessionID == "" {
        return nil, fmt.Errorf("session_id is required")
    }

    // Prepare request payload
    payload := map[string]interface{}{
        "session_id":    sessionID,
        "bridge_secret": bv.validationSecret,
    }

    if opts != nil {
        if opts.IP != "" {
            payload["ip"] = opts.IP
        }
        if opts.UserAgent != "" {
            payload["user_agent"] = opts.UserAgent
        }
        if opts.DeviceID != "" {
            payload["device_id"] = opts.DeviceID
        }
    }

    // Make request to Flowless
    url := fmt.Sprintf("%s/api/bridge/validate", bv.flowlessURL)
    
    bv.logger.Debug("validating session",
        zap.String("session_id", sessionID[:8]+"..."),
        zap.String("url", url),
    )

    resp, err := bv.client.R().
        SetContext(ctx).
        SetHeader("Content-Type", "application/json").
        SetBody(payload).
        Post(url)

    if err != nil {
        bv.logger.Error("bridge validation request failed",
            zap.Error(err),
            zap.String("session_id", sessionID[:8]+"..."),
        )
        return nil, fmt.Errorf("bridge validation failed: %w", err)
    }

    // Check response status
    if resp.StatusCode() != 200 {
        bv.logger.Warn("bridge validation rejected",
            zap.Int("status_code", resp.StatusCode()),
            zap.String("session_id", sessionID[:8]+"..."),
        )
        return nil, fmt.Errorf("invalid session: status %d", resp.StatusCode())
    }

    // Parse response
    var sessionData SessionData
    if err := json.Unmarshal(resp.Body(), &sessionData); err != nil {
        bv.logger.Error("failed to parse session data",
            zap.Error(err),
        )
        return nil, fmt.Errorf("failed to parse session data: %w", err)
    }

    bv.logger.Info("session validated successfully",
        zap.String("user_id", sessionData.UserID),
        zap.String("email", sessionData.Email),
    )

    return &sessionData, nil
}

// ValidateSessionWithRetry validates session with custom retry logic
func (bv *BridgeValidator) ValidateSessionWithRetry(
    ctx context.Context,
    sessionID string,
    opts *ValidationOptions,
    maxRetries int,
) (*SessionData, error) {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff):
            }
        }

        session, err := bv.ValidateSession(ctx, sessionID, opts)
        if err == nil {
            return session, nil
        }

        lastErr = err
        bv.logger.Warn("validation attempt failed",
            zap.Int("attempt", attempt+1),
            zap.Int("max_retries", maxRetries),
            zap.Error(err),
        )
    }

    return nil, fmt.Errorf("validation failed after %d attempts: %w", maxRetries, lastErr)
}
```

---

## 2️⃣ Validation Modes

### internal/lib/auth/validation_mode.go

```go
package auth

import "strings"

// ValidationMode represents the level of session validation
type ValidationMode string

const (
    ValidationModeDisabled ValidationMode = "DISABLED"
    ValidationModeStandard ValidationMode = "STANDARD"
    ValidationModeAdvanced ValidationMode = "ADVANCED"
    ValidationModeStrict   ValidationMode = "STRICT"
)

// ValidationConfig holds validation configuration
type ValidationConfig struct {
    Mode              ValidationMode
    EnableValidation  bool
    IPValidation      bool
    UserAgentValidation bool
    DeviceValidation  bool
    AutoInvalidate    bool
    LogViolations     bool
}

// ShouldValidateIP checks if IP validation is required
func (vc *ValidationConfig) ShouldValidateIP() bool {
    if !vc.EnableValidation || vc.Mode == ValidationModeDisabled {
        return false
    }
    return vc.Mode == ValidationModeStandard ||
           vc.Mode == ValidationModeAdvanced ||
           vc.Mode == ValidationModeStrict
}

// ShouldValidateUserAgent checks if User-Agent validation is required
func (vc *ValidationConfig) ShouldValidateUserAgent() bool {
    if !vc.EnableValidation || vc.Mode == ValidationModeDisabled {
        return false
    }
    return vc.Mode == ValidationModeAdvanced || vc.Mode == ValidationModeStrict
}

// ShouldValidateDevice checks if Device ID validation is required
func (vc *ValidationConfig) ShouldValidateDevice() bool {
    if !vc.EnableValidation || vc.Mode == ValidationModeDisabled {
        return false
    }
    return vc.Mode == ValidationModeStrict
}

// BuildValidationOptions creates ValidationOptions based on config
func (vc *ValidationConfig) BuildValidationOptions(ip, userAgent, deviceID string) *ValidationOptions {
    opts := &ValidationOptions{}

    if vc.ShouldValidateIP() {
        opts.IP = ip
    }

    if vc.ShouldValidateUserAgent() {
        opts.UserAgent = userAgent
    }

    if vc.ShouldValidateDevice() {
        opts.DeviceID = deviceID
    }

    return opts
}

// ParseValidationMode parses a string into ValidationMode
func ParseValidationMode(mode string) ValidationMode {
    switch strings.ToUpper(mode) {
    case "DISABLED":
        return ValidationModeDisabled
    case "STANDARD":
        return ValidationModeStandard
    case "ADVANCED":
        return ValidationModeAdvanced
    case "STRICT":
        return ValidationModeStrict
    default:
        return ValidationModeStandard
    }
}
```

---

## 3️⃣ HybridCache

### internal/lib/cache/hybrid_cache.go

```go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/dgraph-io/ristretto"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

// CacheMetrics tracks cache performance
type CacheMetrics struct {
    RistrettoHits   int64
    RistrettoMisses int64
    RedisHits       int64
    RedisMisses     int64
    DatabaseHits    int64
}

// HybridCache implements 3-tier caching: Ristretto → Redis → Database
type HybridCache struct {
    ristretto *ristretto.Cache
    redis     *redis.Client
    logger    *zap.Logger
    metrics   *CacheMetrics
    enabled   bool
}

// NewHybridCache creates a new HybridCache instance
func NewHybridCache(
    maxSize int64,
    redisClient *redis.Client,
    logger *zap.Logger,
    enabled bool,
) (*HybridCache, error) {
    if !enabled {
        logger.Info("cache disabled")
        return &HybridCache{
            enabled: false,
            logger:  logger,
            metrics: &CacheMetrics{},
        }, nil
    }

    // Configure Ristretto
    ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: maxSize * 10,
        MaxCost:     maxSize,
        BufferItems: 64,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
    }

    logger.Info("hybrid cache initialized",
        zap.Int64("max_size", maxSize),
        zap.Bool("redis_enabled", redisClient != nil),
    )

    return &HybridCache{
        ristretto: ristrettoCache,
        redis:     redisClient,
        logger:    logger,
        metrics:   &CacheMetrics{},
        enabled:   true,
    }, nil
}

// Get retrieves a value from cache (Ristretto → Redis → nil)
func (hc *HybridCache) Get(ctx context.Context, key string) (interface{}, bool) {
    if !hc.enabled {
        return nil, false
    }

    // Try Ristretto first
    if value, found := hc.ristretto.Get(key); found {
        hc.metrics.RistrettoHits++
        hc.logger.Debug("ristretto cache hit", zap.String("key", key))
        return value, true
    }
    hc.metrics.RistrettoMisses++

    // Try Redis
    if hc.redis != nil {
        val, err := hc.redis.Get(ctx, key).Result()
        if err == nil {
            hc.metrics.RedisHits++
            hc.logger.Debug("redis cache hit", zap.String("key", key))

            // Backfill Ristretto
            var data interface{}
            if err := json.Unmarshal([]byte(val), &data); err == nil {
                hc.ristretto.Set(key, data, 1)
            }

            return data, true
        }
        hc.metrics.RedisMisses++
    }

    return nil, false
}

// Set stores a value in all cache tiers
func (hc *HybridCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if !hc.enabled {
        return nil
    }

    // Set in Ristretto
    hc.ristretto.SetWithTTL(key, value, 1, ttl)

    // Set in Redis
    if hc.redis != nil {
        data, err := json.Marshal(value)
        if err != nil {
            hc.logger.Error("failed to marshal value for redis",
                zap.Error(err),
                zap.String("key", key),
            )
            return err
        }

        if err := hc.redis.Set(ctx, key, data, ttl).Err(); err != nil {
            hc.logger.Error("failed to set redis cache",
                zap.Error(err),
                zap.String("key", key),
            )
            return err
        }
    }

    hc.logger.Debug("cache set", zap.String("key", key), zap.Duration("ttl", ttl))
    return nil
}

// Delete removes a value from all cache tiers
func (hc *HybridCache) Delete(ctx context.Context, key string) error {
    if !hc.enabled {
        return nil
    }

    // Delete from Ristretto
    hc.ristretto.Del(key)

    // Delete from Redis
    if hc.redis != nil {
        if err := hc.redis.Del(ctx, key).Err(); err != nil {
            hc.logger.Error("failed to delete from redis",
                zap.Error(err),
                zap.String("key", key),
            )
            return err
        }
    }

    hc.logger.Debug("cache deleted", zap.String("key", key))
    return nil
}

// GetMetrics returns cache metrics
func (hc *HybridCache) GetMetrics() *CacheMetrics {
    return hc.metrics
}

// Close closes the cache
func (hc *HybridCache) Close() {
    if hc.ristretto != nil {
        hc.ristretto.Close()
    }
}
```

---

## 4️⃣ Auth Middleware

### internal/lib/auth/middleware.go

```go
package auth

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"

    "github.com/yourusername/flowfull-go/internal/config"
    "github.com/yourusername/flowfull-go/internal/lib/cache"
)

// AuthMiddleware handles authentication middleware
type AuthMiddleware struct {
    validator      *BridgeValidator
    cache          *cache.HybridCache
    config         *config.Config
    validationCfg  *ValidationConfig
    logger         *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(
    validator *BridgeValidator,
    cache *cache.HybridCache,
    cfg *config.Config,
    logger *zap.Logger,
) *AuthMiddleware {
    validationCfg := &ValidationConfig{
        Mode:                ParseValidationMode(cfg.AuthValidationMode),
        EnableValidation:    cfg.AuthEnableValidationMode,
        IPValidation:        cfg.AuthIPValidation,
        UserAgentValidation: cfg.AuthUserAgentValidation,
        DeviceValidation:    cfg.AuthDeviceValidation,
        AutoInvalidate:      cfg.AuthAutoInvalidate,
        LogViolations:       cfg.AuthLogViolations,
    }

    return &AuthMiddleware{
        validator:     validator,
        cache:         cache,
        config:        cfg,
        validationCfg: validationCfg,
        logger:        logger,
    }
}

// RequireAuth middleware requires authentication
func (am *AuthMiddleware) RequireAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        sessionID := am.extractSessionID(c)
        if sessionID == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "session_id required",
            })
        }

        // Check cache first
        cacheKey := "session:" + sessionID
        if cachedData, found := am.cache.Get(c.Context(), cacheKey); found {
            if session, ok := cachedData.(*SessionData); ok {
                am.setUserContext(c, session)
                return c.Next()
            }
        }

        // Validate with Flowless
        opts := am.buildValidationOptions(c)
        session, err := am.validator.ValidateSession(c.Context(), sessionID, opts)
        if err != nil {
            am.logger.Warn("session validation failed",
                zap.Error(err),
                zap.String("session_id", sessionID[:8]+"..."),
            )
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid session",
            })
        }

        // Cache session
        ttl := time.Duration(am.config.SessionValidationCacheTTL) * time.Second
        am.cache.Set(c.Context(), cacheKey, session, ttl)

        // Set user context
        am.setUserContext(c, session)

        return c.Next()
    }
}

// OptionalAuth middleware allows optional authentication
func (am *AuthMiddleware) OptionalAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        sessionID := am.extractSessionID(c)
        if sessionID == "" {
            return c.Next()
        }

        // Check cache
        cacheKey := "session:" + sessionID
        if cachedData, found := am.cache.Get(c.Context(), cacheKey); found {
            if session, ok := cachedData.(*SessionData); ok {
                am.setUserContext(c, session)
                return c.Next()
            }
        }

        // Validate with Flowless
        opts := am.buildValidationOptions(c)
        session, err := am.validator.ValidateSession(c.Context(), sessionID, opts)
        if err == nil {
            ttl := time.Duration(am.config.SessionValidationCacheTTL) * time.Second
            am.cache.Set(c.Context(), cacheKey, session, ttl)
            am.setUserContext(c, session)
        }

        return c.Next()
    }
}

// RequireUserType middleware requires specific user type
func (am *AuthMiddleware) RequireUserType(userTypes ...string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userType := c.Locals("user_type")
        if userType == nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "authentication required",
            })
        }

        userTypeStr := userType.(string)
        for _, allowedType := range userTypes {
            if userTypeStr == allowedType {
                return c.Next()
            }
        }

        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "insufficient permissions",
        })
    }
}

// extractSessionID extracts session ID from header or cookie
func (am *AuthMiddleware) extractSessionID(c *fiber.Ctx) string {
    // Try header first
    sessionID := c.Get(am.config.SessionHeaderName)
    if sessionID != "" {
        return sessionID
    }

    // Try cookie
    return c.Cookies(am.config.SessionCookieName)
}

// buildValidationOptions builds validation options from request
func (am *AuthMiddleware) buildValidationOptions(c *fiber.Ctx) *ValidationOptions {
    return am.validationCfg.BuildValidationOptions(
        c.IP(),
        c.Get("User-Agent"),
        c.Get("X-Device-ID"),
    )
}

// setUserContext sets user data in Fiber context
func (am *AuthMiddleware) setUserContext(c *fiber.Ctx, session *SessionData) {
    c.Locals("user_id", session.UserID)
    c.Locals("email", session.Email)
    c.Locals("name", session.Name)
    c.Locals("user_type", session.UserType)
    c.Locals("organization_id", session.OrganizationID)
    c.Locals("permissions", session.Permissions)
    c.Locals("session_data", session)
}
```

---

## 5️⃣ Trust Tokens (PASETO)

### internal/lib/tokens/trust_tokens.go

```go
package tokens

import (
    "crypto/ed25519"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/o1egl/paseto"
    "go.uber.org/zap"
)

// TokenClaims represents PASETO token claims
type TokenClaims struct {
    UserID         string    `json:"user_id"`
    Email          string    `json:"email"`
    TokenType      string    `json:"token_type"`
    IssuedAt       time.Time `json:"iat"`
    ExpiresAt      time.Time `json:"exp"`
    AdditionalData map[string]interface{} `json:"data,omitempty"`
}

// TrustTokenManager manages PASETO v4 tokens
type TrustTokenManager struct {
    privateKey ed25519.PrivateKey
    publicKey  ed25519.PublicKey
    logger     *zap.Logger
}

// NewTrustTokenManager creates a new trust token manager
func NewTrustTokenManager(privateKeyHex string, logger *zap.Logger) (*TrustTokenManager, error) {
    privateKeyBytes, err := hex.DecodeString(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("invalid private key: %w", err)
    }

    if len(privateKeyBytes) != ed25519.PrivateKeySize {
        return nil, fmt.Errorf("invalid private key size: expected %d, got %d",
            ed25519.PrivateKeySize, len(privateKeyBytes))
    }

    privateKey := ed25519.PrivateKey(privateKeyBytes)
    publicKey := privateKey.Public().(ed25519.PublicKey)

    return &TrustTokenManager{
        privateKey: privateKey,
        publicKey:  publicKey,
        logger:     logger,
    }, nil
}

// CreateToken creates a new PASETO v4 token
func (ttm *TrustTokenManager) CreateToken(
    userID string,
    email string,
    tokenType string,
    ttl time.Duration,
    additionalData map[string]interface{},
) (string, error) {
    now := time.Now()
    claims := TokenClaims{
        UserID:         userID,
        Email:          email,
        TokenType:      tokenType,
        IssuedAt:       now,
        ExpiresAt:      now.Add(ttl),
        AdditionalData: additionalData,
    }

    v4 := paseto.NewV4()
    token, err := v4.Sign(ttm.privateKey, claims, nil)
    if err != nil {
        ttm.logger.Error("failed to create token",
            zap.Error(err),
            zap.String("user_id", userID),
        )
        return "", fmt.Errorf("failed to create token: %w", err)
    }

    ttm.logger.Debug("token created",
        zap.String("user_id", userID),
        zap.String("token_type", tokenType),
        zap.Duration("ttl", ttl),
    )

    return token, nil
}

// VerifyToken verifies and decodes a PASETO v4 token
func (ttm *TrustTokenManager) VerifyToken(token string) (*TokenClaims, error) {
    var claims TokenClaims
    v4 := paseto.NewV4()

    err := v4.Verify(token, ttm.publicKey, &claims, nil)
    if err != nil {
        ttm.logger.Warn("token verification failed", zap.Error(err))
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    // Check expiration
    if time.Now().After(claims.ExpiresAt) {
        return nil, fmt.Errorf("token expired")
    }

    return &claims, nil
}

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() (string, string, error) {
    publicKey, privateKey, err := ed25519.GenerateKey(nil)
    if err != nil {
        return "", "", err
    }

    return hex.EncodeToString(publicKey),
           hex.EncodeToString(privateKey),
           nil
}
```

---

## Summary

This document covers the implementation of the 7 core concepts:

1. ✅ **Bridge Validation** - Session validation with Flowless
2. ✅ **Validation Modes** - Layered security
3. ✅ **HybridCache** - 3-tier caching with Ristretto
4. ✅ **Auth Middleware** - Fiber middleware for route protection
5. ✅ **Trust Tokens** - PASETO v4 tokens

**Remaining concepts** (covered in other files):
- **Multi-Database** - See `04-USAGE-GUIDE.md`
- **Environment Config** - See `03-ENVIRONMENT.md`

---

**Next**: [03-ENVIRONMENT.md](./03-ENVIRONMENT.md) - Environment configuration

