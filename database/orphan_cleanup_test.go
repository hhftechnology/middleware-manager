package database

import "testing"

func TestCleanupOrphanedRelationshipsKeepsValidRows(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	mustExec(t, db, `INSERT INTO resources (id, host, service_id, org_id, site_id, status) VALUES (?, ?, ?, ?, ?, ?)`,
		"res-1", "app.example.com", "svc", "org", "site", "active")
	mustExec(t, db, `INSERT INTO services (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"svc-1", "App", "loadBalancer", `{}`)
	mustExec(t, db, `INSERT INTO middlewares (id, name, type, config) VALUES (?, ?, ?, ?)`,
		"mw-1", "Headers", "headers", `{}`)
	mustExec(t, db, `INSERT INTO resource_services (resource_id, service_id) VALUES (?, ?)`, "res-1", "svc-1")
	mustExec(t, db, `INSERT INTO resource_services (resource_id, service_id) VALUES (?, ?)`, "res-1", "missing-svc")
	mustExec(t, db, `INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)`, "res-1", "mw-1", 100)
	mustExec(t, db, `INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)`, "res-1", "missing-mw", 100)
	mustExec(t, db, `INSERT INTO resource_middlewares (resource_id, middleware_id, priority) VALUES (?, ?, ?)`, "missing-res", "mw-1", 100)

	opts := DefaultCleanupOptions()
	opts.LogLevel = 0
	if err := db.CleanupOrphanedRelationships(opts); err != nil {
		t.Fatalf("cleanup: %v", err)
	}

	var serviceCount, middlewareCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM resource_services`).Scan(&serviceCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM resource_middlewares`).Scan(&middlewareCount); err != nil {
		t.Fatal(err)
	}
	if serviceCount != 1 {
		t.Fatalf("expected 1 resource_services row, got %d", serviceCount)
	}
	if middlewareCount != 1 {
		t.Fatalf("expected 1 resource_middlewares row, got %d", middlewareCount)
	}

	var keptService string
	if err := db.QueryRow(`SELECT service_id FROM resource_services`).Scan(&keptService); err != nil {
		t.Fatal(err)
	}
	if keptService != "svc-1" {
		t.Fatalf("kept service_id = %s", keptService)
	}
}
