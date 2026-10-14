package services

import (
	"encoding/json"
	"testing"

	"github.com/hhftechnology/middleware-manager/database"
)

func TestResolveMiddlewareRefUsesIDNotDisplayName(t *testing.T) {
	ids := map[string]struct{}{
		"ip-allowlist": {},
		"rate-limit":   {},
	}
	names := map[string]string{
		"IP Allow List": "ip-allowlist",
		"Rate Limit":    "rate-limit",
	}

	cases := []struct {
		in, want string
	}{
		{"IP Allow List", "ip-allowlist"},
		{"ip-allowlist", "ip-allowlist"},
		{"rate-limit@file", "rate-limit@file"},
		{"Rate Limit@file", "rate-limit@file"},
		{"badger@http", "badger@http"},
		{"  IP Allow List  ", "ip-allowlist"},
	}
	for _, tc := range cases {
		got := ResolveMiddlewareRef(tc.in, ids, names)
		if got != tc.want {
			t.Errorf("ResolveMiddlewareRef(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRewriteChainRefsForProxyAndFile(t *testing.T) {
	records := []middlewareRecord{
		{ID: "ip-allowlist", Name: "IP Allow List", Type: "ipAllowList", Config: `{"sourceRange":["10.0.0.0/8"]}`},
		{ID: "security-chain", Name: "Security Chain", Type: "chain", Config: `{"middlewares":["IP Allow List","badger@http"]}`},
	}
	ids, names := indexMiddlewareRecords(records)

	proxyCfg := map[string]interface{}{
		"middlewares": []interface{}{"IP Allow List", "ip-allowlist@file", "badger@http"},
	}
	rewriteChainRefs(proxyCfg, ids, names, chainRefProxy)
	assertChainRefs(t, proxyCfg, []string{"ip-allowlist", "ip-allowlist", "badger@http"})

	fileCfg := map[string]interface{}{
		"middlewares": []interface{}{"IP Allow List", "badger@http"},
	}
	rewriteChainRefs(fileCfg, ids, names, chainRefFile)
	assertChainRefs(t, fileCfg, []string{"ip-allowlist@file", "badger@http"})
}

func TestExpandChainMemberIDsIncludesNestedMembers(t *testing.T) {
	records := []middlewareRecord{
		{ID: "ip-allowlist", Name: "IP Allow List", Type: "ipAllowList", Config: `{}`},
		{ID: "rate-limit", Name: "Rate Limit", Type: "rateLimit", Config: `{}`},
		{ID: "inner", Name: "Inner", Type: "chain", Config: `{"middlewares":["Rate Limit"]}`},
		{ID: "outer", Name: "Outer", Type: "chain", Config: `{"middlewares":["inner","IP Allow List"]}`},
	}
	allowed := expandChainMemberIDs(records, map[string]struct{}{"outer": {}})
	for _, id := range []string{"outer", "inner", "ip-allowlist", "rate-limit"} {
		if _, ok := allowed[id]; !ok {
			t.Errorf("expected %s to be included, got %#v", id, allowed)
		}
	}
}

func TestEmitMiddlewareConfigsKeysByID(t *testing.T) {
	records := []middlewareRecord{
		{ID: "ip-allowlist", Name: "IP Allow List", Type: "ipAllowList", Config: `{"sourceRange":["10.0.0.0/8"]}`},
		{ID: "security-chain", Name: "Security Chain", Type: "chain", Config: `{"middlewares":["IP Allow List"]}`},
	}
	emitted := emitMiddlewareConfigs(records, chainRefProxy)
	if _, ok := emitted["IP Allow List"]; ok {
		t.Fatal("display name must not be a Traefik middleware key")
	}
	chain, ok := emitted["security-chain"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing chain middleware: %#v", emitted)
	}
	body, _ := chain["chain"].(map[string]interface{})
	assertChainRefs(t, body, []string{"ip-allowlist"})
}

func assertChainRefs(t *testing.T, cfg map[string]interface{}, want []string) {
	t.Helper()
	raw, ok := cfg["middlewares"].([]interface{})
	if !ok {
		t.Fatalf("middlewares type %T", cfg["middlewares"])
	}
	if len(raw) != len(want) {
		t.Fatalf("got %d refs %#v, want %v", len(raw), raw, want)
	}
	for i, item := range raw {
		s, _ := item.(string)
		if s != want[i] {
			t.Errorf("ref[%d] = %q, want %q", i, s, want[i])
		}
	}
}

func TestApplyMiddlewaresChainAssignedByDisplayName(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)

	mustExecChain(t, db, `INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"ip-allowlist", "IP Allow List", "ipAllowList", `{"sourceRange":["10.0.0.0/8"]}`)
	mustExecChain(t, db, `INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"security-chain", "Security Chain", "chain", `{"middlewares":["IP Allow List"]}`)
	mustExecChain(t, db, `INSERT INTO resources (id, pangolin_router_id, host, service_id, org_id, site_id, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"res-1", "app-router", "app.example.com", "app", "org", "site", "active")
	mustExecChain(t, db, `INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)`,
		"res-1", "security-chain", 200)

	cp := NewConfigProxy(db, cm, "http://pangolin.invalid")
	config := &ProxiedTraefikConfig{}
	cp.initializeConfigMaps(config)
	config.HTTP.Routers["app-router"] = map[string]interface{}{
		"rule":        "Host(`app.example.com`)",
		"service":     "app",
		"middlewares": []interface{}{"badger"},
		"priority":    100,
	}

	if err := cp.mergeMiddlewareManagerConfig(config); err != nil {
		t.Fatalf("merge: %v", err)
	}

	if _, ok := config.HTTP.Middlewares["IP Allow List"]; ok {
		t.Fatal("emitted display name as middleware key")
	}
	if _, ok := config.HTTP.Middlewares["ip-allowlist"]; !ok {
		t.Fatalf("missing member middleware: %#v", config.HTTP.Middlewares)
	}
	chainEntry, ok := config.HTTP.Middlewares["security-chain"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing chain: %#v", config.HTTP.Middlewares)
	}
	body, _ := chainEntry["chain"].(map[string]interface{})
	assertChainRefs(t, body, []string{"ip-allowlist"})

	router := config.HTTP.Routers["app-router"].(map[string]interface{})
	encoded, _ := json.Marshal(router["middlewares"])
	var names []string
	if err := json.Unmarshal(encoded, &names); err != nil {
		t.Fatalf("router middlewares %#v: %v", router["middlewares"], err)
	}
	foundChain := false
	for _, name := range names {
		if name == "Security Chain" || name == "IP Allow List" {
			t.Fatalf("router referenced display name %q in %#v", name, names)
		}
		if name == "security-chain" {
			foundChain = true
		}
	}
	if !foundChain {
		t.Fatalf("router missing security-chain: %#v", names)
	}
}

func TestProcessMiddlewaresFileModeQualifiesIDs(t *testing.T) {
	db := newTestDB(t)
	cm := newTestConfigManager(t)
	mustExecChain(t, db, `INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"ip-allowlist", "IP Allow List", "ipAllowList", `{"sourceRange":["10.0.0.0/8"]}`)
	mustExecChain(t, db, `INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"security-chain", "Security Chain", "chain", `{"middlewares":["IP Allow List"]}`)

	cg := NewConfigGenerator(db, t.TempDir(), cm)
	cfg := &TraefikConfig{}
	cfg.HTTP.Middlewares = map[string]interface{}{}
	if err := cg.processMiddlewares(cfg); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.HTTP.Middlewares["IP Allow List"]; ok {
		t.Fatal("file config used the display name as the middleware key")
	}
	entry, ok := cfg.HTTP.Middlewares["security-chain"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing chain: %#v", cfg.HTTP.Middlewares)
	}
	body, _ := entry["chain"].(map[string]interface{})
	assertChainRefs(t, body, []string{"ip-allowlist@file"})
}

func mustExecChain(t *testing.T, db *database.DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %s: %v", query, err)
	}
}
