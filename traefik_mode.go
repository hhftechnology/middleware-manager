package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmserver "github.com/hhftechnology/middleware-manager/internal/traefikmanager/server"
)

func startTraefikManager(debug bool) {
	cfg := tmconfig.LoadRuntimeConfig(debug)
	files, err := tmconfig.NewFileStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Traefik Manager file store: %v", err)
	}
	settings := tmconfig.NewSettingsStore(cfg)
	dashboard := tmconfig.NewDashboardStore(cfg)
	client := &http.Client{Timeout: cfg.HTTPClientTimeout}
	server := tmserver.New(cfg, files, settings, dashboard, client)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Traefik Manager server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down Traefik Manager...")
	if err := server.Stop(); err != nil {
		log.Printf("Traefik Manager shutdown error: %v", err)
	}
}
