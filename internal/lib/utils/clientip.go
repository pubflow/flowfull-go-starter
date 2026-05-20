package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ExtractClientIP resolves the real visitor IP when the app runs behind
// Cloudflare, Traefik, Nginx, or Coolify's proxy.
func ExtractClientIP(c *fiber.Ctx) string {
	headers := []string{
		"CF-Connecting-IP",
		"True-Client-IP",
		"Fly-Client-IP",
		"X-Real-IP",
		"X-Client-IP",
		"X-Forwarded-For",
		"Forwarded-For",
		"X-Forwarded",
		"Forwarded",
	}

	for _, header := range headers {
		value := strings.TrimSpace(c.Get(header))
		if value == "" {
			continue
		}

		first := strings.TrimSpace(strings.Split(value, ",")[0])
		normalized := NormalizeIP(first)
		if normalized != "" && normalized != "unknown" {
			return normalized
		}
	}

	return NormalizeIP(c.IP())
}

// NormalizeIP strips proxy-header wrappers, ports, and IPv4-mapped IPv6
// prefixes so downstream validation sees a stable client IP.
func NormalizeIP(value string) string {
	ip := strings.TrimSpace(value)
	if ip == "" {
		return ""
	}

	lower := strings.ToLower(ip)
	if strings.HasPrefix(lower, "for=") {
		ip = strings.TrimSpace(ip[4:])
	}

	ip = strings.Trim(ip, "\"'")
	if idx := strings.Index(ip, ";"); idx >= 0 {
		ip = strings.TrimSpace(ip[:idx])
	}

	if strings.HasPrefix(ip, "[") {
		if end := strings.Index(ip, "]"); end > 1 {
			ip = ip[1:end]
		}
	} else if strings.Count(ip, ":") == 1 && strings.Contains(ip, ".") {
		parts := strings.Split(ip, ":")
		if len(parts) == 2 {
			ip = parts[0]
		}
	}

	lower = strings.ToLower(strings.TrimSpace(ip))
	if strings.HasPrefix(lower, "::ffff:") {
		lower = strings.TrimPrefix(lower, "::ffff:")
	}
	if lower == "::1" {
		return "127.0.0.1"
	}
	return lower
}
