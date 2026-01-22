package handlers

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/hhftechnology/middleware-manager/models"
    "github.com/hhftechnology/middleware-manager/services"
)

// DataSourceHandler handles data source configuration requests
type DataSourceHandler struct {
    ConfigManager *services.ConfigManager
}

// NewDataSourceHandler creates a new data source handler
func NewDataSourceHandler(configManager *services.ConfigManager) *DataSourceHandler {
    return &DataSourceHandler{
        ConfigManager: configManager,
    }
}

// GetDataSources returns all configured data sources
func (h *DataSourceHandler) GetDataSources(c *gin.Context) {
    sources := h.ConfigManager.GetDataSources()
    activeSource := h.ConfigManager.GetActiveSourceName()
    
    // Format sources to mask passwords
    for key, source := range sources {
        source.FormatBasicAuth()
        sources[key] = source
    }
    
    c.JSON(http.StatusOK, gin.H{
        "active_source": activeSource,
        "sources":       sources,
    })
}

// GetActiveDataSource returns the active data source configuration
func (h *DataSourceHandler) GetActiveDataSource(c *gin.Context) {
    sourceConfig, err := h.ConfigManager.GetActiveDataSourceConfig()
    if err != nil {
        ResponseWithError(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "name":   h.ConfigManager.GetActiveSourceName(),
        "config": sourceConfig,
    })
}

// SetActiveDataSource sets the active data source
func (h *DataSourceHandler) SetActiveDataSource(c *gin.Context) {
    var request struct {
        Name string `json:"name" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
        return
    }
    
    if err := h.ConfigManager.SetActiveDataSource(request.Name); err != nil {
        ResponseWithError(c, http.StatusBadRequest, err.Error())
        return
    }
    
    // Log the data source change
    log.Printf("Active data source changed to: %s", request.Name)
    
    c.JSON(http.StatusOK, gin.H{
        "message":        "Data source updated successfully",
        "name":           request.Name,
        "refresh_needed": true,
    })
}

// UpdateDataSource updates a data source configuration
func (h *DataSourceHandler) UpdateDataSource(c *gin.Context) {
    name := c.Param("name")
    if name == "" {
        ResponseWithError(c, http.StatusBadRequest, "Data source name is required")
        return
    }
    
    var config models.DataSourceConfig
    if err := c.ShouldBindJSON(&config); err != nil {
        ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
        return
    }
    
    if err := h.ConfigManager.UpdateDataSource(name, config); err != nil {
        ResponseWithError(c, http.StatusInternalServerError, err.Error())
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Data source updated successfully",
        "name":    name,
        "config":  config,
    })
}

// TestDataSourceConnection tests the connection to a data source
func (h *DataSourceHandler) TestDataSourceConnection(c *gin.Context) {
    name := c.Param("name")
    if name == "" {
        ResponseWithError(c, http.StatusBadRequest, "Data source name is required")
        return
    }
    
    var config models.DataSourceConfig
    if err := c.ShouldBindJSON(&config); err != nil {
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
        return
    }
    
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Test the connection with endpoints that work
    err := testDataSourceConnection(ctx, config)
    if err != nil {
        log.Printf("Connection test failed for %s: %v", name, err)
        ResponseWithError(c, http.StatusBadRequest, fmt.Sprintf("Connection test failed: %v", err))
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Connection test successful",
        "name":    name,
    })
}

// testDataSourceConnection tests the connection to a data source using different endpoints
// based on the data source type
func testDataSourceConnection(ctx context.Context, config models.DataSourceConfig) error {
    client := &http.Client{
        Timeout: 5 * time.Second,
    }
    
    var url string
    switch config.Type {
    case models.PangolinAPI:
        // Use traefik-config endpoint instead of status to test Pangolin
        url = config.URL + "/traefik-config"
    case models.TraefikAPI:
        // Use http/routers endpoint to test Traefik
        url = config.URL + "/api/http/routers"
    default:
        return fmt.Errorf("unsupported data source type: %s", config.Type)
    }
    
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add basic auth if configured
    if config.BasicAuth.Username != "" {
        req.SetBasicAuth(config.BasicAuth.Username, config.BasicAuth.Password)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("connection failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("API returned status code: %d", resp.StatusCode)
    }
    
    return nil
}