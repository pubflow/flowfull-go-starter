package auth

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/pubflow/flowfull-go-starter/internal/config"
	"github.com/pubflow/flowfull-go-starter/internal/lib/cache"
)

// MockBridgeValidator is a mock implementation of BridgeValidator
type MockBridgeValidator struct {
	mock.Mock
}

func (m *MockBridgeValidator) ValidateSession(ctx context.Context, sessionID string, opts *ValidationOptions) (*SessionData, error) {
	args := m.Called(ctx, sessionID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SessionData), args.Error(1)
}

// MockHybridCache is a mock implementation of HybridCache
type MockHybridCache struct {
	data map[string]interface{}
}

func NewMockHybridCache() *MockHybridCache {
	return &MockHybridCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockHybridCache) Get(ctx context.Context, key string) (interface{}, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *MockHybridCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockHybridCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockHybridCache) GetMetrics() *cache.CacheMetrics {
	return &cache.CacheMetrics{}
}

func (m *MockHybridCache) Close() {}

func TestRequireAuth_ValidSession(t *testing.T) {
	// Setup
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	mockValidator := new(MockBridgeValidator)
	mockCache := NewMockHybridCache()

	cfg := &config.Config{
		SessionHeaderName:           "X-Session-Id",
		SessionCookieName:           "session_id",
		AuthValidationMode:          "STANDARD",
		AuthEnableValidationMode:    true,
		SessionValidationCacheTTL:   300,
	}

	authMiddleware := NewAuthMiddleware(mockValidator, mockCache, cfg, logger)

	// Mock session data
	sessionData := &SessionData{
		UserID:         "user123",
		Email:          "test@example.com",
		Name:           "Test User",
		UserType:       "user",
		OrganizationID: "org123",
		Permissions:    []string{"read", "write"},
	}

	mockValidator.On("ValidateSession", mock.Anything, "valid-session-id", mock.Anything).
		Return(sessionData, nil)

	// Setup route
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		return c.JSON(fiber.Map{"user_id": userID})
	})

	// Test
	req := fiber.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI("/protected")
	req.Header.Set("X-Session-Id", "valid-session-id")

	resp, err := app.Test(req, -1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockValidator.AssertExpectations(t)
}

func TestRequireAuth_InvalidSession(t *testing.T) {
	// Setup
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	mockValidator := new(MockBridgeValidator)
	mockCache := NewMockHybridCache()

	cfg := &config.Config{
		SessionHeaderName:           "X-Session-Id",
		SessionCookieName:           "session_id",
		AuthValidationMode:          "STANDARD",
		AuthEnableValidationMode:    true,
		SessionValidationCacheTTL:   300,
	}

	authMiddleware := NewAuthMiddleware(mockValidator, mockCache, cfg, logger)

	mockValidator.On("ValidateSession", mock.Anything, "invalid-session-id", mock.Anything).
		Return(nil, fiber.NewError(fiber.StatusUnauthorized, "invalid session"))

	// Setup route
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Test
	req := fiber.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI("/protected")
	req.Header.Set("X-Session-Id", "invalid-session-id")

	resp, err := app.Test(req, -1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
	mockValidator.AssertExpectations(t)
}

func TestOptionalAuth_WithSession(t *testing.T) {
	// Setup
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	mockValidator := new(MockBridgeValidator)
	mockCache := NewMockHybridCache()

	cfg := &config.Config{
		SessionHeaderName:           "X-Session-Id",
		SessionCookieName:           "session_id",
		AuthValidationMode:          "STANDARD",
		AuthEnableValidationMode:    true,
		SessionValidationCacheTTL:   300,
	}

	authMiddleware := NewAuthMiddleware(mockValidator, mockCache, cfg, logger)

	sessionData := &SessionData{
		UserID: "user123",
		Email:  "test@example.com",
	}

	mockValidator.On("ValidateSession", mock.Anything, "valid-session-id", mock.Anything).
		Return(sessionData, nil)

	// Setup route
	app.Get("/optional", authMiddleware.OptionalAuth(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID != nil {
			return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
		}
		return c.JSON(fiber.Map{"authenticated": false})
	})

	// Test
	req := fiber.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI("/optional")
	req.Header.Set("X-Session-Id", "valid-session-id")

	resp, err := app.Test(req, -1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

