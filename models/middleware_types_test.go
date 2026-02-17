package models

import (
	"testing"
)

func TestGetProcessor(t *testing.T) {
	knownTypes := []string{
		"headers", "basicAuth", "forwardAuth", "digestAuth",
		"redirectRegex", "redirectScheme", "replacePath",
		"replacePathRegex", "stripPrefix", "stripPrefixRegex",
		"chain", "plugin", "rateLimit", "inFlightReq",
		"ipAllowList", "buffering",
	}

	for _, mwType := range knownTypes {
		t.Run(mwType, func(t *testing.T) {
			p := GetProcessor(mwType)
			if p == nil {
				t.Errorf("GetProcessor(%q) returned nil", mwType)
			}
			if _, isDefault := p.(*DefaultProcessor); isDefault {
				t.Errorf("GetProcessor(%q) returned DefaultProcessor, want specific", mwType)
			}
		})
	}

	t.Run("unknown returns DefaultProcessor", func(t *testing.T) {
		p := GetProcessor("unknownType")
		if _, ok := p.(*DefaultProcessor); !ok {
			t.Error("GetProcessor(unknown) should return DefaultProcessor")
		}
	})
}

func TestProcessMiddlewareConfig(t *testing.T) {
	config := map[string]interface{}{"key": "value"}
	result := ProcessMiddlewareConfig("headers", config)
	if result == nil {
		t.Error("ProcessMiddlewareConfig returned nil")
	}
}

func TestBufferingProcessor_Process(t *testing.T) {
	p := &BufferingProcessor{}

	t.Run("float64 whole numbers converted to int64", func(t *testing.T) {
		config := map[string]interface{}{
			"maxRequestBodyBytes":  float64(1048576),
			"memRequestBodyBytes":  float64(524288),
			"maxResponseBodyBytes": float64(2097152),
			"memResponseBodyBytes": float64(1048576),
		}
		result := p.Process(config)

		for _, field := range []string{"maxRequestBodyBytes", "memRequestBodyBytes", "maxResponseBodyBytes", "memResponseBodyBytes"} {
			if _, ok := result[field].(int64); !ok {
				t.Errorf("%s should be int64, got %T", field, result[field])
			}
		}
	})

	t.Run("non-whole float preserved", func(t *testing.T) {
		config := map[string]interface{}{
			"maxRequestBodyBytes": float64(1.5),
		}
		result := p.Process(config)
		if _, ok := result["maxRequestBodyBytes"].(float64); !ok {
			t.Errorf("non-whole float should remain float64, got %T", result["maxRequestBodyBytes"])
		}
	})
}

func TestHeadersProcessor_Process(t *testing.T) {
	p := &HeadersProcessor{}

	config := map[string]interface{}{
		"customResponseHeaders": map[string]interface{}{
			"X-Custom": "value",
			"Server":   "",
		},
		"customRequestHeaders": map[string]interface{}{
			"Authorization": "Bearer token",
		},
		"contentSecurityPolicy": "default-src 'self'",
		"referrerPolicy":        "strict-origin",
	}

	result := p.Process(config)

	respHeaders := result["customResponseHeaders"].(map[string]interface{})
	if respHeaders["X-Custom"] != "value" {
		t.Errorf("X-Custom = %v, want value", respHeaders["X-Custom"])
	}
	if respHeaders["Server"] != "" {
		t.Errorf("Server = %v, want empty string", respHeaders["Server"])
	}
}

func TestAuthProcessor_Process(t *testing.T) {
	p := &AuthProcessor{}

	config := map[string]interface{}{
		"address":            "http://auth-service:9000",
		"trustForwardHeader": true,
		"users":              []interface{}{"admin:$apr1$hash"},
	}

	result := p.Process(config)

	if result["address"] != "http://auth-service:9000" {
		t.Errorf("address = %v", result["address"])
	}
	if result["trustForwardHeader"] != true {
		t.Errorf("trustForwardHeader = %v", result["trustForwardHeader"])
	}
}

func TestPathProcessor_Process(t *testing.T) {
	p := &PathProcessor{}

	config := map[string]interface{}{
		"regex":       `^/old/(.*)`,
		"replacement": "/new/$1",
		"prefixes":    []interface{}{"/api", "/web"},
		"permanent":   true,
		"forceSlash":  false,
	}

	result := p.Process(config)

	if result["regex"] != `^/old/(.*)` {
		t.Errorf("regex = %v", result["regex"])
	}
	if result["replacement"] != "/new/$1" {
		t.Errorf("replacement = %v", result["replacement"])
	}
	if result["permanent"] != true {
		t.Errorf("permanent = %v", result["permanent"])
	}
}

