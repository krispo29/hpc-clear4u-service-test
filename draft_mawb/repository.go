package draft_mawb

import (
	"context"
	"fmt"
	"hpc-express-service/common"
	"hpc-express-service/utils"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

// DraftMAWBFilters represents filters for querying draft MAWB records
type DraftMAWBFilters struct {
	Status        []string   `json:"status"`
	CustomerUUID  string     `json:"customer_uuid"`
	CreatedAfter  *time.Time `json:"created_after"`
	CreatedBefore *time.Time `json:"created_before"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// Repository interface defines the contract for draft MAWB data operations
// Enhanced with complex CRUD operations for nested data handling
type Repository interface {
	// Core CRUD operations
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	GetByUUID(ctx context.Context, uuid string) (*DraftMAWB, error)
	CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) error
	UpdateStatus(ctx context.Context, uuid, status string) error
	Delete(ctx context.Context, uuid string) error

	// Validation operations
	ValidateMAWBExists(ctx context.Context, mawbUUID string) error
	ValidateUUIDExists(ctx context.Context, uuid string) error

	// Batch operations for performance
	GetMultipleByMAWBUUIDs(ctx context.Context, mawbUUIDs []string) ([]*DraftMAWB, error)
	BatchUpdateStatus(ctx context.Context, uuids []string, status string) error

	// Advanced query operations
	GetWithFilters(ctx context.Context, filters DraftMAWBFilters) ([]*DraftMAWB, error)
	GetItemsByDraftMAWBUUID(ctx context.Context, draftMAWBUUID string) ([]DraftMAWBItem, error)
	GetChargesByDraftMAWBUUID(ctx context.Context, draftMAWBUUID string) ([]DraftMAWBCharge, error)
}

type repository struct {
	contextTimeout time.Duration
	queryMonitor   *common.QueryMonitor
}

// NewRepository creates a new draft MAWB repository instance
func NewRepository(timeout time.Duration) Repository {
	// Configure query monitoring: 200ms threshold for complex queries, log slow queries
	queryMonitor := common.NewQueryMonitor(200*time.Millisecond, true, false)

	return &repository{
		contextTimeout: timeout,
		queryMonitor:   queryMonitor,
	}
}

// GetByMAWBUUID retrieves a draft MAWB by MAWB Info UUID with optimized JOIN queries
func (r *repository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for consistent read
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	var draftMAWB DraftMAWB

	// Get draft MAWB with validation
	sqlStr := `
		SELECT 
			dm.uuid,
			dm.mawb_info_uuid,
			dm.customer_uuid,
			dm.airline_logo,
			dm.airline_name,
			dm.mawb,
			dm.hawb,
			dm.shipper_name_and_address,
			dm.consignee_name_and_address,
			dm.issuing_carrier_agent_name_and_city,
			dm.accounting_information,
			dm.agent_iata_code,
			dm.account_no,
			dm.airport_of_departure,
			dm.reference_number,
			dm.to_1,
			dm.by_first_carrier,
			dm.to_2,
			dm.by_2,
			dm.to_3,
			dm.by_3,
			dm.currency,
			dm.chgs_code,
			dm.wt_val_ppd,
			dm.wt_val_coll,
			dm.other_ppd,
			dm.other_coll,
			dm.declared_value_carriage,
			dm.declared_value_customs,
			dm.airport_of_destination,
			dm.flight_no,
			dm.flight_date,
			dm.insurance_amount,
			dm.handling_information,
			dm.sci,
			dm.total_no_of_pieces,
			dm.total_gross_weight,
			dm.total_kg_lb,
			dm.total_rate_class,
			dm.total_chargeable_weight,
			dm.total_rate_charge,
			dm.total_amount,
			dm.shipper_certifies_text,
			dm.executed_on_date,
			dm.executed_at_place,
			dm.signature_of_shipper,
			dm.signature_of_issuing_carrier,
			dm.status,
			dm.created_at,
			dm.updated_at
		FROM draft_mawb dm
		INNER JOIN tbl_mawb_info mi ON dm.mawb_info_uuid = mi.uuid
		WHERE dm.mawb_info_uuid = ?
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	_, err = tx.QueryOneContext(ctx, pg.Scan(
		&draftMAWB.UUID,
		&draftMAWB.MAWBInfoUUID,
		&draftMAWB.CustomerUUID,
		&draftMAWB.AirlineLogo,
		&draftMAWB.AirlineName,
		&draftMAWB.MAWB,
		&draftMAWB.HAWB,
		&draftMAWB.ShipperNameAndAddress,
		&draftMAWB.ConsigneeNameAndAddress,
		&draftMAWB.IssuingCarrierAgentNameAndCity,
		&draftMAWB.AccountingInformation,
		&draftMAWB.AgentIATACode,
		&draftMAWB.AccountNo,
		&draftMAWB.AirportOfDeparture,
		&draftMAWB.ReferenceNumber,
		&draftMAWB.To1,
		&draftMAWB.ByFirstCarrier,
		&draftMAWB.To2,
		&draftMAWB.By2,
		&draftMAWB.To3,
		&draftMAWB.By3,
		&draftMAWB.Currency,
		&draftMAWB.ChgsCode,
		&draftMAWB.WtValPPD,
		&draftMAWB.WtValColl,
		&draftMAWB.OtherPPD,
		&draftMAWB.OtherColl,
		&draftMAWB.DeclaredValueCarriage,
		&draftMAWB.DeclaredValueCustoms,
		&draftMAWB.AirportOfDestination,
		&draftMAWB.FlightNo,
		&draftMAWB.FlightDate,
		&draftMAWB.InsuranceAmount,
		&draftMAWB.HandlingInformation,
		&draftMAWB.SCI,
		&draftMAWB.TotalNoOfPieces,
		&draftMAWB.TotalGrossWeight,
		&draftMAWB.TotalKgLb,
		&draftMAWB.TotalRateClass,
		&draftMAWB.TotalChargeableWeight,
		&draftMAWB.TotalRateCharge,
		&draftMAWB.TotalAmount,
		&draftMAWB.ShipperCertifiesText,
		&draftMAWB.ExecutedOnDate,
		&draftMAWB.ExecutedAtPlace,
		&draftMAWB.SignatureOfShipper,
		&draftMAWB.SignatureOfIssuingCarrier,
		&draftMAWB.Status,
		&draftMAWB.CreatedAt,
		&draftMAWB.UpdatedAt,
	), mawbUUID)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Load nested data efficiently
	if err := r.loadNestedData(ctx, tx, &draftMAWB); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return &draftMAWB, nil
}

// CreateOrUpdate creates a new draft MAWB or updates an existing one with enhanced transaction management
func (r *repository) CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction with proper rollback handling
	tx, err := db.Begin()
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	// Enhanced rollback handling with error capture
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Validate MAWB Info exists before proceeding
	var mawbInfoCount int
	validateSQL := `SELECT COUNT(*) FROM tbl_mawb_info WHERE uuid = ?`
	validateSQL = utils.ReplaceSQL(validateSQL, "?")
	_, err = tx.QueryOneContext(ctx, pg.Scan(&mawbInfoCount), validateSQL, draftMAWB.MAWBInfoUUID)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	if mawbInfoCount == 0 {
		return fmt.Errorf("MAWB Info not found: %s", draftMAWB.MAWBInfoUUID)
	}

	// Check if draft MAWB already exists
	var existingUUID string
	checkSQL := `SELECT uuid FROM draft_mawb WHERE mawb_info_uuid = ?`
	checkSQL = utils.ReplaceSQL(checkSQL, "?")
	_, err = tx.QueryOneContext(ctx, pg.Scan(&existingUUID), checkSQL, draftMAWB.MAWBInfoUUID)

	if err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	if err == pg.ErrNoRows {
		// Create new draft MAWB
		draftMAWB.UUID = uuid.New().String()
		draftMAWB.Status = StatusDraft
		draftMAWB.CreatedAt = time.Now()
		draftMAWB.UpdatedAt = time.Now()

		insertSQL := `
			INSERT INTO draft_mawb (
				uuid, mawb_info_uuid, customer_uuid, airline_logo, airline_name,
				mawb, hawb, shipper_name_and_address, consignee_name_and_address,
				issuing_carrier_agent_name_and_city, accounting_information, agent_iata_code,
				account_no, airport_of_departure, reference_number, to_1, by_first_carrier,
				to_2, by_2, to_3, by_3, currency, chgs_code, wt_val_ppd, wt_val_coll,
				other_ppd, other_coll, declared_value_carriage, declared_value_customs,
				airport_of_destination, flight_no, flight_date, insurance_amount,
				handling_information, sci, total_no_of_pieces, total_gross_weight,
				total_kg_lb, total_rate_class, total_chargeable_weight, total_rate_charge,
				total_amount, shipper_certifies_text, executed_on_date, executed_at_place,
				signature_of_shipper, signature_of_issuing_carrier, status, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		insertSQL = utils.ReplaceSQL(insertSQL, "?")
		_, err = tx.ExecContext(ctx, insertSQL,
			draftMAWB.UUID,
			draftMAWB.MAWBInfoUUID,
			draftMAWB.CustomerUUID,
			draftMAWB.AirlineLogo,
			draftMAWB.AirlineName,
			draftMAWB.MAWB,
			draftMAWB.HAWB,
			draftMAWB.ShipperNameAndAddress,
			draftMAWB.ConsigneeNameAndAddress,
			draftMAWB.IssuingCarrierAgentNameAndCity,
			draftMAWB.AccountingInformation,
			draftMAWB.AgentIATACode,
			draftMAWB.AccountNo,
			draftMAWB.AirportOfDeparture,
			draftMAWB.ReferenceNumber,
			draftMAWB.To1,
			draftMAWB.ByFirstCarrier,
			draftMAWB.To2,
			draftMAWB.By2,
			draftMAWB.To3,
			draftMAWB.By3,
			draftMAWB.Currency,
			draftMAWB.ChgsCode,
			draftMAWB.WtValPPD,
			draftMAWB.WtValColl,
			draftMAWB.OtherPPD,
			draftMAWB.OtherColl,
			draftMAWB.DeclaredValueCarriage,
			draftMAWB.DeclaredValueCustoms,
			draftMAWB.AirportOfDestination,
			draftMAWB.FlightNo,
			draftMAWB.FlightDate,
			draftMAWB.InsuranceAmount,
			draftMAWB.HandlingInformation,
			draftMAWB.SCI,
			draftMAWB.TotalNoOfPieces,
			draftMAWB.TotalGrossWeight,
			draftMAWB.TotalKgLb,
			draftMAWB.TotalRateClass,
			draftMAWB.TotalChargeableWeight,
			draftMAWB.TotalRateCharge,
			draftMAWB.TotalAmount,
			draftMAWB.ShipperCertifiesText,
			draftMAWB.ExecutedOnDate,
			draftMAWB.ExecutedAtPlace,
			draftMAWB.SignatureOfShipper,
			draftMAWB.SignatureOfIssuingCarrier,
			draftMAWB.Status,
			draftMAWB.CreatedAt,
			draftMAWB.UpdatedAt,
		)

		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	} else {
		// Update existing draft MAWB
		draftMAWB.UUID = existingUUID
		draftMAWB.UpdatedAt = time.Now()

		updateSQL := `
			UPDATE draft_mawb SET 
				customer_uuid = ?, airline_logo = ?, airline_name = ?, mawb = ?, hawb = ?,
				shipper_name_and_address = ?, consignee_name_and_address = ?,
				issuing_carrier_agent_name_and_city = ?, accounting_information = ?,
				agent_iata_code = ?, account_no = ?, airport_of_departure = ?,
				reference_number = ?, to_1 = ?, by_first_carrier = ?, to_2 = ?, by_2 = ?,
				to_3 = ?, by_3 = ?, currency = ?, chgs_code = ?, wt_val_ppd = ?,
				wt_val_coll = ?, other_ppd = ?, other_coll = ?, declared_value_carriage = ?,
				declared_value_customs = ?, airport_of_destination = ?, flight_no = ?,
				flight_date = ?, insurance_amount = ?, handling_information = ?, sci = ?,
				total_no_of_pieces = ?, total_gross_weight = ?, total_kg_lb = ?,
				total_rate_class = ?, total_chargeable_weight = ?, total_rate_charge = ?,
				total_amount = ?, shipper_certifies_text = ?, executed_on_date = ?,
				executed_at_place = ?, signature_of_shipper = ?, signature_of_issuing_carrier = ?,
				updated_at = ?
			WHERE uuid = ?
		`

		updateSQL = utils.ReplaceSQL(updateSQL, "?")
		_, err = tx.ExecContext(ctx, updateSQL,
			draftMAWB.CustomerUUID,
			draftMAWB.AirlineLogo,
			draftMAWB.AirlineName,
			draftMAWB.MAWB,
			draftMAWB.HAWB,
			draftMAWB.ShipperNameAndAddress,
			draftMAWB.ConsigneeNameAndAddress,
			draftMAWB.IssuingCarrierAgentNameAndCity,
			draftMAWB.AccountingInformation,
			draftMAWB.AgentIATACode,
			draftMAWB.AccountNo,
			draftMAWB.AirportOfDeparture,
			draftMAWB.ReferenceNumber,
			draftMAWB.To1,
			draftMAWB.ByFirstCarrier,
			draftMAWB.To2,
			draftMAWB.By2,
			draftMAWB.To3,
			draftMAWB.By3,
			draftMAWB.Currency,
			draftMAWB.ChgsCode,
			draftMAWB.WtValPPD,
			draftMAWB.WtValColl,
			draftMAWB.OtherPPD,
			draftMAWB.OtherColl,
			draftMAWB.DeclaredValueCarriage,
			draftMAWB.DeclaredValueCustoms,
			draftMAWB.AirportOfDestination,
			draftMAWB.FlightNo,
			draftMAWB.FlightDate,
			draftMAWB.InsuranceAmount,
			draftMAWB.HandlingInformation,
			draftMAWB.SCI,
			draftMAWB.TotalNoOfPieces,
			draftMAWB.TotalGrossWeight,
			draftMAWB.TotalKgLb,
			draftMAWB.TotalRateClass,
			draftMAWB.TotalChargeableWeight,
			draftMAWB.TotalRateCharge,
			draftMAWB.TotalAmount,
			draftMAWB.ShipperCertifiesText,
			draftMAWB.ExecutedOnDate,
			draftMAWB.ExecutedAtPlace,
			draftMAWB.SignatureOfShipper,
			draftMAWB.SignatureOfIssuingCarrier,
			draftMAWB.UpdatedAt,
			draftMAWB.UUID,
		)

		if err != nil {
			return utils.PostgresErrorTransform(err)
		}

		// Delete existing items and their dimensions (CASCADE will handle dimensions)
		deleteItemsSQL := `DELETE FROM draft_mawb_items WHERE draft_mawb_uuid = ?`
		deleteItemsSQL = utils.ReplaceSQL(deleteItemsSQL, "?")
		_, err = tx.ExecContext(ctx, deleteItemsSQL, draftMAWB.UUID)
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}

		// Delete existing charges
		deleteChargesSQL := `DELETE FROM draft_mawb_charges WHERE draft_mawb_uuid = ?`
		deleteChargesSQL = utils.ReplaceSQL(deleteChargesSQL, "?")
		_, err = tx.ExecContext(ctx, deleteChargesSQL, draftMAWB.UUID)
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}

	// Insert draft MAWB items with enhanced batch processing
	if len(draftMAWB.Items) > 0 {
		if err := r.insertItemsWithDimensions(ctx, tx, draftMAWB.UUID, draftMAWB.Items); err != nil {
			return err
		}
	}

	// Insert draft MAWB charges with enhanced batch processing
	if len(draftMAWB.Charges) > 0 {
		if err := r.insertCharges(ctx, tx, draftMAWB.UUID, draftMAWB.Charges); err != nil {
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return utils.PostgresErrorTransform(err)
	}

	return nil
}

// UpdateStatus updates the status of a draft MAWB
func (r *repository) UpdateStatus(ctx context.Context, uuid, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Validate status
	if !ValidateStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	sqlStr := `
		UPDATE draft_mawb 
		SET status = ?, updated_at = ? 
		WHERE uuid = ?
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, status, time.Now(), uuid)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.ErrRecordNotFound
	}

	return nil
}

