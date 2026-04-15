package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
	"gopkg.in/yaml.v3"
)

type MiddlewareHandler struct {
	files *tmconfig.FileStore
}

func NewMiddlewareHandler(files *tmconfig.FileStore) *MiddlewareHandler {
	return &MiddlewareHandler{files: files}
}

func (h *MiddlewareHandler) List(c *gin.Context) {
	items := make([]tmtypes.MiddlewareEntry, 0)
	for _, path := range h.files.ConfigPaths() {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to load middlewares", err)
			return
		}
		configFile := ""
		if h.files.MultiConfig() || h.files.ActiveConfigDir() != "" {
			configFile = filepath.Base(path)
		}
		middlewares, err := tmconfig.BuildMiddlewares(config, configFile)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to load middlewares", err)
			return
		}
		items = append(items, middlewares...)
	}
	respondJSON(c, items)
}

func (h *MiddlewareHandler) Create(c *gin.Context) {
	var request tmtypes.MiddlewareRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	if err := h.save(request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (h *MiddlewareHandler) Update(c *gin.Context) {
	var request tmtypes.MiddlewareRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}
	if request.OriginalName == "" {
		request.OriginalName = c.Param("name")
	}
	if request.Name == "" {
		request.Name = c.Param("name")
	}
	if err := h.save(request); err != nil {
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	respondJSON(c, gin.H{"ok": true})
}

func (h *MiddlewareHandler) Delete(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		respondError(c, http.StatusBadRequest, "middleware name is required", nil)
		return
	}
	targets := h.files.ConfigPaths()
	if configFile := strings.TrimSpace(c.Query("configFile")); configFile != "" {
		if path := h.files.ResolveConfigPath(configFile); path != "" {
			targets = []string{path}
		}
	}
	for _, path := range targets {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to load config", err)
			return
		}
		httpSection := tmconfig.MapFromAny(config["http"])
		middlewares := tmconfig.MapFromAny(httpSection["middlewares"])
		if _, ok := middlewares[name]; !ok {
			continue
		}
		delete(middlewares, name)
		httpSection["middlewares"] = middlewares
		config["http"] = httpSection
		if _, err := h.files.CreateBackup(path); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to create backup", err)
			return
		}
		if err := h.files.SaveConfig(path, config); err != nil {
			respondError(c, http.StatusInternalServerError, "Failed to save config", err)
			return
		}
		respondJSON(c, gin.H{"ok": true})
		return
	}
	respondError(c, http.StatusNotFound, "middleware not found", nil)
}

func (h *MiddlewareHandler) save(request tmtypes.MiddlewareRequest) error {
	if strings.TrimSpace(request.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(request.YAML) == "" {
		return fmt.Errorf("yaml is required")
	}
	targetPath := h.files.ResolveConfigPath(request.ConfigFile)
	if targetPath == "" {
		targetPath = h.files.ResolveConfigPath("")
	}
	config, err := h.files.LoadConfig(targetPath)
	if err != nil {
		return err
	}
	parsed := map[string]any{}
	if err := yaml.Unmarshal([]byte(request.YAML), &parsed); err != nil {
		return fmt.Errorf("parse middleware yaml: %w", err)
	}
	httpSection := tmconfig.MapFromAny(config["http"])
	middlewares := tmconfig.MapFromAny(httpSection["middlewares"])
	if request.OriginalName != "" && request.OriginalName != request.Name {
		delete(middlewares, request.OriginalName)
	}
	middlewares[request.Name] = parsed
	httpSection["middlewares"] = middlewares
	config["http"] = httpSection
	if _, err := h.files.CreateBackup(targetPath); err != nil {
		return err
	}
	return h.files.SaveConfig(targetPath, config)
}
