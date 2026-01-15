package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hhftechnology/middleware-manager/api"
	"github.com/hhftechnology/middleware-manager/config"
	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/services"
)

// Configuration represents the application configuration
type Configuration struct {
	PangolinAPIURL          string
	TraefikAPIURL           string
	TraefikConfDir          string
	DBPath                  string
	Port                    string
	UIPath                  string
	ConfigDir               string
	CheckInterval           time.Duration
	GenerateInterval        time.Duration
	ServiceInterval         time.Duration
	Debug                   bool
	AllowCORS               bool
	CORSOrigin              string
	ActiveDataSource        string
	TraefikStaticConfigPath string
}

// DiscoverTraefikAPI attempts to discover the Traefik API by trying common URLs
func DiscoverTraefikAPI() (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	urls := []string{
		"http://host.docker.internal:8080",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://traefik:8080",
	}

	for _, url := range urls {
		testURL := url + "/api/version"
		resp, err := client.Get(testURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Printf("Discovered Traefik API at %s", url)
			return url, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return "", nil
}

func main() {
    log.Println("Starting Middleware Manager...")

    var debug bool
    flag.BoolVar(&debug, "debug", false, "Enable debug mode")
    flag.Parse()

    cfg := loadConfiguration(debug)

    if os.Getenv("TRAEFIK_API_URL") == "" {
        if discoveredURL, err := DiscoverTraefikAPI(); err == nil && discoveredURL != "" {
            log.Printf("Auto-discovered Traefik API URL: %s", discoveredURL)
            cfg.TraefikAPIURL = discoveredURL
        }
    }

    db, err := database.InitDB(cfg.DBPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    configDir := cfg.ConfigDir
    if err := config.EnsureConfigDirectory(configDir); err != nil {
        log.Printf("Warning: Failed to create config directory: %v", err)
    }

    if err := config.SaveTemplateFile(configDir); err != nil {
        log.Printf("Warning: Failed to save default middleware templates: %v", err)
    }

    if err := config.LoadDefaultTemplates(db); err != nil {
        log.Printf("Warning: Failed to load default middleware templates: %v", err)
    }

    if err := config.SaveTemplateServicesFile(configDir); err != nil {
        log.Printf("Warning: Failed to save default service templates: %v", err)
    }

    if err := config.LoadDefaultServiceTemplates(db); err != nil {
        log.Printf("Warning: Failed to load default service templates: %v", err)
    }

    // Run comprehensive database cleanup on startup
    log.Println("Performing full database cleanup...")
    cleanupOpts := database.DefaultCleanupOptions()
    cleanupOpts.LogLevel = 2 // More verbose logging during startup
    
    if err := db.PerformFullCleanup(cleanupOpts); err != nil {
        log.Printf("Warning: Database cleanup encountered issues: %v", err)
    } else {
        log.Println("Database cleanup completed successfully")
    }

    configManager, err := services.NewConfigManager(filepath.Join(configDir, "config.json"))
    if err != nil {
        log.Fatalf("Failed to initialize config manager: %v", err)
    }

    configManager.EnsureDefaultDataSources(cfg.PangolinAPIURL, cfg.TraefikAPIURL)

    stopChan := make(chan struct{})

    resourceWatcher, err := services.NewResourceWatcher(db, configManager)
    if err != nil {
        log.Fatalf("Failed to create resource watcher: %v", err)
    }
    go resourceWatcher.Start(cfg.CheckInterval)

    configGenerator := services.NewConfigGenerator(db, cfg.TraefikConfDir, configManager)
    go configGenerator.Start(cfg.GenerateInterval)

    serverConfig := api.ServerConfig{
        Port:        cfg.Port,
        UIPath:      cfg.UIPath,
        Debug:       cfg.Debug,
        AllowCORS:   cfg.AllowCORS,
        CORSOrigin:  cfg.CORSOrigin,
        PangolinURL: cfg.PangolinAPIURL,
    }

    server := api.NewServer(db, serverConfig, configManager, cfg.TraefikStaticConfigPath)
    go func() {
        if err := server.Start(); err != nil {
            log.Printf("Server error: %v", err)
            close(stopChan)
        }
    }()

    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

    serviceWatcher, err := services.NewServiceWatcher(db, configManager)
    if err != nil {
        log.Printf("Warning: Failed to create service watcher: %v", err)
        serviceWatcher = nil
    } else {
        go serviceWatcher.Start(cfg.ServiceInterval)
    }

    select {
    case <-signalChan:
        log.Println("Received shutdown signal")
    case <-stopChan:
        log.Println("Received stop signal from server")
    }

    log.Println("Shutting down...")
    resourceWatcher.Stop()
    if serviceWatcher != nil {
        serviceWatcher.Stop()
    }
    configGenerator.Stop()
    server.Stop()
    log.Println("Middleware Manager stopped")
}

func loadConfiguration(debug bool) Configuration {
	checkInterval := 30 * time.Second
	if intervalStr := getEnv("CHECK_INTERVAL_SECONDS", "30"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			checkInterval = time.Duration(interval) * time.Second
		}
	}

	generateInterval := 10 * time.Second
	if intervalStr := getEnv("GENERATE_INTERVAL_SECONDS", "10"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			generateInterval = time.Duration(interval) * time.Second
		}
	}

	parsedServiceInterval := 30 * time.Second
	if intervalStr := getEnv("SERVICE_INTERVAL_SECONDS", "30"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			parsedServiceInterval = time.Duration(interval) * time.Second
		}
	}

	allowCORS := false
	if corsStr := getEnv("ALLOW_CORS", "false"); corsStr != "" {
		allowCORS = strings.ToLower(corsStr) == "true"
	}

	if debugStr := getEnv("DEBUG", ""); debugStr != "" {
		debug = strings.ToLower(debugStr) == "true"
	}

	return Configuration{
		PangolinAPIURL:          getEnv("PANGOLIN_API_URL", "http://pangolin:3001/api/v1"),
		TraefikAPIURL:           getEnv("TRAEFIK_API_URL", "http://host.docker.internal:8080"),
		TraefikConfDir:          getEnv("TRAEFIK_CONF_DIR", "/conf"),
		DBPath:                  getEnv("DB_PATH", "/data/middleware.db"),
		Port:                    getEnv("PORT", "3456"),
		UIPath:                  getEnv("UI_PATH", "/app/ui/dist"),
		ConfigDir:               getEnv("CONFIG_DIR", "/app/config"),
		ActiveDataSource:        getEnv("ACTIVE_DATA_SOURCE", "pangolin"),
		CheckInterval:           checkInterval,
		GenerateInterval:        generateInterval,
		ServiceInterval:         parsedServiceInterval,
		Debug:                   debug,
		AllowCORS:               allowCORS,
		CORSOrigin:              getEnv("CORS_ORIGIN", ""),
		TraefikStaticConfigPath: getEnv("TRAEFIK_STATIC_CONFIG_PATH", "/etc/traefik/traefik.yml"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}