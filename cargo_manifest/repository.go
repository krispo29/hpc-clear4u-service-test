package cargo_manifest

import (
	"context"
	"fmt"
	"hpc-express-service/common"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

// Repository interface defines the contract for cargo manifest data operations
type Repository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	CreateOrUpdate(ctx context.Context, manifest *CargoManifest) error
	UpdateStatus(ctx context.Context, uuid, status string) error
	ValidateMAWBExists(ctx context.Context, mawbUUID string) error
}

type repository struct {
	contextTimeout time.Duration
	queryMonitor   *common.QueryMonitor
}

// NewRepository creates a new cargo manifest repository instance
func NewRepository(timeout time.Duration) Repository {
	// Configure query monitoring: 100ms threshold, log slow queries, don't log all queries
	queryMonitor := common.NewQueryMonitor(100*time.Millisecond, true, false)

	return &repository{
		contextTimeout: timeout,
		queryMonitor:   queryMonitor,
	}
}

// GetByMAWBUUID retrieves a cargo manifest by MAWB Info UUID
func (r *repository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	var manifest CargoManifest

	// Get cargo manifest with optimized query using covering index
	sqlStr := `
		SELECT 
			uuid,
			mawb_info_uuid,
			mawb_number,
			port_of_discharge,
			flight_no,
			freight_date,
			shipper,
			consignee,
			total_ctn,
			transshipment,
			status,
			created_at,
			updated_at
		FROM cargo_manifest 
		WHERE mawb_info_uuid = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	err = r.queryMonitor.MonitorPreparedQuery(ctx, stmt, "GetCargoManifestByMAWBUUID", func(stmt *pg.Stmt) error {
		_, err := stmt.QueryOneContext(ctx, pg.Scan(
			&manifest.UUID,
			&manifest.MAWBInfoUUID,
			&manifest.MAWBNumber,
			&manifest.PortOfDischarge,
			&manifest.FlightNo,
			&manifest.FreightDate,
			&manifest.Shipper,
			&manifest.Consignee,
			&manifest.TotalCtn,
			&manifest.Transshipment,
			&manifest.Status,
			&manifest.CreatedAt,
			&manifest.UpdatedAt,
		), mawbUUID)
		return err
	})

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Get cargo manifest items with optimized query
	itemsSQL := `
		SELECT 
			id,
			cargo_manifest_uuid,
			hawb_no,
			pkgs,
			gross_weight,
			destination,
			commodity,
			shipper_name_address,
			consignee_name_address,
			created_at
		FROM cargo_manifest_items 
		WHERE cargo_manifest_uuid = ?
		ORDER BY id
	`

	itemsSQL = utils.ReplaceSQL(itemsSQL, "?")
	itemsStmt, err := db.Prepare(itemsSQL)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer itemsStmt.Close()

	var items []CargoManifestItem
	err = r.queryMonitor.MonitorPreparedQuery(ctx, itemsStmt, "GetCargoManifestItems", func(stmt *pg.Stmt) error {
		_, err := stmt.QueryContext(ctx, &items, manifest.UUID)
		return err
	})

	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	manifest.Items = items
	return &manifest, nil
}

// CreateOrUpdate creates a new cargo manifest or updates an existing one
func (r *repository) CreateOrUpdate(ctx context.Context, manifest *CargoManifest) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Start monitored transaction
	return r.queryMonitor.MonitorTransaction(ctx, db, "CreateOrUpdateCargoManifest", func(tx *pg.Tx) error {
		return r.executeCreateOrUpdate(ctx, tx, manifest)
	})
}

// executeCreateOrUpdate performs the actual create or update operation within a transaction
func (r *repository) executeCreateOrUpdate(ctx context.Context, tx *pg.Tx, manifest *CargoManifest) error {

	// Check if cargo manifest already exists using optimized query
	var existingUUID string
	checkSQL := `SELECT uuid FROM cargo_manifest WHERE mawb_info_uuid = ? LIMIT 1`
	checkSQL = utils.ReplaceSQL(checkSQL, "?")
	_, err := tx.QueryOneContext(ctx, pg.Scan(&existingUUID), checkSQL, manifest.MAWBInfoUUID)

	if err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	if err == pg.ErrNoRows {
		// Create new cargo manifest
		manifest.UUID = uuid.New().String()
		manifest.Status = StatusDraft
		manifest.CreatedAt = time.Now()
		manifest.UpdatedAt = time.Now()

		insertSQL := `
			INSERT INTO cargo_manifest (
				uuid, mawb_info_uuid, mawb_number, port_of_discharge, 
				flight_no, freight_date, shipper, consignee, 
				total_ctn, transshipment, status, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		insertSQL = utils.ReplaceSQL(insertSQL, "?")
		_, err = tx.ExecContext(ctx, insertSQL,
			manifest.UUID,
			manifest.MAWBInfoUUID,
			manifest.MAWBNumber,
			manifest.PortOfDischarge,
			manifest.FlightNo,
			manifest.FreightDate,
			manifest.Shipper,
			manifest.Consignee,
			manifest.TotalCtn,
			manifest.Transshipment,
			manifest.Status,
			manifest.CreatedAt,
			manifest.UpdatedAt,
		)

		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	} else {
		// Update existing cargo manifest
		manifest.UUID = existingUUID
		manifest.UpdatedAt = time.Now()

		updateSQL := `
			UPDATE cargo_manifest SET 
				mawb_number = ?, port_of_discharge = ?, flight_no = ?, 
				freight_date = ?, shipper = ?, consignee = ?, 
				total_ctn = ?, transshipment = ?, updated_at = ?
			WHERE uuid = ?
		`

		updateSQL = utils.ReplaceSQL(updateSQL, "?")
		_, err = tx.ExecContext(ctx, updateSQL,
			manifest.MAWBNumber,
			manifest.PortOfDischarge,
			manifest.FlightNo,
			manifest.FreightDate,
			manifest.Shipper,
			manifest.Consignee,
			manifest.TotalCtn,
			manifest.Transshipment,
			manifest.UpdatedAt,
			manifest.UUID,
		)

		if err != nil {
			return utils.PostgresErrorTransform(err)
		}

		// Delete existing items using optimized query
		deleteItemsSQL := `DELETE FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?`
		deleteItemsSQL = utils.ReplaceSQL(deleteItemsSQL, "?")
		_, err = tx.ExecContext(ctx, deleteItemsSQL, manifest.UUID)
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}

	// Insert cargo manifest items using batch processing for better performance
	if len(manifest.Items) > 0 {
		return r.insertItemsBatch(ctx, tx, manifest.UUID, manifest.Items)
	}

	return nil
}

