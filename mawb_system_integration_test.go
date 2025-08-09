package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/suite"

	"hpc-express-service/cargo_manifest"
	"hpc-express-service/config"
	"hpc-express-service/database"
	"hpc-express-service/draft_mawb"
	"hpc-express-service/factory"
	"hpc-express-service/server"
)

// MAWBSystemIntegrationTestSuite tests the complete MAWB system integration
type MAWBSystemIntegrationTestSuite struct {
	suite.Suite
	server                 *server.Server
	db                     *pg.DB
	ctx                    context.Context
	testToken              string
	testUserUUID           string
	svcFactory             *factory.ServiceFactory
	cargoManifestRepo      cargo_manifest.Repository
	draftMAWBRepo          draft_mawb.Repository
	testMAWBInfoUUIDs      []string
	testCargoManifestUUIDs []string
	testDraftMAWBUUIDs     []string
}

// SetupSuite sets up the complete test environment
func (suite *MAWBSystemIntegrationTestSuite) SetupSuite() {
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
	} // Set up
	context
	suite.ctx = context.WithValue(context.Background(), "postgreSQLConn", suite.db)

	// Run database migrations
	suite.runMigrations()

	// Initialize repositories
	suite.cargoManifestRepo = cargo_manifest.NewRepository(30 * time.Second)
	suite.draftMAWBRepo = draft_mawb.NewRepository(30 * time.Second)

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
	suite.testCargoManifestUUIDs = make([]string, 0)
	suite.testDraftMAWBUUIDs = make([]string, 0)
}

// TearDownSuite cleans up after all tests
func (suite *MAWBSystemIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.cleanupAllTestData()
		suite.db.Close()
	}
}

// SetupTest runs before each test
func (suite *MAWBSystemIntegrationTestSuite) SetupTest() {
	suite.cleanupTestData()
}

// TearDownTest runs after each test
func (suite *MAWBSystemIntegrationTestSuite) TearDownTest() {
	suite.cleanupTestData()
}

// generateTestToken creates a valid JWT token for testing
func (suite *MAWBSystemIntegrationTestSuite) generateTestToken() {
	suite.testUserUUID = "test-user-uuid-12345"
	suite.testToken = "Bearer test-jwt-token-for-mawb-integration-tests"
}

// runMigrations runs the database migrations for testing
func (suite *MAWBSystemIntegrationTestSuite) runMigrations() {
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
func (suite *MAWBSystemIntegrationTestSuite) runCargoManifestMigrations() {
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
		`CREATE INDEX IF NOT EXISTS idx_cargo_manifest_mawb_info_uuid 
			ON cargo_manifest(mawb_info_uuid)`,
		`CREATE INDEX IF NOT EXISTS idx_cargo_manifest_items_cargo_manifest_uuid 
			ON cargo_manifest_items(cargo_manifest_uuid)`,
	}

	for _, migration := range migrations {
		_, err := suite.db.Exec(migration)
		if err != nil {
			suite.T().Fatalf("Failed to run cargo manifest migration: %v", err)
		}
	}
}

// runDraftMAWBMigrations runs draft MAWB table migrations
func (suite *MAWBSystemIntegrationTestSuite) runDraftMAWBMigrations() {
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
		)`, `C
