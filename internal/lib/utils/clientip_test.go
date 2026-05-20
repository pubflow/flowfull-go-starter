package utils

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestExtractClientIPUsesForwardedClient(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(ExtractClientIP(c))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 172.18.0.2")

	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "203.0.113.7", string(body))
}

func TestNormalizeIPHandlesForwardedHeader(t *testing.T) {
	assert.Equal(t, "203.0.113.7", NormalizeIP(`for="203.0.113.7:443";proto=https`))
	assert.Equal(t, "127.0.0.1", NormalizeIP("::1"))
}
