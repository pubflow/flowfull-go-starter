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
	UserID         string                 `json:"user_id"`
	Email          string                 `json:"email"`
	TokenType      string                 `json:"token_type"`
	IssuedAt       time.Time              `json:"iat"`
	ExpiresAt      time.Time              `json:"exp"`
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

