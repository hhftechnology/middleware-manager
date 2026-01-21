package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/services"
)

// SecurityHandler handles security-related requests
type SecurityHandler struct {
	DB                *sql.DB
	DuplicateDetector *services.DuplicateDetector
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(db *sql.DB, configManager *services.ConfigManager) *SecurityHandler {
	return &SecurityHandler{
		DB:                db,
		DuplicateDetector: services.NewDuplicateDetector(configManager),
	}
}

// GetConfig returns the current security configuration
func (h *SecurityHandler) GetConfig(c *gin.Context) {
	var config models.SecurityConfig
	var tlsHardeningEnabled, secureHeadersEnabled int

	err := h.DB.QueryRow(`
		SELECT id, tls_hardening_enabled, secure_headers_enabled,
		       secure_headers_x_content_type_options, secure_headers_x_frame_options,
		       secure_headers_x_xss_protection, secure_headers_hsts,
		       secure_headers_referrer_policy, secure_headers_csp,
		       secure_headers_permissions_policy, created_at, updated_at
		FROM security_config WHERE id = 1
	`).Scan(
		&config.ID, &tlsHardeningEnabled, &secureHeadersEnabled,
		&config.SecureHeaders.XContentTypeOptions, &config.SecureHeaders.XFrameOptions,
		&config.SecureHeaders.XXSSProtection, &config.SecureHeaders.HSTS,
		&config.SecureHeaders.ReferrerPolicy, &config.SecureHeaders.CSP,
		&config.SecureHeaders.PermissionsPolicy, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return defaults
			config = models.SecurityConfig{
				ID:                   1,
				TLSHardeningEnabled:  false,
				SecureHeadersEnabled: false,
				SecureHeaders:        models.DefaultSecureHeaders(),
			}
		} else {
			log.Printf("Error getting security config: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to get security configuration")
			return
		}
	} else {
		config.TLSHardeningEnabled = tlsHardeningEnabled == 1
		config.SecureHeadersEnabled = secureHeadersEnabled == 1
	}

	c.JSON(http.StatusOK, config)
}

// EnableTLSHardening enables TLS hardening globally
func (h *SecurityHandler) EnableTLSHardening(c *gin.Context) {
	_, err := h.DB.Exec(`
		UPDATE security_config SET tls_hardening_enabled = 1, updated_at = CURRENT_TIMESTAMP WHERE id = 1
	`)
	if err != nil {
		log.Printf("Error enabling TLS hardening: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to enable TLS hardening")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TLS hardening enabled",
		"enabled": true,
	})
}

// DisableTLSHardening disables TLS hardening globally
func (h *SecurityHandler) DisableTLSHardening(c *gin.Context) {
	_, err := h.DB.Exec(`
		UPDATE security_config SET tls_hardening_enabled = 0, updated_at = CURRENT_TIMESTAMP WHERE id = 1
	`)
	if err != nil {
		log.Printf("Error disabling TLS hardening: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to disable TLS hardening")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TLS hardening disabled",
		"enabled": false,
	})
}

// EnableSecureHeaders enables secure headers globally
func (h *SecurityHandler) EnableSecureHeaders(c *gin.Context) {
	_, err := h.DB.Exec(`
		UPDATE security_config SET secure_headers_enabled = 1, updated_at = CURRENT_TIMESTAMP WHERE id = 1
	`)
	if err != nil {
		log.Printf("Error enabling secure headers: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to enable secure headers")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Secure headers enabled",
		"enabled": true,
	})
}

// DisableSecureHeaders disables secure headers globally
func (h *SecurityHandler) DisableSecureHeaders(c *gin.Context) {
	_, err := h.DB.Exec(`
		UPDATE security_config SET secure_headers_enabled = 0, updated_at = CURRENT_TIMESTAMP WHERE id = 1
	`)
	if err != nil {
		log.Printf("Error disabling secure headers: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to disable secure headers")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Secure headers disabled",
		"enabled": false,
	})
}

// UpdateSecureHeadersConfig updates the secure headers configuration
func (h *SecurityHandler) UpdateSecureHeadersConfig(c *gin.Context) {
	var input models.SecureHeadersConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	_, err := h.DB.Exec(`
		UPDATE security_config SET
		       secure_headers_x_content_type_options = ?,
		       secure_headers_x_frame_options = ?,
		       secure_headers_x_xss_protection = ?,
		       secure_headers_hsts = ?,
		       secure_headers_referrer_policy = ?,
		       secure_headers_csp = ?,
		       secure_headers_permissions_policy = ?,
		       updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, input.XContentTypeOptions, input.XFrameOptions, input.XXSSProtection,
		input.HSTS, input.ReferrerPolicy, input.CSP, input.PermissionsPolicy)

	if err != nil {
		log.Printf("Error updating secure headers config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update secure headers configuration")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Secure headers configuration updated",
	})
}

// CheckMiddlewareDuplicates checks if a middleware name conflicts with existing Traefik middlewares
func (h *SecurityHandler) CheckMiddlewareDuplicates(c *gin.Context) {
	var req models.DuplicateCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	result := h.DuplicateDetector.CheckDuplicates(req.Name, req.PluginName)

	c.JSON(http.StatusOK, result)
}

// UpdateResourceTLSHardening updates TLS hardening for a specific resource
func (h *SecurityHandler) UpdateResourceTLSHardening(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var input models.UpdateResourceSecurityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Check if mTLS is enabled for this resource - TLS hardening should be disabled when mTLS is active
	var mtlsEnabled int
	err := h.DB.QueryRow("SELECT COALESCE(mtls_enabled, 0) FROM resources WHERE id = ?", resourceID).Scan(&mtlsEnabled)
	if err != nil {
		if err == sql.ErrNoRows {
			ResponseWithError(c, http.StatusNotFound, "Resource not found")
			return
		}
		log.Printf("Error checking resource mTLS status: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to check resource")
		return
	}

	if mtlsEnabled == 1 && input.Enabled {
		ResponseWithError(c, http.StatusBadRequest, "Cannot enable TLS hardening when mTLS is active. mTLS already includes TLS hardening.")
		return
	}

	enabledVal := 0
	if input.Enabled {
		enabledVal = 1
	}

	_, err = h.DB.Exec(`
		UPDATE resources SET tls_hardening_enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, enabledVal, resourceID)

	if err != nil {
		log.Printf("Error updating resource TLS hardening: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update TLS hardening")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "TLS hardening updated",
		"resource_id":          resourceID,
		"tls_hardening_enabled": input.Enabled,
	})
}

// UpdateResourceSecureHeaders updates secure headers for a specific resource
func (h *SecurityHandler) UpdateResourceSecureHeaders(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var input models.UpdateResourceSecurityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Check if global secure headers is enabled
	var globalEnabled int
	err := h.DB.QueryRow("SELECT COALESCE(secure_headers_enabled, 0) FROM security_config WHERE id = 1").Scan(&globalEnabled)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking global secure headers status: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to check global configuration")
		return
	}

	if globalEnabled == 0 && input.Enabled {
		ResponseWithError(c, http.StatusBadRequest, "Enable secure headers globally first before enabling per-resource")
		return
	}

	enabledVal := 0
	if input.Enabled {
		enabledVal = 1
	}

	_, err = h.DB.Exec(`
		UPDATE resources SET secure_headers_enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, enabledVal, resourceID)

	if err != nil {
		log.Printf("Error updating resource secure headers: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update secure headers")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                "Secure headers updated",
		"resource_id":            resourceID,
		"secure_headers_enabled": input.Enabled,
	})
}
