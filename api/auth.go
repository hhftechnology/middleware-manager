package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig is optional HTTP basic auth for the admin UI and API.
// It is disabled unless MM_AUTH_ENABLED is set, so existing deployments
// that rely on network isolation keep working.
type AuthConfig struct {
	Enabled  bool
	Username string
	Token    string
}

// AuthConfigFromEnv reads MM_AUTH_ENABLED, MM_AUTH_USER, and MM_AUTH_TOKEN.
// Enabling auth without both a username and a token is a configuration error.
func AuthConfigFromEnv() (AuthConfig, error) {
	raw := strings.TrimSpace(os.Getenv("MM_AUTH_ENABLED"))
	enabled := strings.EqualFold(raw, "true") || raw == "1"
	username := os.Getenv("MM_AUTH_USER")
	token := os.Getenv("MM_AUTH_TOKEN")
	if !enabled {
		return AuthConfig{}, nil
	}
	if strings.TrimSpace(username) == "" || token == "" {
		return AuthConfig{}, fmt.Errorf("MM_AUTH_ENABLED requires MM_AUTH_USER and MM_AUTH_TOKEN")
	}
	return AuthConfig{Enabled: true, Username: username, Token: token}, nil
}

// basicAuthMiddleware requires HTTP basic auth on every route except /health.
// The container healthcheck calls /health without credentials. Traefik's HTTP
// provider must send the same Authorization header when it polls
// /api/v1/traefik-config.
func basicAuthMiddleware(username, token string) gin.HandlerFunc {
	inner := gin.BasicAuth(gin.Accounts{username: token})
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		inner(c)
	}
}
