package database

import (
	"strings"
	"testing"
)

func TestParseLibSQLConnectionSeparatesAuthTokenFromURL(t *testing.T) {
	safeURL, authToken, err := parseLibSQLConnection("libsql://example.turso.io?authToken=secret-jwt&tls=1", "")
	if err != nil {
		t.Fatalf("parseLibSQLConnection returned error: %v", err)
	}

	if authToken != "secret-jwt" {
		t.Fatalf("expected auth token to be extracted, got %q", authToken)
	}
	if strings.Contains(safeURL, "secret-jwt") || strings.Contains(safeURL, "authToken") {
		t.Fatalf("expected safe URL to omit auth token, got %q", safeURL)
	}
	if !strings.Contains(safeURL, "tls=1") {
		t.Fatalf("expected non-secret query params to be preserved, got %q", safeURL)
	}
}

func TestParseLibSQLConnectionUsesFallbackAuthToken(t *testing.T) {
	safeURL, authToken, err := parseLibSQLConnection("libsql://example.turso.io", "separate-token")
	if err != nil {
		t.Fatalf("parseLibSQLConnection returned error: %v", err)
	}

	if authToken != "separate-token" {
		t.Fatalf("expected fallback auth token, got %q", authToken)
	}
	if strings.Contains(safeURL, "separate-token") {
		t.Fatalf("expected safe URL to omit fallback token, got %q", safeURL)
	}
}
