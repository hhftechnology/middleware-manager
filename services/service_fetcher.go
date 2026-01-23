package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hhftechnology/middleware-manager/models"
)

// ServiceFetcher defines the interface for fetching services
type ServiceFetcher interface {
	FetchServices(ctx context.Context) (*models.ServiceCollection, error)
}

// ServiceFetcherFactory creates the appropriate service fetcher based on type
func NewServiceFetcher(config models.DataSourceConfig) (ServiceFetcher, error) {
	switch config.Type {
	case models.PangolinAPI:
		return NewPangolinServiceFetcher(config), nil
	case models.TraefikAPI:
		return NewTraefikServiceFetcher(config), nil
	default:
		return nil, fmt.Errorf("unknown data source type: %s", config.Type)
	}
}

// PangolinServiceFetcher fetches services from Pangolin API
type PangolinServiceFetcher struct {
	config     models.DataSourceConfig
	httpClient *http.Client
}

// NewPangolinServiceFetcher creates a new Pangolin API fetcher for services
func NewPangolinServiceFetcher(config models.DataSourceConfig) *PangolinServiceFetcher {
	return &PangolinServiceFetcher{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchServices fetches services from Pangolin API
func (f *PangolinServiceFetcher) FetchServices(ctx context.Context) (*models.ServiceCollection, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.config.URL+"/traefik-config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if configured
	if f.config.BasicAuth.Username != "" {
		req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
	}

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Process response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the Pangolin config (which includes services)
	var config models.PangolinTraefikConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert Pangolin services to our internal model
	services := &models.ServiceCollection{
		Services: make([]models.Service, 0),
	}

	// Extract HTTP services
	for id, service := range config.HTTP.Services {
		// Skip system services
		if isPangolinSystemService(id) {
			continue
		}

		// Extract service configuration
		serviceConfig := make(map[string]interface{})

		// Determine service type based on structure
		serviceType := determineServiceType(service)

		// Extract the appropriate configuration based on type
		switch serviceType {
		case string(models.LoadBalancerType):
			if lb, ok := service.LoadBalancer.(map[string]interface{}); ok {
				serviceConfig = lb
			}
		case string(models.WeightedType):
			if w, ok := service.Weighted.(map[string]interface{}); ok {
				serviceConfig = w
			}
		case string(models.MirroringType):
			if m, ok := service.Mirroring.(map[string]interface{}); ok {
				serviceConfig = m
			}
		case string(models.FailoverType):
			if f, ok := service.Failover.(map[string]interface{}); ok {
				serviceConfig = f
			}
		}

		// Create new service
		configJSON, _ := json.Marshal(serviceConfig)

		newService := models.Service{
			ID:        id,
			Name:      id, // Use ID as name by default
			Type:      serviceType,
			Config:    string(configJSON),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		services.Services = append(services.Services, newService)
	}

	// TODO: Extract TCP and UDP services if they exist in the Pangolin API response

	log.Printf("Fetched %d services from Pangolin API", len(services.Services))
	return services, nil
}

// determineServiceType determines the service type from the service structure
func determineServiceType(service models.PangolinService) string {
	// Check for various service types
	if service.LoadBalancer != nil {
		return string(models.LoadBalancerType)
	}
	if service.Weighted != nil {
		return string(models.WeightedType)
	}
	if service.Mirroring != nil {
		return string(models.MirroringType)
	}
	if service.Failover != nil {
		return string(models.FailoverType)
	}

	// Default to LoadBalancer if can't determine
	return string(models.LoadBalancerType)
}

// isPangolinSystemService checks if a service is a Pangolin system service (to be skipped)
func isPangolinSystemService(serviceID string) bool {
	systemPrefixes := []string{
		"api-service",
		"next-service",
		"noop",
	}

	for _, prefix := range systemPrefixes {
		if strings.Contains(serviceID, prefix) {
			return true
		}
	}

	return false
}

// TraefikServiceFetcher fetches services from Traefik API
type TraefikServiceFetcher struct {
	config     models.DataSourceConfig
	httpClient *http.Client
}

// NewTraefikServiceFetcher creates a new Traefik API fetcher for services
func NewTraefikServiceFetcher(config models.DataSourceConfig) *TraefikServiceFetcher {
	return &TraefikServiceFetcher{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchServices fetches services from Traefik API with fallback options
func (f *TraefikServiceFetcher) FetchServices(ctx context.Context) (*models.ServiceCollection, error) {
	log.Println("Fetching services from Traefik API...")

	// Try the configured URL first
	services, err := f.fetchServicesFromURL(ctx, f.config.URL)
	if err == nil {
		log.Printf("Successfully fetched services from %s", f.config.URL)
		return services, nil
	}

	// Log the initial error
	log.Printf("Failed to connect to primary Traefik API URL %s: %v", f.config.URL, err)

	// Try common fallback URLs
	fallbackURLs := []string{
		"http://traefik:8080",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://host.docker.internal:8080",
	}

	// Don't try the same URL twice
	if f.config.URL != "" {
		for i := len(fallbackURLs) - 1; i >= 0; i-- {
			if fallbackURLs[i] == f.config.URL {
				fallbackURLs = append(fallbackURLs[:i], fallbackURLs[i+1:]...)
			}
		}
	}

	// Try each fallback URL
	var lastErr error
	for _, url := range fallbackURLs {
		log.Printf("Trying fallback Traefik API URL for services: %s", url)
		services, err := f.fetchServicesFromURL(ctx, url)
		if err == nil {
			// Success with fallback - remember this URL for next time
			f.suggestURLUpdate(url)
			return services, nil
		}
		lastErr = err
		log.Printf("Fallback URL %s failed: %v", url, err)
	}

	// All fallbacks failed
	return nil, fmt.Errorf("all Traefik API connection attempts failed, last error: %w", lastErr)
}

// fetchServicesFromURL fetches services from a specific URL
func (f *TraefikServiceFetcher) fetchServicesFromURL(ctx context.Context, baseURL string) (*models.ServiceCollection, error) {
	// Fetch HTTP services
	httpServices, err := f.fetchHTTPServices(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTTP services: %w", err)
	}

	// Try to fetch TCP services if available
	tcpServices, err := f.fetchTCPServices(ctx, baseURL)
	if err != nil {
		// Log but don't fail - TCP services are optional
		log.Printf("Warning: Failed to fetch TCP services: %v", err)
	}

	// Try to fetch UDP services if available (may not be supported in all Traefik versions)
	udpServices, err := f.fetchUDPServices(ctx, baseURL)
	if err != nil {
		// Log but don't fail - UDP services are optional
		log.Printf("Warning: Failed to fetch UDP services: %v", err)
	}

	// Combine all services
	services := &models.ServiceCollection{
		Services: make([]models.Service, 0, len(httpServices)+len(tcpServices)+len(udpServices)),
	}

	// Add HTTP services
	services.Services = append(services.Services, httpServices...)

	// Add TCP services
	services.Services = append(services.Services, tcpServices...)

	// Add UDP services
	services.Services = append(services.Services, udpServices...)

	log.Printf("Fetched %d total services from Traefik API (%d HTTP, %d TCP, %d UDP)",
		len(services.Services), len(httpServices), len(tcpServices), len(udpServices))

	return services, nil
}

// Update the fetchHTTPServices function with these changes:

func (f *TraefikServiceFetcher) fetchHTTPServices(ctx context.Context, baseURL string) ([]models.Service, error) {
	// Create HTTP request to fetch HTTP services
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/http/services", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if configured
	if f.config.BasicAuth.Username != "" {
		req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
	}

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// First try to parse as an array of services
	var traefikServicesArray []models.TraefikService
	err = json.Unmarshal(body, &traefikServicesArray)

	services := make([]models.Service, 0)

	if err == nil {
		// Successfully parsed as array
		for _, traefikService := range traefikServicesArray {
			// Skip internal services
			if traefikService.Provider == "internal" {
				continue
			}

			// Process each service
			service := processTraefikService(traefikService)
			if service != nil {
				services = append(services, *service)
			}
		}
	} else {
		// Try parsing as a map
		var traefikServicesMap map[string]models.TraefikService
		if jsonErr := json.Unmarshal(body, &traefikServicesMap); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse services JSON: %w", jsonErr)
		}

		// Process each service in the map
		for name, traefikService := range traefikServicesMap {
			// Skip internal services
			if traefikService.Provider == "internal" {
				continue
			}

			// Set the name from the map key
			traefikService.Name = name

			// Process the service
			service := processTraefikService(traefikService)
			if service != nil {
				services = append(services, *service)
			}
		}
	}

	return services, nil
}

// fetchTCPServices fetches TCP services from Traefik API
func (f *TraefikServiceFetcher) fetchTCPServices(ctx context.Context, baseURL string) ([]models.Service, error) {
	// Create HTTP request to fetch TCP services
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/tcp/services", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if configured
	if f.config.BasicAuth.Username != "" {
		req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
	}

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response (similar to HTTP services but adapt for TCP)
	// This is a simplified implementation - would need to adapt to actual TCP service structure

	services := make([]models.Service, 0)

	// Try parsing as an array first
	var tcpServicesArray []interface{}
	err = json.Unmarshal(body, &tcpServicesArray)

	if err == nil {
		// Successfully parsed as array
		for i, tcpService := range tcpServicesArray {
			serviceMap, ok := tcpService.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip internal services
			provider, _ := serviceMap["provider"].(string)
			if provider == "internal" {
				continue
			}

			name, _ := serviceMap["name"].(string)
			if name == "" {
				name = fmt.Sprintf("tcp-service-%d", i)
			}

			// Extract loadBalancer config
			var config map[string]interface{}
			if lb, ok := serviceMap["loadBalancer"].(map[string]interface{}); ok {
				config = lb
			} else {
				// Try other service types if needed
				config = serviceMap
			}

			// Create service
			configJSON, _ := json.Marshal(config)

			services = append(services, models.Service{
				ID:        name,
				Name:      name,
				Type:      string(models.LoadBalancerType), // Most TCP services are loadBalancers
				Config:    string(configJSON),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	} else {
		// Try parsing as a map
		var tcpServicesMap map[string]interface{}
		if jsonErr := json.Unmarshal(body, &tcpServicesMap); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse TCP services JSON: %w", jsonErr)
		}

		// Process each service in the map
		for name, tcpService := range tcpServicesMap {
			serviceMap, ok := tcpService.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip internal services
			provider, _ := serviceMap["provider"].(string)
			if provider == "internal" {
				continue
			}

			// Extract loadBalancer config
			var config map[string]interface{}
			if lb, ok := serviceMap["loadBalancer"].(map[string]interface{}); ok {
				config = lb
			} else {
				// Try other service types if needed
				config = serviceMap
			}

			// Create service
			configJSON, _ := json.Marshal(config)

			services = append(services, models.Service{
				ID:        name,
				Name:      name,
				Type:      string(models.LoadBalancerType), // Most TCP services are loadBalancers
				Config:    string(configJSON),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	return services, nil
}

// fetchUDPServices fetches UDP services from Traefik API
func (f *TraefikServiceFetcher) fetchUDPServices(ctx context.Context, baseURL string) ([]models.Service, error) {
	// Create HTTP request to fetch UDP services
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/udp/services", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if configured
	if f.config.BasicAuth.Username != "" {
		req.SetBasicAuth(f.config.BasicAuth.Username, f.config.BasicAuth.Password)
	}

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Some Traefik versions may not support UDP services
		if resp.StatusCode == http.StatusNotFound {
			return []models.Service{}, nil
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Similar to TCP services but adapted for UDP
	services := make([]models.Service, 0)

	// Try parsing as an array first
	var udpServicesArray []interface{}
	err = json.Unmarshal(body, &udpServicesArray)

	if err == nil {
		// Successfully parsed as array
		for i, udpService := range udpServicesArray {
			serviceMap, ok := udpService.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip internal services
			provider, _ := serviceMap["provider"].(string)
			if provider == "internal" {
				continue
			}

			name, _ := serviceMap["name"].(string)
			if name == "" {
				name = fmt.Sprintf("udp-service-%d", i)
			}

			// Extract loadBalancer config
			var config map[string]interface{}
			if lb, ok := serviceMap["loadBalancer"].(map[string]interface{}); ok {
				config = lb
			} else {
				// Try other service types if needed
				config = serviceMap
			}

			// Create service
			configJSON, _ := json.Marshal(config)

			services = append(services, models.Service{
				ID:        name,
				Name:      name,
				Type:      string(models.LoadBalancerType), // Most UDP services are loadBalancers
				Config:    string(configJSON),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	} else {
		// Try parsing as a map
		var udpServicesMap map[string]interface{}
		if jsonErr := json.Unmarshal(body, &udpServicesMap); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse UDP services JSON: %w", jsonErr)
		}

		// Process each service in the map
		for name, udpService := range udpServicesMap {
			serviceMap, ok := udpService.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip internal services
			provider, _ := serviceMap["provider"].(string)
			if provider == "internal" {
				continue
			}

			// Extract loadBalancer config
			var config map[string]interface{}
			if lb, ok := serviceMap["loadBalancer"].(map[string]interface{}); ok {
				config = lb
			} else {
				// Try other service types if needed
				config = serviceMap
			}

			// Create service
			configJSON, _ := json.Marshal(config)

			services = append(services, models.Service{
				ID:        name,
				Name:      name,
				Type:      string(models.LoadBalancerType), // Most UDP services are loadBalancers
				Config:    string(configJSON),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	return services, nil
}

// processTraefikService extracts service information from Traefik API response
// Update the processTraefikService function signature and implementation:

func processTraefikService(traefikService models.TraefikService) *models.Service {
	// Skip system services
	if isTraefikSystemService(traefikService.Name) {
		return nil
	}

	// Determine service type and extract config
	serviceType := string(models.LoadBalancerType) // Default
	var config map[string]interface{}

	if traefikService.LoadBalancer != nil {
		// Most common case: LoadBalancer
		config = make(map[string]interface{})

		// Extract servers
		if traefikService.LoadBalancer.Servers != nil {
			config["servers"] = traefikService.LoadBalancer.Servers
		}

		// Extract other loadbalancer properties
		if traefikService.LoadBalancer.PassHostHeader != nil {
			config["passHostHeader"] = traefikService.LoadBalancer.PassHostHeader
		}

		if traefikService.LoadBalancer.Sticky != nil {
			config["sticky"] = traefikService.LoadBalancer.Sticky
		}

		if traefikService.LoadBalancer.HealthCheck != nil {
			config["healthCheck"] = traefikService.LoadBalancer.HealthCheck
		}
	} else if traefikService.Weighted != nil {
		// Weighted service
		serviceType = string(models.WeightedType)
		config = make(map[string]interface{})

		// Extract weighted service properties
		if traefikService.Weighted.Services != nil {
			config["services"] = traefikService.Weighted.Services
		}

		if traefikService.Weighted.Sticky != nil {
			config["sticky"] = traefikService.Weighted.Sticky
		}

		if traefikService.Weighted.HealthCheck != nil {
			config["healthCheck"] = traefikService.Weighted.HealthCheck
		}
	} else if traefikService.Mirroring != nil {
		// Mirroring service
		serviceType = string(models.MirroringType)
		config = make(map[string]interface{})

		// Extract mirroring service properties
		if traefikService.Mirroring.Service != "" {
			config["service"] = traefikService.Mirroring.Service
		}

		if traefikService.Mirroring.Mirrors != nil {
			config["mirrors"] = traefikService.Mirroring.Mirrors
		}

		if traefikService.Mirroring.MaxBodySize != nil {
			config["maxBodySize"] = traefikService.Mirroring.MaxBodySize
		}

		if traefikService.Mirroring.MirrorBody != nil {
			config["mirrorBody"] = traefikService.Mirroring.MirrorBody
		}

		if traefikService.Mirroring.HealthCheck != nil {
			config["healthCheck"] = traefikService.Mirroring.HealthCheck
		}
	} else if traefikService.Failover != nil {
		// Failover service
		serviceType = string(models.FailoverType)
		config = make(map[string]interface{})

		// Extract failover service properties
		if traefikService.Failover.Service != "" {
			config["service"] = traefikService.Failover.Service
		}

		if traefikService.Failover.Fallback != "" {
			config["fallback"] = traefikService.Failover.Fallback
		}

		if traefikService.Failover.HealthCheck != nil {
			config["healthCheck"] = traefikService.Failover.HealthCheck
		}
	} else {
		// Unknown service type or empty config
		config = make(map[string]interface{})
	}

	// Convert config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		log.Printf("Error marshaling service config: %v", err)
		configJSON = []byte("{}")
	}

	return &models.Service{
		ID:        traefikService.Name,
		Name:      traefikService.Name,
		Type:      serviceType,
		Config:    string(configJSON),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// suggestURLUpdate logs a message suggesting the URL should be updated
func (f *TraefikServiceFetcher) suggestURLUpdate(workingURL string) {
	log.Printf("IMPORTANT: Consider updating the Traefik API URL to %s in the settings", workingURL)
}

// isTraefikSystemService checks if a service is a Traefik system service (to be skipped)
func isTraefikSystemService(serviceID string) bool {
	systemPrefixes := []string{
		"api@internal",
		"dashboard@internal",
		"noop@internal",
		"acme-http@internal",
	}

	for _, prefix := range systemPrefixes {
		if strings.Contains(serviceID, prefix) {
			return true
		}
	}

	return false
}
