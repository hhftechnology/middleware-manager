package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResourceHandler handles resource-related requests
type ResourceHandler struct {
	DB *sql.DB
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(db *sql.DB) *ResourceHandler {
	return &ResourceHandler{DB: db}
}

// GetResources returns all resources and their assigned middlewares
// GetResources returns all resources and their assigned middlewares
func (h *ResourceHandler) GetResources(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT r.id, r.host, r.service_id, r.org_id, r.site_id, r.status, 
		       r.entrypoints, r.tls_domains, r.tcp_enabled, r.tcp_entrypoints, r.tcp_sni_rule,
		       r.custom_headers, r.router_priority, r.source_type,
		       GROUP_CONCAT(m.id || ':' || m.name || ':' || rm.priority, ',') as middlewares
		FROM resources r
		LEFT JOIN resource_middlewares rm ON r.id = rm.resource_id
		LEFT JOIN middlewares m ON rm.middleware_id = m.id
		GROUP BY r.id
	`)
	if err != nil {
		log.Printf("Error fetching resources: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch resources")
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}
	for rows.Next() {
		var id, host, serviceID, orgID, siteID, status, entrypoints, tlsDomains, tcpEntrypoints, tcpSNIRule, customHeaders, sourceType string
		var tcpEnabled int
		var routerPriority sql.NullInt64
		var middlewares sql.NullString
		
		// Fixed scan operation to match the exact order and number of columns in the query
		if err := rows.Scan(&id, &host, &serviceID, &orgID, &siteID, &status, 
				&entrypoints, &tlsDomains, &tcpEnabled, &tcpEntrypoints, &tcpSNIRule, 
				&customHeaders, &routerPriority, &sourceType, &middlewares); err != nil {
			log.Printf("Error scanning resource row: %v", err)
			continue
		}
		
		// Use default priority if null
		priority := 200 // Default value
		if routerPriority.Valid {
			priority = int(routerPriority.Int64)
		}
		
		resource := map[string]interface{}{
			"id":              id,
			"host":            host,
			"service_id":      serviceID,
			"org_id":          orgID,
			"site_id":         siteID,
			"status":          status,
			"entrypoints":     entrypoints,
			"tls_domains":     tlsDomains,
			"tcp_enabled":     tcpEnabled > 0,
			"tcp_entrypoints": tcpEntrypoints,
			"tcp_sni_rule":    tcpSNIRule,
			"custom_headers":  customHeaders,
			"router_priority": priority,
			"source_type":     sourceType, // Make sure this is included in the returned resource
		}
		
		if middlewares.Valid {
			resource["middlewares"] = middlewares.String
		} else {
			resource["middlewares"] = ""
		}
		
		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error during resource rows iteration: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch resources")
		return
	}

	c.JSON(http.StatusOK, resources)
}

// GetResource returns a specific resource
// GetResource returns a specific resource
func (h *ResourceHandler) GetResource(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }

    var host, serviceID, orgID, siteID, status, entrypoints, tlsDomains, tcpEntrypoints, tcpSNIRule, customHeaders, sourceType string
    var tcpEnabled int
    var routerPriority sql.NullInt64
    var middlewares sql.NullString

    err := h.DB.QueryRow(`
        SELECT r.host, r.service_id, r.org_id, r.site_id, r.status,
               r.entrypoints, r.tls_domains, r.tcp_enabled, r.tcp_entrypoints, r.tcp_sni_rule,
               r.custom_headers, r.router_priority, r.source_type,
               GROUP_CONCAT(m.id || ':' || m.name || ':' || rm.priority, ',') as middlewares
        FROM resources r
        LEFT JOIN resource_middlewares rm ON r.id = rm.resource_id
        LEFT JOIN middlewares m ON rm.middleware_id = m.id
        WHERE r.id = ?
        GROUP BY r.id
    `, id).Scan(&host, &serviceID, &orgID, &siteID, &status, 
            &entrypoints, &tlsDomains, &tcpEnabled, &tcpEntrypoints, &tcpSNIRule, 
            &customHeaders, &routerPriority, &sourceType, &middlewares)

    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, fmt.Sprintf("Resource not found: %s", id))
        return
    } else if err != nil {
        log.Printf("Error fetching resource: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch resource")
        return
    }
    
    // Use default priority if null
    priority := 200 // Default value
    if routerPriority.Valid {
        priority = int(routerPriority.Int64)
    }

    resource := map[string]interface{}{
        "id":              id,
        "host":            host,
        "service_id":      serviceID,
        "org_id":          orgID,
        "site_id":         siteID,
        "status":          status,
        "entrypoints":     entrypoints,
        "tls_domains":     tlsDomains,
        "tcp_enabled":     tcpEnabled > 0,
        "tcp_entrypoints": tcpEntrypoints,
        "tcp_sni_rule":    tcpSNIRule,
        "custom_headers":  customHeaders,
        "router_priority": priority,
        "source_type":     sourceType, // Make sure this is included
    }

    if middlewares.Valid {
        resource["middlewares"] = middlewares.String
    } else {
        resource["middlewares"] = ""
    }

    c.JSON(http.StatusOK, resource)
}

// DeleteResource deletes a resource from the database
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	// Check if resource exists and its status
	var status string
	err := h.DB.QueryRow("SELECT status FROM resources WHERE id = ?", id).Scan(&status)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Resource not found")
		return
	} else if err != nil {
		log.Printf("Error checking resource existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Only allow deletion of disabled resources
	if status != "disabled" {
		ResponseWithError(c, http.StatusBadRequest, "Only disabled resources can be deleted")
		return
	}

	// Delete the resource using a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	// If something goes wrong, rollback
	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back due to error: %v", txErr)
		}
	}()
	
	// First delete any middleware relationships
	log.Printf("Removing middleware relationships for resource %s", id)
	_, txErr = tx.Exec("DELETE FROM resource_middlewares WHERE resource_id = ?", id)
	if txErr != nil {
		log.Printf("Error removing resource middlewares: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete resource")
		return
	}
	
	// Then delete the resource
	log.Printf("Deleting resource %s", id)
	result, txErr := tx.Exec("DELETE FROM resources WHERE id = ?", id)
	if txErr != nil {
		log.Printf("Error deleting resource: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete resource")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	if rowsAffected == 0 {
		ResponseWithError(c, http.StatusNotFound, "Resource not found")
		return
	}
	
	log.Printf("Delete affected %d rows", rowsAffected)
	
	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully deleted resource %s", id)
	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
}

// AssignMiddleware assigns a middleware to a resource
func (h *ResourceHandler) AssignMiddleware(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var input struct {
		MiddlewareID string `json:"middleware_id" binding:"required"`
		Priority     int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Default priority is 200 if not specified
	if input.Priority <= 0 {
		input.Priority = 200
	}

	// Verify resource exists
	var exists int
	var status string
	err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", resourceID).Scan(&exists, &status)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Resource not found")
		return
	} else if err != nil {
		log.Printf("Error checking resource existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	// Don't allow attaching middlewares to disabled resources
	if status == "disabled" {
		ResponseWithError(c, http.StatusBadRequest, "Cannot assign middleware to a disabled resource")
		return
	}

	// Verify middleware exists
	err = h.DB.QueryRow("SELECT 1 FROM middlewares WHERE id = ?", input.MiddlewareID).Scan(&exists)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Middleware not found")
		return
	} else if err != nil {
		log.Printf("Error checking middleware existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Insert or update the resource middleware relationship using a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	// If something goes wrong, rollback
	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back due to error: %v", txErr)
		}
	}()
	
	// First delete any existing relationship
	log.Printf("Removing existing middleware relationship: resource=%s, middleware=%s",
		resourceID, input.MiddlewareID)
	_, txErr = tx.Exec(
		"DELETE FROM resource_middlewares WHERE resource_id = ? AND middleware_id = ?",
		resourceID, input.MiddlewareID,
	)
	if txErr != nil {
		log.Printf("Error removing existing relationship: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}
	
	// Then insert the new relationship
	log.Printf("Creating new middleware relationship: resource=%s, middleware=%s, priority=%d",
		resourceID, input.MiddlewareID, input.Priority)
	result, txErr := tx.Exec(
		"INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)",
		resourceID, input.MiddlewareID, input.Priority,
	)
	if txErr != nil {
		log.Printf("Error assigning middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to assign middleware")
		return
	}
	
	rowsAffected, err := result.RowsAffected()
	if err == nil {
		log.Printf("Insert affected %d rows", rowsAffected)
	}
	
	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully assigned middleware %s to resource %s with priority %d",
		input.MiddlewareID, resourceID, input.Priority)
	c.JSON(http.StatusOK, gin.H{
		"resource_id":   resourceID,
		"middleware_id": input.MiddlewareID,
		"priority":      input.Priority,
	})
}

// AssignMultipleMiddlewares assigns multiple middlewares to a resource in one operation
func (h *ResourceHandler) AssignMultipleMiddlewares(c *gin.Context) {
    resourceID := c.Param("id")
    if resourceID == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
        return
    }

    var input struct {
        Middlewares []struct {
            MiddlewareID string `json:"middleware_id" binding:"required"`
            Priority     int    `json:"priority"`
        } `json:"middlewares" binding:"required"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }

    // Verify resource exists and is active
    var exists int
    var status string
    err := h.DB.QueryRow("SELECT 1, status FROM resources WHERE id = ?", resourceID).Scan(&exists, &status)
    if err == sql.ErrNoRows {
        ResponseWithError(c, http.StatusNotFound, "Resource not found")
        return
    } else if err != nil {
        log.Printf("Error checking resource existence: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // Don't allow attaching middlewares to disabled resources
    if status == "disabled" {
        ResponseWithError(c, http.StatusBadRequest, "Cannot assign middlewares to a disabled resource")
        return
    }

    // Start a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // If something goes wrong, rollback
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()

    // Process each middleware
    successful := make([]map[string]interface{}, 0)
    log.Printf("Assigning %d middlewares to resource %s", len(input.Middlewares), resourceID)
    
    for _, mw := range input.Middlewares {
        // Default priority is 200 if not specified
        if mw.Priority <= 0 {
            mw.Priority = 200
        }

        // Verify middleware exists
        var middlewareExists int
        err := h.DB.QueryRow("SELECT 1 FROM middlewares WHERE id = ?", mw.MiddlewareID).Scan(&middlewareExists)
        if err == sql.ErrNoRows {
            // Skip this middleware but don't fail the entire request
            log.Printf("Middleware %s not found, skipping", mw.MiddlewareID)
            continue
        } else if err != nil {
            log.Printf("Error checking middleware existence: %v", err)
            ResponseWithError(c, http.StatusInternalServerError, "Database error")
            return
        }

        // First delete any existing relationship
        log.Printf("Removing existing relationship: resource=%s, middleware=%s",
            resourceID, mw.MiddlewareID)
        _, txErr = tx.Exec(
            "DELETE FROM resource_middlewares WHERE resource_id = ? AND middleware_id = ?",
            resourceID, mw.MiddlewareID,
        )
        if txErr != nil {
            log.Printf("Error removing existing relationship: %v", txErr)
            ResponseWithError(c, http.StatusInternalServerError, "Database error")
            return
        }
        
        // Then insert the new relationship
        log.Printf("Creating new relationship: resource=%s, middleware=%s, priority=%d",
            resourceID, mw.MiddlewareID, mw.Priority)
        result, txErr := tx.Exec(
            "INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)",
            resourceID, mw.MiddlewareID, mw.Priority,
        )
        if txErr != nil {
            log.Printf("Error assigning middleware: %v", txErr)
            ResponseWithError(c, http.StatusInternalServerError, "Failed to assign middleware")
            return
        }
        
        rowsAffected, err := result.RowsAffected()
        if err == nil && rowsAffected > 0 {
            log.Printf("Successfully assigned middleware %s with priority %d", 
                mw.MiddlewareID, mw.Priority)
            successful = append(successful, map[string]interface{}{
                "middleware_id": mw.MiddlewareID,
                "priority": mw.Priority,
            })
        } else {
            log.Printf("Warning: Insertion query succeeded but affected %d rows", rowsAffected)
        }
    }
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }

    log.Printf("Successfully assigned %d middlewares to resource %s", len(successful), resourceID)
    c.JSON(http.StatusOK, gin.H{
        "resource_id": resourceID,
        "middlewares": successful,
    })
}

// RemoveMiddleware removes a middleware from a resource
func (h *ResourceHandler) RemoveMiddleware(c *gin.Context) {
    resourceID := c.Param("id")
    middlewareID := c.Param("middlewareId")
    
    if resourceID == "" || middlewareID == "" {
        ResponseWithError(c, http.StatusBadRequest, "Resource ID and Middleware ID are required")
        return
    }

    log.Printf("Removing middleware %s from resource %s", middlewareID, resourceID)

    // Delete the relationship using a transaction
    tx, err := h.DB.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    // If something goes wrong, rollback
    var txErr error
    defer func() {
        if txErr != nil {
            tx.Rollback()
            log.Printf("Transaction rolled back due to error: %v", txErr)
        }
    }()
    
    result, txErr := tx.Exec(
        "DELETE FROM resource_middlewares WHERE resource_id = ? AND middleware_id = ?",
        resourceID, middlewareID,
    )
    
    if txErr != nil {
        log.Printf("Error removing middleware: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Failed to remove middleware")
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Printf("Error getting rows affected: %v", err)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }
    
    if rowsAffected == 0 {
        log.Printf("No relationship found between resource %s and middleware %s", resourceID, middlewareID)
        ResponseWithError(c, http.StatusNotFound, "Resource middleware relationship not found")
        return
    }
    
    log.Printf("Delete affected %d rows", rowsAffected)
    
    // Commit the transaction
    if txErr = tx.Commit(); txErr != nil {
        log.Printf("Error committing transaction: %v", txErr)
        ResponseWithError(c, http.StatusInternalServerError, "Database error")
        return
    }

    log.Printf("Successfully removed middleware %s from resource %s", middlewareID, resourceID)
    c.JSON(http.StatusOK, gin.H{"message": "Middleware removed from resource successfully"})
}