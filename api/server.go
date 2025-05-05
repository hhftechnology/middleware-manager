package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/api/handlers"
	"github.com/hhftechnology/middleware-manager/services"
)

// Server represents the API server
type Server struct {
	db                *sql.DB
	router            *gin.Engine
	srv               *http.Server
	middlewareHandler *handlers.MiddlewareHandler
	resourceHandler   *handlers.ResourceHandler
	configHandler     *handlers.ConfigHandler
	dataSourceHandler *handlers.DataSourceHandler
	serviceHandler    *handlers.ServiceHandler
	configManager     *services.ConfigManager
}

// ServerConfig contains configuration options for the server
type ServerConfig struct {
	Port       string
	UIPath     string
	Debug      bool
	AllowCORS  bool
	CORSOrigin string
}

// NewServer creates a new API server
func NewServer(db *sql.DB, config ServerConfig, configManager *services.ConfigManager) *Server {
	// Set gin mode based on debug flag
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	
	// Use recovery and logger middleware
	router.Use(gin.Recovery())
	if config.Debug {
		router.Use(gin.Logger())
	} else {
		// In production, use a custom minimal logger
		router.Use(minimalLogger())
	}

	// CORS middleware if enabled
	if config.AllowCORS {
		corsConfig := cors.DefaultConfig()
		
		// If a specific origin is provided, use it
		if config.CORSOrigin != "" {
			corsConfig.AllowOrigins = []string{config.CORSOrigin}
		} else {
			corsConfig.AllowAllOrigins = true
		}
		
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
		corsConfig.ExposeHeaders = []string{"Content-Length"}
		corsConfig.AllowCredentials = true
		corsConfig.MaxAge = 12 * time.Hour
		
		router.Use(cors.New(corsConfig))
	}

	// Create request handlers
	middlewareHandler := handlers.NewMiddlewareHandler(db)
	resourceHandler := handlers.NewResourceHandler(db)
	configHandler := handlers.NewConfigHandler(db)
	dataSourceHandler := handlers.NewDataSourceHandler(configManager)
	serviceHandler := handlers.NewServiceHandler(db)

	// Setup server with all handlers
	server := &Server{
		db:                db,
		router:            router,
		middlewareHandler: middlewareHandler,
		resourceHandler:   resourceHandler,
		configHandler:     configHandler,
		dataSourceHandler: dataSourceHandler,
		serviceHandler:    serviceHandler,
		configManager:     configManager,
		srv: &http.Server{
			Addr:              ":" + config.Port,
			Handler:           router,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	// Configure routes
	server.setupRoutes(config.UIPath)

	return server
}

// setupRoutes configures all the routes for the API server
func (s *Server) setupRoutes(uiPath string) {
	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// API routes
	api := s.router.Group("/api")
	{
		// Middleware routes
		middlewares := api.Group("/middlewares")
		{
			middlewares.GET("", s.middlewareHandler.GetMiddlewares)
			middlewares.POST("", s.middlewareHandler.CreateMiddleware)
			middlewares.GET("/:id", s.middlewareHandler.GetMiddleware)
			middlewares.PUT("/:id", s.middlewareHandler.UpdateMiddleware)
			middlewares.DELETE("/:id", s.middlewareHandler.DeleteMiddleware)
		}

		// Service routes
		services := api.Group("/services")
		{
			services.GET("", s.serviceHandler.GetServices)
			services.POST("", s.serviceHandler.CreateService)
			services.GET("/:id", s.serviceHandler.GetService)
			services.PUT("/:id", s.serviceHandler.UpdateService)
			services.DELETE("/:id", s.serviceHandler.DeleteService)
		}

		// Resource routes
		resources := api.Group("/resources")
		{
			resources.GET("", s.resourceHandler.GetResources)
			resources.GET("/:id", s.resourceHandler.GetResource)
			resources.DELETE("/:id", s.resourceHandler.DeleteResource)
			
			// Middleware assignments
			resources.POST("/:id/middlewares", s.resourceHandler.AssignMiddleware)
			resources.POST("/:id/middlewares/bulk", s.resourceHandler.AssignMultipleMiddlewares)
			resources.DELETE("/:id/middlewares/:middlewareId", s.resourceHandler.RemoveMiddleware)
			
			// Service assignments
			resources.GET("/:id/service", s.serviceHandler.GetResourceService)
			resources.POST("/:id/service", s.serviceHandler.AssignServiceToResource)
			resources.DELETE("/:id/service", s.serviceHandler.RemoveServiceFromResource)
			
			// Router configuration routes
			resources.PUT("/:id/config/http", s.configHandler.UpdateHTTPConfig)
			resources.PUT("/:id/config/tls", s.configHandler.UpdateTLSConfig)
			resources.PUT("/:id/config/tcp", s.configHandler.UpdateTCPConfig)
			resources.PUT("/:id/config/headers", s.configHandler.UpdateHeadersConfig)
			resources.PUT("/:id/config/priority", s.configHandler.UpdateRouterPriority)
		}

		// Data source routes
		datasource := api.Group("/datasource")
		{
			datasource.GET("", s.dataSourceHandler.GetDataSources)
			datasource.GET("/active", s.dataSourceHandler.GetActiveDataSource)
			datasource.PUT("/active", s.dataSourceHandler.SetActiveDataSource)
			datasource.PUT("/:name", s.dataSourceHandler.UpdateDataSource)
			datasource.POST("/:name/test", s.dataSourceHandler.TestDataSourceConnection)
		}
	}

	// Serve the React app
	uiPathToUse := uiPath
	if uiPathToUse == "" {
		// Default UI path
		uiPathToUse = "/app/ui/build"
	}
	
	// Check if UI path exists and is a directory
	if stat, err := os.Stat(uiPathToUse); err == nil && stat.IsDir() {
		s.router.Use(static.Serve("/", static.LocalFile(uiPathToUse, false)))
		
		// Handle all other routes by serving the index.html file
		s.router.NoRoute(func(c *gin.Context) {
			// API routes should 404 when not found
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
				return
			}
			
			// Non-API routes serve the SPA
			c.File(uiPathToUse + "/index.html")
		})
	} else {
		log.Printf("Warning: UI path %s doesn't exist or is not a directory. Web UI will not be available.", uiPathToUse)
	}
}

// Start starts the API server with graceful shutdown
func (s *Server) Start() error {
	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("API server listening on %s", s.srv.Addr)
		serverErrors <- s.srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for either a server error or a signal.
	select {
	case err := <-serverErrors:
		// Non-nil error from ListenAndServe.
		return err

	case <-shutdown:
		log.Println("Shutdown signal received")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := s.srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout.
			log.Printf("Graceful shutdown failed: %v", err)
			if err := s.srv.Close(); err != nil {
				log.Printf("Error during forced shutdown: %v", err)
			}
			return err
		}

		log.Println("API server stopped gracefully")
	}

	return nil
}

// Stop gracefully stops the API server
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Printf("Failed to gracefully shutdown server: %v", err)
		if err := s.srv.Close(); err != nil {
			log.Printf("Error during forced shutdown: %v", err)
		}
	} else {
		log.Println("API server stopped gracefully")
	}
}

// minimalLogger returns a Gin middleware for minimal request logging
func minimalLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Log only when path is not being probed by health checkers
		if c.Request.URL.Path != "/health" && c.Request.URL.Path != "/ping" {
			// Log only requests with errors or non-standard responses
			if c.Writer.Status() >= 400 || len(c.Errors) > 0 {
				log.Printf("[GIN] %s | %d | %v | %s | %s",
					c.Request.Method,
					c.Writer.Status(),
					time.Since(start),
					c.ClientIP(),
					c.Request.URL.Path,
				)
			}
		}
	}
}