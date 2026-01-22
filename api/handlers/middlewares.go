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

// MiddlewareHandler handles middleware-related requests
type MiddlewareHandler struct {
	DB *sql.DB
}

// NewMiddlewareHandler creates a new middleware handler
func NewMiddlewareHandler(db *sql.DB) *MiddlewareHandler {
	return &MiddlewareHandler{DB: db}
}

// GetMiddlewares returns all middleware configurations
// Supports pagination via ?page=N&page_size=M query parameters
func (h *MiddlewareHandler) GetMiddlewares(c *gin.Context) {
	usePagination := IsPaginationRequested(c)
	params := GetPaginationParams(c)

	var total int
	if usePagination {
		err := h.DB.QueryRow("SELECT COUNT(*) FROM middlewares").Scan(&total)
		if err != nil {
			log.Printf("Error counting middlewares: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to count middlewares")
			return
		}
	}

	query := "SELECT id, name, type, config FROM middlewares ORDER BY name"
	var rows *sql.Rows
	var err error

	if usePagination {
		query += " LIMIT ? OFFSET ?"
		rows, err = h.DB.Query(query, params.PageSize, params.Offset)
	} else {
		rows, err = h.DB.Query(query)
	}

	if err != nil {
		log.Printf("Error fetching middlewares: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch middlewares")
		return
	}
	defer rows.Close()

	middlewares := []map[string]interface{}{}
	for rows.Next() {
		var id, name, typ, configStr string
		if err := rows.Scan(&id, &name, &typ, &configStr); err != nil {
			log.Printf("Error scanning middleware row: %v", err)
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			log.Printf("Error parsing middleware config: %v", err)
			config = map[string]interface{}{}
		}

		middlewares = append(middlewares, map[string]interface{}{
			"id":     id,
			"name":   name,
			"type":   typ,
			"config": config,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating middleware rows: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error while fetching middlewares")
		return
	}

	if usePagination {
		c.JSON(http.StatusOK, NewPaginatedResponse(middlewares, total, params))
	} else {
		c.JSON(http.StatusOK, middlewares)
	}
}

// CreateMiddleware creates a new middleware configuration
func (h *MiddlewareHandler) CreateMiddleware(c *gin.Context) {
	var middleware struct {
		Name   string                 `json:"name" binding:"required"`
		Type   string                 `json:"type" binding:"required"`
		Config map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&middleware); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Validate middleware type
	if !isValidMiddlewareType(middleware.Type) {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid middleware type: %s", middleware.Type))
		return
	}

	// Generate a unique ID
	id, err := generateID()
	if err != nil {
		log.Printf("Error generating ID: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to generate ID")
		return
	}

	// Convert config to JSON string
	configJSON, err := json.Marshal(middleware.Config)
	if err != nil {
		log.Printf("Error encoding config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to encode config")
		return
	}

	// Insert into database using a transaction
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
	
	log.Printf("Attempting to insert middleware with ID=%s, name=%s, type=%s", 
		id, middleware.Name, middleware.Type)
	
	result, txErr := tx.Exec(
		"INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)",
		id, middleware.Name, middleware.Type, string(configJSON),
	)
	
	if txErr != nil {
		log.Printf("Error inserting middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to save middleware")
		return
	}
	
	rowsAffected, err := result.RowsAffected()
	if err == nil {
		log.Printf("Insert affected %d rows", rowsAffected)
	}

	// Remove from deleted_templates if it was previously deleted (user is re-creating it)
	_, _ = tx.Exec("DELETE FROM deleted_templates WHERE id = ? AND type = 'middleware'", id)

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully created middleware %s (%s)", middleware.Name, id)
	c.JSON(http.StatusCreated, gin.H{
		"id":     id,
		"name":   middleware.Name,
		"type":   middleware.Type,
		"config": middleware.Config,
	})
}

// GetMiddleware returns a specific middleware configuration
func (h *MiddlewareHandler) GetMiddleware(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Middleware ID is required")
		return
	}

	var name, typ, configStr string
	err := h.DB.QueryRow("SELECT name, type, config FROM middlewares WHERE id = ?", id).Scan(&name, &typ, &configStr)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Middleware not found")
		return
	} else if err != nil {
		log.Printf("Error fetching middleware: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch middleware")
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		log.Printf("Error parsing middleware config: %v", err)
		config = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"name":   name,
		"type":   typ,
		"config": config,
	})
}

