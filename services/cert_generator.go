package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hhftechnology/middleware-manager/models"
	"software.sslmate.com/src/go-pkcs12"
)

// CertGenerator handles certificate generation and management
type CertGenerator struct {
	db *sql.DB
}

// NewCertGenerator creates a new certificate generator
func NewCertGenerator(db *sql.DB) *CertGenerator {
	return &CertGenerator{db: db}
}

// GenerateCA creates a new Certificate Authority
func (cg *CertGenerator) GenerateCA(req models.CreateCARequest, basePath string) (*models.MTLSConfig, error) {
	// Set defaults
	if req.ValidityDays <= 0 {
		req.ValidityDays = 1825 // 5 years
	}

	// Generate RSA 4096-bit private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Build subject
	subject := pkix.Name{
		CommonName: req.CommonName,
	}
	if req.Organization != "" {
		subject.Organization = []string{req.Organization}
	}
	if req.Country != "" {
		subject.Country = []string{req.Country}
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 0, req.ValidityDays)

	// Create CA certificate template
	caTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	// Self-sign the CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Encode certificate to PEM
	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	})

	// Encode private key to PEM
	caKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Build subject string for display
	subjectStr := fmt.Sprintf("CN=%s", req.CommonName)
	if req.Organization != "" {
		subjectStr += fmt.Sprintf(", O=%s", req.Organization)
	}
	if req.Country != "" {
		subjectStr += fmt.Sprintf(", C=%s", req.Country)
	}

	// Prepare certificate path
	certPath := filepath.Join(basePath, "ca", "ca.crt")

	// Create mTLS config
	config := &models.MTLSConfig{
		ID:            1,
		Enabled:       false,
		CACert:        string(caCertPEM),
		CAKey:         string(caKeyPEM),
		CACertPath:    certPath,
		CASubject:     subjectStr,
		CAExpiry:      &notAfter,
		CertsBasePath: basePath,
		HasCA:         true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to database
	_, err = cg.db.Exec(`
		UPDATE mtls_config SET
			ca_cert = ?,
			ca_key = ?,
			ca_cert_path = ?,
			ca_subject = ?,
			ca_expiry = ?,
			certs_base_path = ?,
			updated_at = ?
		WHERE id = 1
	`, config.CACert, config.CAKey, config.CACertPath, config.CASubject, config.CAExpiry, config.CertsBasePath, config.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save CA to database: %w", err)
	}

	// Write CA certificate to filesystem for Traefik
	if err := cg.WriteCACertToFilesystem(basePath, caCertPEM); err != nil {
		log.Printf("Warning: Failed to write CA cert to filesystem: %v", err)
		// Don't fail - cert is in database, user can manually extract
	}

	return config, nil
}