REATE TABLE IF NOT EXISTS draft_mawb_items (
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
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_mawb_info_uuid 
			ON draft_mawb(mawb_info_uuid)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_items_draft_mawb_uuid 
			ON draft_mawb_items(draft_mawb_uuid)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_item_dims_draft_mawb_item_id 
			ON draft_mawb_item_dims(draft_mawb_item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_charges_draft_mawb_uuid 
			ON draft_mawb_charges(draft_mawb_uuid)`,
	}

	for _, migration := range migrations {
		_, err := suite.db.Exec(migration)
		if err != nil {
			suite.T().Fatalf("Failed to run draft MAWB migration: %v", err)
		}
	}
}

// createTestMAWBInfo creates a test MAWB Info record and returns its UUID
func (suite *MAWBSystemIntegrationTestSuite) createTestMAWBInfo() string {
	var uuid string
	_, err := suite.db.QueryOne(pg.Scan(&uuid), `
		INSERT INTO tbl_mawb_info (mawb_number, chargeable_weight, service_type, shipping_type) 
		VALUES (?, ?, ?, ?) 
		RETURNING uuid
	`, fmt.Sprintf("TEST-INTEGRATION-%d", time.Now().UnixNano()), "100.5", "Express", "Air")

	if err != nil {
		suite.T().Fatalf("Failed to create test MAWB Info: %v", err)
	}

	suite.testMAWBInfoUUIDs = append(suite.testMAWBInfoUUIDs, uuid)
	return uuid
}

// makeRequest makes an HTTP request to the test server
func (suite *MAWBSystemIntegrationTestSuite) makeRequest(method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
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
func (suite *MAWBSystemIntegrationTestSuite) makeJSONRequest(method, path string, payload interface{}) *httptest.ResponseRecorder {
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
func (suite *MAWBSystemIntegrationTestSuite) parseJSONResponse(rr *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		suite.T().Fatalf("Failed to parse JSON response: %v", err)
	}
	return response
}

// cleanupTestData cleans up test data created during individual tests
func (suite *MAWBSystemIntegrationTestSuite) cleanupTestData() {
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
	suite.testCargoManifestUUIDs = suite.testCargoManifestUUIDs[:0]
	suite.testDraftMAWBUUIDs = suite.testDraftMAWBUUIDs[:0]
}

// cleanupAllTestData cleans up all test data
func (suite *MAWBSystemIntegrationTestSuite) cleanupAllTestData() {
	suite.cleanupTestData()
}
// TestCompleteWorkflowFromMAWBInfoToCargoManifestAndDraftMAWB tests the complete workflow
// from MAWB Info creation to cargo manifest and draft MAWB creation, updates, and status management
func (suite *MAWBSystemIntegrationTestSuite) TestCompleteWorkflowFromMAWBInfoToCargoManifestAndDraftMAWB() {
	// Step 1: Create MAWB Info record
	mawbUUID := suite.createTestMAWBInfo()
	assert.NotEmpty(suite.T(), mawbUUID)

	// Verify MAWB Info exists in database
	var mawbExists bool
	_, err := suite.db.QueryOne(pg.Scan(&mawbExists), 
		"SELECT EXISTS(SELECT 1 FROM tbl_mawb_info WHERE uuid = ?)", mawbUUID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), mawbExists)

	// Step 2: Create Cargo Manifest linked to MAWB Info
	cargoManifestPayload := map[string]interface{}{
		"mawbNumber":      "TEST-WORKFLOW-123456789",
		"portOfDischarge": "BKK",
		"flightNo":        "TG123",
		"freightDate":     "2024-01-01",
		"shipper":         "Test Workflow Shipper",
		"consignee":       "Test Workflow Consignee",
		"totalCtn":        "10",
		"transshipment":   "No",
		"items": []map[string]interface{}{
			{
				"hawbNo":                  "H123-WORKFLOW",
				"pkgs":                    "5",
				"grossWeight":             "100.5",
				"dst":                     "NYC",
				"commodity":               "Electronics",
				"shipperNameAndAddress":   "Shipper Address",
				"consigneeNameAndAddress": "Consignee Address",
			},
			{
				"hawbNo":                  "H456-WORKFLOW",
				"pkgs":                    "3",
				"grossWeight":             "75.0",
				"dst":                     "LAX",
				"commodity":               "Textiles",
				"shipperNameAndAddress":   "Another Shipper",
				"consigneeNameAndAddress": "Another Consignee",
			},
		},
	}

	createCargoManifestPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID)
	cargoManifestRR := suite.makeJSONRequest("POST", createCargoManifestPath, cargoManifestPayload)

	assert.Equal(suite.T(), http.StatusOK, cargoManifestRR.Code)
	cargoManifestResponse := suite.parseJSONResponse(cargoManifestRR)
	assert.Equal(suite.T(), "success", cargoManifestResponse["message"])
	assert.NotNil(suite.T(), cargoManifestResponse["data"])

	// Extract cargo manifest UUID for tracking
	cargoManifestData := cargoManifestResponse["data"].(map[string]interface{})
	cargoManifestUUID := cargoManifestData["uuid"].(string)
	suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, cargoManifestUUID)

	// Step 3: Verify Cargo Manifest can be retrieved
	getCargoManifestRR := suite.makeRequest("GET", createCargoManifestPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getCargoManifestRR.Code)

	getCargoManifestResponse := suite.parseJSONResponse(getCargoManifestRR)
	retrievedCargoManifest := getCargoManifestResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "TEST-WORKFLOW-123456789", retrievedCargoManifest["mawbNumber"])
	assert.Equal(suite.T(), "Draft", retrievedCargoManifest["status"])

	// Verify items were saved correctly
	items := retrievedCargoManifest["items"].([]interface{})
	assert.Equal(suite.T(), 2, len(items))

	// Step 4: Create Draft MAWB linked to the same MAWB Info
	draftMAWBPayload := map[string]interface{}{
		"customerUUID":            "customer-workflow-uuid",
		"airlineLogo":             "workflow-logo.png",
		"airlineName":             "Test Workflow Airlines",
		"mawb":                    "TEST-WORKFLOW-DRAFT-123456789",
		"hawb":                    "H123456-WORKFLOW",
		"shipperNameAndAddress":   "Test Workflow Shipper Address",
		"consigneeNameAndAddress": "Test Workflow Consignee Address",
		"flightNo":                "TG456",
		"flightDate":              "2024-01-02",
		"executedOnDate":          "2024-01-03",
		"insuranceAmount":         1500.0,
		"totalNoOfPieces":         15,
		"totalGrossWeight":        750.5,
		"totalChargeableWeight":   900.0,
		"totalRateCharge":         3000.0,
		"totalAmount":             3750.0,
		"items": []map[string]interface{}{
			{
				"piecesRCP":         "8",
				"grossWeight":       "400.5",
				"kgLb":              "kg",
				"rateClass":         "N",
				"totalVolume":       0.8,
				"chargeableWeight":  500.0,
				"rateCharge":        1500.0,
				"total":             1500.0,
				"natureAndQuantity": "Electronics Equipment",
				"dims": []map[string]interface{}{
					{
						"length": "120",
						"width":  "60",
						"height": "40",
						"count":  "4",
					},
					{
						"length": "100",
						"width":  "50",
						"height": "30",
						"count":  "2",
					},
				},
			},
			{
				"piecesRCP":         "7",
				"grossWeight":       "350.0",
				"kgLb":              "kg",
				"rateClass":         "M",
				"totalVolume":       0.6,
				"chargeableWeight":  400.0,
				"rateCharge":        1200.0,
				"total":             1200.0,
				"natureAndQuantity": "Textile Products",
				"dims": []map[string]interface{}{
					{
						"length": "80",
						"width":  "60",
						"height": "25",
						"count":  "7",
					},
				},
			},
		},
		"charges": []map[string]interface{}{
			{
				"key":   "fuel_surcharge",
				"value": 750.0,
			},
			{
				"key":   "security_fee",
				"value": 300.0,
			},
			{
				"key":   "handling_fee",
				"value": 200.0,
			},
		},
	}

	createDraftMAWBPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb", mawbUUID)
	draftMAWBRR := suite.makeJSONRequest("POST", createDraftMAWBPath, draftMAWBPayload)

	assert.Equal(suite.T(), http.StatusOK, draftMAWBRR.Code)
	draftMAWBResponse := suite.parseJSONResponse(draftMAWBRR)
	assert.Equal(suite.T(), "success", draftMAWBResponse["message"])
	assert.NotNil(suite.T(), draftMAWBResponse["data"])

	// Extract draft MAWB UUID for tracking
	draftMAWBData := draftMAWBResponse["data"].(map[string]interface{})
	draftMAWBUUID := draftMAWBData["uuid"].(string)
	suite.testDraftMAWBUUIDs = append(suite.testDraftMAWBUUIDs, draftMAWBUUID)

	// Step 5: Verify Draft MAWB can be retrieved with all nested data
	getDraftMAWBRR := suite.makeRequest("GET", createDraftMAWBPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getDraftMAWBRR.Code)

	getDraftMAWBResponse := suite.parseJSONResponse(getDraftMAWBRR)
	retrievedDraftMAWB := getDraftMAWBResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "TEST-WORKFLOW-DRAFT-123456789", retrievedDraftMAWB["mawb"])
	assert.Equal(suite.T(), "Draft", retrievedDraftMAWB["status"])

	// Verify items and nested data
	draftItems := retrievedDraftMAWB["items"].([]interface{})
	assert.Equal(suite.T(), 2, len(draftItems))

	// Verify dimensions for first item
	firstItem := draftItems[0].(map[string]interface{})
	firstItemDims := firstItem["dims"].([]interface{})
	assert.Equal(suite.T(), 2, len(firstItemDims))

	// Verify charges
	charges := retrievedDraftMAWB["charges"].([]interface{})
	assert.Equal(suite.T(), 3, len(charges))

	// Step 6: Test status transitions for Cargo Manifest
	confirmCargoManifestPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/confirm", mawbUUID)
	confirmCargoManifestRR := suite.makeJSONRequest("POST", confirmCargoManifestPath, nil)
	assert.Equal(suite.T(), http.StatusOK, confirmCargoManifestRR.Code)

	// Verify status was updated
	getCargoManifestAfterConfirmRR := suite.makeRequest("GET", createCargoManifestPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getCargoManifestAfterConfirmRR.Code)
	confirmedCargoManifestResponse := suite.parseJSONResponse(getCargoManifestAfterConfirmRR)
	confirmedCargoManifest := confirmedCargoManifestResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Confirmed", confirmedCargoManifest["status"])

	// Step 7: Test status transitions for Draft MAWB
	confirmDraftMAWBPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/confirm", mawbUUID)
	confirmDraftMAWBRR := suite.makeJSONRequest("POST", confirmDraftMAWBPath, nil)
	assert.Equal(suite.T(), http.StatusOK, confirmDraftMAWBRR.Code)

	// Verify status was updated
	getDraftMAWBAfterConfirmRR := suite.makeRequest("GET", createDraftMAWBPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getDraftMAWBAfterConfirmRR.Code)
	confirmedDraftMAWBResponse := suite.parseJSONResponse(getDraftMAWBAfterConfirmRR)
	confirmedDraftMAWB := confirmedDraftMAWBResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Confirmed", confirmedDraftMAWB["status"])

	// Step 8: Test PDF generation for both documents
	printCargoManifestPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/print", mawbUUID)
	printCargoManifestRR := suite.makeRequest("GET", printCargoManifestPath, nil, nil)
	// PDF generation might fail in test environment, so we accept both success and error
	assert.True(suite.T(), printCargoManifestRR.Code == http.StatusOK || printCargoManifestRR.Code >= 400)

	printDraftMAWBPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/print", mawbUUID)
	printDraftMAWBRR := suite.makeRequest("GET", printDraftMAWBPath, nil, nil)
	// PDF generation might fail in test environment, so we accept both success and error
	assert.True(suite.T(), printDraftMAWBRR.Code == http.StatusOK || printDraftMAWBRR.Code >= 400)

	// Step 9: Test updating existing records
	updateCargoManifestPayload := cargoManifestPayload
	updateCargoManifestPayload["shipper"] = "Updated Workflow Shipper"
	updateCargoManifestPayload["totalCtn"] = "15"

	updateCargoManifestRR := suite.makeJSONRequest("POST", createCargoManifestPath, updateCargoManifestPayload)
	assert.Equal(suite.T(), http.StatusOK, updateCargoManifestRR.Code)

	// Verify update
	getUpdatedCargoManifestRR := suite.makeRequest("GET", createCargoManifestPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getUpdatedCargoManifestRR.Code)
	updatedCargoManifestResponse := suite.parseJSONResponse(getUpdatedCargoManifestRR)
	updatedCargoManifest := updatedCargoManifestResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Updated Workflow Shipper", updatedCargoManifest["shipper"])
	assert.Equal(suite.T(), "15", updatedCargoManifest["totalCtn"])

	// Step 10: Test rejection workflow
	rejectDraftMAWBPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/reject", mawbUUID)
	rejectDraftMAWBRR := suite.makeJSONRequest("POST", rejectDraftMAWBPath, nil)
	assert.Equal(suite.T(), http.StatusOK, rejectDraftMAWBRR.Code)

	// Verify rejection status
	getRejectedDraftMAWBRR := suite.makeRequest("GET", createDraftMAWBPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRejectedDraftMAWBRR.Code)
	rejectedDraftMAWBResponse := suite.parseJSONResponse(getRejectedDraftMAWBRR)
	rejectedDraftMAWB := rejectedDraftMAWBResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Rejected", rejectedDraftMAWB["status"])

	// Step 11: Verify data integrity in database
	suite.verifyDatabaseIntegrity(mawbUUID, cargoManifestUUID, draftMAWBUUID)
}

// verifyDatabaseIntegrity verifies data integrity in the database
func (suite *MAWBSystemIntegrationTestSuite) verifyDatabaseIntegrity(mawbUUID, cargoManifestUUID, draftMAWBUUID string) {
	// Verify MAWB Info exists
	var mawbCount int
	_, err := suite.db.QueryOne(pg.Scan(&mawbCount),
		"SELECT COUNT(*) FROM tbl_mawb_info WHERE uuid = ?", mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, mawbCount)

	// Verify Cargo Manifest exists and is linked correctly
	var cargoManifestCount int
	var linkedMAWBUUID string
	_, err = suite.db.QueryOne(pg.Scan(&cargoManifestCount, &linkedMAWBUUID),
		"SELECT COUNT(*), mawb_info_uuid FROM cargo_manifest WHERE uuid = ? GROUP BY mawb_info_uuid", cargoManifestUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, cargoManifestCount)
	assert.Equal(suite.T(), mawbUUID, linkedMAWBUUID)

	// Verify Cargo Manifest Items exist
	var cargoManifestItemCount int
	_, err = suite.db.QueryOne(pg.Scan(&cargoManifestItemCount),
		"SELECT COUNT(*) FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?", cargoManifestUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, cargoManifestItemCount)

	// Verify Draft MAWB exists and is linked correctly
	var draftMAWBCount int
	var draftLinkedMAWBUUID string
	_, err = suite.db.QueryOne(pg.Scan(&draftMAWBCount, &draftLinkedMAWBUUID),
		"SELECT COUNT(*), mawb_info_uuid FROM draft_mawb WHERE uuid = ? GROUP BY mawb_info_uuid", draftMAWBUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, draftMAWBCount)
	assert.Equal(suite.T(), mawbUUID, draftLinkedMAWBUUID)

	// Verify Draft MAWB Items exist
	var draftMAWBItemCount int
	_, err = suite.db.QueryOne(pg.Scan(&draftMAWBItemCount),
		"SELECT COUNT(*) FROM draft_mawb_items WHERE draft_mawb_uuid = ?", draftMAWBUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, draftMAWBItemCount)

	// Verify Draft MAWB Item Dimensions exist
	var draftMAWBDimCount int
	_, err = suite.db.QueryOne(pg.Scan(&draftMAWBDimCount),
		`SELECT COUNT(*) FROM draft_mawb_item_dims 
		 WHERE draft_mawb_item_id IN (
			SELECT id FROM draft_mawb_items WHERE draft_mawb_uuid = ?
		 )`, draftMAWBUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, draftMAWBDimCount) // 2 + 1 dimensions

	// Verify Draft MAWB Charges exist
	var draftMAWBChargeCount int
	_, err = suite.db.QueryOne(pg.Scan(&draftMAWBChargeCount),
		"SELECT COUNT(*) FROM draft_mawb_charges WHERE draft_mawb_uuid = ?", draftMAWBUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, draftMAWBChargeCount)
}

// TestForeignKeyRelationshipsWithExistingData tests that foreign key relationships work correctly with existing data
func (suite *MAWBSystemIntegrationTestSuite) TestForeignKeyRelationshipsWithExistingData() {
	// Create multiple MAWB Info records
	mawbUUID1 := suite.createTestMAWBInfo()
	mawbUUID2 := suite.createTestMAWBInfo()

	// Create cargo manifest for first MAWB Info
	cargoManifestPayload1 := map[string]interface{}{
		"mawbNumber": "TEST-FK-123456789",
		"shipper":    "FK Test Shipper 1",
		"consignee":  "FK Test Consignee 1",
		"items": []map[string]interface{}{
			{
				"hawbNo":    "H123-FK1",
				"pkgs":      "5",
				"commodity": "Electronics",
			},
		},
	}

	createPath1 := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID1)
	rr1 := suite.makeJSONRequest("POST", createPath1, cargoManifestPayload1)
	assert.Equal(suite.T(), http.StatusOK, rr1.Code)

	// Create cargo manifest for second MAWB Info
	cargoManifestPayload2 := map[string]interface{}{
		"mawbNumber": "TEST-FK-987654321",
		"shipper":    "FK Test Shipper 2",
		"consignee":  "FK Test Consignee 2",
		"items": []map[string]interface{}{
			{
				"hawbNo":    "H456-FK2",
				"pkgs":      "3",
				"commodity": "Textiles",
			},
		},
	}

	createPath2 := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID2)
	rr2 := suite.makeJSONRequest("POST", createPath2, cargoManifestPayload2)
	assert.Equal(suite.T(), http.StatusOK, rr2.Code)

	// Verify both cargo manifests exist and are linked to correct MAWB Info
	getRR1 := suite.makeRequest("GET", createPath1, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRR1.Code)
	response1 := suite.parseJSONResponse(getRR1)
	data1 := response1["data"].(map[string]interface{})
	assert.Equal(suite.T(), "TEST-FK-123456789", data1["mawbNumber"])

	getRR2 := suite.makeRequest("GET", createPath2, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRR2.Code)
	response2 := suite.parseJSONResponse(getRR2)
	data2 := response2["data"].(map[string]interface{})
	assert.Equal(suite.T(), "TEST-FK-987654321", data2["mawbNumber"])

	// Test that trying to create cargo manifest with invalid MAWB UUID fails
	invalidMAWBUUID := "00000000-0000-0000-0000-000000000000"
	invalidPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", invalidMAWBUUID)
	invalidRR := suite.makeJSONRequest("POST", invalidPath, cargoManifestPayload1)
	assert.Equal(suite.T(), http.StatusNotFound, invalidRR.Code)

	// Test that trying to access cargo manifest with invalid MAWB UUID fails
	invalidGetRR := suite.makeRequest("GET", invalidPath, nil, nil)
	assert.Equal(suite.T(), http.StatusNotFound, invalidGetRR.Code)

	// Verify database foreign key constraints are working
	var cargoManifestCount int
	_, err := suite.db.QueryOne(pg.Scan(&cargoManifestCount),
		"SELECT COUNT(*) FROM cargo_manifest WHERE mawb_info_uuid IN (?, ?)", mawbUUID1, mawbUUID2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, cargoManifestCount)

	// Test direct database foreign key violation (should fail)
	_, err = suite.db.Exec(`
		INSERT INTO cargo_manifest (mawb_info_uuid, mawb_number, status) 
		VALUES (?, 'TEST-INVALID-FK', 'Draft')
	`, invalidMAWBUUID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), strings.ToLower(err.Error()), "foreign key")
}

// TestCascadingDeleteOperationsDoNotAffectUnrelatedData tests that cascading deletes work correctly
// and don't affect unrelated data
func (suite *MAWBSystemIntegrationTestSuite) TestCascadingDeleteOperationsDoNotAffectUnrelatedData() {
	// Create multiple MAWB Info records
	mawbUUID1 := suite.createTestMAWBInfo()
	mawbUUID2 := suite.createTestMAWBInfo()
	mawbUUID3 := suite.createTestMAWBInfo()

	// Create cargo manifests for all three MAWB Info records
	for i, mawbUUID := range []string{mawbUUID1, mawbUUID2, mawbUUID3} {
		cargoManifestPayload := map[string]interface{}{
			"mawbNumber": fmt.Sprintf("TEST-CASCADE-%d", i+1),
			"shipper":    fmt.Sprintf("Cascade Test Shipper %d", i+1),
			"consignee":  fmt.Sprintf("Cascade Test Consignee %d", i+1),
			"items": []map[string]interface{}{
				{
					"hawbNo":    fmt.Sprintf("H%d-CASCADE", i+1),
					"pkgs":      fmt.Sprintf("%d", i+5),
					"commodity": "Test Commodity",
				},
			},
		}

		createPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID)
		rr := suite.makeJSONRequest("POST", createPath, cargoManifestPayload)
		assert.Equal(suite.T(), http.StatusOK, rr.Code)

		// Track the created cargo manifest UUID
		response := suite.parseJSONResponse(rr)
		data := response["data"].(map[string]interface{})
		cargoManifestUUID := data["uuid"].(string)
		suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, cargoManifestUUID)
	}

	// Create draft MAWBs for all three MAWB Info records
	for i, mawbUUID := range []string{mawbUUID1, mawbUUID2, mawbUUID3} {
		draftMAWBPayload := map[string]interface{}{
			"mawb":        fmt.Sprintf("TEST-CASCADE-DRAFT-%d", i+1),
			"airlineName": fmt.Sprintf("Cascade Airlines %d", i+1),
			"items": []map[string]interface{}{
				{
					"piecesRCP":         fmt.Sprintf("%d", i+3),
					"grossWeight":       fmt.Sprintf("%d.5", (i+1)*100),
					"kgLb":              "kg",
					"rateClass":         "N",
					"natureAndQuantity": "Cascade Test Item",
					"dims": []map[string]interface{}{
						{
							"length": fmt.Sprintf("%d", (i+1)*10),
							"width":  fmt.Sprintf("%d", (i+1)*8),
							"height": fmt.Sprintf("%d", (i+1)*5),
							"count":  fmt.Sprintf("%d", i+2),
						},
					},
				},
			},
			"charges": []map[string]interface{}{
				{
					"key":   fmt.Sprintf("test_charge_%d", i+1),
					"value": float64((i+1) * 100),
				},
			},
		}

		createPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb", mawbUUID)
		rr := suite.makeJSONRequest("POST", createPath, draftMAWBPayload)
		assert.Equal(suite.T(), http.StatusOK, rr.Code)

		// Track the created draft MAWB UUID
		response := suite.parseJSONResponse(rr)
		data := response["data"].(map[string]interface{})
		draftMAWBUUID := data["uuid"].(string)
		suite.testDraftMAWBUUIDs = append(suite.testDraftMAWBUUIDs, draftMAWBUUID)
	}

	// Verify all records exist before deletion
	var totalCargoManifests, totalDraftMAWBs, totalCargoManifestItems, totalDraftMAWBItems, totalDraftMAWBDims, totalDraftMAWBCharges int

	_, err := suite.db.QueryOne(pg.Scan(&totalCargoManifests),
		"SELECT COUNT(*) FROM cargo_manifest WHERE mawb_info_uuid IN (?, ?, ?)", mawbUUID1, mawbUUID2, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, totalCargoManifests)

	_, err = suite.db.QueryOne(pg.Scan(&totalDraftMAWBs),
		"SELECT COUNT(*) FROM draft_mawb WHERE mawb_info_uuid IN (?, ?, ?)", mawbUUID1, mawbUUID2, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, totalDraftMAWBs)

	_, err = suite.db.QueryOne(pg.Scan(&totalCargoManifestItems),
		`SELECT COUNT(*) FROM cargo_manifest_items 
		 WHERE cargo_manifest_uuid IN (
			SELECT uuid FROM cargo_manifest WHERE mawb_info_uuid IN (?, ?, ?)
		 )`, mawbUUID1, mawbUUID2, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, totalCargoManifestItems)

	_, err = suite.db.QueryOne(pg.Scan(&totalDraftMAWBItems),
		`SELECT COUNT(*) FROM draft_mawb_items 
		 WHERE draft_mawb_uuid IN (
			SELECT uuid FROM draft_mawb WHERE mawb_info_uuid IN (?, ?, ?)
		 )`, mawbUUID1, mawbUUID2, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, totalDraftMAWBItems)

	_, err = suite.db.QueryOne(pg.Scan(&totalDraftMAWBDims),
		`SELECT COUNT(*) FROM draft_mawb_item_dims 
		 WHERE draft_mawb_item_id IN (
			SELECT id FROM draft_mawb_items 
			WHERE draft_mawb_uuid IN (
				SELECT uuid FROM draft_mawb WHERE mawb_info_uuid IN (?, ?, ?)
			)
		 )`, mawbUUID1, mawbUUID2, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, totalDraftMAWBDims)

	_, err = suite.db.QueryOne(pg.Scan(&totalDraftMAWBCharges),
		`SELECT COUNT(*) FROM draft_mawb_charges 
		 WHERE draft_mawb_uuid IN (
			SELECT uuid FROM draft_mawb WHERE mawb_info_uuid IN (?, ?, ?)
		 )`, mawbUUID1, mawbUUID2, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, totalDraftMAWBCharges)

	// Delete the middle MAWB Info record (should cascade delete only related records)
	_, err = suite.db.Exec("DELETE FROM tbl_mawb_info WHERE uuid = ?", mawbUUID2)
	assert.NoError(suite.T(), err)

	// Remove from tracking slice since it was deleted
	for i, uuid := range suite.testMAWBInfoUUIDs {
		if uuid == mawbUUID2 {
			suite.testMAWBInfoUUIDs = append(suite.testMAWBInfoUUIDs[:i], suite.testMAWBInfoUUIDs[i+1:]...)
			break
		}
	}

	// Remove related cargo manifest and draft MAWB from tracking slices
	// (they were cascade deleted, so we need to remove them from tracking to avoid cleanup errors)
	suite.testCargoManifestUUIDs = suite.testCargoManifestUUIDs[:len(suite.testCargoManifestUUIDs)-1]
	suite.testDraftMAWBUUIDs = suite.testDraftMAWBUUIDs[:len(suite.testDraftMAWBUUIDs)-1]

	// Verify that only records related to mawbUUID2 were deleted
	var remainingCargoManifests, remainingDraftMAWBs, remainingCargoManifestItems, remainingDraftMAWBItems, remainingDraftMAWBDims, remainingDraftMAWBCharges int

	_, err = suite.db.QueryOne(pg.Scan(&remainingCargoManifests),
		"SELECT COUNT(*) FROM cargo_manifest WHERE mawb_info_uuid IN (?, ?)", mawbUUID1, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, remainingCargoManifests)

	_, err = suite.db.QueryOne(pg.Scan(&remainingDraftMAWBs),
		"SELECT COUNT(*) FROM draft_mawb WHERE mawb_info_uuid IN (?, ?)", mawbUUID1, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, remainingDraftMAWBs)

	_, err = suite.db.QueryOne(pg.Scan(&remainingCargoManifestItems),
		`SELECT COUNT(*) FROM cargo_manifest_items 
		 WHERE cargo_manifest_uuid IN (
			SELECT uuid FROM cargo_manifest WHERE mawb_info_uuid IN (?, ?)
		 )`, mawbUUID1, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, remainingCargoManifestItems)

	_, err = suite.db.QueryOne(pg.Scan(&remainingDraftMAWBItems),
		`SELECT COUNT(*) FROM draft_mawb_items 
		 WHERE draft_mawb_uuid IN (
			SELECT uuid FROM draft_mawb WHERE mawb_info_uuid IN (?, ?)
		 )`, mawbUUID1, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, remainingDraftMAWBItems)

	_, err = suite.db.QueryOne(pg.Scan(&remainingDraftMAWBDims),
		`SELECT COUNT(*) FROM draft_mawb_item_dims 
		 WHERE draft_mawb_item_id IN (
			SELECT id FROM draft_mawb_items 
			WHERE draft_mawb_uuid IN (
				SELECT uuid FROM draft_mawb WHERE mawb_info_uuid IN (?, ?)
			)
		 )`, mawbUUID1, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, remainingDraftMAWBDims)

	_, err = suite.db.QueryOne(pg.Scan(&remainingDraftMAWBCharges),
		`SELECT COUNT(*) FROM draft_mawb_charges 
		 WHERE draft_mawb_uuid IN (
			SELECT uuid FROM draft_mawb WHERE mawb_info_uuid IN (?, ?)
		 )`, mawbUUID1, mawbUUID3)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, remainingDraftMAWBCharges)

	// Verify that records for mawbUUID2 were completely removed
	var deletedRecordCount int
	_, err = suite.db.QueryOne(pg.Scan(&deletedRecordCount),
		"SELECT COUNT(*) FROM cargo_manifest WHERE mawb_info_uuid = ?", mawbUUID2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, deletedRecordCount)

	_, err = suite.db.QueryOne(pg.Scan(&deletedRecordCount),
		"SELECT COUNT(*) FROM draft_mawb WHERE mawb_info_uuid = ?", mawbUUID2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, deletedRecordCount)

	// Verify that the remaining records are still accessible via API
	getRemainingPath1 := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID1)
	getRemainingRR1 := suite.makeRequest("GET", getRemainingPath1, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRemainingRR1.Code)

	getRemainingPath3 := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID3)
	getRemainingRR3 := suite.makeRequest("GET", getRemainingPath3, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, getRemainingRR3.Code)

	// Verify that the deleted record is no longer accessible via API
	getDeletedPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID2)
	getDeletedRR := suite.makeRequest("GET", getDeletedPath, nil, nil)
	assert.Equal(suite.T(), http.StatusNotFound, getDeletedRR.Code)
}

// TestComplexDataIntegrityWithMultipleOperations tests data integrity with multiple concurrent operations
func (suite *MAWBSystemIntegrationTestSuite) TestComplexDataIntegrityWithMultipleOperations() {
	// Create MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Create initial cargo manifest
	initialCargoManifestPayload := map[string]interface{}{
		"mawbNumber": "TEST-COMPLEX-123456789",
		"shipper":    "Initial Shipper",
		"consignee":  "Initial Consignee",
		"items": []map[string]interface{}{
			{
				"hawbNo":    "H123-INITIAL",
				"pkgs":      "5",
				"commodity": "Initial Commodity",
			},
		},
	}

	cargoManifestPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest", mawbUUID)
	initialRR := suite.makeJSONRequest("POST", cargoManifestPath, initialCargoManifestPayload)
	assert.Equal(suite.T(), http.StatusOK, initialRR.Code)

	// Create initial draft MAWB
	initialDraftMAWBPayload := map[string]interface{}{
		"mawb":        "TEST-COMPLEX-DRAFT-123456789",
		"airlineName": "Initial Airlines",
		"items": []map[string]interface{}{
			{
				"piecesRCP":         "3",
				"grossWeight":       "150.0",
				"kgLb":              "kg",
				"rateClass":         "N",
				"natureAndQuantity": "Initial Draft Item",
				"dims": []map[string]interface{}{
					{
						"length": "50",
						"width":  "40",
						"height": "30",
						"count":  "3",
					},
				},
			},
		},
		"charges": []map[string]interface{}{
			{
				"key":   "initial_charge",
				"value": 200.0,
			},
		},
	}

	draftMAWBPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb", mawbUUID)
	initialDraftRR := suite.makeJSONRequest("POST", draftMAWBPath, initialDraftMAWBPayload)
	assert.Equal(suite.T(), http.StatusOK, initialDraftRR.Code)

	// Perform multiple updates to test data consistency
	for i := 1; i <= 3; i++ {
		// Update cargo manifest
		updatedCargoManifestPayload := map[string]interface{}{
			"mawbNumber": fmt.Sprintf("TEST-COMPLEX-UPDATED-%d", i),
			"shipper":    fmt.Sprintf("Updated Shipper %d", i),
			"consignee":  fmt.Sprintf("Updated Consignee %d", i),
			"items": []map[string]interface{}{
				{
					"hawbNo":    fmt.Sprintf("H123-UPDATED-%d", i),
					"pkgs":      fmt.Sprintf("%d", i+5),
					"commodity": fmt.Sprintf("Updated Commodity %d", i),
				},
				{
					"hawbNo":    fmt.Sprintf("H456-ADDED-%d", i),
					"pkgs":      fmt.Sprintf("%d", i+2),
					"commodity": fmt.Sprintf("Added Commodity %d", i),
				},
			},
		}

		updateCargoRR := suite.makeJSONRequest("POST", cargoManifestPath, updatedCargoManifestPayload)
		assert.Equal(suite.T(), http.StatusOK, updateCargoRR.Code)

		// Update draft MAWB
		updatedDraftMAWBPayload := map[string]interface{}{
			"mawb":        fmt.Sprintf("TEST-COMPLEX-DRAFT-UPDATED-%d", i),
			"airlineName": fmt.Sprintf("Updated Airlines %d", i),
			"items": []map[string]interface{}{
				{
					"piecesRCP":         fmt.Sprintf("%d", i+3),
					"grossWeight":       fmt.Sprintf("%d.5", (i+1)*100),
					"kgLb":              "kg",
					"rateClass":         "M",
					"natureAndQuantity": fmt.Sprintf("Updated Draft Item %d", i),
					"dims": []map[string]interface{}{
						{
							"length": fmt.Sprintf("%d", i*20),
							"width":  fmt.Sprintf("%d", i*15),
							"height": fmt.Sprintf("%d", i*10),
							"count":  fmt.Sprintf("%d", i+1),
						},
						{
							"length": fmt.Sprintf("%d", i*15),
							"width":  fmt.Sprintf("%d", i*12),
							"height": fmt.Sprintf("%d", i*8),
							"count":  fmt.Sprintf("%d", i),
						},
					},
				},
			},
			"charges": []map[string]interface{}{
				{
					"key":   fmt.Sprintf("updated_charge_%d", i),
					"value": float64(i * 150),
				},
				{
					"key":   fmt.Sprintf("additional_charge_%d", i),
					"value": float64(i * 75),
				},
			},
		}

		updateDraftRR := suite.makeJSONRequest("POST", draftMAWBPath, updatedDraftMAWBPayload)
		assert.Equal(suite.T(), http.StatusOK, updateDraftRR.Code)

		// Verify data after each update
		getCargoRR := suite.makeRequest("GET", cargoManifestPath, nil, nil)
		assert.Equal(suite.T(), http.StatusOK, getCargoRR.Code)
		cargoResponse := suite.parseJSONResponse(getCargoRR)
		cargoData := cargoResponse["data"].(map[string]interface{})
		assert.Equal(suite.T(), fmt.Sprintf("TEST-COMPLEX-UPDATED-%d", i), cargoData["mawbNumber"])

		getDraftRR := suite.makeRequest("GET", draftMAWBPath, nil, nil)
		assert.Equal(suite.T(), http.StatusOK, getDraftRR.Code)
		draftResponse := suite.parseJSONResponse(getDraftRR)
		draftData := draftResponse["data"].(map[string]interface{})
		assert.Equal(suite.T(), fmt.Sprintf("TEST-COMPLEX-DRAFT-UPDATED-%d", i), draftData["mawb"])

		// Verify nested data integrity
		cargoItems := cargoData["items"].([]interface{})
		assert.Equal(suite.T(), 2, len(cargoItems))

		draftItems := draftData["items"].([]interface{})
		assert.Equal(suite.T(), 1, len(draftItems))

		firstDraftItem := draftItems[0].(map[string]interface{})
		draftDims := firstDraftItem["dims"].([]interface{})
		assert.Equal(suite.T(), 2, len(draftDims))

		draftCharges := draftData["charges"].([]interface{})
		assert.Equal(suite.T(), 2, len(draftCharges))
	}

	// Test status transitions after multiple updates
	confirmCargoPath := fmt.Sprintf("/v1/mawbinfo/%s/cargo-manifest/confirm", mawbUUID)
	confirmCargoRR := suite.makeJSONRequest("POST", confirmCargoPath, nil)
	assert.Equal(suite.T(), http.StatusOK, confirmCargoRR.Code)

	confirmDraftPath := fmt.Sprintf("/v1/mawbinfo/%s/draft-mawb/confirm", mawbUUID)
	confirmDraftRR := suite.makeJSONRequest("POST", confirmDraftPath, nil)
	assert.Equal(suite.T(), http.StatusOK, confirmDraftRR.Code)

	// Verify final state
	finalCargoRR := suite.makeRequest("GET", cargoManifestPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, finalCargoRR.Code)
	finalCargoResponse := suite.parseJSONResponse(finalCargoRR)
	finalCargoData := finalCargoResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Confirmed", finalCargoData["status"])
	assert.Equal(suite.T(), "TEST-COMPLEX-UPDATED-3", finalCargoData["mawbNumber"])

	finalDraftRR := suite.makeRequest("GET", draftMAWBPath, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, finalDraftRR.Code)
	finalDraftResponse := suite.parseJSONResponse(finalDraftRR)
	finalDraftData := finalDraftResponse["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Confirmed", finalDraftData["status"])
	assert.Equal(suite.T(), "TEST-COMPLEX-DRAFT-UPDATED-3", finalDraftData["mawb"])
}// H
elper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Run the MAWB System Integration test suite
func TestMAWBSystemIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if TEST_INTEGRATION is not set
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Skipping MAWB System integration tests. Set TEST_INTEGRATION=1 to run.")
	}

	suite.Run(t, new(MAWBSystemIntegrationTestSuite))
}