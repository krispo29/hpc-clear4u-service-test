package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"hpc-express-service/cargo_manifest"
	"hpc-express-service/config"
	"hpc-express-service/database"
	"hpc-express-service/draft_mawb"
)

// DatabaseIntegrationTestSuite defines the integration test suite for database operations
type DatabaseIntegrationTestSuite struct {
	suite.Suite
	db                     *pg.DB
	ctx                    context.Context
	cargoManifestRepo      cargo_manifest.Repository
	draftMAWBRepo          draft_mawb.Repository
	testMAWBInfoUUIDs      []string
	testCargoManifestUUIDs []string
	testDraftMAWBUUIDs     []string
}

// SetupSuite sets up the test database connection and repositories
func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	// Load test configuration
	testConfig := &config.Config{
		PostgreSQLHost:     getEnvOrDefault("TEST_POSTGRESQL_HOST", "localhost"),
		PostgreSQLUser:     getEnvOrDefault("TEST_POSTGRESQL_USER", "postgres"),
		PostgreSQLPassword: getEnvOrDefault("TEST_POSTGRESQL_PASSWORD", "password"),
		PostgreSQLName:     getEnvOrDefault("TEST_POSTGRESQL_NAME", "test_db"),
		PostgreSQLPort:     getEnvOrDefault("TEST_POSTGRESQL_PORT", "5432"),
		PostgreSQLSSLMode:  getEnvOrDefault("TEST_POSTGRESQL_SSLMODE", "false"),
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

	// Set up context with database connection
	suite.ctx = context.WithValue(context.Background(), "postgreSQLConn", suite.db)

	// Initialize repositories
	suite.cargoManifestRepo = cargo_manifest.NewRepository(30 * time.Second)
	suite.draftMAWBRepo = draft_mawb.NewRepository(30 * time.Second)

	// Initialize slices for tracking test data
	suite.testMAWBInfoUUIDs = make([]string, 0)
	suite.testCargoManifestUUIDs = make([]string, 0)
	suite.testDraftMAWBUUIDs = make([]string, 0)

	// Run database migrations if needed
	suite.runMigrations()
}

// TearDownSuite cleans up after all tests
func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		// Clean up all test data
		suite.cleanupAllTestData()
		suite.db.Close()
	}
}

// SetupTest runs before each test
func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Start a transaction for each test (optional, for isolation)
	// This can be implemented if needed for better test isolation
}

// TearDownTest runs after each test
func (suite *DatabaseIntegrationTestSuite) TearDownTest() {
	// Clean up test data created during the test
	suite.cleanupTestData()
}

