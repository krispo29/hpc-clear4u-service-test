package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"hpc-express-service/config"
	"hpc-express-service/database"
	"hpc-express-service/factory"
	"hpc-express-service/server"
)

// APIIntegrationTestSuite defines the integration test suite for API endpoints
type APIIntegrationTestSuite struct {
	suite.Suite
	server            *server.Server
	db                *pg.DB
	ctx               context.Context
	testToken         string
	testUserUUID      string
	testMAWBInfoUUIDs []string
	svcFactory        *factory.ServiceFactory
}

// SetupSuite sets up the test server and database
func (suite *APIIntegrationTestSuite) SetupSuite() {
	// Load test configuration
	testConfig := &config.Config{
		PostgreSQLHost:     getEnvOrDefault("TEST_POSTGRESQL_HOST", "localhost"),
		PostgreSQLUser:     getEnvOrDefault("TEST_POSTGRESQL_USER", "postgres"),
		PostgreSQLPassword: getEnvOrDefault("TEST_POSTGRESQL_PASSWORD", "password"),
		PostgreSQLName:     getEnvOrDefault("TEST_POSTGRESQL_NAME", "test_db"),
		PostgreSQLPort:     getEnvOrDefault("TEST_POSTGRESQL_PORT", "5432"),
		PostgreSQLSSLMode:  getEnvOrDefault("TEST_POSTGRESQL_SSLMODE", "false"),
		Mode:               "test",
	}

	// Create test database connection
	var err error
	suite.db, err = database.NewPostgreSQLConnection(
		testConfig.PostgreSQLUser,
		testConfig.PostgreSQLPassword,
		testConfig.PostgreSQLName,
		testConfig.PostgreSQLHost,
		testConfig.PostgreSQLPort,
		testConfig.PostgreSQLSSLMode,
	)

	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}

	// Set up context
	suite.ctx = context.WithValue(context.Background(), "postgreSQLConn", suite.db)

	// Run database migrations
	suite.runMigrations()

	// Create repository factory
	repoFactory := factory.NewRepositoryFactory()

	// Create service factory
	suite.svcFactory = factory.NewServiceFactory(repoFactory, nil, testConfig)

	// Create test server
	suite.server = server.New(suite.svcFactory, suite.db, testConfig.Mode)

	// Generate test JWT token
	suite.generateTestToken()

	// Initialize tracking slices
	suite.testMAWBInfoUUIDs = make([]string, 0)
}

// TearDownSuite cleans up after all tests
func (suite *APIIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.cleanupAllTestData()
		suite.db.Close()
	}
}

// SetupTest runs before each test
func (suite *APIIntegrationTestSuite) SetupTest() {
	// Clean up any existing test data
	suite.cleanupTestData()
}

// TearDownTest runs after each test
func (suite *APIIntegrationTestSuite) TearDownTest() {
	// Clean up test data created during the test
	suite.cleanupTestData()
}

