package clientip

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestParseRejectsInvalidEntry(t *testing.T) {
	_, err := Parse("10.0.0.0/8,not-an-ip")
	require.Error(t, err)
}

func TestClientIPUsesNearestUntrustedAddress(t *testing.T) {
	list, err := Parse(DefaultTrustedProxies)
	require.NoError(t, err)

	ip, ok := clientIP("9.9.9.9, 203.0.113.5, 172.18.0.2", "", list)
	require.True(t, ok)
	require.Equal(t, "203.0.113.5", ip)
}

func TestClientIPIgnoresSpoofedHeaderFromUntrustedPeer(t *testing.T) {
	app := newIPApp(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderXForwardedFor, "203.0.113.5")

	body := do(t, app, req)
	require.Equal(t, "0.0.0.0", body)
}

func TestClientIPFromForwardedChain(t *testing.T) {
	app := newIPApp(t, "0.0.0.0/32,10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderXForwardedFor, "9.9.9.9, 203.0.113.5, 10.1.0.4")

	body := do(t, app, req)
	require.Equal(t, "203.0.113.5", body)
}

func TestClientIPFromRealIP(t *testing.T) {
	app := newIPApp(t, "0.0.0.0/32,10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.8")

	body := do(t, app, req)
	require.Equal(t, "203.0.113.8", body)
}

func TestClientIPKeepsSocketWhenHeaderIsOnlyProxies(t *testing.T) {
	app := newIPApp(t, "0.0.0.0/32,10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderXForwardedFor, "10.1.0.4")

	body := do(t, app, req)
	require.Equal(t, "0.0.0.0", body)
}

func TestResolveUsesCloudflareHeaderFromPrivateProxy(t *testing.T) {
	list, err := Parse(DefaultTrustedProxies)
	require.NoError(t, err)
	ip, ok := Resolve(net.ParseIP("10.0.0.5"), "203.0.113.9", "9.9.9.9", "", list)
	require.True(t, ok)
	require.Equal(t, "203.0.113.9", ip)
}

func TestResolveUsesCloudflareHeaderFromCloudflarePeer(t *testing.T) {
	list, err := Parse("127.0.0.1,::1")
	require.NoError(t, err)
	ip, ok := Resolve(net.ParseIP("104.16.1.1"), "203.0.113.9", "", "", list)
	require.True(t, ok)
	require.Equal(t, "203.0.113.9", ip)
}

func TestResolveIgnoresHeadersFromPublicPeer(t *testing.T) {
	list, err := Parse(DefaultTrustedProxies)
	require.NoError(t, err)
	_, ok := Resolve(net.ParseIP("198.51.100.7"), "203.0.113.9", "203.0.113.5", "", list)
	require.False(t, ok)
}

func newIPApp(t *testing.T, trusted string) *fiber.App {
	t.Helper()
	list, err := Parse(trusted)
	require.NoError(t, err)

	app := fiber.New(fiber.Config{
		ProxyHeader:             fiber.HeaderXForwardedFor,
		EnableTrustedProxyCheck: true,
		EnableIPValidation:      true,
		TrustedProxies:          list.FiberStrings(),
	})
	app.Use(Middleware(list))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(c.IP())
	})
	return app
}

func do(t *testing.T, app *fiber.App, req *http.Request) string {
	t.Helper()
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return string(body)
}
