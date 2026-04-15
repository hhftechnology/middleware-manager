package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	"github.com/hhftechnology/middleware-manager/internal/traefikmanager/handlers"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

type Server struct {
	router *gin.Engine
	srv    *http.Server
}

func New(cfg tmtypes.RuntimeConfig, files *tmconfig.FileStore, settings *tmconfig.SettingsStore, dashboard *tmconfig.DashboardStore, client *http.Client) *Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	if cfg.Debug {
		router.Use(gin.Logger())
	}
	if cfg.AllowCORS {
		corsConfig := cors.DefaultConfig()
		if cfg.CORSOrigin != "" {
			corsConfig.AllowOrigins = []string{cfg.CORSOrigin}
		} else {
			corsConfig.AllowAllOrigins = true
		}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
		router.Use(cors.New(corsConfig))
	}

	routeHandler := handlers.NewRouteHandler(files, settings)
	middlewareHandler := handlers.NewMiddlewareHandler(files)
	settingsHandler := handlers.NewSettingsHandler(settings, files, client)
	backupHandler := handlers.NewBackupHandler(files)
	traefikHandler := handlers.NewTraefikHandler(settings, files, client)
	dashboardHandler := handlers.NewDashboardHandler(dashboard, client)
	managerHandler := handlers.NewManagerHandler(client, cfg.GitHubRepo)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "mode": "traefik-manager"})
	})

	api := router.Group("/api")
	{
		api.GET("/routes", routeHandler.List)
		api.GET("/routes/all", routeHandler.ListAll)
		api.POST("/routes", routeHandler.Create)
		api.PUT("/routes/:id", routeHandler.Update)
		api.DELETE("/routes/:id", routeHandler.Delete)
		api.POST("/routes/:id/toggle", routeHandler.Toggle)
		api.GET("/configs", routeHandler.Configs)

		api.GET("/middlewares", middlewareHandler.List)
		api.POST("/middlewares", middlewareHandler.Create)
		api.PUT("/middlewares/:name", middlewareHandler.Update)
		api.DELETE("/middlewares/:name", middlewareHandler.Delete)

		api.GET("/settings", settingsHandler.Get)
		api.POST("/settings", settingsHandler.Save)
		api.POST("/settings/tabs", settingsHandler.SaveTabs)
		api.GET("/settings/self-route", settingsHandler.GetSelfRoute)
		api.POST("/settings/self-route", settingsHandler.SaveSelfRoute)
		api.POST("/settings/test-connection", settingsHandler.TestConnection)

		api.GET("/backups", backupHandler.List)
		api.POST("/backup/create", backupHandler.Create)
		api.POST("/restore/:filename", backupHandler.Restore)
		api.POST("/backup/delete/:filename", backupHandler.Delete)

		api.GET("/traefik/overview", traefikHandler.Overview)
		api.GET("/traefik/routers", traefikHandler.Routers)
		api.GET("/traefik/services", traefikHandler.Services)
		api.GET("/traefik/middlewares", traefikHandler.Middlewares)
		api.GET("/traefik/entrypoints", traefikHandler.Entrypoints)
		api.GET("/traefik/version", traefikHandler.Version)
		api.GET("/traefik/ping", traefikHandler.Ping)
		api.GET("/traefik/certs", traefikHandler.Certs)
		api.GET("/traefik/logs", traefikHandler.Logs)
		api.GET("/traefik/plugins", traefikHandler.Plugins)
		api.GET("/traefik/router/:protocol/:name", traefikHandler.RouterDetail)

		api.GET("/dashboard/config", dashboardHandler.GetConfig)
		api.POST("/dashboard/config", dashboardHandler.SaveConfig)
		api.GET("/dashboard/icon/:slug", dashboardHandler.Icon)

		api.GET("/manager/version", managerHandler.Version)
		api.GET("/manager/router-names", routeHandler.RouterNames)
	}

	uiPath := cfg.UIPath
	if stat, err := os.Stat(uiPath); err == nil && stat.IsDir() {
		router.Use(static.Serve("/", static.LocalFile(uiPath, false)))
		router.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "API endpoint not found"})
				return
			}
			c.File(uiPath + "/index.html")
		})
	} else {
		log.Printf("Traefik UI path %s missing; UI disabled", uiPath)
	}

	return &Server{
		router: router,
		srv: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           router,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("Traefik Manager API listening on %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
