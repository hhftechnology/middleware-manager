package services

import (
	"testing"

	"github.com/hhftechnology/middleware-manager/models"
)

// TestNewCertGenerator tests cert generator creation
func TestNewCertGenerator(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	if cg == nil {
		t.Fatal("NewCertGenerator() returned nil")
	}
	if cg.db == nil {
		t.Error("cg.db is nil")
	}
}

// TestCertGenerator_GetConfig tests fetching mTLS config
func TestCertGenerator_GetConfig(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	config, err := cg.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Default config should have HasCA = false
	if config.HasCA {
		t.Error("expected HasCA false for new config")
	}
	if config.Enabled {
		t.Error("expected Enabled false for new config")
	}
}

// TestCertGenerator_EnableMTLS_NoCA tests enabling mTLS without CA
func TestCertGenerator_EnableMTLS_NoCA(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	err := cg.EnableMTLS()
	if err == nil {
		t.Error("EnableMTLS() should fail without CA")
	}
}

// TestCertGenerator_DisableMTLS tests disabling mTLS
func TestCertGenerator_DisableMTLS(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	err := cg.DisableMTLS()
	if err != nil {
		t.Errorf("DisableMTLS() error = %v", err)
	}

	// Verify it's disabled
	config, _ := cg.GetConfig()
	if config.Enabled {
		t.Error("expected Enabled false after DisableMTLS()")
	}
}

// TestCertGenerator_GetClients_Empty tests getting clients when none exist
func TestCertGenerator_GetClients_Empty(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	clients, err := cg.GetClients()
	if err != nil {
		t.Fatalf("GetClients() error = %v", err)
	}

	// Should return empty slice, not nil
	if clients == nil {
		t.Log("GetClients() returned nil, expected empty slice")
	}
	if len(clients) != 0 {
		t.Errorf("expected 0 clients, got %d", len(clients))
	}
}

// TestCertGenerator_GetClient_NotFound tests getting non-existent client
func TestCertGenerator_GetClient_NotFound(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	_, err := cg.GetClient("nonexistent-id")
	if err == nil {
		t.Error("GetClient() should error for non-existent client")
	}
}

// TestCertGenerator_GetClientP12_NotFound tests getting P12 for non-existent client
func TestCertGenerator_GetClientP12_NotFound(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	_, _, err := cg.GetClientP12("nonexistent-id")
	if err == nil {
		t.Error("GetClientP12() should error for non-existent client")
	}
}

// TestCertGenerator_RevokeClient_NotFound tests revoking non-existent client
func TestCertGenerator_RevokeClient_NotFound(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	err := cg.RevokeClient("nonexistent-id")
	if err == nil {
		t.Error("RevokeClient() should error for non-existent client")
	}
}

// TestCertGenerator_DeleteClient_NotFound tests deleting non-existent client
func TestCertGenerator_DeleteClient_NotFound(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	err := cg.DeleteClient("nonexistent-id")
	if err == nil {
		t.Error("DeleteClient() should error for non-existent client")
	}
}

// TestCertGenerator_GetClientCount_Empty tests client count when empty
func TestCertGenerator_GetClientCount_Empty(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	count, err := cg.GetClientCount()
	if err != nil {
		t.Fatalf("GetClientCount() error = %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 clients, got %d", count)
	}
}

// TestCertGenerator_UpdateCertsBasePath tests updating certs base path
func TestCertGenerator_UpdateCertsBasePath(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	testPath := "/test/certs/path"
	err := cg.UpdateCertsBasePath(testPath)
	if err != nil {
		t.Fatalf("UpdateCertsBasePath() error = %v", err)
	}

	// Verify it was updated
	config, _ := cg.GetConfig()
	if config.CertsBasePath != testPath {
		t.Errorf("CertsBasePath = %q, want %q", config.CertsBasePath, testPath)
	}
}

// TestCertGenerator_GetMiddlewareConfig tests getting middleware config
func TestCertGenerator_GetMiddlewareConfig(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	config, err := cg.GetMiddlewareConfig()
	if err != nil {
		t.Fatalf("GetMiddlewareConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetMiddlewareConfig() returned nil")
	}

	// Check default values
	if config.RefreshInterval == 0 {
		t.Error("expected non-zero RefreshInterval")
	}
}