// ValidateMAWBExists checks if the MAWB Info UUID exists in the database
func (r *repository) ValidateMAWBExists(ctx context.Context, mawbUUID string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	var count int
	sqlStr := `SELECT COUNT(*) FROM tbl_mawb_info WHERE uuid = ?`
	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	_, err := db.QueryOneContext(ctx, pg.Scan(&count), sqlStr, mawbUUID)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	if count == 0 {
		return fmt.Errorf("MAWB Info not found: %s", mawbUUID)
	}

	return nil
}

// loadNestedData efficiently loads items, dimensions, and charges for a draft MAWB
func (r *repository) loadNestedData(ctx context.Context, tx *pg.Tx, draftMAWB *DraftMAWB) error {
	// Load items first
	itemsSQL := `
		SELECT 
			i.id,
			i.draft_mawb_uuid,
			i.pieces_rcp,
			i.gross_weight,
			i.kg_lb,
			i.rate_class,
			i.total_volume,
			i.chargeable_weight,
			i.rate_charge,
			i.total,
			i.nature_and_quantity,
			i.created_at
		FROM draft_mawb_items i
		WHERE i.draft_mawb_uuid = ?
		ORDER BY i.id
	`

	itemsSQL = utils.ReplaceSQL(itemsSQL, "?")
	var items []DraftMAWBItem
	_, err := tx.QueryContext(ctx, &items, draftMAWB.UUID)
	if err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	// Load dimensions for each item
	for i := range items {
		dimsSQL := `
			SELECT 
				id,
				draft_mawb_item_id,
				length,
				width,
				height,
				count,
				created_at
			FROM draft_mawb_item_dims 
			WHERE draft_mawb_item_id = ?
			ORDER BY id
		`

		dimsSQL = utils.ReplaceSQL(dimsSQL, "?")
		var dims []DraftMAWBItemDim
		_, err := tx.QueryContext(ctx, &dims, items[i].ID)
		if err != nil && err != pg.ErrNoRows {
			return utils.PostgresErrorTransform(err)
		}

		items[i].Dims = dims
	}

	draftMAWB.Items = items

	// Load charges
	chargesSQL := `
		SELECT 
			id,
			draft_mawb_uuid,
			charge_key,
			charge_value,
			created_at
		FROM draft_mawb_charges 
		WHERE draft_mawb_uuid = ?
		ORDER BY id
	`

	chargesSQL = utils.ReplaceSQL(chargesSQL, "?")
	var charges []DraftMAWBCharge
	_, err = tx.QueryContext(ctx, &charges, draftMAWB.UUID)
	if err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}
	draftMAWB.Charges = charges

	return nil
}