// runMigrations runs the database migrations for testing
func (suite *DatabaseIntegrationTestSuite) runMigrations() {
	// Create test MAWB Info table if it doesn't exist (simplified for testing)
	_, err := suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS tbl_mawb_info (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mawb_number VARCHAR(255) NOT NULL,
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
func (suite *DatabaseIntegrationTestSuite) runCargoManifestMigrations() {
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
func (suite *DatabaseIntegrationTestSuite) runDraftMAWBMigrations() {
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
func (suite *DatabaseIntegrationTestSuite) createTestMAWBInfo() string {
	var uuid string
	_, err := suite.db.QueryOne(pg.Scan(&uuid), `
		INSERT INTO tbl_mawb_info (mawb_number) 
		VALUES (?) 
		RETURNING uuid
	`, fmt.Sprintf("TEST-MAWB-%d", time.Now().UnixNano()))

	if err != nil {
		suite.T().Fatalf("Failed to create test MAWB Info: %v", err)
	}

	suite.testMAWBInfoUUIDs = append(suite.testMAWBInfoUUIDs, uuid)
	return uuid
}

// createTestCargoManifest creates a test cargo manifest
func (suite *DatabaseIntegrationTestSuite) createTestCargoManifest(mawbUUID string) *cargo_manifest.CargoManifest {
	manifest := &cargo_manifest.CargoManifest{
		MAWBInfoUUID:    mawbUUID,
		MAWBNumber:      "TEST-123456789",
		PortOfDischarge: "BKK",
		FlightNo:        "TG123",
		FreightDate:     "2024-01-01",
		Shipper:         "Test Shipper",
		Consignee:       "Test Consignee",
		TotalCtn:        "10",
		Transshipment:   "No",
		Items: []cargo_manifest.CargoManifestItem{
			{
				HAWBNo:                  "H123",
				Pkgs:                    "5",
				GrossWeight:             "100.5",
				Destination:             "NYC",
				Commodity:               "Electronics",
				ShipperNameAndAddress:   "Shipper Address",
				ConsigneeNameAndAddress: "Consignee Address",
			},
			{
				HAWBNo:                  "H456",
				Pkgs:                    "3",
				GrossWeight:             "75.0",
				Destination:             "LAX",
				Commodity:               "Textiles",
				ShipperNameAndAddress:   "Another Shipper",
				ConsigneeNameAndAddress: "Another Consignee",
			},
		},
	}

	return manifest
}

// createTestDraftMAWB creates a test draft MAWB
func (suite *DatabaseIntegrationTestSuite) createTestDraftMAWB(mawbUUID string) *draft_mawb.DraftMAWB {
	flightDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	executedDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	draftMAWB := &draft_mawb.DraftMAWB{
		MAWBInfoUUID:            mawbUUID,
		CustomerUUID:            "customer-uuid",
		AirlineLogo:             "logo.png",
		AirlineName:             "Test Airlines",
		MAWB:                    "TEST-123456789",
		HAWB:                    "H123456",
		ShipperNameAndAddress:   "Test Shipper Address",
		ConsigneeNameAndAddress: "Test Consignee Address",
		FlightNo:                "TG123",
		FlightDate:              &flightDate,
		ExecutedOnDate:          &executedDate,
		InsuranceAmount:         1000.0,
		TotalNoOfPieces:         10,
		TotalGrossWeight:        500.5,
		TotalChargeableWeight:   600.0,
		TotalRateCharge:         2000.0,
		TotalAmount:             2500.0,
		Items: []draft_mawb.DraftMAWBItem{
			{
				PiecesRCP:         "5",
				GrossWeight:       "250.5",
				KgLb:              "kg",
				RateClass:         "N",
				TotalVolume:       0.5,
				ChargeableWeight:  300.0,
				RateCharge:        1000.0,
				Total:             1000.0,
				NatureAndQuantity: "Electronics",
				Dims: []draft_mawb.DraftMAWBItemDim{
					{
						Length: "100",
						Width:  "50",
						Height: "30",
						Count:  "2",
					},
					{
						Length: "80",
						Width:  "40",
						Height: "25",
						Count:  "1",
					},
				},
			},
			{
				PiecesRCP:         "3",
				GrossWeight:       "150.0",
				KgLb:              "kg",
				RateClass:         "M",
				TotalVolume:       0.3,
				ChargeableWeight:  200.0,
				RateCharge:        800.0,
				Total:             800.0,
				NatureAndQuantity: "Textiles",
				Dims: []draft_mawb.DraftMAWBItemDim{
					{
						Length: "60",
						Width:  "40",
						Height: "20",
						Count:  "3",
					},
				},
			},
		},
		Charges: []draft_mawb.DraftMAWBCharge{
			{
				Key:   "fuel_surcharge",
				Value: 500.0,
			},
			{
				Key:   "security_fee",
				Value: 200.0,
			},
		},
	}

	return draftMAWB
}

// cleanupTestData cleans up test data created during individual tests
func (suite *DatabaseIntegrationTestSuite) cleanupTestData() {
	// Clean up in reverse dependency order

	// Clean up draft MAWB charges
	if len(suite.testDraftMAWBUUIDs) > 0 {
		for _, uuid := range suite.testDraftMAWBUUIDs {
			suite.db.Exec("DELETE FROM draft_mawb_charges WHERE draft_mawb_uuid = ?", uuid)
		}
	}

	// Clean up draft MAWB item dimensions
	suite.db.Exec(`
		DELETE FROM draft_mawb_item_dims 
		WHERE draft_mawb_item_id IN (
			SELECT id FROM draft_mawb_items 
			WHERE draft_mawb_uuid = ANY(?)
		)
	`, suite.testDraftMAWBUUIDs)

	// Clean up draft MAWB items
	if len(suite.testDraftMAWBUUIDs) > 0 {
		for _, uuid := range suite.testDraftMAWBUUIDs {
			suite.db.Exec("DELETE FROM draft_mawb_items WHERE draft_mawb_uuid = ?", uuid)
		}
	}

	// Clean up draft MAWBs
	if len(suite.testDraftMAWBUUIDs) > 0 {
		for _, uuid := range suite.testDraftMAWBUUIDs {
			suite.db.Exec("DELETE FROM draft_mawb WHERE uuid = ?", uuid)
		}
	}

	// Clean up cargo manifest items
	if len(suite.testCargoManifestUUIDs) > 0 {
		for _, uuid := range suite.testCargoManifestUUIDs {
			suite.db.Exec("DELETE FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?", uuid)
		}
	}

	// Clean up cargo manifests
	if len(suite.testCargoManifestUUIDs) > 0 {
		for _, uuid := range suite.testCargoManifestUUIDs {
			suite.db.Exec("DELETE FROM cargo_manifest WHERE uuid = ?", uuid)
		}
	}

	// Clean up MAWB Info records
	if len(suite.testMAWBInfoUUIDs) > 0 {
		for _, uuid := range suite.testMAWBInfoUUIDs {
			suite.db.Exec("DELETE FROM tbl_mawb_info WHERE uuid = ?", uuid)
		}
	}

	// Reset tracking slices
	suite.testMAWBInfoUUIDs = suite.testMAWBInfoUUIDs[:0]
	suite.testCargoManifestUUIDs = suite.testCargoManifestUUIDs[:0]
	suite.testDraftMAWBUUIDs = suite.testDraftMAWBUUIDs[:0]
}

// cleanupAllTestData cleans up all test data
func (suite *DatabaseIntegrationTestSuite) cleanupAllTestData() {
	// Clean up all test data in reverse dependency order
	suite.db.Exec("DELETE FROM draft_mawb_charges WHERE charge_key LIKE 'test_%'")
	suite.db.Exec("DELETE FROM draft_mawb_item_dims WHERE length LIKE 'test_%'")
	suite.db.Exec("DELETE FROM draft_mawb_items WHERE pieces_rcp LIKE 'test_%'")
	suite.db.Exec("DELETE FROM draft_mawb WHERE mawb LIKE 'TEST-%'")
	suite.db.Exec("DELETE FROM cargo_manifest_items WHERE hawb_no LIKE 'H%'")
	suite.db.Exec("DELETE FROM cargo_manifest WHERE mawb_number LIKE 'TEST-%'")
	suite.db.Exec("DELETE FROM tbl_mawb_info WHERE mawb_number LIKE 'TEST-%'")
}

// Test Cases for Database Integration

// TestCargoManifestCRUDOperations tests complete CRUD operations for cargo manifest
func (suite *DatabaseIntegrationTestSuite) TestCargoManifestCRUDOperations() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Test Create
	manifest := suite.createTestCargoManifest(mawbUUID)
	err := suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), manifest.UUID)
	assert.Equal(suite.T(), "Draft", manifest.Status)
	suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, manifest.UUID)

	// Test Read
	retrievedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedManifest)
	assert.Equal(suite.T(), manifest.UUID, retrievedManifest.UUID)
	assert.Equal(suite.T(), manifest.MAWBNumber, retrievedManifest.MAWBNumber)
	assert.Equal(suite.T(), len(manifest.Items), len(retrievedManifest.Items))

	// Test Update
	manifest.MAWBNumber = "UPDATED-123456789"
	manifest.Shipper = "Updated Shipper"
	err = suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
	assert.NoError(suite.T(), err)

	// Verify update
	updatedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "UPDATED-123456789", updatedManifest.MAWBNumber)
	assert.Equal(suite.T(), "Updated Shipper", updatedManifest.Shipper)

	// Test Status Update
	err = suite.cargoManifestRepo.UpdateStatus(suite.ctx, manifest.UUID, "Confirmed")
	assert.NoError(suite.T(), err)

	// Verify status update
	confirmedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Confirmed", confirmedManifest.Status)
}

