package clientip

import (
	"fmt"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const DefaultTrustedProxies = "172.16.0.0/12,10.0.0.0/8,192.168.0.0/16,127.0.0.1,::1"

// List is the set of reverse proxies allowed to supply the client IP.
type List struct {
	entries []string
	nets    []*net.IPNet
}

// Parse builds a trust list from a comma-separated set of CIDRs and IPs.
// An empty value uses DefaultTrustedProxies.
func Parse(raw string) (*List, error) {
	if strings.TrimSpace(raw) == "" {
		raw = DefaultTrustedProxies
	}

	parts := strings.Split(raw, ",")
	list := &List{
		entries: make([]string, 0, len(parts)),
		nets:    make([]*net.IPNet, 0, len(parts)),
	}
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			return nil, fmt.Errorf("trusted proxy entry is empty")
		}
		network, err := parseEntry(entry)
		if err != nil {
			return nil, err
		}
		list.entries = append(list.entries, entry)
		list.nets = append(list.nets, network)
	}
	return list, nil
}

// Strings returns the configured entries for Fiber's TrustedProxies.
func (l *List) Strings() []string {
	if l == nil {
		return nil
	}
	out := make([]string, len(l.entries))
	copy(out, l.entries)
	return out
}

// Contains reports whether ip belongs to a trusted proxy.
func (l *List) Contains(ip net.IP) bool {
	if l == nil || ip == nil {
		return false
	}
	for _, network := range l.nets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// Middleware rewrites X-Forwarded-For to a single client IP when the TCP peer
// is a trusted proxy. Direct clients keep their socket address.
func Middleware(list *List) fiber.Handler {
	return func(c *fiber.Ctx) error {
		remote := c.Context().RemoteIP()
		if !list.Contains(remote) {
			return c.Next()
		}
		forwardedFor := c.Get(fiber.HeaderXForwardedFor)
		if ip, ok := clientIP(forwardedFor, c.Get("X-Real-IP"), list); ok {
			c.Request().Header.Set(fiber.HeaderXForwardedFor, ip)
		} else if forwardedFor != "" {
			c.Request().Header.Del(fiber.HeaderXForwardedFor)
		}
		return c.Next()
	}
}

// clientIP walks X-Forwarded-For from the proxy toward the client and returns
// the nearest address that is not itself a trusted proxy. When that header is
// absent, a non-proxy X-Real-IP is used.
func clientIP(forwardedFor, realIP string, list *List) (string, bool) {
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := net.ParseIP(strings.TrimSpace(parts[i]))
			if ip == nil {
				continue
			}
			if list.Contains(ip) {
				continue
			}
			return ip.String(), true
		}
		return "", false
	}

	ip := net.ParseIP(strings.TrimSpace(realIP))
	if ip == nil || list.Contains(ip) {
		return "", false
	}
	return ip.String(), true
}

func parseEntry(entry string) (*net.IPNet, error) {
	if strings.Contains(entry, "/") {
		_, network, err := net.ParseCIDR(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy %q: %w", entry, err)
		}
		return network, nil
	}

	ip := net.ParseIP(entry)
	if ip == nil {
		return nil, fmt.Errorf("invalid trusted proxy %q", entry)
	}
	if v4 := ip.To4(); v4 != nil {
		return &net.IPNet{IP: v4, Mask: net.CIDRMask(32, 32)}, nil
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}, nil
}