// GetByUUID retrieves a draft MAWB by its UUID
func (r *repository) GetByUUID(ctx context.Context, uuid string) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for consistent read
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	var draftMAWB DraftMAWB

	// Get draft MAWB by UUID
	sqlStr := `
		SELECT 
			uuid,
			mawb_info_uuid,
			customer_uuid,
			airline_logo,
			airline_name,
			mawb,
			hawb,
			shipper_name_and_address,
			consignee_name_and_address,
			issuing_carrier_agent_name_and_city,
			accounting_information,
			agent_iata_code,
			account_no,
			airport_of_departure,
			reference_number,
			to_1,
			by_first_carrier,
			to_2,
			by_2,
			to_3,
			by_3,
			currency,
			chgs_code,
			wt_val_ppd,
			wt_val_coll,
			other_ppd,
			other_coll,
			declared_value_carriage,
			declared_value_customs,
			airport_of_destination,
			flight_no,
			flight_date,
			insurance_amount,
			handling_information,
			sci,
			total_no_of_pieces,
			total_gross_weight,
			total_kg_lb,
			total_rate_class,
			total_chargeable_weight,
			total_rate_charge,
			total_amount,
			shipper_certifies_text,
			executed_on_date,
			executed_at_place,
			signature_of_shipper,
			signature_of_issuing_carrier,
			status,
			created_at,
			updated_at
		FROM draft_mawb 
		WHERE uuid = ?
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	_, err = tx.QueryOneContext(ctx, pg.Scan(
		&draftMAWB.UUID,
		&draftMAWB.MAWBInfoUUID,
		&draftMAWB.CustomerUUID,
		&draftMAWB.AirlineLogo,
		&draftMAWB.AirlineName,
		&draftMAWB.MAWB,
		&draftMAWB.HAWB,
		&draftMAWB.ShipperNameAndAddress,
		&draftMAWB.ConsigneeNameAndAddress,
		&draftMAWB.IssuingCarrierAgentNameAndCity,
		&draftMAWB.AccountingInformation,
		&draftMAWB.AgentIATACode,
		&draftMAWB.AccountNo,
		&draftMAWB.AirportOfDeparture,
		&draftMAWB.ReferenceNumber,
		&draftMAWB.To1,
		&draftMAWB.ByFirstCarrier,
		&draftMAWB.To2,
		&draftMAWB.By2,
		&draftMAWB.To3,
		&draftMAWB.By3,
		&draftMAWB.Currency,
		&draftMAWB.ChgsCode,
		&draftMAWB.WtValPPD,
		&draftMAWB.WtValColl,
		&draftMAWB.OtherPPD,
		&draftMAWB.OtherColl,
		&draftMAWB.DeclaredValueCarriage,
		&draftMAWB.DeclaredValueCustoms,
		&draftMAWB.AirportOfDestination,
		&draftMAWB.FlightNo,
		&draftMAWB.FlightDate,
		&draftMAWB.InsuranceAmount,
		&draftMAWB.HandlingInformation,
		&draftMAWB.SCI,
		&draftMAWB.TotalNoOfPieces,
		&draftMAWB.TotalGrossWeight,
		&draftMAWB.TotalKgLb,
		&draftMAWB.TotalRateClass,
		&draftMAWB.TotalChargeableWeight,
		&draftMAWB.TotalRateCharge,
		&draftMAWB.TotalAmount,
		&draftMAWB.ShipperCertifiesText,
		&draftMAWB.ExecutedOnDate,
		&draftMAWB.ExecutedAtPlace,
		&draftMAWB.SignatureOfShipper,
		&draftMAWB.SignatureOfIssuingCarrier,
		&draftMAWB.Status,
		&draftMAWB.CreatedAt,
		&draftMAWB.UpdatedAt,
	), uuid)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Load nested data efficiently
	if err := r.loadNestedData(ctx, tx, &draftMAWB); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return &draftMAWB, nil
}

// Delete removes a draft MAWB and all related data (CASCADE DELETE)
func (r *repository) Delete(ctx context.Context, uuid string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for atomic delete
	tx, err := db.Begin()
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// Verify draft MAWB exists before deletion
	var count int
	checkSQL := `SELECT COUNT(*) FROM draft_mawb WHERE uuid = ?`
	checkSQL = utils.ReplaceSQL(checkSQL, "?")
	_, err = tx.QueryOneContext(ctx, pg.Scan(&count), checkSQL, uuid)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	if count == 0 {
		return utils.ErrRecordNotFound
	}

	// Delete draft MAWB (CASCADE will handle related tables)
	deleteSQL := `DELETE FROM draft_mawb WHERE uuid = ?`
	deleteSQL = utils.ReplaceSQL(deleteSQL, "?")
	result, err := tx.ExecContext(ctx, deleteSQL, uuid)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.ErrRecordNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return utils.PostgresErrorTransform(err)
	}

	return nil
}

// ValidateUUIDExists checks if a draft MAWB UUID exists
func (r *repository) ValidateUUIDExists(ctx context.Context, uuid string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	var count int
	sqlStr := `SELECT COUNT(*) FROM draft_mawb WHERE uuid = ?`
	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	_, err := db.QueryOneContext(ctx, pg.Scan(&count), sqlStr, uuid)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	if count == 0 {
		return fmt.Errorf("Draft MAWB not found: %s", uuid)
	}

	return nil
}

// GetMultipleByMAWBUUIDs retrieves multiple draft MAWBs by MAWB Info UUIDs
func (r *repository) GetMultipleByMAWBUUIDs(ctx context.Context, mawbUUIDs []string) ([]*DraftMAWB, error) {
	if len(mawbUUIDs) == 0 {
		return []*DraftMAWB{}, nil
	}

	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for consistent read
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// Use individual queries for simplicity and maintainability

	// For simplicity, let's use individual queries for each UUID
	// This is more maintainable and follows the existing patterns
	var draftMAWBs []*DraftMAWB
	for _, mawbUUID := range mawbUUIDs {
		draftMAWB, err := r.GetByMAWBUUID(ctx, mawbUUID)
		if err != nil && err != utils.ErrRecordNotFound {
			return nil, err
		}
		if draftMAWB != nil {
			draftMAWBs = append(draftMAWBs, draftMAWB)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return draftMAWBs, nil
}

// BatchUpdateStatus updates the status of multiple draft MAWBs
func (r *repository) BatchUpdateStatus(ctx context.Context, uuids []string, status string) error {
	if len(uuids) == 0 {
		return nil
	}

	// Validate status
	if !ValidateStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for atomic update
	tx, err := db.Begin()
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// Build IN clause with placeholders
	placeholders := make([]string, len(uuids))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	sqlStr := fmt.Sprintf(`
		UPDATE draft_mawb 
		SET status = ?, updated_at = ? 
		WHERE uuid IN (%s)
	`, inClause)

	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	// Prepare arguments: status, updated_at, then all UUIDs
	args := make([]interface{}, 2+len(uuids))
	args[0] = status
	args[1] = time.Now()
	for i, uuid := range uuids {
		args[2+i] = uuid
	}

	result, err := tx.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.ErrRecordNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return utils.PostgresErrorTransform(err)
	}

	return nil
}

// GetWithFilters retrieves draft MAWBs with filtering and pagination
func (r *repository) GetWithFilters(ctx context.Context, filters DraftMAWBFilters) ([]*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for consistent read
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// Build dynamic query using go-pg ORM for better maintainability

	// Use go-pg ORM pattern for simpler query handling
	var draftMAWBs []DraftMAWB
	query := tx.Model(&draftMAWBs)

	// Apply filters
	if len(filters.Status) > 0 {
		query = query.Where("status IN (?)", pg.In(filters.Status))
	}

	if filters.CustomerUUID != "" {
		query = query.Where("customer_uuid = ?", filters.CustomerUUID)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	// Apply ordering and pagination
	query = query.Order("created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	err = query.Select()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Convert to pointer slice and load nested data
	var result []*DraftMAWB
	for i := range draftMAWBs {
		// Load nested data for each draft MAWB
		if err := r.loadNestedData(ctx, tx, &draftMAWBs[i]); err != nil {
			return nil, err
		}
		result = append(result, &draftMAWBs[i])
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return result, nil
}

// GetItemsByDraftMAWBUUID retrieves items for a specific draft MAWB
func (r *repository) GetItemsByDraftMAWBUUID(ctx context.Context, draftMAWBUUID string) ([]DraftMAWBItem, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start transaction for consistent read
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// Create a temporary draft MAWB to use loadNestedData
	tempDraftMAWB := &DraftMAWB{UUID: draftMAWBUUID}
	if err := r.loadNestedData(ctx, tx, tempDraftMAWB); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return tempDraftMAWB.Items, nil
}

// GetChargesByDraftMAWBUUID retrieves charges for a specific draft MAWB
func (r *repository) GetChargesByDraftMAWBUUID(ctx context.Context, draftMAWBUUID string) ([]DraftMAWBCharge, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	chargesSQL := `
		SELECT 
			id,
			draft_mawb_uuid,
			charge_key,
			charge_value,
			created_at
		FROM draft_mawb_charges 
		WHERE draft_mawb_uuid = ?
		ORDER BY id
	`

	chargesSQL = utils.ReplaceSQL(chargesSQL, "?")
	var charges []DraftMAWBCharge
	_, err := db.QueryContext(ctx, &charges, draftMAWBUUID)
	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	return charges, nil
}

// insertItemsWithDimensions efficiently inserts items and their dimensions in batches
func (r *repository) insertItemsWithDimensions(ctx context.Context, tx *pg.Tx, draftMAWBUUID string, items []DraftMAWBItem) error {
	for _, item := range items {
		item.DraftMAWBUUID = draftMAWBUUID
		item.CreatedAt = time.Now()

		// Insert item and get ID
		itemSQL := `
			INSERT INTO draft_mawb_items (
				draft_mawb_uuid, pieces_rcp, gross_weight, kg_lb, rate_class,
				total_volume, chargeable_weight, rate_charge, total, nature_and_quantity, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			RETURNING id
		`

		itemSQL = utils.ReplaceSQL(itemSQL, "?")
		var itemID int
		_, err := tx.QueryOneContext(ctx, pg.Scan(&itemID), itemSQL,
			item.DraftMAWBUUID,
			item.PiecesRCP,
			item.GrossWeight,
			item.KgLb,
			item.RateClass,
			item.TotalVolume,
			item.ChargeableWeight,
			item.RateCharge,
			item.Total,
			item.NatureAndQuantity,
			item.CreatedAt,
		)

		if err != nil {
			return utils.PostgresErrorTransform(err)
		}

		// Insert dimensions for this item if any exist
		if len(item.Dims) > 0 {
			if err := r.insertDimensions(ctx, tx, itemID, item.Dims); err != nil {
				return err
			}
		}
	}
	return nil
}

// insertDimensions efficiently inserts dimensions for an item
func (r *repository) insertDimensions(ctx context.Context, tx *pg.Tx, itemID int, dims []DraftMAWBItemDim) error {
	// Batch insert dimensions if there are multiple
	if len(dims) == 1 {
		// Single dimension - use simple insert
		dim := dims[0]
		dim.DraftMAWBItemID = itemID
		dim.CreatedAt = time.Now()

		dimSQL := `
			INSERT INTO draft_mawb_item_dims (
				draft_mawb_item_id, length, width, height, count, created_at
			) VALUES (?, ?, ?, ?, ?, ?)
		`

		dimSQL = utils.ReplaceSQL(dimSQL, "?")
		_, err := tx.ExecContext(ctx, dimSQL,
			dim.DraftMAWBItemID,
			dim.Length,
			dim.Width,
			dim.Height,
			dim.Count,
			dim.CreatedAt,
		)

		return utils.PostgresErrorTransform(err)
	}

	// Multiple dimensions - use batch insert
	valueStrings := make([]string, 0, len(dims))
	valueArgs := make([]interface{}, 0, len(dims)*6)

	for _, dim := range dims {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			itemID,
			dim.Length,
			dim.Width,
			dim.Height,
			dim.Count,
			time.Now(),
		)
	}

	batchSQL := fmt.Sprintf(`
		INSERT INTO draft_mawb_item_dims (
			draft_mawb_item_id, length, width, height, count, created_at
		) VALUES %s
	`, strings.Join(valueStrings, ","))

	batchSQL = utils.ReplaceSQL(batchSQL, "?")
	_, err := tx.ExecContext(ctx, batchSQL, valueArgs...)

	return utils.PostgresErrorTransform(err)
}

// insertCharges efficiently inserts charges in batches
func (r *repository) insertCharges(ctx context.Context, tx *pg.Tx, draftMAWBUUID string, charges []DraftMAWBCharge) error {
	if len(charges) == 1 {
		// Single charge - use simple insert
		charge := charges[0]
		charge.DraftMAWBUUID = draftMAWBUUID
		charge.CreatedAt = time.Now()

		chargeSQL := `
			INSERT INTO draft_mawb_charges (
				draft_mawb_uuid, charge_key, charge_value, created_at
			) VALUES (?, ?, ?, ?)
		`

		chargeSQL = utils.ReplaceSQL(chargeSQL, "?")
		_, err := tx.ExecContext(ctx, chargeSQL,
			charge.DraftMAWBUUID,
			charge.Key,
			charge.Value,
			charge.CreatedAt,
		)

		return utils.PostgresErrorTransform(err)
	}

	// Multiple charges - use batch insert
	valueStrings := make([]string, 0, len(charges))
	valueArgs := make([]interface{}, 0, len(charges)*4)

	for _, charge := range charges {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			draftMAWBUUID,
			charge.Key,
			charge.Value,
			time.Now(),
		)
	}

	batchSQL := fmt.Sprintf(`
		INSERT INTO draft_mawb_charges (
			draft_mawb_uuid, charge_key, charge_value, created_at
		) VALUES %s
	`, strings.Join(valueStrings, ","))

	batchSQL = utils.ReplaceSQL(batchSQL, "?")
	_, err := tx.ExecContext(ctx, batchSQL, valueArgs...)

	return utils.PostgresErrorTransform(err)
}
