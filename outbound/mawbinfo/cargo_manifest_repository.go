package mawbinfo

import (
	"context"
	"database/sql"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type CargoManifestRepository interface {
	GetByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error)
	CreateOrUpdate(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	UpdateStatus(ctx context.Context, mawbInfoUUID, status string) error
}

type cargoManifestRepository struct {
	contextTimeout time.Duration
}

func NewCargoManifestRepository(timeout time.Duration) CargoManifestRepository {
	return &cargoManifestRepository{
		contextTimeout: timeout,
	}
}

func (r *cargoManifestRepository) GetByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	if err := r.createTablesIfNotExists(ctx, db); err != nil {
		return nil, err
	}

	var manifest CargoManifest
	selectSQL := `SELECT uuid, mawb_info_uuid, mawb_number, port_of_discharge, flight_no, freight_date, shipper, consignee, total_ctn, transshipment, status, created_at, updated_at FROM cargo_manifest WHERE mawb_info_uuid = ?`
	_, err := db.QueryOneContext(ctx, &manifest, selectSQL, mawbInfoUUID)

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found is not an error here
		}
		return nil, utils.PostgresErrorTransform(err)
	}

	// Now, load the items for the manifest
	selectItemsSQL := `SELECT id, cargo_manifest_uuid, hawb_no, pkgs, gross_weight, destination, commodity, shipper_name_address, consignee_name_address FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?`
	_, err = db.QueryContext(ctx, &manifest.Items, selectItemsSQL, manifest.UUID)

	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	return &manifest, nil
}

func (r *cargoManifestRepository) CreateOrUpdate(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	if err := r.createTablesIfNotExists(ctx, db); err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// Check if a manifest already exists for this mawb_info_uuid
	var existingUUID string
	checkSQL := `SELECT uuid FROM cargo_manifest WHERE mawb_info_uuid = ?`
	_, err = tx.QueryOneContext(ctx, pg.Scan(&existingUUID), checkSQL, manifest.MAWBInfoUUID)

	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	if existingUUID != "" { // Update existing
		manifest.UUID = existingUUID
		updateSQL := `UPDATE cargo_manifest SET
			mawb_number = ?, port_of_discharge = ?, flight_no = ?, freight_date = ?, shipper = ?, consignee = ?,
			total_ctn = ?, transshipment = ?, status = ?, updated_at = ?
			WHERE uuid = ?`
		_, err = tx.ExecContext(ctx, updateSQL,
			manifest.MAWBNumber, manifest.PortOfDischarge, manifest.FlightNo, manifest.FreightDate, manifest.Shipper, manifest.Consignee,
			manifest.TotalCtn, manifest.Transshipment, manifest.Status, time.Now(), manifest.UUID)
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}

		// Delete old items
		deleteItemsSQL := `DELETE FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?`
		_, err = tx.ExecContext(ctx, deleteItemsSQL, manifest.UUID)
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
	} else { // Insert new
		insertSQL := `INSERT INTO cargo_manifest (
			mawb_info_uuid, mawb_number, port_of_discharge, flight_no, freight_date, shipper, consignee,
			total_ctn, transshipment, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING uuid`
		_, err = tx.QueryOneContext(ctx, pg.Scan(&manifest.UUID), insertSQL,
			manifest.MAWBInfoUUID, manifest.MAWBNumber, manifest.PortOfDischarge, manifest.FlightNo, manifest.FreightDate, manifest.Shipper, manifest.Consignee,
			manifest.TotalCtn, manifest.Transshipment, manifest.Status)
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
	}

	// Insert new items
	if len(manifest.Items) > 0 {
		insertItemSQL := `INSERT INTO cargo_manifest_items (
			cargo_manifest_uuid, hawb_no, pkgs, gross_weight, destination, commodity, shipper_name_address, consignee_name_address
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		stmt, err := tx.Prepare(insertItemSQL) // Corrected: removed Context
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
		defer stmt.Close()

		for _, item := range manifest.Items {
			_, err = stmt.ExecContext(ctx,
				manifest.UUID, item.HAWBNo, item.Pkgs, item.GrossWeight, item.Destination, item.Commodity, item.ShipperNameAndAddress, item.ConsigneeNameAndAddress)
			if err != nil {
				return nil, utils.PostgresErrorTransform(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Refetch to get all default values and timestamps
	return r.GetByMAWBInfoUUID(ctx, manifest.MAWBInfoUUID)
}

func (r *cargoManifestRepository) UpdateStatus(ctx context.Context, mawbInfoUUID, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	if err := r.createTablesIfNotExists(ctx, db); err != nil {
		return err
	}

	updateSQL := `UPDATE cargo_manifest SET status = ?, updated_at = ? WHERE mawb_info_uuid = ?`
	res, err := db.ExecContext(ctx, updateSQL, status, time.Now(), mawbInfoUUID)

	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	rowsAffected := res.RowsAffected() // Corrected: single return value
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *cargoManifestRepository) createTablesIfNotExists(ctx context.Context, db orm.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS cargo_manifest (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mawb_info_uuid UUID NOT NULL,
			mawb_number VARCHAR(255) NOT NULL,
			port_of_discharge VARCHAR(255),
			flight_no VARCHAR(100),
			freight_date VARCHAR(50),
			shipper TEXT,
			consignee TEXT,
			total_ctn VARCHAR(100),
			transshipment TEXT,
			status VARCHAR(50) DEFAULT 'Draft',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_mawb_info
				FOREIGN KEY(mawb_info_uuid)
				REFERENCES tbl_mawb_info(uuid)
				ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cargo_manifest_mawb_info_uuid ON cargo_manifest(mawb_info_uuid)`,
		`CREATE TABLE IF NOT EXISTS cargo_manifest_items (
			id SERIAL PRIMARY KEY,
			cargo_manifest_uuid UUID NOT NULL,
			hawb_no VARCHAR(255),
			pkgs VARCHAR(100),
			gross_weight VARCHAR(100),
			destination VARCHAR(100),
			commodity TEXT,
			shipper_name_address TEXT,
			consignee_name_address TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_cargo_manifest
				FOREIGN KEY(cargo_manifest_uuid)
				REFERENCES cargo_manifest(uuid)
				ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cargo_manifest_items_uuid ON cargo_manifest_items(cargo_manifest_uuid)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}
	return nil
}
