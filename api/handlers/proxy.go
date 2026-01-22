package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/services"
)

// ProxyHandler handles the config proxy endpoint for Traefik
type ProxyHandler struct {
	ConfigProxy *services.ConfigProxy
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(configProxy *services.ConfigProxy) *ProxyHandler {
	return &ProxyHandler{
		ConfigProxy: configProxy,
	}
}

// GetTraefikConfig returns merged Pangolin + MW-manager configuration
// This endpoint is designed to be used by Traefik's HTTP provider
// GET /api/traefik-config
func (h *ProxyHandler) GetTraefikConfig(c *gin.Context) {
	config, err := h.ConfigProxy.GetMergedConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get Traefik configuration",
			"details": err.Error(),
		})
		return
	}

	// Return the merged configuration
	// Traefik expects a JSON response with http, tcp, udp, tls sections
	c.JSON(http.StatusOK, config)
}

// InvalidateCache forces the proxy to fetch fresh configuration
// POST /api/traefik-config/invalidate
func (h *ProxyHandler) InvalidateCache(c *gin.Context) {
	h.ConfigProxy.InvalidateCache()
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache invalidated successfully",
	})
}

// GetProxyStatus returns the current status of the config proxy
// GET /api/traefik-config/status
func (h *ProxyHandler) GetProxyStatus(c *gin.Context) {
	// Try to get config to check if everything is working
	_, err := h.ConfigProxy.GetMergedConfig()

	status := "healthy"
	var errorMsg string
	if err != nil {
		status = "unhealthy"
		errorMsg = err.Error()
	}

	response := gin.H{
		"status":  status,
		"message": "Config proxy is operational",
	}

	if errorMsg != "" {
		response["error"] = errorMsg
	}

	c.JSON(http.StatusOK, response)
}
