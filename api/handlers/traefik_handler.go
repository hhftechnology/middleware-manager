package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/models"
	"github.com/hhftechnology/middleware-manager/services"
)

// TraefikHandler handles Traefik API proxy requests
type TraefikHandler struct {
	DB            *sql.DB
	ConfigManager *services.ConfigManager
}

// NewTraefikHandler creates a new Traefik handler
func NewTraefikHandler(db *sql.DB, configManager *services.ConfigManager) *TraefikHandler {
	return &TraefikHandler{
		DB:            db,
		ConfigManager: configManager,
	}
}

// getFetcher gets the appropriate fetcher based on active data source
func (h *TraefikHandler) getFetcher() (*services.TraefikFetcher, error) {
	config, err := h.ConfigManager.GetActiveDataSourceConfig()
	if err != nil {
		return nil, err
	}

	// For Traefik data, always use TraefikFetcher
	// If current source is Pangolin, we'll use the Traefik config from data sources
	if config.Type == models.PangolinAPI {
		// Try to get Traefik config if available
		sources := h.ConfigManager.GetDataSources()
		if traefikConfig, ok := sources["traefik"]; ok {
			config = traefikConfig
		}
	}

	return services.NewTraefikFetcher(config), nil
}

// GetOverview returns the Traefik overview
func (h *TraefikHandler) GetOverview(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	overview, err := fetcher.GetOverview(ctx)
	if err != nil {
		log.Printf("Error fetching Traefik overview: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch Traefik overview")
		return
	}

	if overview == nil {
		c.JSON(http.StatusOK, gin.H{
			"http": gin.H{
				"routers":     gin.H{"total": 0, "warnings": 0, "errors": 0},
				"services":    gin.H{"total": 0, "warnings": 0, "errors": 0},
				"middlewares": gin.H{"total": 0, "warnings": 0, "errors": 0},
			},
			"tcp": gin.H{
				"routers":     gin.H{"total": 0, "warnings": 0, "errors": 0},
				"services":    gin.H{"total": 0, "warnings": 0, "errors": 0},
				"middlewares": gin.H{"total": 0, "warnings": 0, "errors": 0},
			},
			"udp": gin.H{
				"routers":     gin.H{"total": 0, "warnings": 0, "errors": 0},
				"services":    gin.H{"total": 0, "warnings": 0, "errors": 0},
				"middlewares": gin.H{"total": 0, "warnings": 0, "errors": 0},
			},
			"features":  gin.H{},
			"providers": []string{},
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetVersion returns the Traefik version
func (h *TraefikHandler) GetVersion(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	version, err := fetcher.GetVersion(ctx)
	if err != nil {
		log.Printf("Error fetching Traefik version: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch Traefik version")
		return
	}

	if version == nil {
		c.JSON(http.StatusOK, gin.H{
			"version":  "unknown",
			"codename": "unknown",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// GetEntrypoints returns Traefik entrypoints
func (h *TraefikHandler) GetEntrypoints(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	entrypoints, err := fetcher.GetEntrypoints(ctx)
	if err != nil {
		log.Printf("Error fetching Traefik entrypoints: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch Traefik entrypoints")
		return
	}

	c.JSON(http.StatusOK, entrypoints)
}

// GetRouters returns Traefik routers with optional protocol filter
func (h *TraefikHandler) GetRouters(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	protocolType := c.Query("type")

	switch protocolType {
	case "http", "":
		routers, err := fetcher.GetTraefikRouters(ctx)
		if err != nil {
			log.Printf("Error fetching HTTP routers: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch HTTP routers")
			return
		}
		c.JSON(http.StatusOK, routers)

	case "tcp":
		routers, err := fetcher.GetTCPRouters(ctx)
		if err != nil {
			log.Printf("Error fetching TCP routers: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch TCP routers")
			return
		}
		c.JSON(http.StatusOK, routers)

	case "udp":
		routers, err := fetcher.GetUDPRouters(ctx)
		if err != nil {
			log.Printf("Error fetching UDP routers: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch UDP routers")
			return
		}
		c.JSON(http.StatusOK, routers)

	case "all":
		data, err := fetcher.FetchFullData(ctx)
		if err != nil {
			log.Printf("Error fetching all routers: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch routers")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"http": data.HTTPRouters,
			"tcp":  data.TCPRouters,
			"udp":  data.UDPRouters,
			"total": gin.H{
				"http": len(data.HTTPRouters),
				"tcp":  len(data.TCPRouters),
				"udp":  len(data.UDPRouters),
			},
		})

	default:
		ResponseWithError(c, http.StatusBadRequest, "Invalid protocol type. Use: http, tcp, udp, or all")
	}
}

// GetServices returns Traefik services with optional protocol filter
func (h *TraefikHandler) GetServices(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	protocolType := c.Query("type")

	switch protocolType {
	case "http", "":
		services, err := fetcher.GetTraefikServices(ctx)
		if err != nil {
			log.Printf("Error fetching HTTP services: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch HTTP services")
			return
		}
		c.JSON(http.StatusOK, services)

	case "tcp":
		services, err := fetcher.GetTCPServices(ctx)
		if err != nil {
			log.Printf("Error fetching TCP services: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch TCP services")
			return
		}
		c.JSON(http.StatusOK, services)

	case "udp":
		services, err := fetcher.GetUDPServices(ctx)
		if err != nil {
			log.Printf("Error fetching UDP services: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch UDP services")
			return
		}
		c.JSON(http.StatusOK, services)

	case "all":
		data, err := fetcher.FetchFullData(ctx)
		if err != nil {
			log.Printf("Error fetching all services: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch services")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"http": data.HTTPServices,
			"tcp":  data.TCPServices,
			"udp":  data.UDPServices,
			"total": gin.H{
				"http": len(data.HTTPServices),
				"tcp":  len(data.TCPServices),
				"udp":  len(data.UDPServices),
			},
		})

	default:
		ResponseWithError(c, http.StatusBadRequest, "Invalid protocol type. Use: http, tcp, udp, or all")
	}
}

// GetMiddlewares returns Traefik middlewares with optional protocol filter
func (h *TraefikHandler) GetMiddlewares(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	protocolType := c.Query("type")

	switch protocolType {
	case "http", "":
		middlewares, err := fetcher.GetTraefikMiddlewares(ctx)
		if err != nil {
			log.Printf("Error fetching HTTP middlewares: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch HTTP middlewares")
			return
		}
		c.JSON(http.StatusOK, middlewares)

	case "tcp":
		middlewares, err := fetcher.GetTCPMiddlewares(ctx)
		if err != nil {
			log.Printf("Error fetching TCP middlewares: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch TCP middlewares")
			return
		}
		c.JSON(http.StatusOK, middlewares)

	case "all":
		data, err := fetcher.FetchFullData(ctx)
		if err != nil {
			log.Printf("Error fetching all middlewares: %v", err)
			ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch middlewares")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"http": data.HTTPMiddlewares,
			"tcp":  data.TCPMiddlewares,
			"total": gin.H{
				"http": len(data.HTTPMiddlewares),
				"tcp":  len(data.TCPMiddlewares),
			},
		})

	default:
		ResponseWithError(c, http.StatusBadRequest, "Invalid protocol type. Use: http, tcp, or all")
	}
}

// GetFullData returns all Traefik data in one request
func (h *TraefikHandler) GetFullData(c *gin.Context) {
	fetcher, err := h.getFetcher()
	if err != nil {
		log.Printf("Error getting fetcher: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to get data source configuration")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	data, err := fetcher.FetchFullData(ctx)
	if err != nil {
		log.Printf("Error fetching full Traefik data: %v", err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to fetch Traefik data")
		return
	}

	c.JSON(http.StatusOK, data)
}