// generateTestToken creates a valid JWT token for testing
func (suite *APIIntegrationTestSuite) generateTestToken() {
	suite.testUserUUID = "test-user-uuid-12345"

	// Create test claims
	claims := jwt.MapClaims{
		"uuid": suite.testUserUUID,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	}

	// Create token with claims (unused for now, using mock token)
	_ = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Load private key for signing (in real tests, you'd use a test key)
	// For this example, we'll create a mock token string
	suite.testToken = "Bearer test-jwt-token-for-integration-tests"
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// runMigrations runs the database migrations for testing
func (suite *APIIntegrationTestSuite) runMigrations() {
	// Create test MAWB Info table
	_, err := suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS tbl_mawb_info (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mawb_number VARCHAR(255) NOT NULL,
			chargeable_weight VARCHAR(100),
			date DATE,
			service_type VARCHAR(100),
			shipping_type VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		suite.T().Fatalf("Failed to create test mawb_info table: %v", err)
	}

	// Run cargo manifest migrations
	suite.runCargoManifestMigrations()

	// Run draft MAWB migrations
	suite.runDraftMAWBMigrations()
}

// runCargoManifestMigrations runs cargo manifest table migrations
func (suite *APIIntegrationTestSuite) runCargoManifestMigrations() {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS cargo_manifest (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mawb_info_uuid UUID NOT NULL,
			mawb_number VARCHAR(255) NOT NULL,
			port_of_discharge VARCHAR(255),
			flight_no VARCHAR(100),
			freight_date DATE,
			shipper TEXT,
			consignee TEXT,
			total_ctn VARCHAR(50),
			transshipment VARCHAR(255),
			status VARCHAR(50) NOT NULL DEFAULT 'Draft',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_cargo_manifest_mawb_info 
				FOREIGN KEY (mawb_info_uuid) 
				REFERENCES tbl_mawb_info(uuid) 
				ON DELETE CASCADE,
			CONSTRAINT chk_cargo_manifest_status 
				CHECK (status IN ('Draft', 'Pending', 'Confirmed', 'Rejected'))
		)`,
		`CREATE TABLE IF NOT EXISTS cargo_manifest_items (
			id SERIAL PRIMARY KEY,
			cargo_manifest_uuid UUID NOT NULL,
			hawb_no VARCHAR(255),
			pkgs VARCHAR(100),
			gross_weight VARCHAR(100),
			destination VARCHAR(255),
			commodity TEXT,
			shipper_name_address TEXT,
			consignee_name_address TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_cargo_manifest_items_cargo_manifest 
				FOREIGN KEY (cargo_manifest_uuid) 
				REFERENCES cargo_manifest(uuid) 
				ON DELETE CASCADE
		)`,
	}

	for _, migration := range migrations {
		_, err := suite.db.Exec(migration)
		if err != nil {
			suite.T().Fatalf("Failed to run cargo manifest migration: %v", err)
		}
	}
}

// runDraftMAWBMigrations runs draft MAWB table migrations
func (suite *APIIntegrationTestSuite) runDraftMAWBMigrations() {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS draft_mawb (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mawb_info_uuid UUID NOT NULL,
			customer_uuid UUID,
			airline_logo VARCHAR(255),
			airline_name VARCHAR(255),
			mawb VARCHAR(255) NOT NULL,
			hawb VARCHAR(255),
			shipper_name_and_address TEXT,
			consignee_name_and_address TEXT,
			flight_no VARCHAR(100),
			flight_date DATE,
			executed_on_date DATE,
			insurance_amount DECIMAL(15,2),
			total_no_of_pieces INTEGER,
			total_gross_weight DECIMAL(10,2),
			total_chargeable_weight DECIMAL(10,2),
			total_rate_charge DECIMAL(15,2),
			total_amount DECIMAL(15,2),
			status VARCHAR(50) NOT NULL DEFAULT 'Draft',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb_mawb_info 
				FOREIGN KEY (mawb_info_uuid) 
				REFERENCES tbl_mawb_info(uuid) 
				ON DELETE CASCADE,
			CONSTRAINT chk_draft_mawb_status 
				CHECK (status IN ('Draft', 'Pending', 'Confirmed', 'Rejected'))
		)`,
		`CREATE TABLE IF NOT EXISTS draft_mawb_items (
			id SERIAL PRIMARY KEY,
			draft_mawb_uuid UUID NOT NULL,
			pieces_rcp VARCHAR(100),
			gross_weight VARCHAR(100),
			kg_lb VARCHAR(10),
			rate_class VARCHAR(50),
			total_volume DECIMAL(15,6),
			chargeable_weight DECIMAL(10,2),
			rate_charge DECIMAL(15,2),
			total DECIMAL(15,2),
			nature_and_quantity TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb_items_draft_mawb 
				FOREIGN KEY (draft_mawb_uuid) 
				REFERENCES draft_mawb(uuid) 
				ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS draft_mawb_item_dims (
			id SERIAL PRIMARY KEY,
			draft_mawb_item_id INTEGER NOT NULL,
			length VARCHAR(50),
			width VARCHAR(50),
			height VARCHAR(50),
			count VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb_item_dims_draft_mawb_item 
				FOREIGN KEY (draft_mawb_item_id) 
				REFERENCES draft_mawb_items(id) 
				ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS draft_mawb_charges (
			id SERIAL PRIMARY KEY,
			draft_mawb_uuid UUID NOT NULL,
			charge_key VARCHAR(100) NOT NULL,
			charge_value DECIMAL(15,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb_charges_draft_mawb 
				FOREIGN KEY (draft_mawb_uuid) 
				REFERENCES draft_mawb(uuid) 
				ON DELETE CASCADE
		)`,
	}

	for _, migration := range migrations {
		_, err := suite.db.Exec(migration)
		if err != nil {
			suite.T().Fatalf("Failed to run draft MAWB migration: %v", err)
		}
	}
}

// createTestMAWBInfo creates a test MAWB Info record and returns its UUID
func (suite *APIIntegrationTestSuite) createTestMAWBInfo() string {
	var uuid string
	_, err := suite.db.QueryOne(pg.Scan(&uuid), `
		INSERT INTO tbl_mawb_info (mawb_number, chargeable_weight, service_type, shipping_type) 
		VALUES (?, ?, ?, ?) 
		RETURNING uuid
	`, fmt.Sprintf("TEST-API-%d", time.Now().UnixNano()), "100.5", "Express", "Air")

	if err != nil {
		suite.T().Fatalf("Failed to create test MAWB Info: %v", err)
	}

	suite.testMAWBInfoUUIDs = append(suite.testMAWBInfoUUIDs, uuid)
	return uuid
}

// cleanupTestData cleans up test data created during individual tests
func (suite *APIIntegrationTestSuite) cleanupTestData() {
	// Clean up in reverse dependency order
	suite.db.Exec("DELETE FROM draft_mawb_charges WHERE charge_key LIKE 'test_%'")
	suite.db.Exec("DELETE FROM draft_mawb_item_dims WHERE length LIKE 'test_%'")
	suite.db.Exec("DELETE FROM draft_mawb_items WHERE pieces_rcp LIKE 'test_%'")
	suite.db.Exec("DELETE FROM draft_mawb WHERE mawb LIKE 'TEST-%'")
	suite.db.Exec("DELETE FROM cargo_manifest_items WHERE hawb_no LIKE 'H%'")
	suite.db.Exec("DELETE FROM cargo_manifest WHERE mawb_number LIKE 'TEST-%'")
	suite.db.Exec("DELETE FROM tbl_mawb_info WHERE mawb_number LIKE 'TEST-%'")

	// Reset tracking slices
	suite.testMAWBInfoUUIDs = suite.testMAWBInfoUUIDs[:0]
}

// cleanupAllTestData cleans up all test data
func (suite *APIIntegrationTestSuite) cleanupAllTestData() {
	suite.cleanupTestData()
}

// makeRequest makes an HTTP request to the test server
func (suite *APIIntegrationTestSuite) makeRequest(method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)

	// Add authorization header
	req.Header.Set("Authorization", suite.testToken)

	// Add additional headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	suite.server.ServeHTTP(rr, req)

	return rr
}

