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
	Mode                ValidationMode
	EnableValidation    bool
	IPValidation        bool
	UserAgentValidation bool
	DeviceValidation    bool
	AutoInvalidate      bool
	LogViolations       bool
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

