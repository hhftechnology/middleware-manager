package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

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
// Supports pagination via ?page=N&page_size=M query parameters
// Supports filtering by source_type via ?source_type=pangolin|traefik
// Supports filtering by status via ?status=active|disabled (default: active)
func (h *ResourceHandler) GetResources(c *gin.Context) {
	// Check if pagination is requested
	usePagination := IsPaginationRequested(c)
	params := GetPaginationParams(c)

	// Get optional filters
	sourceType := c.Query("source_type")
	statusFilter := c.DefaultQuery("status", "active") // Default to active resources only

	// Build WHERE clause for filters
	whereClause := ""
	var filterArgs []interface{}

	if statusFilter != "" && statusFilter != "all" {
		whereClause = " WHERE r.status = ?"
		filterArgs = append(filterArgs, statusFilter)
	}

	if sourceType != "" {
		if whereClause == "" {
			whereClause = " WHERE r.source_type = ?"
		} else {
			whereClause += " AND r.source_type = ?"
		}
		filterArgs = append(filterArgs, sourceType)
	}

	var total int
	if usePagination {
		// Get total count for pagination with filters
		countQuery := "SELECT COUNT(*) FROM resources r" + whereClause
		err := h.DB.QueryRow(countQuery, filterArgs...).Scan(&total)
		if err != nil {
			log.Printf("Error counting resources: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to count resources")
			return
		}
	}

	// Build query with optional pagination and filters
	query := `
		SELECT r.id, COALESCE(r.pangolin_router_id, r.id), r.host, r.service_id, r.org_id, r.site_id, r.status,
		       r.entrypoints, r.tls_domains, r.tcp_enabled, r.tcp_entrypoints, r.tcp_sni_rule,
		       r.custom_headers, r.mtls_enabled, r.router_priority, r.source_type,
		       r.mtls_rules, r.mtls_request_headers, r.mtls_reject_message, r.mtls_reject_code,
		       r.mtls_refresh_interval, r.mtls_external_data,
		       COALESCE(r.tls_hardening_enabled, 0), COALESCE(r.secure_headers_enabled, 0),
		       GROUP_CONCAT(m.id || ':' || m.name || ':' || rm.priority, ',') as middlewares
		FROM resources r
		LEFT JOIN resource_middlewares rm ON r.id = rm.resource_id
		LEFT JOIN middlewares m ON rm.middleware_id = m.id
	` + whereClause + `
		GROUP BY r.id
		ORDER BY r.id
	`

	var rows *sql.Rows
	var err error

	if usePagination {
		query += " LIMIT ? OFFSET ?"
		args := append(filterArgs, params.PageSize, params.Offset)
		rows, err = h.DB.Query(query, args...)
	} else {
		rows, err = h.DB.Query(query, filterArgs...)
	}

	if err != nil {
		log.Printf("Error fetching resources: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch resources")
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}
	for rows.Next() {
		var id, pangolinRouterID, host, serviceID, orgID, siteID, status, entrypoints, tlsDomains, tcpEntrypoints, tcpSNIRule, customHeaders, sourceType string
		var tcpEnabled int
		var mtlsEnabled int
		var tlsHardeningEnabled, secureHeadersEnabled int
		var routerPriority sql.NullInt64
		var middlewares sql.NullString
		var mtlsRules, mtlsRequestHeaders, mtlsRejectMessage, mtlsRefreshInterval, mtlsExternalData sql.NullString
		var mtlsRejectCode sql.NullInt64

		if err := rows.Scan(&id, &pangolinRouterID, &host, &serviceID, &orgID, &siteID, &status,
			&entrypoints, &tlsDomains, &tcpEnabled, &tcpEntrypoints, &tcpSNIRule,
			&customHeaders, &mtlsEnabled, &routerPriority, &sourceType,
			&mtlsRules, &mtlsRequestHeaders, &mtlsRejectMessage, &mtlsRejectCode,
			&mtlsRefreshInterval, &mtlsExternalData,
			&tlsHardeningEnabled, &secureHeadersEnabled,
			&middlewares); err != nil {
			log.Printf("Error scanning resource row: %v", err)
			continue
		}

		priority := 200
		if routerPriority.Valid {
			priority = int(routerPriority.Int64)
		}

		resource := map[string]interface{}{
			"id":                     id,
			"pangolin_router_id":     pangolinRouterID,
			"host":                   host,
			"service_id":             serviceID,
			"org_id":                 orgID,
			"site_id":                siteID,
			"status":                 status,
			"entrypoints":            entrypoints,
			"tls_domains":            tlsDomains,
			"tcp_enabled":            tcpEnabled > 0,
			"tcp_entrypoints":        tcpEntrypoints,
			"tcp_sni_rule":           tcpSNIRule,
			"custom_headers":         customHeaders,
			"mtls_enabled":           mtlsEnabled > 0,
			"router_priority":        priority,
			"source_type":            sourceType,
			"tls_hardening_enabled":  tlsHardeningEnabled > 0,
			"secure_headers_enabled": secureHeadersEnabled > 0,
		}

		if mtlsRules.Valid {
			resource["mtls_rules"] = mtlsRules.String
		}
		if mtlsRequestHeaders.Valid {
			resource["mtls_request_headers"] = mtlsRequestHeaders.String
		}
		if mtlsRejectMessage.Valid {
			resource["mtls_reject_message"] = mtlsRejectMessage.String
		}
		if mtlsRejectCode.Valid {
			resource["mtls_reject_code"] = mtlsRejectCode.Int64
		}
		if mtlsRefreshInterval.Valid {
			resource["mtls_refresh_interval"] = mtlsRefreshInterval.String
		}
		if mtlsExternalData.Valid {
			resource["mtls_external_data"] = mtlsExternalData.String
		}

		if middlewares.Valid {
			resource["middlewares"] = middlewares.String
		} else {
			resource["middlewares"] = ""
		}

		resource["external_middlewares"] = ""
		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error during resource rows iteration: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch resources")
		return
	}

	// Batch-load external middlewares for all resources
	if len(resources) > 0 {
		extRows, err := h.DB.Query(
			"SELECT resource_id, middleware_name, priority, provider FROM resource_external_middlewares ORDER BY resource_id, priority DESC",
		)
		if err != nil {
			log.Printf("Warning: failed to fetch external middlewares: %v", err)
		} else {
			defer extRows.Close()
			extMap := make(map[string][]string)
			for extRows.Next() {
				var resID, name, provider string
				var priority int
				if err := extRows.Scan(&resID, &name, &priority, &provider); err != nil {
					log.Printf("Error scanning external middleware: %v", err)
					continue
				}
				extMap[resID] = append(extMap[resID], fmt.Sprintf("%s:%d:%s", name, priority, provider))
			}
			for i, res := range resources {
				if parts, ok := extMap[res["id"].(string)]; ok {
					resources[i]["external_middlewares"] = strings.Join(parts, ",")
				}
			}
		}
	}

	// Return paginated or regular response
	if usePagination {
		c.JSON(http.StatusOK, NewPaginatedResponse(resources, total, params))
	} else {
		c.JSON(http.StatusOK, resources)
	}
}