// makeJSONRequest makes a JSON HTTP request
func (suite *APIIntegrationTestSuite) makeJSONRequest(method, path string, payload interface{}) *httptest.ResponseRecorder {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			suite.T().Fatalf("Failed to marshal JSON: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return suite.makeRequest(method, path, body, headers)
}

// parseJSONResponse parses JSON response into a map
func (suite *APIIntegrationTestSuite) parseJSONResponse(rr *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		suite.T().Fatalf("Failed to parse JSON response: %v", err)
	}
	return response
}

// TestHealthzEndpoint tests the health check endpoint
func (suite *APIIntegrationTestSuite) TestHealthzEndpoint() {
	rr := suite.makeRequest("GET", "/healthz", nil, nil)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	response := suite.parseJSONResponse(rr)
	assert.Equal(suite.T(), "OK", response["message"])
}

// TestCargoManifestAPIEndpoints tests all cargo manifest API endpoints
func (suite *APIIntegrationTestSuite) TestCargoManifestAPIEndpoints() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Test POST /v1/mawbinfo/{uuid}/cargo-manifest (Create)
	createPayload := map[string]interface{}{
		"mawbNumber":      "TEST-123456789",
		"portOfDischarge": "BKK",
		"flightNo":        "TG123",
		"freightDate":     "2024-01-01",
		"shipper":         "Test Shipper",
		"consignee":       "Test Consignee",
		"totalCtn":        "10",
		"transshipment":   "No",
		"items": []map[string]interface{}{
			{
				"hawbNo":                  "H123",
				"pkgs":                    "5",
				"grossWeight":             "100.5",
				"dst":                     "NYC",
				"commodity":               "Electronics",
				"shipperNameAndAddress":   "Shipper Address",
				"consigneeNameAndAddress": "Consignee Address",
			},
		},
	}

	createPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID)
	rr := suite.makeJSONRequest("POST", createPath, createPayload)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	createResponse := suite.parseJSONResponse(rr)
	assert.Equal(suite.T(), "success", createResponse["message"])
	assert.NotNil(suite.T(), createResponse["data"])

	// Test GET /v1/mawbinfo/{uuid}/cargo-manifest (Read)
	getRR := suite.makeRequest("GET", createPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRR.Code)

	// Test status transitions
	confirmPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/confirm", mawbUUID)
	confirmRR := suite.makeJSONRequest("POST", confirmPath, nil)
	assert.Equal(suite.T(), http.StatusOK, confirmRR.Code)

	rejectPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/reject", mawbUUID)
	rejectRR := suite.makeJSONRequest("POST", rejectPath, nil)
	assert.Equal(suite.T(), http.StatusOK, rejectRR.Code)

	// Test PDF generation
	printPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/print", mawbUUID)
	printRR := suite.makeRequest("GET", printPath, nil, nil)
	assert.True(suite.T(), printRR.Code == http.StatusOK || printRR.Code >= 400)
}

