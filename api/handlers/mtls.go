package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/services"
	"gopkg.in/yaml.v3"
)

// MTLSHandler handles mTLS-related requests
type MTLSHandler struct {
	DB                      *sql.DB
	CertGenerator           *services.CertGenerator
	TraefikStaticConfigPath string
}

// NewMTLSHandler creates a new mTLS handler
func NewMTLSHandler(db *sql.DB) *MTLSHandler {
	return &MTLSHandler{
		DB:            db,
		CertGenerator: services.NewCertGenerator(db),
	}
}

// SetTraefikConfigPath sets the Traefik static config path
func (h *MTLSHandler) SetTraefikConfigPath(path string) {
	h.TraefikStaticConfigPath = path
}

// GetConfig returns the current mTLS configuration
func (h *MTLSHandler) GetConfig(c *gin.Context) {
	config, err := h.CertGenerator.GetConfig()
	if err != nil {
		log.Printf("Error getting mTLS config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get mTLS configuration")
		return
	}

	// Get client count
	clientCount, err := h.CertGenerator.GetClientCount()
	if err != nil {
		log.Printf("Error getting client count: %v", err)
		clientCount = 0
	}

	response := models.MTLSConfigResponse{
		MTLSConfig:  *config,
		ClientCount: clientCount,
	}

	// Don't return the actual certificate content for list view
	response.CACert = ""

	c.JSON(http.StatusOK, response)
}

// EnableMTLS enables mTLS globally
func (h *MTLSHandler) EnableMTLS(c *gin.Context) {
	if err := h.CertGenerator.EnableMTLS(); err != nil {
		log.Printf("Error enabling mTLS: %v", err)
		ResponseWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "mTLS enabled successfully",
		"enabled": true,
	})
}

// DisableMTLS disables mTLS globally
func (h *MTLSHandler) DisableMTLS(c *gin.Context) {
	if err := h.CertGenerator.DisableMTLS(); err != nil {
		log.Printf("Error disabling mTLS: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to disable mTLS")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "mTLS disabled successfully",
		"enabled": false,
	})
}

// CreateCA creates a new Certificate Authority
func (h *MTLSHandler) CreateCA(c *gin.Context) {
	var req models.CreateCARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get current config for base path
	config, err := h.CertGenerator.GetConfig()
	if err != nil {
		log.Printf("Error getting mTLS config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get mTLS configuration")
		return
	}

	basePath := config.CertsBasePath
	if basePath == "" {
		basePath = "/etc/traefik/certs"
	}

	newConfig, err := h.CertGenerator.GenerateCA(req, basePath)
	if err != nil {
		log.Printf("Error creating CA: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to create CA: "+err.Error())
		return
	}

	// Don't return the certificate content
	newConfig.CACert = ""

	c.JSON(http.StatusCreated, newConfig)
}

// DeleteCA deletes the CA and all client certificates
func (h *MTLSHandler) DeleteCA(c *gin.Context) {
	if err := h.CertGenerator.DeleteCA(); err != nil {
		log.Printf("Error deleting CA: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete CA: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CA and all client certificates deleted successfully",
	})
}

// GetClients returns all client certificates
func (h *MTLSHandler) GetClients(c *gin.Context) {
	clients, err := h.CertGenerator.GetClients()
	if err != nil {
		log.Printf("Error getting clients: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get clients")
		return
	}

	// Don't return certificate content in list
	for i := range clients {
		clients[i].Cert = ""
	}

	if clients == nil {
		clients = []models.MTLSClient{}
	}

	c.JSON(http.StatusOK, clients)
}

// CreateClient creates a new client certificate
func (h *MTLSHandler) CreateClient(c *gin.Context) {
	var req models.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	client, err := h.CertGenerator.GenerateClientCert(req)
	if err != nil {
		log.Printf("Error creating client certificate: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to create client certificate: "+err.Error())
		return
	}

	// Don't return the certificate content
	client.Cert = ""

	c.JSON(http.StatusCreated, client)
}

