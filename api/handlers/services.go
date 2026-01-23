package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/util"
)

// ServiceHandler handles service-related requests
type ServiceHandler struct {
	DB *sql.DB
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(db *sql.DB) *ServiceHandler {
	return &ServiceHandler{DB: db}
}

// GetServices returns all service configurations
// Supports pagination via ?page=N&page_size=M query parameters
// By default only returns active services; use ?status=all to include disabled
func (h *ServiceHandler) GetServices(c *gin.Context) {
	usePagination := IsPaginationRequested(c)
	params := GetPaginationParams(c)

	// Filter by status - default to active only
	statusFilter := c.DefaultQuery("status", "active")
	statusCondition := "WHERE status = 'active'"
	if statusFilter == "all" {
		statusCondition = ""
	} else if statusFilter == "disabled" {
		statusCondition = "WHERE status = 'disabled'"
	}

	var total int
	if usePagination {
		countQuery := "SELECT COUNT(*) FROM services " + statusCondition
		err := h.DB.QueryRow(countQuery).Scan(&total)
		if err != nil {
			log.Printf("Error counting services: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to count services")
			return
		}
	}

	query := "SELECT id, name, type, config, COALESCE(status, 'active') as status, COALESCE(source_type, '') as source_type FROM services " + statusCondition + " ORDER BY name"
	var rows *sql.Rows
	var err error

	if usePagination {
		query += " LIMIT ? OFFSET ?"
		rows, err = h.DB.Query(query, params.PageSize, params.Offset)
	} else {
		rows, err = h.DB.Query(query)
	}

	if err != nil {
		log.Printf("Error fetching services: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch services")
		return
	}
	defer rows.Close()

	services := []map[string]interface{}{}
	for rows.Next() {
		var id, name, typ, configStr, status, sourceType string
		if err := rows.Scan(&id, &name, &typ, &configStr, &status, &sourceType); err != nil {
			log.Printf("Error scanning service row: %v", err)
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			log.Printf("Error parsing service config: %v", err)
			config = map[string]interface{}{}
		}

		services = append(services, map[string]interface{}{
			"id":          id,
			"name":        name,
			"type":        typ,
			"config":      config,
			"status":      status,
			"source_type": sourceType,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating service rows: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error while fetching services")
		return
	}

	if usePagination {
		c.JSON(http.StatusOK, NewPaginatedResponse(services, total, params))
	} else {
		c.JSON(http.StatusOK, services)
	}
}

// CreateService creates a new service configuration
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var service struct {
		Name   string                 `json:"name" binding:"required"`
		Type   string                 `json:"type" binding:"required"`
		Config map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&service); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Validate service type
	if !models.IsValidServiceType(service.Type) {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid service type: %s", service.Type))
		return
	}

	// Generate a unique ID
	id, err := generateID()
	if err != nil {
		log.Printf("Error generating ID: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to generate ID")
		return
	}

	// Process the service configuration based on the type
	service.Config = models.ProcessServiceConfig(service.Type, service.Config)

	// Convert config to JSON string
	configJSON, err := json.Marshal(service.Config)
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

	log.Printf("Attempting to insert service with ID=%s, name=%s, type=%s",
		id, service.Name, service.Type)

	result, txErr := tx.Exec(
		"INSERT INTO services (id, name, type, config, status, source_type) VALUES (?, ?, ?, ?, 'active', 'manual')",
		id, service.Name, service.Type, string(configJSON),
	)

	if txErr != nil {
		log.Printf("Error inserting service: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to save service")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil {
		log.Printf("Insert affected %d rows", rowsAffected)
	}

	// Remove from deleted_templates if it was previously deleted (user is re-creating it)
	_, _ = tx.Exec("DELETE FROM deleted_templates WHERE id = ? AND type = 'service'", id)

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully created service %s (%s)", service.Name, id)
	c.JSON(http.StatusCreated, gin.H{
		"id":     id,
		"name":   service.Name,
		"type":   service.Type,
		"config": service.Config,
	})
}

// GetService returns a specific service configuration
func (h *ServiceHandler) GetService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Service ID is required")
		return
	}

	rec, err := h.findServiceByID(id)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Service not found")
		return
	} else if err != nil {
		log.Printf("Error fetching service: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch service")
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(rec.Config), &config); err != nil {
		log.Printf("Error parsing service config: %v", err)
		config = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          rec.ID,
		"name":        rec.Name,
		"type":        rec.Type,
		"config":      config,
		"status":      rec.Status,
		"source_type": rec.SourceType,
	})
}