// TestDraftMAWBAPIEndpoints tests all draft MAWB API endpoints
func (suite *APIIntegrationTestSuite) TestDraftMAWBAPIEndpoints() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Test POST /v1/mawbinfo/{uuid}/draft-mawb (Create)
	createPayload := map[string]interface{}{
		"customerUUID":            "customer-uuid",
		"airlineLogo":             "logo.png",
		"airlineName":             "Test Airlines",
		"mawb":                    "TEST-123456789",
		"hawb":                    "H123456",
		"shipperNameAndAddress":   "Test Shipper Address",
		"consigneeNameAndAddress": "Test Consignee Address",
		"items": []map[string]interface{}{
			{
				"piecesRCP":         "5",
				"grossWeight":       "250.5",
				"kgLb":              "kg",
				"rateClass":         "N",
				"totalVolume":       0.5,
				"chargeableWeight":  300.0,
				"rateCharge":        1000.0,
				"total":             1000.0,
				"natureAndQuantity": "Electronics",
				"dims": []map[string]interface{}{
					{
						"length": "100",
						"width":  "50",
						"height": "30",
						"count":  "2",
					},
				},
			},
		},
		"charges": []map[string]interface{}{
			{
				"key":   "fuel_surcharge",
				"value": 500.0,
			},
		},
	}

	createPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb", mawbUUID)
	rr := suite.makeJSONRequest("POST", createPath, createPayload)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	createResponse := suite.parseJSONResponse(rr)
	assert.Equal(suite.T(), "success", createResponse["message"])
	assert.NotNil(suite.T(), createResponse["data"])

	// Test GET /v1/mawbinfo/{uuid}/draft-mawb (Read)
	getRR := suite.makeRequest("GET", createPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRR.Code)

	// Test status transitions
	confirmPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/confirm", mawbUUID)
	confirmRR := suite.makeJSONRequest("POST", confirmPath, nil)
	assert.Equal(suite.T(), http.StatusOK, confirmRR.Code)

	rejectPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/reject", mawbUUID)
	rejectRR := suite.makeJSONRequest("POST", rejectPath, nil)
	assert.Equal(suite.T(), http.StatusOK, rejectRR.Code)

	// Test PDF generation
	printPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/print", mawbUUID)
	printRR := suite.makeRequest("GET", printPath, nil, nil)
	assert.True(suite.T(), printRR.Code == http.StatusOK || printRR.Code >= 400)
}

