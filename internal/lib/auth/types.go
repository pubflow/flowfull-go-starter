package auth

import "time"

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