// UpdateService updates a service configuration
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Service ID is required")
		return
	}

	var service struct {
		Name   string                 `json:"name" binding:"required"`
		Type   string                 `json:"type" binding:"required"`
		Config map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&service); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Validate service type
	if !models.IsValidServiceType(service.Type) {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid service type: %s", service.Type))
		return
	}

	rec, err := h.findServiceByID(id)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Service not found")
		return
	} else if err != nil {
		log.Printf("Error checking service existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Process the service configuration based on the type
	service.Config = models.ProcessServiceConfig(service.Type, service.Config)

	// Convert config to JSON string
	configJSON, err := json.Marshal(service.Config)
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

	log.Printf("Attempting to update service %s with name=%s, type=%s",
		id, service.Name, service.Type)

	result, txErr := tx.Exec(
		"UPDATE services SET name = ?, type = ?, config = ?, updated_at = ? WHERE id = ?",
		service.Name, service.Type, string(configJSON), time.Now(), rec.ID,
	)

	if txErr != nil {
		log.Printf("Error updating service: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update service")
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

	// Double-check that the service was updated
	var updatedName string
	err = h.DB.QueryRow("SELECT name FROM services WHERE id = ?", rec.ID).Scan(&updatedName)
	if err != nil {
		log.Printf("Warning: Could not verify service update: %v", err)
	} else if updatedName != service.Name {
		log.Printf("Warning: Name mismatch after update. Expected '%s', got '%s'", service.Name, updatedName)
	} else {
		log.Printf("Successfully verified service update for %s", rec.ID)
	}

	// Return the updated service
	c.JSON(http.StatusOK, gin.H{
		"id":     rec.ID,
		"name":   service.Name,
		"type":   service.Type,
		"config": service.Config,
	})
}

// DeleteService deletes a service configuration
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Service ID is required")
		return
	}

	rec, err := h.findServiceByID(id)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Service not found")
		return
	} else if err != nil {
		log.Printf("Error fetching service for delete: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Check for dependencies first - resources using this service
	var count int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM resource_services WHERE service_id = ?", rec.ID).Scan(&count)
	if err != nil {
		log.Printf("Error checking service dependencies: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if count > 0 {
		ResponseWithError(c, http.StatusConflict, fmt.Sprintf("Cannot delete service because it is used by %d resources", count))
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

	log.Printf("Attempting to delete service %s", rec.ID)

	result, txErr := tx.Exec("DELETE FROM services WHERE id = ?", rec.ID)
	if txErr != nil {
		log.Printf("Error deleting service: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete service")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if rowsAffected == 0 {
		ResponseWithError(c, http.StatusNotFound, "Service not found")
		return
	}

	// Track deletion to prevent template from being re-created on restart
	_, txErr = tx.Exec("INSERT OR REPLACE INTO deleted_templates (id, type) VALUES (?, 'service')", rec.ID)
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

	log.Printf("Successfully deleted service %s", rec.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

// AssignServiceToResource assigns a service to a resource
func (h *ServiceHandler) AssignServiceToResource(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var input struct {
		ServiceID string `json:"service_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
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

	// Don't allow attaching services to disabled resources
	if status == "disabled" {
		ResponseWithError(c, http.StatusBadRequest, "Cannot assign service to a disabled resource")
		return
	}

	// Verify service exists (supports normalized IDs / provider suffixes)
	serviceRec, err := h.findServiceByID(input.ServiceID)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Service not found")
		return
	} else if err != nil {
		log.Printf("Error checking service existence: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Insert or update the resource service relationship using a transaction
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
	log.Printf("Removing existing service relationship: resource=%s", resourceID)
	_, txErr = tx.Exec(
		"DELETE FROM resource_services WHERE resource_id = ?",
		resourceID,
	)
	if txErr != nil {
		log.Printf("Error removing existing relationship: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Then insert the new relationship
	log.Printf("Creating new service relationship: resource=%s, service=%s",
		resourceID, serviceRec.ID)
	result, txErr := tx.Exec(
		"INSERT INTO resource_services (resource_id, service_id) VALUES (?, ?)",
		resourceID, serviceRec.ID,
	)

	if txErr != nil {
		log.Printf("Error assigning service: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to assign service")
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

	log.Printf("Successfully assigned service %s to resource %s",
		serviceRec.ID, resourceID)
	c.JSON(http.StatusOK, gin.H{
		"resource_id": resourceID,
		"service_id":  serviceRec.ID,
	})
}

// RemoveServiceFromResource removes a service from a resource
func (h *ServiceHandler) RemoveServiceFromResource(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	log.Printf("Removing service from resource %s", resourceID)

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
		"DELETE FROM resource_services WHERE resource_id = ?",
		resourceID,
	)

	if txErr != nil {
		log.Printf("Error removing service: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to remove service")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if rowsAffected == 0 {
		log.Printf("No service assignment found for resource %s", resourceID)
		ResponseWithError(c, http.StatusNotFound, "Resource service relationship not found")
		return
	}

	log.Printf("Delete affected %d rows", rowsAffected)

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully removed service from resource %s", resourceID)
	c.JSON(http.StatusOK, gin.H{"message": "Service removed from resource successfully"})
}

// GetResourceService returns the service associated with a resource
func (h *ServiceHandler) GetResourceService(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var serviceID string
	err := h.DB.QueryRow("SELECT service_id FROM resource_services WHERE resource_id = ?", resourceID).Scan(&serviceID)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "No service assigned to this resource")
		return
	} else if err != nil {
		log.Printf("Error fetching resource service: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch resource service")
		return
	}

	// Get service details
	var name, typ, configStr string
	err = h.DB.QueryRow("SELECT name, type, config FROM services WHERE id = ?", serviceID).Scan(&name, &typ, &configStr)
	if err == sql.ErrNoRows {
		ResponseWithError(c, http.StatusNotFound, "Service not found")
		return
	} else if err != nil {
		log.Printf("Error fetching service: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch service")
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		log.Printf("Error parsing service config: %v", err)
		config = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"resource_id": resourceID,
		"service": gin.H{
			"id":     serviceID,
			"name":   name,
			"type":   typ,
			"config": config,
		},
	})
}

type serviceRecord struct {
	ID         string
	Name       string
	Type       string
	Config     string
	Status     string
	SourceType string
}

// findServiceByID resolves a service by exact ID, normalized ID, or provider-suffixed variants.
func (h *ServiceHandler) findServiceByID(id string) (serviceRecord, error) {
	candidates := []string{id}
	normalized := util.NormalizeID(id)
	if normalized != id {
		candidates = append(candidates, normalized)
	}
	if !strings.Contains(normalized, "@") {
		candidates = append(candidates, normalized+"@%")
	}

	var rec serviceRecord
	for _, candidate := range candidates {
		var err error
		if strings.Contains(candidate, "%") {
			err = h.DB.QueryRow(
				"SELECT id, name, type, config, COALESCE(status, 'active'), COALESCE(source_type, '') FROM services WHERE id LIKE ? LIMIT 1",
				candidate,
			).Scan(&rec.ID, &rec.Name, &rec.Type, &rec.Config, &rec.Status, &rec.SourceType)
		} else {
			err = h.DB.QueryRow(
				"SELECT id, name, type, config, COALESCE(status, 'active'), COALESCE(source_type, '') FROM services WHERE id = ?",
				candidate,
			).Scan(&rec.ID, &rec.Name, &rec.Type, &rec.Config, &rec.Status, &rec.SourceType)
		}

		if err == nil {
			return rec, nil
		}
		if err != sql.ErrNoRows {
			return serviceRecord{}, err
		}
	}

	return serviceRecord{}, sql.ErrNoRows
}