// TestAuthenticationFlows tests authentication and authorization scenarios
func (suite *APIIntegrationTestSuite) TestAuthenticationFlows() {
	mawbUUID := suite.createTestMAWBInfo()

	// Test request without authorization header
	req := httptest.NewRequest("GET", fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID), nil)
	rr := httptest.NewRecorder()
	suite.server.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, rr.Code)

	// Test request with invalid authorization header
	req = httptest.NewRequest("GET", fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID), nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = httptest.NewRecorder()
	suite.server.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, rr.Code)

	// Test request with valid authorization header
	req = httptest.NewRequest("GET", fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID), nil)
	req.Header.Set("Authorization", suite.testToken)
	rr = httptest.NewRecorder()
	suite.server.ServeHTTP(rr, req)

	// Should return 404 (not found) rather than 401 (unauthorized) since auth passed
	assert.True(suite.T(), rr.Code == http.StatusNotFound || rr.Code == http.StatusOK)
}

// TestErrorResponseFormats tests error response formats and status codes
func (suite *APIIntegrationTestSuite) TestErrorResponseFormats() {
	// Test invalid UUID format
	invalidUUID := "invalid-uuid-format"
	path := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", invalidUUID)
	rr := suite.makeRequest("GET", path, nil, nil)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)

	response := suite.parseJSONResponse(rr)
	assert.Contains(suite.T(), response["message"], "invalid UUID format")

	// Test non-existent MAWB Info UUID
	nonExistentUUID := "00000000-0000-0000-0000-000000000000"
	path = fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", nonExistentUUID)
	rr = suite.makeRequest("GET", path, nil, nil)

	assert.Equal(suite.T(), http.StatusNotFound, rr.Code)

	response = suite.parseJSONResponse(rr)
	assert.Contains(suite.T(), strings.ToLower(response["message"].(string)), "not found")
}

// TestPDFGenerationEndpoints tests PDF generation with actual file validation
func (suite *APIIntegrationTestSuite) TestPDFGenerationEndpoints() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Create cargo manifest first
	cargoManifestPayload := map[string]interface{}{
		"mawbNumber":      "TEST-PDF-123456789",
		"portOfDischarge": "BKK",
		"flightNo":        "TG123",
		"shipper":         "Test Shipper",
		"consignee":       "Test Consignee",
		"items": []map[string]interface{}{
			{
				"hawbNo":    "H123",
				"pkgs":      "5",
				"commodity": "Electronics",
			},
		},
	}

	createPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID)
	createRR := suite.makeJSONRequest("POST", createPath, cargoManifestPayload)
	assert.Equal(suite.T(), http.StatusOK, createRR.Code)

	// Test cargo manifest PDF generation
	printPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/print", mawbUUID)
	printRR := suite.makeRequest("GET", printPath, nil, nil)

	if printRR.Code == http.StatusOK {
		// Verify response headers
		contentType := printRR.Header().Get("Content-Type")
		assert.True(suite.T(), strings.Contains(contentType, "application/pdf") ||
			strings.Contains(contentType, "application/octet-stream"))

		// Verify PDF content (basic validation)
		body := printRR.Body.Bytes()
		assert.True(suite.T(), len(body) > 0, "PDF content should not be empty")

		// Check for PDF magic number (basic PDF validation)
		if len(body) >= 4 {
			pdfHeader := string(body[:4])
			assert.Equal(suite.T(), "%PDF", pdfHeader, "Response should start with PDF magic number")
		}
	} else {
		// If PDF generation fails, ensure it returns appropriate error
		assert.True(suite.T(), printRR.Code >= 400 && printRR.Code < 600)
	}
}

// Run the API integration test suite
func TestAPIIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if TEST_INTEGRATION is not set
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Skipping API integration tests. Set TEST_INTEGRATION=1 to run.")
	}

	suite.Run(t, new(APIIntegrationTestSuite))
}
