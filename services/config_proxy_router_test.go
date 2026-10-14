package services

import (
	"testing"
)

func routerWithRule(rule string, priority int) map[string]interface{} {
	return map[string]interface{}{
		"rule":     rule,
		"priority": priority,
		"service":  "svc",
	}
}

func TestFindMatchingRouterSingleHostIsStable(t *testing.T) {
	cp := &ConfigProxy{}
	routers := map[string]interface{}{
		"app-redirect": routerWithRule("Host(`app.example.com`)", 1),
		"app":          routerWithRule("Host(`app.example.com`)", 100),
		"other":        routerWithRule("Host(`other.example.com`)", 100),
	}

	for i := 0; i < 30; i++ {
		name, router := cp.findMatchingRouter(routers, "app.example.com")
		if name != "app" || router == nil {
			t.Fatalf("iteration %d: expected router app, got %q", i, name)
		}
	}

	name, router := cp.findMatchingRouter(routers, "missing.example.com")
	if name != "" || router != nil {
		t.Fatalf("expected no match for unknown host, got %q", name)
	}
}

func TestFindMatchingRouterAmbiguousMultiTargetHost(t *testing.T) {
	cp := &ConfigProxy{}
	routers := map[string]interface{}{
		"catch-all": routerWithRule("Host(`app.example.com`)", 10),
		"specific":  routerWithRule("Host(`app.example.com`) && PathPrefix(`/app`)", 200),
		"app-redirect": routerWithRule("Host(`app.example.com`)", 1),
	}

	for i := 0; i < 30; i++ {
		name, router := cp.findMatchingRouter(routers, "app.example.com")
		if name != "" || router != nil {
			t.Fatalf("iteration %d: ambiguous host must not match, got %q", i, name)
		}
	}
}

func TestFindMatchingRouterRedirectOnlyIsDeterministic(t *testing.T) {
	cp := &ConfigProxy{}
	routers := map[string]interface{}{
		"z-redirect": routerWithRule("Host(`app.example.com`)", 1),
		"a-redirect": routerWithRule("Host(`app.example.com`)", 1),
	}

	for i := 0; i < 30; i++ {
		name, router := cp.findMatchingRouter(routers, "app.example.com")
		if name != "a-redirect" || router == nil {
			t.Fatalf("iteration %d: expected a-redirect, got %q", i, name)
		}
	}
}

func TestApplyResourceOverridesSkipsAmbiguousMultiTargetHost(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	cp := NewConfigProxy(db, cm, "http://pangolin.invalid")

	config := &ProxiedTraefikConfig{}
	cp.initializeConfigMaps(config)
	config.HTTP.Routers["catch-all"] = routerWithRule("Host(`app.example.com`)", 10)
	config.HTTP.Routers["specific"] = routerWithRule("Host(`app.example.com`) && PathPrefix(`/app`)", 200)

	resource := &resourceData{
		ID:               "res-1",
		PangolinRouterID: "does-not-exist",
		Host:             "app.example.com",
		RouterPriority:   50,
	}

	if err := cp.applyResourceOverrides(config, []*resourceData{resource}, nil, nil); err != nil {
		t.Fatalf("applyResourceOverrides: %v", err)
	}

	catch := config.HTTP.Routers["catch-all"].(map[string]interface{})
	specific := config.HTTP.Routers["specific"].(map[string]interface{})
	if catch["priority"] != 10 {
		t.Fatalf("catch-all priority changed to %#v", catch["priority"])
	}
	if specific["priority"] != 200 {
		t.Fatalf("specific priority changed to %#v", specific["priority"])
	}
}

func TestApplyResourceOverridesSetsPriorityForSingleMatch(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	cp := NewConfigProxy(db, cm, "http://pangolin.invalid")

	config := &ProxiedTraefikConfig{}
	cp.initializeConfigMaps(config)
	config.HTTP.Routers["app"] = routerWithRule("Host(`app.example.com`)", 100)

	resource := &resourceData{
		ID:               "res-1",
		PangolinRouterID: "does-not-exist",
		Host:             "app.example.com",
		RouterPriority:   250,
	}

	if err := cp.applyResourceOverrides(config, []*resourceData{resource}, nil, nil); err != nil {
		t.Fatalf("applyResourceOverrides: %v", err)
	}

	app := config.HTTP.Routers["app"].(map[string]interface{})
	if app["priority"] != 250 {
		t.Fatalf("expected priority 250, got %#v", app["priority"])
	}
}