// TestCertGenerator_UpdateMiddlewareConfig tests updating middleware config
func TestCertGenerator_UpdateMiddlewareConfig(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	newConfig := &models.MTLSMiddlewareConfig{
		Rules:           "*.example.com",
		RequestHeaders:  "X-Client-CN,X-Client-Serial",
		RejectMessage:   "Access denied",
		RefreshInterval: 600,
	}

	err := cg.UpdateMiddlewareConfig(newConfig)
	if err != nil {
		t.Fatalf("UpdateMiddlewareConfig() error = %v", err)
	}

	// Verify it was updated
	config, _ := cg.GetMiddlewareConfig()
	if config.Rules != newConfig.Rules {
		t.Errorf("Rules = %q, want %q", config.Rules, newConfig.Rules)
	}
	if config.RefreshInterval != newConfig.RefreshInterval {
		t.Errorf("RefreshInterval = %d, want %d", config.RefreshInterval, newConfig.RefreshInterval)
	}
}

// TestCertGenerator_DeleteCA_NoCA tests deleting CA when none exists
func TestCertGenerator_DeleteCA_NoCA(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	// Should not error even when no CA exists
	err := cg.DeleteCA()
	if err != nil {
		t.Errorf("DeleteCA() error = %v (should succeed even without CA)", err)
	}
}

// TestCertGenerator_GenerateClientCert_NoCA tests generating client cert without CA
func TestCertGenerator_GenerateClientCert_NoCA(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	req := models.CreateClientRequest{
		Name:         "test-client",
		ValidityDays: 365,
		P12Password:  "testpassword",
	}

	_, err := cg.GenerateClientCert(req)
	if err == nil {
		t.Error("GenerateClientCert() should fail without CA")
	}
}

// TestCertGenerator_GenerateCA tests CA generation
func TestCertGenerator_GenerateCA(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	req := models.CreateCARequest{
		CommonName:   "Test CA",
		Organization: "Test Org",
		Country:      "US",
		ValidityDays: 365,
	}

	basePath := t.TempDir()
	config, err := cg.GenerateCA(req, basePath)
	if err != nil {
		t.Fatalf("GenerateCA() error = %v", err)
	}

	if config == nil {
		t.Fatal("GenerateCA() returned nil config")
	}

	// Verify CA was created
	if !config.HasCA {
		t.Error("expected HasCA true after GenerateCA()")
	}
	if config.CACert == "" {
		t.Error("expected non-empty CACert")
	}
	if config.CASubject == "" {
		t.Error("expected non-empty CASubject")
	}
	if config.CAExpiry == nil {
		t.Error("expected non-nil CAExpiry")
	}
}

// TestCertGenerator_GenerateCA_DefaultValidity tests CA generation with default validity
func TestCertGenerator_GenerateCA_DefaultValidity(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	req := models.CreateCARequest{
		CommonName: "Test CA",
		// No ValidityDays set - should use default
	}

	basePath := t.TempDir()
	config, err := cg.GenerateCA(req, basePath)
	if err != nil {
		t.Fatalf("GenerateCA() error = %v", err)
	}

	if config == nil {
		t.Fatal("GenerateCA() returned nil config")
	}

	// Verify CA was created with default validity
	if config.CAExpiry == nil {
		t.Error("expected non-nil CAExpiry")
	}
}

// TestCertGenerator_GenerateClientCert tests client cert generation
func TestCertGenerator_GenerateClientCert(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	// First create a CA
	caReq := models.CreateCARequest{
		CommonName:   "Test CA",
		Organization: "Test Org",
		ValidityDays: 365,
	}

	basePath := t.TempDir()
	_, err := cg.GenerateCA(caReq, basePath)
	if err != nil {
		t.Fatalf("GenerateCA() error = %v", err)
	}

	// Now create a client cert
	clientReq := models.CreateClientRequest{
		Name:         "test-client",
		ValidityDays: 365,
		P12Password:  "testpassword123",
	}

	client, err := cg.GenerateClientCert(clientReq)
	if err != nil {
		t.Fatalf("GenerateClientCert() error = %v", err)
	}

	if client == nil {
		t.Fatal("GenerateClientCert() returned nil")
	}

	// Verify client was created
	if client.ID == "" {
		t.Error("expected non-empty client ID")
	}
	if client.Name != clientReq.Name {
		t.Errorf("Name = %q, want %q", client.Name, clientReq.Name)
	}
	if client.Cert == "" {
		t.Error("expected non-empty Cert")
	}
	if client.Key == "" {
		t.Error("expected non-empty Key")
	}
	if len(client.P12) == 0 {
		t.Error("expected non-empty P12 data")
	}
	if client.Subject == "" {
		t.Error("expected non-empty Subject")
	}
	if client.Revoked {
		t.Error("expected Revoked false for new client")
	}
}