// insertItemsBatch performs batch insert of cargo manifest items for better performance
func (r *repository) insertItemsBatch(ctx context.Context, tx *pg.Tx, manifestUUID string, items []CargoManifestItem) error {
	const batchSize = 100 // Process items in batches of 100

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		if err := r.insertItemsBatchChunk(ctx, tx, manifestUUID, batch); err != nil {
			return err
		}
	}

	return nil
}

// insertItemsBatchChunk inserts a chunk of items
func (r *repository) insertItemsBatchChunk(ctx context.Context, tx *pg.Tx, manifestUUID string, items []CargoManifestItem) error {
	for _, item := range items {
		item.CargoManifestUUID = manifestUUID
		item.CreatedAt = time.Now()

		itemSQL := `
			INSERT INTO cargo_manifest_items (
				cargo_manifest_uuid, hawb_no, pkgs, gross_weight, 
				destination, commodity, shipper_name_address, 
				consignee_name_address, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		itemSQL = utils.ReplaceSQL(itemSQL, "?")
		_, err := tx.ExecContext(ctx, itemSQL,
			item.CargoManifestUUID,
			item.HAWBNo,
			item.Pkgs,
			item.GrossWeight,
			item.Destination,
			item.Commodity,
			item.ShipperNameAndAddress,
			item.ConsigneeNameAndAddress,
			item.CreatedAt,
		)

		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}

	return nil
}

// UpdateStatus updates the status of a cargo manifest
func (r *repository) UpdateStatus(ctx context.Context, uuid, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), r.contextTimeout)
	defer cancel()

	// Validate status
	if !ValidateStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	sqlStr := `
		UPDATE cargo_manifest 
		SET status = ?, updated_at = ? 
		WHERE uuid = ?
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	var result pg.Result
	err = r.queryMonitor.MonitorPreparedQuery(ctx, stmt, "UpdateCargoManifestStatus", func(stmt *pg.Stmt) error {
		var err error
		result, err = stmt.ExecContext(ctx, status, time.Now(), uuid)
		return err
	})

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
	sqlStr := `SELECT COUNT(*) FROM tbl_mawb_info WHERE uuid = ? LIMIT 1`
	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	err := r.queryMonitor.MonitorQuery(ctx, "ValidateMAWBExists", func() error {
		_, err := db.QueryOneContext(ctx, pg.Scan(&count), sqlStr, mawbUUID)
		return err
	})

	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	if count == 0 {
		return fmt.Errorf("MAWB Info not found: %s", mawbUUID)
	}

	return nil
}