// GetResource returns a specific resource
func (h *ResourceHandler) GetResource(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var pangolinRouterID, host, serviceID, orgID, siteID, status, entrypoints, tlsDomains, tcpEntrypoints, tcpSNIRule, customHeaders, sourceType string
	var tcpEnabled int
	var mtlsEnabled int
	var tlsHardeningEnabled, secureHeadersEnabled int
	var routerPriority sql.NullInt64
	var middlewares sql.NullString
	var mtlsRules, mtlsRequestHeaders, mtlsRejectMessage, mtlsRefreshInterval, mtlsExternalData sql.NullString
	var mtlsRejectCode sql.NullInt64

	err := h.DB.QueryRow(`
        SELECT COALESCE(r.pangolin_router_id, r.id), r.host, r.service_id, r.org_id, r.site_id, r.status,
               r.entrypoints, r.tls_domains, r.tcp_enabled, r.tcp_entrypoints, r.tcp_sni_rule,
               r.custom_headers, r.mtls_enabled, r.router_priority, r.source_type,
               r.mtls_rules, r.mtls_request_headers, r.mtls_reject_message, r.mtls_reject_code,
               r.mtls_refresh_interval, r.mtls_external_data,
               COALESCE(r.tls_hardening_enabled, 0), COALESCE(r.secure_headers_enabled, 0),
               GROUP_CONCAT(m.id || ':' || m.name || ':' || rm.priority, ',') as middlewares
        FROM resources r
        LEFT JOIN resource_middlewares rm ON r.id = rm.resource_id
        LEFT JOIN middlewares m ON rm.middleware_id = m.id
        WHERE r.id = ?
        GROUP BY r.id
    `, id).Scan(&pangolinRouterID, &host, &serviceID, &orgID, &siteID, &status,
		&entrypoints, &tlsDomains, &tcpEnabled, &tcpEntrypoints, &tcpSNIRule,
		&customHeaders, &mtlsEnabled, &routerPriority, &sourceType,
		&mtlsRules, &mtlsRequestHeaders, &mtlsRejectMessage, &mtlsRejectCode,
		&mtlsRefreshInterval, &mtlsExternalData,
		&tlsHardeningEnabled, &secureHeadersEnabled,
		&middlewares)

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
		"id":                     id,
		"pangolin_router_id":     pangolinRouterID,
		"host":                   host,
		"service_id":             serviceID,
		"org_id":                 orgID,
		"site_id":                siteID,
		"status":                 status,
		"entrypoints":            entrypoints,
		"tls_domains":            tlsDomains,
		"tcp_enabled":            tcpEnabled > 0,
		"tcp_entrypoints":        tcpEntrypoints,
		"tcp_sni_rule":           tcpSNIRule,
		"custom_headers":         customHeaders,
		"mtls_enabled":           mtlsEnabled > 0,
		"router_priority":        priority,
		"source_type":            sourceType,
		"tls_hardening_enabled":  tlsHardeningEnabled > 0,
		"secure_headers_enabled": secureHeadersEnabled > 0,
	}

	if mtlsRules.Valid {
		resource["mtls_rules"] = mtlsRules.String
	}
	if mtlsRequestHeaders.Valid {
		resource["mtls_request_headers"] = mtlsRequestHeaders.String
	}
	if mtlsRejectMessage.Valid {
		resource["mtls_reject_message"] = mtlsRejectMessage.String
	}
	if mtlsRejectCode.Valid {
		resource["mtls_reject_code"] = mtlsRejectCode.Int64
	}
	if mtlsRefreshInterval.Valid {
		resource["mtls_refresh_interval"] = mtlsRefreshInterval.String
	}
	if mtlsExternalData.Valid {
		resource["mtls_external_data"] = mtlsExternalData.String
	}

	if middlewares.Valid {
		resource["middlewares"] = middlewares.String
	} else {
		resource["middlewares"] = ""
	}

	// Fetch external (Traefik-native) middlewares assigned to this resource
	extRows, err := h.DB.Query(
		"SELECT middleware_name, priority, provider FROM resource_external_middlewares WHERE resource_id = ? ORDER BY priority DESC",
		id,
	)
	if err != nil {
		log.Printf("Error fetching external middlewares for resource %s: %v", id, err)
		resource["external_middlewares"] = ""
	} else {
		defer extRows.Close()
		var extParts []string
		for extRows.Next() {
			var name, provider string
			var priority int
			if err := extRows.Scan(&name, &priority, &provider); err != nil {
				log.Printf("Error scanning external middleware row: %v", err)
				continue
			}
			extParts = append(extParts, fmt.Sprintf("%s:%d:%s", name, priority, provider))
		}
		resource["external_middlewares"] = strings.Join(extParts, ",")
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

// DeleteDisabledResources deletes a list of disabled resources (bulk).
func (h *ResourceHandler) DeleteDisabledResources(c *gin.Context) {
	var payload struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil || len(payload.IDs) == 0 {
		ResponseWithError(c, http.StatusBadRequest, "IDs are required")
		return
	}

	// Use a transaction to remove relationships then resources
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

	placeholders := strings.Repeat("?,", len(payload.IDs))
	placeholders = strings.TrimSuffix(placeholders, ",")

	// Ensure all IDs are disabled before deleting
	query := fmt.Sprintf("SELECT id, status FROM resources WHERE id IN (%s)", placeholders)
	args := make([]interface{}, len(payload.IDs))
	for i, v := range payload.IDs {
		args[i] = v
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		txErr = err
		log.Printf("Error checking resource statuses: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	allowed := map[string]struct{}{}
	for rows.Next() {
		var rid, status string
		if err := rows.Scan(&rid, &status); err != nil {
			txErr = err
			log.Printf("Error scanning resource row: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Database error")
			return
		}
		if status == "disabled" {
			allowed[rid] = struct{}{}
		}
	}

	// Filter IDs to disabled ones
	disabledIDs := make([]string, 0, len(allowed))
	for _, id := range payload.IDs {
		if _, ok := allowed[id]; ok {
			disabledIDs = append(disabledIDs, id)
		}
	}

	if len(disabledIDs) == 0 {
		ResponseWithError(c, http.StatusBadRequest, "No disabled resources to delete")
		return
	}

	// Build placeholders for disabled IDs
	dPlaceholders := strings.Repeat("?,", len(disabledIDs))
	dPlaceholders = strings.TrimSuffix(dPlaceholders, ",")
	dArgs := make([]interface{}, len(disabledIDs))
	for i, v := range disabledIDs {
		dArgs[i] = v
	}

	// Delete resource_middlewares first
	rmQuery := fmt.Sprintf("DELETE FROM resource_middlewares WHERE resource_id IN (%s)", dPlaceholders)
	if _, txErr = tx.Exec(rmQuery, dArgs...); txErr != nil {
		log.Printf("Error deleting resource_middlewares: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete resources")
		return
	}

	// Delete resources
	resQuery := fmt.Sprintf("DELETE FROM resources WHERE id IN (%s) AND status = 'disabled'", dPlaceholders)
	result, txErr := tx.Exec(resQuery, dArgs...)
	if txErr != nil {
		log.Printf("Error deleting resources: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete resources")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		txErr = err
		log.Printf("Error getting rows affected: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted": rowsAffected,
		"ids":     disabledIDs,
	})
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
				"priority":      mw.Priority,
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

// AssignExternalMiddleware assigns a Traefik-native middleware to a resource by name
func (h *ResourceHandler) AssignExternalMiddleware(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	var input struct {
		MiddlewareName string `json:"middleware_name" binding:"required"`
		Priority       int    `json:"priority"`
		Provider       string `json:"provider"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Validate middleware name is not empty after trimming
	input.MiddlewareName = strings.TrimSpace(input.MiddlewareName)
	if input.MiddlewareName == "" {
		ResponseWithError(c, http.StatusBadRequest, "Middleware name is required")
		return
	}

	// Default priority
	if input.Priority <= 0 {
		input.Priority = 100
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

	if status == "disabled" {
		ResponseWithError(c, http.StatusBadRequest, "Cannot assign middleware to a disabled resource")
		return
	}

	// Insert or update using a transaction
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

	// Delete any existing relationship first
	_, txErr = tx.Exec(
		"DELETE FROM resource_external_middlewares WHERE resource_id = ? AND middleware_name = ?",
		resourceID, input.MiddlewareName,
	)
	if txErr != nil {
		log.Printf("Error removing existing external middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Insert new relationship
	_, txErr = tx.Exec(
		"INSERT INTO resource_external_middlewares (resource_id, middleware_name, priority, provider) VALUES (?, ?, ?, ?)",
		resourceID, input.MiddlewareName, input.Priority, input.Provider,
	)
	if txErr != nil {
		log.Printf("Error assigning external middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to assign external middleware")
		return
	}

	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully assigned external middleware %s to resource %s with priority %d",
		input.MiddlewareName, resourceID, input.Priority)
	c.JSON(http.StatusOK, gin.H{
		"resource_id":     resourceID,
		"middleware_name": input.MiddlewareName,
		"priority":        input.Priority,
		"provider":        input.Provider,
	})
}

// RemoveExternalMiddleware removes a Traefik-native middleware from a resource
func (h *ResourceHandler) RemoveExternalMiddleware(c *gin.Context) {
	resourceID := c.Param("id")
	middlewareName := c.Param("name")

	if resourceID == "" || middlewareName == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID and Middleware name are required")
		return
	}

	log.Printf("Removing external middleware %s from resource %s", middlewareName, resourceID)

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

	result, txErr := tx.Exec(
		"DELETE FROM resource_external_middlewares WHERE resource_id = ? AND middleware_name = ?",
		resourceID, middlewareName,
	)
	if txErr != nil {
		log.Printf("Error removing external middleware: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to remove external middleware")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if rowsAffected == 0 {
		ResponseWithError(c, http.StatusNotFound, "External middleware assignment not found")
		return
	}

	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Error committing transaction: %v", txErr)
		ResponseWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	log.Printf("Successfully removed external middleware %s from resource %s", middlewareName, resourceID)
	c.JSON(http.StatusOK, gin.H{"message": "External middleware removed from resource successfully"})
}

// GetExternalMiddlewares returns all external middlewares assigned to a resource
func (h *ResourceHandler) GetExternalMiddlewares(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		ResponseWithError(c, http.StatusBadRequest, "Resource ID is required")
		return
	}

	rows, err := h.DB.Query(
		"SELECT middleware_name, priority, provider, created_at FROM resource_external_middlewares WHERE resource_id = ? ORDER BY priority DESC",
		resourceID,
	)
	if err != nil {
		log.Printf("Error fetching external middlewares: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch external middlewares")
		return
	}
	defer rows.Close()

	var externalMiddlewares []gin.H
	for rows.Next() {
		var name, provider string
		var priority int
		var createdAt string
		if err := rows.Scan(&name, &priority, &provider, &createdAt); err != nil {
			log.Printf("Error scanning external middleware row: %v", err)
			continue
		}
		externalMiddlewares = append(externalMiddlewares, gin.H{
			"resource_id":     resourceID,
			"middleware_name": name,
			"priority":        priority,
			"provider":        provider,
			"created_at":      createdAt,
		})
	}

	if externalMiddlewares == nil {
		externalMiddlewares = []gin.H{}
	}

	c.JSON(http.StatusOK, externalMiddlewares)
}