// UpdateMiddleware updates a middleware configuration
func (h *MiddlewareHandler) UpdateMiddleware(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Middleware ID is required")
		return
	}

	var middleware struct {
		Name   string                 `json:"name" binding:"required"`
		Type   string                 `json:"type" binding:"required"`
		Config map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&middleware); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Validate middleware type
	if !isValidMiddlewareType(middleware.Type) {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid middleware type: %s", middleware.Type))
		return
	}

	// Check if middleware exists
	var exists int
	err := h.DB.QueryRow("SELECT 1 FROM middlewares WHERE id = ?", id).Scan(&exists)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Middleware not found")
		return
	} else if err != nil {
		log.Printf("Error checking middleware existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Convert config to JSON string
	configJSON, err := json.Marshal(middleware.Config)
	if err != nil {
		log.Printf("Error encoding config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to encode config")
		return
	}

	// Update in database using a transaction
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
	
	log.Printf("Attempting to update middleware %s with name=%s, type=%s", 
		id, middleware.Name, middleware.Type)
	
	result, txErr := tx.Exec(
		"UPDATE middlewares SET name = ?, type = ?, config = ?, updated_at = ? WHERE id = ?",
		middleware.Name, middleware.Type, string(configJSON), time.Now(), id,
	)
	
	if txErr != nil {
		log.Printf("Error updating middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update middleware")
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

	// Double-check that the middleware was updated
	var updatedName string
	err = h.DB.QueryRow("SELECT name FROM middlewares WHERE id = ?", id).Scan(&updatedName)
	if err != nil {
		log.Printf("Warning: Could not verify middleware update: %v", err)
	} else if updatedName != middleware.Name {
		log.Printf("Warning: Name mismatch after update. Expected '%s', got '%s'", middleware.Name, updatedName)
	} else {
		log.Printf("Successfully verified middleware update for %s", id)
	}

	// Return the updated middleware
	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"name":   middleware.Name,
		"type":   middleware.Type,
		"config": middleware.Config,
	})
}

// DeleteMiddleware deletes a middleware configuration
func (h *MiddlewareHandler) DeleteMiddleware(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Middleware ID is required")
		return
	}

	// Check for dependencies first
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM resource_middlewares WHERE middleware_id = ?", id).Scan(&count)
	if err != nil {
		log.Printf("Error checking middleware dependencies: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if count > 0 {
		ResponseWithError(c, http.StatusConflict, fmt.Sprintf("Cannot delete middleware because it is used by %d resources", count))
		return
	}

	// Delete from database using a transaction
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
	
	log.Printf("Attempting to delete middleware %s", id)

	result, txErr := tx.Exec("DELETE FROM middlewares WHERE id = ?", id)
	if txErr != nil {
		log.Printf("Error deleting middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete middleware")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if rowsAffected == 0 {
		ResponseWithError(c, http.StatusNotFound, "Middleware not found")
		return
	}

	// Track deletion to prevent template from being re-created on restart
	_, txErr = tx.Exec("INSERT OR REPLACE INTO deleted_templates (id, type) VALUES (?, 'middleware')", id)
	if txErr != nil {
		log.Printf("Warning: Failed to track deleted template: %v", txErr)
		// Continue anyway - this is not critical
	}

	log.Printf("Delete affected %d rows", rowsAffected)

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully deleted middleware %s", id)
	c.JSON(http.StatusOK, gin.H{"message": "Middleware deleted successfully"})
}