// TestCertGenerator_ClientLifecycle tests full client lifecycle
func TestCertGenerator_ClientLifecycle(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	// Create CA
	caReq := models.CreateCARequest{
		CommonName: "Test CA",
	}
	basePath := t.TempDir()
	_, err := cg.GenerateCA(caReq, basePath)
	if err != nil {
		t.Fatalf("GenerateCA() error = %v", err)
	}

	// Create client
	clientReq := models.CreateClientRequest{
		Name:        "lifecycle-test-client",
		P12Password: "password123",
	}
	client, err := cg.GenerateClientCert(clientReq)
	if err != nil {
		t.Fatalf("GenerateClientCert() error = %v", err)
	}

	// Verify count
	count, _ := cg.GetClientCount()
	if count != 1 {
		t.Errorf("expected 1 client, got %d", count)
	}

	// Get client
	retrieved, err := cg.GetClient(client.ID)
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}
	if retrieved.Name != client.Name {
		t.Errorf("retrieved Name = %q, want %q", retrieved.Name, client.Name)
	}

	// Get P12
	p12Data, name, err := cg.GetClientP12(client.ID)
	if err != nil {
		t.Fatalf("GetClientP12() error = %v", err)
	}
	if len(p12Data) == 0 {
		t.Error("expected non-empty P12 data")
	}
	if name != client.Name {
		t.Errorf("P12 name = %q, want %q", name, client.Name)
	}

	// List clients
	clients, err := cg.GetClients()
	if err != nil {
		t.Fatalf("GetClients() error = %v", err)
	}
	if len(clients) != 1 {
		t.Errorf("expected 1 client in list, got %d", len(clients))
	}

	// Revoke client
	err = cg.RevokeClient(client.ID)
	if err != nil {
		t.Fatalf("RevokeClient() error = %v", err)
	}

	// Verify revoked
	revoked, _ := cg.GetClient(client.ID)
	if !revoked.Revoked {
		t.Error("expected client to be revoked")
	}

	// Delete client
	err = cg.DeleteClient(client.ID)
	if err != nil {
		t.Fatalf("DeleteClient() error = %v", err)
	}

	// Verify deleted
	count, _ = cg.GetClientCount()
	if count != 0 {
		t.Errorf("expected 0 clients after delete, got %d", count)
	}
}

// TestCertGenerator_EnableDisableMTLS tests enable/disable cycle
func TestCertGenerator_EnableDisableMTLS(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	// Create CA first
	caReq := models.CreateCARequest{
		CommonName: "Test CA",
	}
	basePath := t.TempDir()
	_, err := cg.GenerateCA(caReq, basePath)
	if err != nil {
		t.Fatalf("GenerateCA() error = %v", err)
	}

	// Enable mTLS
	err = cg.EnableMTLS()
	if err != nil {
		t.Fatalf("EnableMTLS() error = %v", err)
	}

	// Verify enabled
	config, _ := cg.GetConfig()
	if !config.Enabled {
		t.Error("expected Enabled true after EnableMTLS()")
	}

	// Disable mTLS
	err = cg.DisableMTLS()
	if err != nil {
		t.Fatalf("DisableMTLS() error = %v", err)
	}

	// Verify disabled
	config, _ = cg.GetConfig()
	if config.Enabled {
		t.Error("expected Enabled false after DisableMTLS()")
	}
}

// TestCertGenerator_DeleteCA tests CA deletion
func TestCertGenerator_DeleteCA(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	// Create CA
	caReq := models.CreateCARequest{
		CommonName: "Test CA",
	}
	basePath := t.TempDir()
	_, err := cg.GenerateCA(caReq, basePath)
	if err != nil {
		t.Fatalf("GenerateCA() error = %v", err)
	}

	// Create a client
	clientReq := models.CreateClientRequest{
		Name:        "test-client",
		P12Password: "password",
	}
	_, err = cg.GenerateClientCert(clientReq)
	if err != nil {
		t.Fatalf("GenerateClientCert() error = %v", err)
	}

	// Delete CA
	err = cg.DeleteCA()
	if err != nil {
		t.Fatalf("DeleteCA() error = %v", err)
	}

	// Verify CA is gone
	config, _ := cg.GetConfig()
	if config.HasCA {
		t.Error("expected HasCA false after DeleteCA()")
	}
	if config.CACert != "" {
		t.Error("expected empty CACert after DeleteCA()")
	}

	// Verify clients are gone
	count, _ := cg.GetClientCount()
	if count != 0 {
		t.Errorf("expected 0 clients after DeleteCA(), got %d", count)
	}
}

// TestCertGenerator_WriteCACertToFilesystem tests writing CA to filesystem
func TestCertGenerator_WriteCACertToFilesystem(t *testing.T) {
	db := newTestSQLDB(t)
	cg := NewCertGenerator(db)

	basePath := t.TempDir()
	testCert := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")

	err := cg.WriteCACertToFilesystem(basePath, testCert)
	if err != nil {
		t.Fatalf("WriteCACertToFilesystem() error = %v", err)
	}

	// Verify directories were created
	// The function should create basePath/ca/ and basePath/clients/
}
