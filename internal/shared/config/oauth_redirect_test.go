package config

import (
	"testing"
)

func TestLoadOAuthRedirectURIsFromEnv(t *testing.T) {
	t.Setenv("RADIUS_OAUTH_ALLOWEDREDIRECTURIS", "http://localhost:3001/auth/callback/google,http://localhost:3001/auth/callback/github")
	t.Setenv("RADIUS_DATABASE_USER", "u")
	t.Setenv("RADIUS_DATABASE_PASSWORD", "p")
	t.Setenv("RADIUS_DATABASE_NAME", "d")
	t.Setenv("RADIUS_JWT_SECRETKEY", "secret-key-at-least-32-chars-long!!")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.OAuth.AllowedRedirectURIs) != 2 {
		t.Fatalf("got %d uris: %#v", len(cfg.OAuth.AllowedRedirectURIs), cfg.OAuth.AllowedRedirectURIs)
	}
}