// TestDraftMAWBCRUDOperations tests complete CRUD operations for draft MAWB
func (suite *DatabaseIntegrationTestSuite) TestDraftMAWBCRUDOperations() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Test Create
	draftMAWB := suite.createTestDraftMAWB(mawbUUID)
	err := suite.draftMAWBRepo.CreateOrUpdate(suite.ctx, draftMAWB)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), draftMAWB.UUID)
	assert.Equal(suite.T(), "Draft", draftMAWB.Status)
	suite.testDraftMAWBUUIDs = append(suite.testDraftMAWBUUIDs, draftMAWB.UUID)

	// Test Read
	retrievedDraftMAWB, err := suite.draftMAWBRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedDraftMAWB)
	assert.Equal(suite.T(), draftMAWB.UUID, retrievedDraftMAWB.UUID)
	assert.Equal(suite.T(), draftMAWB.MAWB, retrievedDraftMAWB.MAWB)
	assert.Equal(suite.T(), len(draftMAWB.Items), len(retrievedDraftMAWB.Items))
	assert.Equal(suite.T(), len(draftMAWB.Charges), len(retrievedDraftMAWB.Charges))

	// Verify nested data
	if len(retrievedDraftMAWB.Items) > 0 {
		assert.Equal(suite.T(), len(draftMAWB.Items[0].Dims), len(retrievedDraftMAWB.Items[0].Dims))
	}

	// Test Update
	draftMAWB.MAWB = "UPDATED-123456789"
	draftMAWB.AirlineName = "Updated Airlines"
	draftMAWB.InsuranceAmount = 2000.0
	err = suite.draftMAWBRepo.CreateOrUpdate(suite.ctx, draftMAWB)
	assert.NoError(suite.T(), err)

	// Verify update
	updatedDraftMAWB, err := suite.draftMAWBRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "UPDATED-123456789", updatedDraftMAWB.MAWB)
	assert.Equal(suite.T(), "Updated Airlines", updatedDraftMAWB.AirlineName)
	assert.Equal(suite.T(), 2000.0, updatedDraftMAWB.InsuranceAmount)

	// Test Status Update
	err = suite.draftMAWBRepo.UpdateStatus(suite.ctx, draftMAWB.UUID, "Confirmed")
	assert.NoError(suite.T(), err)

	// Verify status update
	confirmedDraftMAWB, err := suite.draftMAWBRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Confirmed", confirmedDraftMAWB.Status)
}

