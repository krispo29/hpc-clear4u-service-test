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
	err := db.Model(&manifest).
		Table("cargo_manifest").
		Where("mawb_info_uuid = ?", mawbInfoUUID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found is not an error here
		}
		return nil, utils.PostgresErrorTransform(err)
	}

	// Now, load the items for the manifest
	err = db.Model(&manifest.Items).
		Table("cargo_manifest_items").
		Where("cargo_manifest_uuid = ?", manifest.UUID).
		Select()

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
	existingManifest := &CargoManifest{}
	err = tx.Model(existingManifest).Table("cargo_manifest").Where("mawb_info_uuid = ?", manifest.MAWBInfoUUID).Select()

	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	if existingManifest.UUID != "" { // Update existing
		manifest.UUID = existingManifest.UUID
		manifest.CreatedAt = existingManifest.CreatedAt
		_, err = tx.Model(manifest).Table("cargo_manifest").WherePK().Update()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}

		// Delete old items
		_, err = tx.Model(&CargoManifestItem{}).Table("cargo_manifest_items").Where("cargo_manifest_uuid = ?", manifest.UUID).Delete()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
	} else { // Insert new
		_, err = tx.Model(manifest).Table("cargo_manifest").Insert()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
	}

	// Insert new items
	for i := range manifest.Items {
		manifest.Items[i].CargoManifestUUID = manifest.UUID
	}

	if len(manifest.Items) > 0 {
		_, err = tx.Model(&manifest.Items).Table("cargo_manifest_items").Insert()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
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

	res, err := db.Model(&CargoManifest{}).
		Table("cargo_manifest").
		Set("status = ?, updated_at = ?", status, time.Now()).
		Where("mawb_info_uuid = ?", mawbInfoUUID).
		Update()

	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	if res.RowsAffected() == 0 {
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
