package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hhftechnology/middleware-manager/database"
	"gopkg.in/yaml.v3"
)

// DefaultService represents a default service template
type DefaultService struct {
	ID     string                 `yaml:"id"`
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// DefaultServiceTemplates represents the structure of the templates_services.yaml file
type DefaultServiceTemplates struct {
	Services []DefaultService `yaml:"services"`
}

// LoadDefaultServiceTemplates loads the default service templates
func LoadDefaultServiceTemplates(db *database.DB) error {
	// Determine the path to the templates file
	templatesFile := "config/templates_services.yaml"
	
	// Check if the file exists in the current directory
	if _, err := os.Stat(templatesFile); os.IsNotExist(err) {
		// Try to find it in different locations
		possiblePaths := []string{
			"/app/config/templates_services.yaml",  // Docker container path
			"templates_services.yaml",              // Current directory
		}
		
		found := false
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				templatesFile = path
				found = true
				break
			}
		}
		
		if !found {
			log.Printf("Warning: templates_services.yaml not found, skipping default service templates")
			return nil
		}
	}
	
	// Read the templates file
	data, err := os.ReadFile(templatesFile)
	if err != nil {
		return err
	}
	
	// Parse the YAML
	var templates DefaultServiceTemplates
	if err := yaml.Unmarshal(data, &templates); err != nil {
		return err
	}
	
	// Process templates to ensure proper value preservation
	for i := range templates.Services {
		// Apply service-specific processing based on type
		switch templates.Services[i].Type {
		case "loadBalancer":
			processLoadBalancerService(&templates.Services[i].Config)
		case "weighted":
			processWeightedService(&templates.Services[i].Config)
		case "mirroring":
			processMirroringService(&templates.Services[i].Config)
		case "failover":
			processFailoverService(&templates.Services[i].Config)
		default:
			// General processing for other service types
			templates.Services[i].Config = preserveTraefikValues(templates.Services[i].Config).(map[string]interface{})
		}
	}

	// Add templates to the database if they don't exist
	for _, service := range templates.Services {
		// Check if the service already exists
		var exists int
		err := db.QueryRow("SELECT 1 FROM services WHERE id = ?", service.ID).Scan(&exists)
		if err == nil {
			// Service exists, skip
			continue
		}
		
		// Convert config to JSON string
		configJSON, err := json.Marshal(service.Config)
		if err != nil {
			log.Printf("Failed to marshal config for %s: %v", service.Name, err)
			continue
		}
		
		// Insert into database
		_, err = db.Exec(
			"INSERT INTO services (id, name, type, config) VALUES (?, ?, ?, ?)",
			service.ID, service.Name, service.Type, string(configJSON),
		)
		
		if err != nil {
			log.Printf("Failed to insert service %s: %v", service.Name, err)
			continue
		}
		
		log.Printf("Added default service: %s", service.Name)
	}
	
	return nil
}