// TestForeignKeyRelationships tests foreign key relationships and constraints
func (suite *DatabaseIntegrationTestSuite) TestForeignKeyRelationships() {
	// Test creating cargo manifest with invalid MAWB Info UUID
	invalidMAWBUUID := "00000000-0000-0000-0000-000000000000"
	manifest := suite.createTestCargoManifest(invalidMAWBUUID)

	err := suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "foreign key")

	// Test creating draft MAWB with invalid MAWB Info UUID
	draftMAWB := suite.createTestDraftMAWB(invalidMAWBUUID)

	err = suite.draftMAWBRepo.CreateOrUpdate(suite.ctx, draftMAWB)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "foreign key")

	// Test valid foreign key relationship
	validMAWBUUID := suite.createTestMAWBInfo()

	// Test MAWB validation
	err = suite.cargoManifestRepo.ValidateMAWBExists(suite.ctx, validMAWBUUID)
	assert.NoError(suite.T(), err)

	err = suite.draftMAWBRepo.ValidateMAWBExists(suite.ctx, validMAWBUUID)
	assert.NoError(suite.T(), err)

	// Test invalid MAWB validation
	err = suite.cargoManifestRepo.ValidateMAWBExists(suite.ctx, invalidMAWBUUID)
	assert.Error(suite.T(), err)

	err = suite.draftMAWBRepo.ValidateMAWBExists(suite.ctx, invalidMAWBUUID)
	assert.Error(suite.T(), err)
}

