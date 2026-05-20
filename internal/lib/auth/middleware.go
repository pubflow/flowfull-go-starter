package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/pubflow/flowfull-go-starter/internal/config"
	"github.com/pubflow/flowfull-go-starter/internal/lib/cache"
	"github.com/pubflow/flowfull-go-starter/internal/lib/utils"
)

// AuthMiddleware handles authentication middleware
type AuthMiddleware struct {
	validator     *BridgeValidator
	cache         *cache.HybridCache
	config        *config.Config
	validationCfg *ValidationConfig
	logger        *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(
	validator *BridgeValidator,
	hybridCache *cache.HybridCache,
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
		cache:         hybridCache,
		config:        cfg,
		validationCfg: validationCfg,
		logger:        logger,
	}
}

// RequireAuth middleware requires authentication
func (am *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		am.logger.Info("🔐 RequireAuth middleware started",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		sessionID := am.extractSessionID(c)
		am.logger.Info("📝 Session ID extracted",
			zap.String("session_id", sessionID),
			zap.Bool("is_empty", sessionID == ""),
		)

		if sessionID == "" {
			am.logger.Warn("❌ No session ID provided")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "session_id required",
			})
		}

		// Check cache first
		cacheKey := "session:" + sessionID
		am.logger.Info("🔍 Checking cache",
			zap.String("cache_key", cacheKey),
		)

		if cachedData, found := am.cache.Get(c.Context(), cacheKey); found {
			am.logger.Info("✅ Cache HIT - session found in cache")
			if session, ok := cachedData.(*SessionData); ok {
				am.setUserContext(c, session)
				return c.Next()
			}
			am.logger.Warn("⚠️ Cache data type mismatch")
		} else {
			am.logger.Info("❌ Cache MISS - session not in cache")
		}

		// Validate with Flowless
		opts := am.buildValidationOptions(c)
		am.logger.Info("🌐 Validating with bridge",
			zap.String("validation_mode", string(am.validationCfg.Mode)),
			zap.String("flowless_url", am.config.FlowlessAPIURL),
		)

		session, err := am.validator.ValidateSession(c.Context(), sessionID, opts)
		if err != nil {
			am.logger.Warn("❌ Session validation failed",
				zap.Error(err),
				zap.String("session_id", sessionID[:8]+"..."),
				zap.String("validation_mode", string(am.validationCfg.Mode)),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid session",
			})
		}

		am.logger.Info("✅ Session validated successfully",
			zap.String("user_id", session.UserID),
			zap.String("email", session.Email),
		)

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
	am.logger.Info("🔍 Checking header for session",
		zap.String("header_name", am.config.SessionHeaderName),
		zap.String("header_value", sessionID),
		zap.Bool("found_in_header", sessionID != ""),
	)
	if sessionID != "" {
		return sessionID
	}

	// Try cookie
	cookieValue := c.Cookies(am.config.SessionCookieName)
	am.logger.Info("🍪 Checking cookie for session",
		zap.String("cookie_name", am.config.SessionCookieName),
		zap.String("cookie_value", cookieValue),
		zap.Bool("found_in_cookie", cookieValue != ""),
	)

	// Try query parameter as fallback
	queryValue := c.Query("session_id")
	am.logger.Info("🔗 Checking query parameter for session",
		zap.String("query_param", "session_id"),
		zap.String("query_value", queryValue),
		zap.Bool("found_in_query", queryValue != ""),
	)

	if cookieValue != "" {
		return cookieValue
	}

	return queryValue
}

// buildValidationOptions builds validation options from request
func (am *AuthMiddleware) buildValidationOptions(c *fiber.Ctx) *ValidationOptions {
	return am.validationCfg.BuildValidationOptions(
		utils.ExtractClientIP(c),
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
