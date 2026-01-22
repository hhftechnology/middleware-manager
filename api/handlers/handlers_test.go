package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

func TestMiddlewareHandlerCreateAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := testutil.NewTempDB(t)
	handler := NewMiddlewareHandler(db.DB)

	body := bytes.NewBufferString(`{"name":"test-mw","type":"headers","config":{"customRequestHeaders":{"X-Test":"1"}}}`)
	c, rec := testutil.NewContext(t, http.MethodPost, "/api/middlewares", body)
	handler.CreateMiddleware(c)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var created map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected generated id, got %v", created["id"])
	}

	cGet, recGet := testutil.NewContext(t, http.MethodGet, "/api/middlewares/"+id, nil)
	cGet.Params = gin.Params{{Key: "id", Value: id}}
	handler.GetMiddleware(cGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recGet.Code)
	}

	var fetched map[string]interface{}
	if err := json.Unmarshal(recGet.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("failed to parse get response: %v", err)
	}
	if fetched["name"] != "test-mw" {
		t.Fatalf("expected middleware name to match, got %v", fetched["name"])
	}
}

func TestDataSourceHandlerGetAndSetActive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cm := testutil.NewTestConfigManager(t)
	handler := NewDataSourceHandler(cm)

	// Initial active should be pangolin
	cActive, recActive := testutil.NewContext(t, http.MethodGet, "/api/datasource/active", nil)
	handler.GetActiveDataSource(cActive)
	if recActive.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recActive.Code)
	}

	// Switch to traefik
	cSet, recSet := testutil.NewContext(t, http.MethodPut, "/api/datasource/active", bytes.NewBufferString(`{"name":"traefik"}`))
	handler.SetActiveDataSource(cSet)
	if recSet.Code != http.StatusOK {
		t.Fatalf("expected 200 from SetActiveDataSource, got %d", recSet.Code)
	}

	cActive2, recActive2 := testutil.NewContext(t, http.MethodGet, "/api/datasource/active", nil)
	handler.GetActiveDataSource(cActive2)
	if recActive2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second active, got %d", recActive2.Code)
	}

	var activeResp map[string]interface{}
	if err := json.Unmarshal(recActive2.Body.Bytes(), &activeResp); err != nil {
		t.Fatalf("failed to parse active response: %v", err)
	}
	if activeResp["name"] != "traefik" {
		t.Fatalf("expected active source traefik, got %v", activeResp["name"])
	}
}
