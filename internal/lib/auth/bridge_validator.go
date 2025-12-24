package auth

import (
	"context"
	"fmt"
	"time"

	flowfull "github.com/pubflow/flowfull-go"
	"go.uber.org/zap"

	"github.com/pubflow/flowfull-go-starter/internal/config"
)

// BridgeValidator handles session validation with Flowless using the official flowfull-go client
type BridgeValidator struct {
	client *flowfull.FlowfullClient
	logger *zap.Logger
}

// NewBridgeValidator creates a new bridge validator instance
func NewBridgeValidator(cfg *config.Config, logger *zap.Logger) *BridgeValidator {
	// Create flowfull client with bridge secret
	client := flowfull.NewClient(
		cfg.FlowlessAPIURL,
		flowfull.WithTimeout(time.Duration(cfg.BridgeValidationTimeout)*time.Millisecond),
		flowfull.WithHeaders(map[string]string{
			"X-Bridge-Secret": cfg.BridgeValidationSecret,
		}),
		flowfull.WithRetry(cfg.BridgeRetryAttempts, 100*time.Millisecond, true),
	)

	return &BridgeValidator{
		client: client,
		logger: logger,
	}
}

// ValidateSession validates a session with Flowless using the official client
func (bv *BridgeValidator) ValidateSession(
	ctx context.Context,
	sessionID string,
	opts *ValidationOptions,
) (*SessionData, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	// Use the flowfull-go client's BridgeValidate method
	result, err := bv.client.Auth.BridgeValidate(sessionID)
	if err != nil {
		bv.logger.Error("bridge validation failed",
			zap.Error(err),
			zap.String("session_id", sessionID[:8]+"..."),
		)
		return nil, fmt.Errorf("bridge validation failed: %w", err)
	}

	// Check if validation was successful
	if !result.Success {
		bv.logger.Warn("session validation failed",
			zap.String("session_id", sessionID[:8]+"..."),
		)
		return nil, fmt.Errorf("session validation failed")
	}

	// Convert flowfull.User to our SessionData
	if result.User == nil {
		return nil, fmt.Errorf("no user data in validation result")
	}

	sessionData := &SessionData{
		UserID:         result.User.ID,
		Email:          result.User.Email,
		Name:           result.User.Name,
		UserType:       result.User.UserType,
		OrganizationID: nil,
		Permissions:    []string{},
		ExpiresAt:      time.Time{},
		ValidatedAt:    time.Now(),
	}

	// Set ExpiresAt from session if available
	if result.Session != nil {
		sessionData.ExpiresAt = result.Session.ExpiresAt
	}

	return sessionData, nil
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

