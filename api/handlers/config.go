package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ConfigHandler handles configuration-related requests
type ConfigHandler struct {
	DB *sql.DB
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(db *sql.DB) *ConfigHandler {
	return &ConfigHandler{DB: db}
}

// UpdateRouterPriority updates the router priority for a resource
func (h *ConfigHandler) UpdateRouterPriority(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }
    
    var input struct {
        RouterPriority int `json:"router_priority" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }
    
    // Verify resource exists and is active
    var exists int
    var status string
    err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", id).Scan(&exists, &status)
    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, "Resource not found")
        return
    } else if err != nil {
        log.Printf("Error checking resource existence: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Don't allow updating disabled resources
    if status == "disabled" {
        ResponseWithError(c, http.StatusBadRequest, "Cannot update a disabled resource")
        return
    }
    
    // Update the resource within a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()
    
    log.Printf("Updating router priority for resource %s to %d", id, input.RouterPriority)
    
    result, txErr := tx.Exec(
        "UPDATE resources SET router_priority = ?, updated_at = ? WHERE id = ?",
        input.RouterPriority, time.Now(), id,
    )
    
    if txErr != nil {
        log.Printf("Error updating router priority: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to update router priority")
        return
    }
    
    rowsAffected, err := result.RowsAffected()
    if err == nil {
        log.Printf("Update affected %d rows", rowsAffected)
    }
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    log.Printf("Successfully updated router priority for resource %s", id)
    c.JSON(http.StatusOK, gin.H{
        "id": id,
        "router_priority": input.RouterPriority,
    })
}

// UpdateHTTPConfig updates the HTTP router entrypoints configuration
func (h *ConfigHandler) UpdateHTTPConfig(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }
    
    var input struct {
        Entrypoints string `json:"entrypoints"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }
    
    // Verify resource exists and is active
    var exists int
    var status string
    err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", id).Scan(&exists, &status)
    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, "Resource not found")
        return
    } else if err != nil {
        log.Printf("Error checking resource existence: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Don't allow updating disabled resources
    if status == "disabled" {
        ResponseWithError(c, http.StatusBadRequest, "Cannot update a disabled resource")
        return
    }
    
    // Validate entrypoints - should be comma-separated list
    if input.Entrypoints == "" {
        input.Entrypoints = "websecure" // Default
    }
    
    // Update the resource within a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()
    
    log.Printf("Updating HTTP entrypoints for resource %s: %s", id, input.Entrypoints)
    
    result, txErr := tx.Exec(
        "UPDATE resources SET entrypoints = ?, updated_at = ? WHERE id = ?",
        input.Entrypoints, time.Now(), id,
    )
    
    if txErr != nil {
        log.Printf("Error updating resource entrypoints: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to update resource")
        return
    }
    
    rowsAffected, err := result.RowsAffected()
    if err == nil {
        log.Printf("Update affected %d rows", rowsAffected)
        if rowsAffected == 0 {
            log.Printf("Warning: Update query succeeded but no rows were affected")
        }
    }
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    log.Printf("Successfully updated HTTP entrypoints for resource %s", id)
    c.JSON(http.StatusOK, gin.H{
        "id": id,
        "entrypoints": input.Entrypoints,
    })
}

// UpdateTLSConfig updates the TLS certificate domains configuration
func (h *ConfigHandler) UpdateTLSConfig(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }
    
    var input struct {
        TLSDomains string `json:"tls_domains"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }
    
    // Verify resource exists and is active
    var exists int
    var status string
    err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", id).Scan(&exists, &status)
    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, "Resource not found")
        return
    } else if err != nil {
        log.Printf("Error checking resource existence: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Don't allow updating disabled resources
    if status == "disabled" {
        ResponseWithError(c, http.StatusBadRequest, "Cannot update a disabled resource")
        return
    }
    
    // Update the resource within a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()
    
    log.Printf("Updating TLS domains for resource %s: %s", id, input.TLSDomains)
    
    result, txErr := tx.Exec(
        "UPDATE resources SET tls_domains = ?, updated_at = ? WHERE id = ?",
        input.TLSDomains, time.Now(), id,
    )
    
    if txErr != nil {
        log.Printf("Error updating TLS domains: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to update TLS domains")
        return
    }
    
    rowsAffected, err := result.RowsAffected()
    if err == nil {
        log.Printf("Update affected %d rows", rowsAffected)
        if rowsAffected == 0 {
            log.Printf("Warning: Update query succeeded but no rows were affected")
        }
    }
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    log.Printf("Successfully updated TLS domains for resource %s", id)
    c.JSON(http.StatusOK, gin.H{
        "id": id,
        "tls_domains": input.TLSDomains,
    })
}

// UpdateTCPConfig updates the TCP SNI router configuration
func (h *ConfigHandler) UpdateTCPConfig(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }
    
    var input struct {
        TCPEnabled     bool   `json:"tcp_enabled"`
        TCPEntrypoints string `json:"tcp_entrypoints"`
        TCPSNIRule     string `json:"tcp_sni_rule"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }
    
    // Verify resource exists and is active
    var exists int
    var status string
    err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", id).Scan(&exists, &status)
    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, "Resource not found")
        return
    } else if err != nil {
        log.Printf("Error checking resource existence: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Don't allow updating disabled resources
    if status == "disabled" {
        ResponseWithError(c, http.StatusBadRequest, "Cannot update a disabled resource")
        return
    }
    
    // Validate TCP entrypoints if provided
    if input.TCPEntrypoints == "" {
        input.TCPEntrypoints = "tcp" // Default
    }
    
    // Convert boolean to integer for SQLite
    tcpEnabled := 0
    if input.TCPEnabled {
        tcpEnabled = 1
    }
    
    // Update the resource within a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()
    
    log.Printf("Updating TCP config for resource %s: enabled=%t, entrypoints=%s", 
        id, input.TCPEnabled, input.TCPEntrypoints)
    
    result, txErr := tx.Exec(
        "UPDATE resources SET tcp_enabled = ?, tcp_entrypoints = ?, tcp_sni_rule = ?, updated_at = ? WHERE id = ?",
        tcpEnabled, input.TCPEntrypoints, input.TCPSNIRule, time.Now(), id,
    )
    
    if txErr != nil {
        log.Printf("Error updating TCP config: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to update TCP configuration")
        return
    }
    
    rowsAffected, err := result.RowsAffected()
    if err == nil {
        log.Printf("Update affected %d rows", rowsAffected)
        if rowsAffected == 0 {
            log.Printf("Warning: Update query succeeded but no rows were affected")
        }
    }
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    log.Printf("Successfully updated TCP configuration for resource %s", id)
    c.JSON(http.StatusOK, gin.H{
        "id":              id,
        "tcp_enabled":     input.TCPEnabled,
        "tcp_entrypoints": input.TCPEntrypoints,
        "tcp_sni_rule":    input.TCPSNIRule,
    })
}

// UpdateMTLSConfig updates the mTLS configuration for a resource
func (h *ConfigHandler) UpdateMTLSConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var input struct {
		MTLSEnabled bool `json:"mtls_enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Verify resource exists and is active
	var exists int
	var status string
	err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", id).Scan(&exists, &status)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Resource not found")
		return
	} else if err != nil {
		log.Printf("Error checking resource existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Don't allow updating disabled resources
	if status == "disabled" {
		ResponseWithError(c, http.StatusBadRequest, "Cannot update a disabled resource")
		return
	}

	// If enabling mTLS, check that global mTLS is enabled
	if input.MTLSEnabled {
		var globalEnabled int
		err := h.DB.QueryRow("SELECT enabled FROM mtls_config WHERE id = 1").Scan(&globalEnabled)
		if err != nil {
			log.Printf("Error checking global mTLS status: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to check mTLS status")
			return
		}
		if globalEnabled != 1 {
			ResponseWithError(c, http.StatusBadRequest, "Cannot enable mTLS on resource: global mTLS is not enabled")
			return
		}
	}

	// Convert boolean to integer for SQLite
	mtlsEnabled := 0
	if input.MTLSEnabled {
		mtlsEnabled = 1
	}

	// Update the resource within a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back due to error: %v", txErr)
		}
	}()

	log.Printf("Updating mTLS config for resource %s: enabled=%t", id, input.MTLSEnabled)

	result, txErr := tx.Exec(
		"UPDATE resources SET mtls_enabled = ?, updated_at = ? WHERE id = ?",
		mtlsEnabled, time.Now(), id,
	)

	if txErr != nil {
		log.Printf("Error updating mTLS config: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update mTLS configuration")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil {
		log.Printf("Update affected %d rows", rowsAffected)
		if rowsAffected == 0 {
			log.Printf("Warning: Update query succeeded but no rows were affected")
		}
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully updated mTLS configuration for resource %s", id)
	c.JSON(http.StatusOK, gin.H{
		"id":           id,
		"mtls_enabled": input.MTLSEnabled,
	})
}

// UpdateHeadersConfig updates the custom headers configuration
func (h *ConfigHandler) UpdateHeadersConfig(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }
    
    var input struct {
        CustomHeaders map[string]string `json:"custom_headers" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }
    
    // Verify resource exists and is active
    var exists int
    var status string
    err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", id).Scan(&exists, &status)
    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, "Resource not found")
        return
    } else if err != nil {
        log.Printf("Error checking resource existence: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Don't allow updating disabled resources
    if status == "disabled" {
        ResponseWithError(c, http.StatusBadRequest, "Cannot update a disabled resource")
        return
    }
    
    // Convert headers to JSON for storage
    headersJSON, err := json.Marshal(input.CustomHeaders)
    if err != nil {
        log.Printf("Error encoding headers: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to encode headers")
        return
    }
    
    // Update the resource within a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()
    
    log.Printf("Updating custom headers for resource %s with %d headers", 
        id, len(input.CustomHeaders))
    
    result, txErr := tx.Exec(
        "UPDATE resources SET custom_headers = ?, updated_at = ? WHERE id = ?",
        string(headersJSON), time.Now(), id,
    )
    
    if txErr != nil {
        log.Printf("Error updating custom headers: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to update custom headers")
        return
    }
    
    rowsAffected, err := result.RowsAffected()
    if err == nil {
        log.Printf("Update affected %d rows", rowsAffected)
        if rowsAffected == 0 {
            log.Printf("Warning: Update query succeeded but no rows were affected")
        }
    }
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Verify the update by reading back the custom_headers
    var storedHeaders string
    verifyErr := h.DB.QueryRow("SELECT custom_headers FROM resources WHERE id = ?", id).Scan(&storedHeaders)
    if verifyErr != nil {
        log.Printf("Warning: Could not verify headers update: %v", verifyErr)
    } else if storedHeaders == "" {
        log.Printf("Warning: Headers may be empty after update for resource %s", id)
    } else {
        log.Printf("Successfully verified headers update for resource %s", id)
    }
    
    log.Printf("Successfully updated custom headers for resource %s", id)
    c.JSON(http.StatusOK, gin.H{
        "id": id,
        "custom_headers": input.CustomHeaders,
    })
}