func TestChainProcessor_Process(t *testing.T) {
	p := &ChainProcessor{}

	config := map[string]interface{}{
		"middlewares": []interface{}{"auth", "rate-limit", "headers"},
	}

	result := p.Process(config)

	mws, ok := result["middlewares"].([]interface{})
	if !ok {
		t.Fatal("middlewares should be []interface{}")
	}
	if len(mws) != 3 {
		t.Errorf("len(middlewares) = %d, want 3", len(mws))
	}
}

func TestRateLimitProcessor_Process(t *testing.T) {
	p := &RateLimitProcessor{}

	config := map[string]interface{}{
		"average": float64(100),
		"burst":   float64(50),
		"sourceCriterion": map[string]interface{}{
			"ipStrategy": map[string]interface{}{
				"depth": float64(2),
			},
		},
	}

	result := p.Process(config)

	if avg, ok := result["average"].(int); !ok || avg != 100 {
		t.Errorf("average = %v (%T), want 100 (int)", result["average"], result["average"])
	}
	if burst, ok := result["burst"].(int); !ok || burst != 50 {
		t.Errorf("burst = %v (%T), want 50 (int)", result["burst"], result["burst"])
	}

	sc := result["sourceCriterion"].(map[string]interface{})
	ipStrat := sc["ipStrategy"].(map[string]interface{})
	if depth, ok := ipStrat["depth"].(int); !ok || depth != 2 {
		t.Errorf("depth = %v (%T), want 2 (int)", ipStrat["depth"], ipStrat["depth"])
	}
}

func TestIPFilterProcessor_Process(t *testing.T) {
	p := &IPFilterProcessor{}

	config := map[string]interface{}{
		"sourceRange": []interface{}{"10.0.0.0/8", "192.168.1.0/24"},
	}

	result := p.Process(config)

	sr, ok := result["sourceRange"].([]interface{})
	if !ok {
		t.Fatal("sourceRange should be []interface{}")
	}
	if len(sr) != 2 {
		t.Errorf("len(sourceRange) = %d, want 2", len(sr))
	}
}

func TestPluginProcessor_Process(t *testing.T) {
	p := &PluginProcessor{}

	config := map[string]interface{}{
		"myPlugin": map[string]interface{}{
			"crowdsecLapiKey": "abc123secret",
			"enabled":         true,
		},
	}

	result := p.Process(config)

	plugin := result["myPlugin"].(map[string]interface{})
	if plugin["crowdsecLapiKey"] != "abc123secret" {
		t.Errorf("API key not preserved: %v", plugin["crowdsecLapiKey"])
	}
}

func TestPreserveTraefikValues(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := preserveTraefikValues(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("URL and path fields preserved", func(t *testing.T) {
		config := map[string]interface{}{
			"url":     "http://example.com",
			"address": "http://localhost:8080",
			"path":    "/api/v1",
		}
		result := preserveTraefikValues(config).(map[string]interface{})
		if result["url"] != "http://example.com" {
			t.Errorf("url = %v", result["url"])
		}
	})

	t.Run("boolean string conversion", func(t *testing.T) {
		config := map[string]interface{}{
			"enabled": "true",
		}
		result := preserveTraefikValues(config).(map[string]interface{})
		if result["enabled"] != true {
			t.Errorf("enabled = %v (%T), want true (bool)", result["enabled"], result["enabled"])
		}
	})

	t.Run("float64 to int for integer keys", func(t *testing.T) {
		config := map[string]interface{}{
			"statusCode": float64(301),
			"depth":      float64(3),
		}
		result := preserveTraefikValues(config).(map[string]interface{})
		if _, ok := result["statusCode"].(int); !ok {
			t.Errorf("statusCode should be int, got %T", result["statusCode"])
		}
	})

	t.Run("nested maps", func(t *testing.T) {
		config := map[string]interface{}{
			"nested": map[string]interface{}{
				"url": "http://inner.com",
			},
		}
		result := preserveTraefikValues(config).(map[string]interface{})
		nested := result["nested"].(map[string]interface{})
		if nested["url"] != "http://inner.com" {
			t.Errorf("nested url = %v", nested["url"])
		}
	})

	t.Run("arrays", func(t *testing.T) {
		config := map[string]interface{}{
			"items": []interface{}{"a", "b", "c"},
		}
		result := preserveTraefikValues(config).(map[string]interface{})
		items := result["items"].([]interface{})
		if len(items) != 3 {
			t.Errorf("len(items) = %d, want 3", len(items))
		}
	})
}