// GetClient returns a specific client certificate
func (h *MTLSHandler) GetClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Client ID is required")
		return
	}

	client, err := h.CertGenerator.GetClient(id)
	if err != nil {
		log.Printf("Error getting client: %v", err)
		ResponseWithError(c, http.StatusNotFound, "Client not found")
		return
	}

	c.JSON(http.StatusOK, client)
}

// DownloadClientP12 downloads the PKCS#12 file for a client
func (h *MTLSHandler) DownloadClientP12(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Client ID is required")
		return
	}

	p12Data, name, err := h.CertGenerator.GetClientP12(id)
	if err != nil {
		log.Printf("Error getting client P12: %v", err)
		ResponseWithError(c, http.StatusNotFound, "Client not found")
		return
	}

	filename := name + ".p12"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/x-pkcs12")
	c.Data(http.StatusOK, "application/x-pkcs12", p12Data)
}

// RevokeClient revokes a client certificate
func (h *MTLSHandler) RevokeClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Client ID is required")
		return
	}

	if err := h.CertGenerator.RevokeClient(id); err != nil {
		log.Printf("Error revoking client: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to revoke client: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Client certificate revoked successfully",
		"id":      id,
	})
}

// DeleteClient deletes a client certificate
func (h *MTLSHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ResponseWithError(c, http.StatusBadRequest, "Client ID is required")
		return
	}

	if err := h.CertGenerator.DeleteClient(id); err != nil {
		log.Printf("Error deleting client: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete client: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Client certificate deleted successfully",
		"id":      id,
	})
}

// UpdateCertsBasePath updates the certificates base path
func (h *MTLSHandler) UpdateCertsBasePath(c *gin.Context) {
	var input struct {
		CertsBasePath string `json:"certs_base_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.CertGenerator.UpdateCertsBasePath(input.CertsBasePath); err != nil {
		log.Printf("Error updating certs base path: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update certificates path")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Certificates base path updated successfully",
		"certs_base_path": input.CertsBasePath,
	})
}

// CheckPlugin checks if the mtlswhitelist plugin is installed
func (h *MTLSHandler) CheckPlugin(c *gin.Context) {
	installed, version := h.isPluginInstalled("mtlswhitelist")

	c.JSON(http.StatusOK, gin.H{
		"installed":   installed,
		"plugin_name": "mtlswhitelist",
		"version":     version,
	})
}

// isPluginInstalled checks if a specific plugin is installed in Traefik static config
func (h *MTLSHandler) isPluginInstalled(pluginName string) (bool, string) {
	if h.TraefikStaticConfigPath == "" {
		return false, ""
	}

	cleanPath := filepath.Clean(h.TraefikStaticConfigPath)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		log.Printf("Warning: Could not read Traefik config for plugin check: %v", err)
		return false, ""
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Printf("Warning: Could not parse Traefik config for plugin check: %v", err)
		return false, ""
	}

	// Check experimental.plugins section
	if experimentalSection, ok := config["experimental"].(map[string]interface{}); ok {
		if pluginsConfig, ok := experimentalSection["plugins"].(map[string]interface{}); ok {
			// Check for the plugin by name (lowercase comparison)
			for key, pluginData := range pluginsConfig {
				if strings.EqualFold(key, pluginName) {
					version := ""
					if pluginEntry, ok := pluginData.(map[string]interface{}); ok {
						if v, ok := pluginEntry["version"].(string); ok {
							version = v
						}
					}
					return true, version
				}
			}
		}
	}

	return false, ""
}

// GetMiddlewareConfig returns the mTLS middleware configuration
func (h *MTLSHandler) GetMiddlewareConfig(c *gin.Context) {
	config, err := h.CertGenerator.GetMiddlewareConfig()
	if err != nil {
		log.Printf("Error getting middleware config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get middleware configuration")
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateMiddlewareConfig updates the mTLS middleware configuration
func (h *MTLSHandler) UpdateMiddlewareConfig(c *gin.Context) {
	var input models.MTLSMiddlewareConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		ResponseWithError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.CertGenerator.UpdateMiddlewareConfig(&input); err != nil {
		log.Printf("Error updating middleware config: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to update middleware configuration: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Middleware configuration updated successfully",
	})
}