// processLoadBalancerService handles loadBalancer service special processing
func processLoadBalancerService(config *map[string]interface{}) {
	// Process servers array
	if servers, ok := (*config)["servers"].([]interface{}); ok {
		for i, server := range servers {
			if serverMap, ok := server.(map[string]interface{}); ok {
				// Process URL if present
				if url, ok := serverMap["url"].(string); ok && url != "" {
					serverMap["url"] = url
				}
				
				// Process address if present
				if address, ok := serverMap["address"].(string); ok && address != "" {
					serverMap["address"] = address
				}
				
				// Process weight if present
				if weight, ok := serverMap["weight"].(float64); ok {
					if weight == float64(int(weight)) {
						serverMap["weight"] = int(weight)
					}
				}
				
				// Process tls flag if present
				if tls, ok := serverMap["tls"].(bool); ok {
					serverMap["tls"] = tls
				}
				
				// Process preservePath if present
				if preservePath, ok := serverMap["preservePath"].(bool); ok {
					serverMap["preservePath"] = preservePath
				}
				
				servers[i] = serverMap
			}
		}
	}
	
	// Process healthCheck if present
	if healthCheck, ok := (*config)["healthCheck"].(map[string]interface{}); ok {
		// Process path
		if path, ok := healthCheck["path"].(string); ok && path != "" {
			healthCheck["path"] = path
		}
		
		// Process interval
		if interval, ok := healthCheck["interval"].(string); ok && interval != "" {
			healthCheck["interval"] = interval
		}
		
		// Process timeout
		if timeout, ok := healthCheck["timeout"].(string); ok && timeout != "" {
			healthCheck["timeout"] = timeout
		}
		
		// Process port
		if port, ok := healthCheck["port"].(float64); ok {
			if port == float64(int(port)) {
				healthCheck["port"] = int(port)
			}
		}
		
		// Process scheme
		if scheme, ok := healthCheck["scheme"].(string); ok && scheme != "" {
			healthCheck["scheme"] = scheme
		}
		
		// Process headers
		if headers, ok := healthCheck["headers"].(map[string]interface{}); ok {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					headers[key] = strValue
				}
			}
		}
	}
	
	// Process sticky if present
	if sticky, ok := (*config)["sticky"].(map[string]interface{}); ok {
		// Process cookie if present
		if cookie, ok := sticky["cookie"].(map[string]interface{}); ok {
			// Process name
			if name, ok := cookie["name"].(string); ok && name != "" {
				cookie["name"] = name
			}
			
			// Process secure flag
			if secure, ok := cookie["secure"].(bool); ok {
				cookie["secure"] = secure
			}
			
			// Process httpOnly flag
			if httpOnly, ok := cookie["httpOnly"].(bool); ok {
				cookie["httpOnly"] = httpOnly
			}
		}
	}
	
	// Process passHostHeader flag
	if passHostHeader, ok := (*config)["passHostHeader"].(bool); ok {
		(*config)["passHostHeader"] = passHostHeader
	}
	
	// Process responseForwarding if present
	if responseForwarding, ok := (*config)["responseForwarding"].(map[string]interface{}); ok {
		// Process flushInterval
		if flushInterval, ok := responseForwarding["flushInterval"].(string); ok && flushInterval != "" {
			responseForwarding["flushInterval"] = flushInterval
		}
	}
	
	// Process other fields
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processWeightedService handles weighted service special processing
func processWeightedService(config *map[string]interface{}) {
	// Process services array
	if services, ok := (*config)["services"].([]interface{}); ok {
		for i, service := range services {
			if serviceMap, ok := service.(map[string]interface{}); ok {
				// Process name
				if name, ok := serviceMap["name"].(string); ok && name != "" {
					serviceMap["name"] = name
				}
				
				// Process weight
				if weight, ok := serviceMap["weight"].(float64); ok {
					if weight == float64(int(weight)) {
						serviceMap["weight"] = int(weight)
					}
				}
				
				services[i] = serviceMap
			}
		}
	}
	
	// Process healthCheck if present
	if healthCheck, ok := (*config)["healthCheck"].(map[string]interface{}); ok {
		// Just ensure it's preserved if empty
		(*config)["healthCheck"] = healthCheck
	}
	
	// Process sticky if present
	if sticky, ok := (*config)["sticky"].(map[string]interface{}); ok {
		// Process cookie if present
		if cookie, ok := sticky["cookie"].(map[string]interface{}); ok {
			// Process name
			if name, ok := cookie["name"].(string); ok && name != "" {
				cookie["name"] = name
			}
			
			// Process secure flag
			if secure, ok := cookie["secure"].(bool); ok {
				cookie["secure"] = secure
			}
			
			// Process httpOnly flag
			if httpOnly, ok := cookie["httpOnly"].(bool); ok {
				cookie["httpOnly"] = httpOnly
			}
		}
	}
	
	// Process other fields
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processMirroringService handles mirroring service special processing
func processMirroringService(config *map[string]interface{}) {
	// Process service name
	if service, ok := (*config)["service"].(string); ok && service != "" {
		(*config)["service"] = service
	}
	
	// Process mirrors array
	if mirrors, ok := (*config)["mirrors"].([]interface{}); ok {
		for i, mirror := range mirrors {
			if mirrorMap, ok := mirror.(map[string]interface{}); ok {
				// Process name
				if name, ok := mirrorMap["name"].(string); ok && name != "" {
					mirrorMap["name"] = name
				}
				
				// Process percent
				if percent, ok := mirrorMap["percent"].(float64); ok {
					if percent == float64(int(percent)) {
						mirrorMap["percent"] = int(percent)
					}
				}
				
				mirrors[i] = mirrorMap
			}
		}
	}
	
	// Process mirrorBody flag
	if mirrorBody, ok := (*config)["mirrorBody"].(bool); ok {
		(*config)["mirrorBody"] = mirrorBody
	}
	
	// Process maxBodySize
	if maxBodySize, ok := (*config)["maxBodySize"].(float64); ok {
		if maxBodySize == float64(int(maxBodySize)) {
			(*config)["maxBodySize"] = int(maxBodySize)
		}
	}
	
	// Process healthCheck if present
	if healthCheck, ok := (*config)["healthCheck"].(map[string]interface{}); ok {
		// Just ensure it's preserved if empty
		(*config)["healthCheck"] = healthCheck
	}
	
	// Process other fields
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// processFailoverService handles failover service special processing
func processFailoverService(config *map[string]interface{}) {
	// Process service name
	if service, ok := (*config)["service"].(string); ok && service != "" {
		(*config)["service"] = service
	}
	
	// Process fallback name
	if fallback, ok := (*config)["fallback"].(string); ok && fallback != "" {
		(*config)["fallback"] = fallback
	}
	
	// Process healthCheck if present
	if healthCheck, ok := (*config)["healthCheck"].(map[string]interface{}); ok {
		// Just ensure it's preserved if empty
		(*config)["healthCheck"] = healthCheck
	}
	
	// Process other fields
	*config = preserveTraefikValues(*config).(map[string]interface{})
}

// SaveTemplateServicesFile saves the default services templates file if it doesn't exist
func SaveTemplateServicesFile(templatesDir string) error {
	templatesFile := filepath.Join(templatesDir, "templates_services.yaml")
	
	// Check if file already exists
	if _, err := os.Stat(templatesFile); err == nil {
		// File exists, skip
		return nil
	}
	
	// Create default templates
	templates := DefaultServiceTemplates{
		Services: []DefaultService{
			// LoadBalancer services
			{
				ID:   "simple-http",
				Name: "Simple HTTP LoadBalancer",
				Type: "loadBalancer",
				Config: map[string]interface{}{
					"servers": []map[string]interface{}{
						{
							"url": "http://localhost:8080",
						},
					},
				},
			},
			{
				ID:   "multi-server-http",
				Name: "Multi-Server HTTP LoadBalancer",
				Type: "loadBalancer",
				Config: map[string]interface{}{
					"servers": []map[string]interface{}{
						{
							"url": "http://server1:8080",
						},
						{
							"url": "http://server2:8080",
						},
					},
				},
			},
			{
				ID:   "health-check",
				Name: "HTTP Service with Health Check",
				Type: "loadBalancer",
				Config: map[string]interface{}{
					"servers": []map[string]interface{}{
						{
							"url": "http://backend:8080",
						},
					},
					"healthCheck": map[string]interface{}{
						"path":     "/health",
						"interval": "10s",
						"timeout":  "3s",
					},
				},
			},
			
			// TCP service
			{
				ID:   "tcp-service",
				Name: "TCP Service",
				Type: "loadBalancer",
				Config: map[string]interface{}{
					"servers": []map[string]interface{}{
						{
							"address": "backend:9000",
						},
					},
				},
			},
			
			// UDP service
			{
				ID:   "udp-service",
				Name: "UDP Service",
				Type: "loadBalancer",
				Config: map[string]interface{}{
					"servers": []map[string]interface{}{
						{
							"address": "backend:53",
						},
					},
				},
			},
			
			// Weighted service
			{
				ID:   "weighted-service",
				Name: "Weighted Service",
				Type: "weighted",
				Config: map[string]interface{}{
					"services": []map[string]interface{}{
						{
							"name":   "service1@file",
							"weight": 3,
						},
						{
							"name":   "service2@file",
							"weight": 1,
						},
					},
				},
			},
			
			// Mirroring service
			{
				ID:   "traffic-mirror",
				Name: "Traffic Mirroring Service",
				Type: "mirroring",
				Config: map[string]interface{}{
					"service": "main-service@file",
					"mirrors": []map[string]interface{}{
						{
							"name":    "test-service@file",
							"percent": 10,
						},
					},
				},
			},
			
			// Failover service
			{
				ID:   "failover-service",
				Name: "Failover Service",
				Type: "failover",
				Config: map[string]interface{}{
					"service":  "main-service@file",
					"fallback": "backup-service@file",
				},
			},
		},
	}
	
	// Process all templates to ensure proper value preservation
	for i := range templates.Services {
		// Apply service-specific processing based on type
		switch templates.Services[i].Type {
		case "loadBalancer":
			processLoadBalancerService(&templates.Services[i].Config)
		case "weighted":
			processWeightedService(&templates.Services[i].Config)
		case "mirroring":
			processMirroringService(&templates.Services[i].Config)
		case "failover":
			processFailoverService(&templates.Services[i].Config)
		default:
			// General processing for other service types
			templates.Services[i].Config = preserveTraefikValues(templates.Services[i].Config).(map[string]interface{})
		}
	}
	
	// Create a custom YAML encoder that preserves string formatting
	yamlNode := &yaml.Node{}
	err := yamlNode.Encode(templates)
	if err != nil {
		return fmt.Errorf("failed to encode templates to YAML node: %w", err)
	}
	
	// Apply additional string preservation to the YAML node
	preserveStringsInYamlNode(yamlNode)
	
	// Marshal the processed node
	data, err := yaml.Marshal(yamlNode)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML node: %w", err)
	}
	
	// Write to file
	return os.WriteFile(templatesFile, data, 0644)
}