// TestCascadingDeletes tests cascading delete operations
func (suite *DatabaseIntegrationTestSuite) TestCascadingDeletes() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Create cargo manifest with items
	manifest := suite.createTestCargoManifest(mawbUUID)
	err := suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
	assert.NoError(suite.T(), err)
	suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, manifest.UUID)

	// Create draft MAWB with items, dimensions, and charges
	draftMAWB := suite.createTestDraftMAWB(mawbUUID)
	err = suite.draftMAWBRepo.CreateOrUpdate(suite.ctx, draftMAWB)
	assert.NoError(suite.T(), err)
	suite.testDraftMAWBUUIDs = append(suite.testDraftMAWBUUIDs, draftMAWB.UUID)

	// Verify data exists
	retrievedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedManifest)

	retrievedDraftMAWB, err := suite.draftMAWBRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedDraftMAWB)

	// Delete MAWB Info (should cascade delete related records)
	_, err = suite.db.Exec("DELETE FROM tbl_mawb_info WHERE uuid = ?", mawbUUID)
	assert.NoError(suite.T(), err)

	// Verify cascading deletes worked
	retrievedManifest, err = suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedManifest)

	retrievedDraftMAWB, err = suite.draftMAWBRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedDraftMAWB)

	// Verify related items were also deleted
	var itemCount int
	_, err = suite.db.QueryOne(pg.Scan(&itemCount),
		"SELECT COUNT(*) FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?",
		manifest.UUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, itemCount)

	_, err = suite.db.QueryOne(pg.Scan(&itemCount),
		"SELECT COUNT(*) FROM draft_mawb_items WHERE draft_mawb_uuid = ?",
		draftMAWB.UUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, itemCount)

	// Remove from tracking slices since they were cascade deleted
	suite.testMAWBInfoUUIDs = suite.testMAWBInfoUUIDs[:len(suite.testMAWBInfoUUIDs)-1]
	suite.testCargoManifestUUIDs = suite.testCargoManifestUUIDs[:len(suite.testCargoManifestUUIDs)-1]
	suite.testDraftMAWBUUIDs = suite.testDraftMAWBUUIDs[:len(suite.testDraftMAWBUUIDs)-1]
}

// TestTransactionRollback tests transaction rollback scenarios
func (suite *DatabaseIntegrationTestSuite) TestTransactionRollback() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Start a transaction
	tx, err := suite.db.Begin()
	assert.NoError(suite.T(), err)

	// Create context with transaction
	txCtx := context.WithValue(context.Background(), "postgreSQLConn", tx)

	// Create cargo manifest in transaction
	manifest := suite.createTestCargoManifest(mawbUUID)
	err = suite.cargoManifestRepo.CreateOrUpdate(txCtx, manifest)
	assert.NoError(suite.T(), err)

	// Verify data exists in transaction
	retrievedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(txCtx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedManifest)

	// Rollback transaction
	err = tx.Rollback()
	assert.NoError(suite.T(), err)

	// Verify data was rolled back (doesn't exist outside transaction)
	retrievedManifest, err = suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedManifest)
}

// TestDataConsistency tests data consistency across related tables
func (suite *DatabaseIntegrationTestSuite) TestDataConsistency() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Create cargo manifest with multiple items
	manifest := suite.createTestCargoManifest(mawbUUID)
	manifest.Items = append(manifest.Items, cargo_manifest.CargoManifestItem{
		HAWBNo:                  "H789",
		Pkgs:                    "2",
		GrossWeight:             "50.0",
		Destination:             "SFO",
		Commodity:               "Documents",
		ShipperNameAndAddress:   "Document Shipper",
		ConsigneeNameAndAddress: "Document Consignee",
	})

	err := suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
	assert.NoError(suite.T(), err)
	suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, manifest.UUID)

	// Verify all items were saved
	retrievedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(retrievedManifest.Items))

	// Create draft MAWB with complex nested data
	draftMAWB := suite.createTestDraftMAWB(mawbUUID)

	// Add more items and dimensions
	draftMAWB.Items = append(draftMAWB.Items, draft_mawb.DraftMAWBItem{
		PiecesRCP:         "2",
		GrossWeight:       "75.0",
		KgLb:              "lb",
		RateClass:         "Q",
		TotalVolume:       0.2,
		ChargeableWeight:  100.0,
		RateCharge:        500.0,
		Total:             500.0,
		NatureAndQuantity: "Documents",
		Dims: []draft_mawb.DraftMAWBItemDim{
			{
				Length: "30",
				Width:  "20",
				Height: "5",
				Count:  "10",
			},
		},
	})

	// Add more charges
	draftMAWB.Charges = append(draftMAWB.Charges, draft_mawb.DraftMAWBCharge{
		Key:   "handling_fee",
		Value: 150.0,
	})

	err = suite.draftMAWBRepo.CreateOrUpdate(suite.ctx, draftMAWB)
	assert.NoError(suite.T(), err)
	suite.testDraftMAWBUUIDs = append(suite.testDraftMAWBUUIDs, draftMAWB.UUID)

	// Verify all nested data was saved correctly
	retrievedDraftMAWB, err := suite.draftMAWBRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(retrievedDraftMAWB.Items))
	assert.Equal(suite.T(), 3, len(retrievedDraftMAWB.Charges))

	// Verify dimensions for each item
	totalDims := 0
	for _, item := range retrievedDraftMAWB.Items {
		totalDims += len(item.Dims)
	}
	assert.Equal(suite.T(), 4, totalDims) // 2 + 1 + 1 dimensions

	// Test data consistency after updates
	manifest.Items[0].GrossWeight = "999.9"
	err = suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
	assert.NoError(suite.T(), err)

	// Verify update consistency
	updatedManifest, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "999.9", updatedManifest.Items[0].GrossWeight)
	assert.Equal(suite.T(), 3, len(updatedManifest.Items)) // Should still have all items
}