// GenerateClientCert creates a new client certificate signed by the CA
func (cg *CertGenerator) GenerateClientCert(req models.CreateClientRequest) (*models.MTLSClient, error) {
	// Set defaults
	if req.ValidityDays <= 0 {
		req.ValidityDays = 730 // 2 years
	}

	// Get CA from database
	var caCertPEM, caKeyPEM, caSubject string
	err := cg.db.QueryRow(`
		SELECT ca_cert, ca_key, ca_subject FROM mtls_config WHERE id = 1
	`).Scan(&caCertPEM, &caKeyPEM, &caSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to get CA from database: %w", err)
	}

	if caCertPEM == "" || caKeyPEM == "" {
		return nil, fmt.Errorf("CA not configured - please create a CA first")
	}

	// Parse CA certificate
	caCertBlock, _ := pem.Decode([]byte(caCertPEM))
	if caCertBlock == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse CA private key
	caKeyBlock, _ := pem.Decode([]byte(caKeyPEM))
	if caKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Generate client private key
	clientKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client private key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Build client subject based on CA subject but with client name
	clientSubject := pkix.Name{
		CommonName: req.Name + "." + caCert.Subject.CommonName,
	}
	if len(caCert.Subject.Organization) > 0 {
		clientSubject.Organization = caCert.Subject.Organization
	}
	if len(caCert.Subject.Country) > 0 {
		clientSubject.Country = caCert.Subject.Country
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 0, req.ValidityDays)

	// Create client certificate template
	clientTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      clientSubject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Sign client certificate with CA
	clientCertDER, err := x509.CreateCertificate(rand.Reader, &clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate: %w", err)
	}

	// Encode client certificate to PEM
	clientCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertDER,
	})

	// Encode client private key to PEM
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})

	// Parse the client certificate for PKCS#12
	clientCert, err := x509.ParseCertificate(clientCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	// Generate PKCS#12 (.p12) file
	p12Data, err := pkcs12.Modern.Encode(clientKey, clientCert, []*x509.Certificate{caCert}, req.P12Password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCS#12: %w", err)
	}

	// Build subject string for display
	subjectStr := fmt.Sprintf("CN=%s", clientSubject.CommonName)
	if len(clientSubject.Organization) > 0 {
		subjectStr += fmt.Sprintf(", O=%s", clientSubject.Organization[0])
	}
	if len(clientSubject.Country) > 0 {
		subjectStr += fmt.Sprintf(", C=%s", clientSubject.Country[0])
	}

	// Generate unique ID
	clientID := uuid.New().String()

	// Create password hint (first and last character)
	passwordHint := ""
	if len(req.P12Password) >= 2 {
		passwordHint = string(req.P12Password[0]) + "***" + string(req.P12Password[len(req.P12Password)-1])
	}

	// Create client record
	client := &models.MTLSClient{
		ID:              clientID,
		Name:            req.Name,
		Cert:            string(clientCertPEM),
		Key:             string(clientKeyPEM),
		P12:             p12Data,
		P12PasswordHint: passwordHint,
		Subject:         subjectStr,
		Expiry:          &notAfter,
		Revoked:         false,
		CreatedAt:       time.Now(),
	}

	// Save to database
	_, err = cg.db.Exec(`
		INSERT INTO mtls_clients (id, name, cert, key, p12, p12_password_hint, subject, expiry, revoked, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, client.ID, client.Name, client.Cert, client.Key, client.P12, client.P12PasswordHint, client.Subject, client.Expiry, 0, client.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save client to database: %w", err)
	}

	return client, nil
}

// GetConfig retrieves the current mTLS configuration
func (cg *CertGenerator) GetConfig() (*models.MTLSConfig, error) {
	var config models.MTLSConfig
	var caExpiry sql.NullTime
	var enabled int

	err := cg.db.QueryRow(`
		SELECT id, enabled, ca_cert, ca_cert_path, ca_subject, ca_expiry, certs_base_path, created_at, updated_at
		FROM mtls_config WHERE id = 1
	`).Scan(&config.ID, &enabled, &config.CACert, &config.CACertPath, &config.CASubject, &caExpiry, &config.CertsBasePath, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get mTLS config: %w", err)
	}

	config.Enabled = enabled == 1
	config.HasCA = config.CACert != ""
	if caExpiry.Valid {
		config.CAExpiry = &caExpiry.Time
	}

	return &config, nil
}

// EnableMTLS enables mTLS globally
func (cg *CertGenerator) EnableMTLS() error {
	// Check if CA exists
	config, err := cg.GetConfig()
	if err != nil {
		return err
	}
	if !config.HasCA {
		return fmt.Errorf("cannot enable mTLS: CA not configured")
	}

	_, err = cg.db.Exec(`UPDATE mtls_config SET enabled = 1, updated_at = ? WHERE id = 1`, time.Now())
	if err != nil {
		return fmt.Errorf("failed to enable mTLS: %w", err)
	}

	// Ensure CA cert is written to filesystem
	if err := cg.WriteCACertToFilesystem(config.CertsBasePath, []byte(config.CACert)); err != nil {
		log.Printf("Warning: Failed to write CA cert to filesystem: %v", err)
	}

	return nil
}

// DisableMTLS disables mTLS globally
func (cg *CertGenerator) DisableMTLS() error {
	_, err := cg.db.Exec(`UPDATE mtls_config SET enabled = 0, updated_at = ? WHERE id = 1`, time.Now())
	if err != nil {
		return fmt.Errorf("failed to disable mTLS: %w", err)
	}
	return nil
}

// DeleteCA removes the CA and all client certificates
func (cg *CertGenerator) DeleteCA() error {
	tx, err := cg.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all clients
	_, err = tx.Exec(`DELETE FROM mtls_clients`)
	if err != nil {
		return fmt.Errorf("failed to delete clients: %w", err)
	}

	// Reset CA config
	_, err = tx.Exec(`
		UPDATE mtls_config SET
			enabled = 0,
			ca_cert = '',
			ca_key = '',
			ca_cert_path = '',
			ca_subject = '',
			ca_expiry = NULL,
			updated_at = ?
		WHERE id = 1
	`, time.Now())
	if err != nil {
		return fmt.Errorf("failed to reset CA config: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetClients retrieves all client certificates
func (cg *CertGenerator) GetClients() ([]models.MTLSClient, error) {
	rows, err := cg.db.Query(`
		SELECT id, name, cert, p12_password_hint, subject, expiry, revoked, revoked_at, created_at
		FROM mtls_clients
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []models.MTLSClient
	for rows.Next() {
		var client models.MTLSClient
		var expiry, revokedAt sql.NullTime
		var revoked int

		err := rows.Scan(&client.ID, &client.Name, &client.Cert, &client.P12PasswordHint, &client.Subject, &expiry, &revoked, &revokedAt, &client.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client row: %w", err)
		}

		client.Revoked = revoked == 1
		if expiry.Valid {
			client.Expiry = &expiry.Time
		}
		if revokedAt.Valid {
			client.RevokedAt = &revokedAt.Time
		}

		clients = append(clients, client)
	}

	return clients, nil
}

// GetClient retrieves a specific client certificate
func (cg *CertGenerator) GetClient(id string) (*models.MTLSClient, error) {
	var client models.MTLSClient
	var expiry, revokedAt sql.NullTime
	var revoked int

	err := cg.db.QueryRow(`
		SELECT id, name, cert, p12_password_hint, subject, expiry, revoked, revoked_at, created_at
		FROM mtls_clients WHERE id = ?
	`, id).Scan(&client.ID, &client.Name, &client.Cert, &client.P12PasswordHint, &client.Subject, &expiry, &revoked, &revokedAt, &client.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	client.Revoked = revoked == 1
	if expiry.Valid {
		client.Expiry = &expiry.Time
	}
	if revokedAt.Valid {
		client.RevokedAt = &revokedAt.Time
	}

	return &client, nil
}

// GetClientP12 retrieves the PKCS#12 data for a client
func (cg *CertGenerator) GetClientP12(id string) ([]byte, string, error) {
	var p12Data []byte
	var name string

	err := cg.db.QueryRow(`SELECT p12, name FROM mtls_clients WHERE id = ?`, id).Scan(&p12Data, &name)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get client P12: %w", err)
	}

	return p12Data, name, nil
}

// RevokeClient marks a client certificate as revoked
func (cg *CertGenerator) RevokeClient(id string) error {
	result, err := cg.db.Exec(`
		UPDATE mtls_clients SET revoked = 1, revoked_at = ? WHERE id = ?
	`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke client: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("client not found: %s", id)
	}

	return nil
}

// DeleteClient removes a client certificate
func (cg *CertGenerator) DeleteClient(id string) error {
	result, err := cg.db.Exec(`DELETE FROM mtls_clients WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("client not found: %s", id)
	}

	return nil
}

// WriteCACertToFilesystem writes the CA certificate to the filesystem for Traefik
func (cg *CertGenerator) WriteCACertToFilesystem(basePath string, caCertPEM []byte) error {
	// Create directories
	caDir := filepath.Join(basePath, "ca")
	if err := os.MkdirAll(caDir, 0755); err != nil {
		return fmt.Errorf("failed to create CA directory: %w", err)
	}

	clientsDir := filepath.Join(basePath, "clients")
	if err := os.MkdirAll(clientsDir, 0755); err != nil {
		return fmt.Errorf("failed to create clients directory: %w", err)
	}

	// Write CA certificate
	certPath := filepath.Join(caDir, "ca.crt")
	if err := os.WriteFile(certPath, caCertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	log.Printf("CA certificate written to %s", certPath)
	return nil
}

// GetClientCount returns the number of client certificates
func (cg *CertGenerator) GetClientCount() (int, error) {
	var count int
	err := cg.db.QueryRow(`SELECT COUNT(*) FROM mtls_clients`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count clients: %w", err)
	}
	return count, nil
}

// UpdateCertsBasePath updates the base path for certificates
func (cg *CertGenerator) UpdateCertsBasePath(basePath string) error {
	_, err := cg.db.Exec(`
		UPDATE mtls_config SET certs_base_path = ?, updated_at = ? WHERE id = 1
	`, basePath, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update certs base path: %w", err)
	}
	return nil
}

// GetMiddlewareConfig retrieves the mTLS middleware plugin configuration
func (cg *CertGenerator) GetMiddlewareConfig() (*models.MTLSMiddlewareConfig, error) {
	var config models.MTLSMiddlewareConfig
	var rules, requestHeaders, rejectMessage sql.NullString
	var refreshInterval sql.NullInt64

	err := cg.db.QueryRow(`
		SELECT middleware_rules, middleware_request_headers, middleware_reject_message, middleware_refresh_interval
		FROM mtls_config WHERE id = 1
	`).Scan(&rules, &requestHeaders, &rejectMessage, &refreshInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to get middleware config: %w", err)
	}

	if rules.Valid {
		config.Rules = rules.String
	}
	if requestHeaders.Valid {
		config.RequestHeaders = requestHeaders.String
	}
	if rejectMessage.Valid {
		config.RejectMessage = rejectMessage.String
	} else {
		config.RejectMessage = "Access denied: Valid client certificate required"
	}
	if refreshInterval.Valid {
		config.RefreshInterval = int(refreshInterval.Int64)
	} else {
		config.RefreshInterval = 300
	}

	return &config, nil
}

// UpdateMiddlewareConfig updates the mTLS middleware plugin configuration
func (cg *CertGenerator) UpdateMiddlewareConfig(config *models.MTLSMiddlewareConfig) error {
	_, err := cg.db.Exec(`
		UPDATE mtls_config SET
			middleware_rules = ?,
			middleware_request_headers = ?,
			middleware_reject_message = ?,
			middleware_refresh_interval = ?,
			updated_at = ?
		WHERE id = 1
	`, config.Rules, config.RequestHeaders, config.RejectMessage, config.RefreshInterval, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update middleware config: %w", err)
	}
	return nil
}
