package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

var iconSlugPattern = regexp.MustCompile(`[^a-z0-9-]`)

type DashboardHandler struct {
	store  *tmconfig.DashboardStore
	client *http.Client
}

func NewDashboardHandler(store *tmconfig.DashboardStore, client *http.Client) *DashboardHandler {
	return &DashboardHandler{store: store, client: client}
}

func (h *DashboardHandler) GetConfig(c *gin.Context) {
	config, err := h.store.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load dashboard config", err)
		return
	}
	respondJSON(c, config)
}

func (h *DashboardHandler) SaveConfig(c *gin.Context) {
	var request tmtypes.DashboardConfig
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	if err := h.store.Save(request); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save dashboard config", err)
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *DashboardHandler) Icon(c *gin.Context) {
	slug := iconSlugPattern.ReplaceAllString(c.Param("slug"), "")
	if slug == "" {
		c.Status(http.StatusNotFound)
		return
	}
	cacheDir, err := h.store.EnsureCacheDir()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	cachePath := filepath.Join(cacheDir, slug+".png")
	missPath := filepath.Join(cacheDir, slug+".404")
	if _, err := os.Stat(cachePath); err == nil {
		c.File(cachePath)
		return
	}
	if _, err := os.Stat(missPath); err == nil {
		c.Status(http.StatusNotFound)
		return
	}
	resp, err := h.client.Get("https://cdn.jsdelivr.net/gh/selfhst/icons/png/" + slug + ".png")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		_ = os.WriteFile(missPath, []byte(""), 0o644)
		c.Status(http.StatusNotFound)
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	_ = os.WriteFile(cachePath, data, 0o644)
	c.Data(http.StatusOK, "image/png", data)
}