// TestConcurrentAccess tests concurrent access to the database
func (suite *DatabaseIntegrationTestSuite) TestConcurrentAccess() {
	// Create test MAWB Info
	mawbUUID := suite.createTestMAWBInfo()

	// Create multiple goroutines that try to create/update manifests concurrently
	const numGoroutines = 5
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			manifest := suite.createTestCargoManifest(mawbUUID)
			manifest.MAWBNumber = fmt.Sprintf("CONCURRENT-%d-%d", index, time.Now().UnixNano())

			err := suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
			if err == nil {
				suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, manifest.UUID)
			}
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete and check for errors
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// In a properly implemented system, we should have minimal or no errors
	// The exact assertion depends on the concurrency control implementation
	assert.True(suite.T(), len(errors) < numGoroutines, "Too many concurrent access errors")
}

// TestDatabaseIndexes tests that database indexes are working properly
func (suite *DatabaseIntegrationTestSuite) TestDatabaseIndexes() {
	// Create multiple test MAWB Info records
	mawbUUIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		mawbUUIDs[i] = suite.createTestMAWBInfo()
	}

	// Create cargo manifests for each MAWB Info
	for i, mawbUUID := range mawbUUIDs {
		manifest := suite.createTestCargoManifest(mawbUUID)
		manifest.MAWBNumber = fmt.Sprintf("INDEX-TEST-%d", i)
		manifest.Status = []string{"Draft", "Pending", "Confirmed", "Rejected"}[i%4]

		err := suite.cargoManifestRepo.CreateOrUpdate(suite.ctx, manifest)
		assert.NoError(suite.T(), err)
		suite.testCargoManifestUUIDs = append(suite.testCargoManifestUUIDs, manifest.UUID)
	}

	// Test index on mawb_info_uuid (should be fast)
	start := time.Now()
	for _, mawbUUID := range mawbUUIDs {
		_, err := suite.cargoManifestRepo.GetByMAWBUUID(suite.ctx, mawbUUID)
		assert.NoError(suite.T(), err)
	}
	indexedQueryTime := time.Since(start)

	// The indexed queries should complete quickly
	assert.True(suite.T(), indexedQueryTime < 100*time.Millisecond,
		"Indexed queries took too long: %v", indexedQueryTime)

	// Test that we can query by status efficiently (status index)
	var count int
	_, err := suite.db.QueryOne(pg.Scan(&count),
		"SELECT COUNT(*) FROM cargo_manifest WHERE status = ?", "Draft")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), count >= 1)
}

// Run the integration test suite
func TestDatabaseIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if TEST_INTEGRATION is not set
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Skipping integration tests. Set TEST_INTEGRATION=1 to run.")
	}

	suite.Run(t, new(DatabaseIntegrationTestSuite))
}
