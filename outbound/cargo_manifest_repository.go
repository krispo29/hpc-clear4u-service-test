package outbound

import (
	"context"
	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
	"time"
)

type CargoManifestRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	CreateOrUpdate(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
}

type cargoManifestRepository struct{}

func NewCargoManifestRepository() CargoManifestRepository {
	return &cargoManifestRepository{}
}

func (r *cargoManifestRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	query := `SELECT uuid, mawb_info_uuid, mawb_number, port_of_discharge, flight_no, freight_date, shipper, consignee, total_ctn, transshipment, status, created_at, updated_at FROM cargo_manifest WHERE mawb_info_uuid = $1`

	manifest := &CargoManifest{}
	_, err := db.Query(manifest, query, mawbUUID)

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found is not an error here
		}
		return nil, err
	}

	itemsQuery := `SELECT id, cargo_manifest_uuid, hawb_no, pkgs, gross_weight, destination, commodity, shipper_name_address, consignee_name_address FROM cargo_manifest_items WHERE cargo_manifest_uuid = $1`
	_, err = db.Query(&manifest.Items, itemsQuery, manifest.UUID)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func (r *cargoManifestRepository) CreateOrUpdate(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := r.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
	if err != nil {
		// Note: GetByMAWBUUID uses the non-transactional DB connection from context.
		// For a transactional read, you'd need to pass `tx` or use a different approach.
		// For this upsert logic, a pre-check is okay.
	}

	if existing != nil {
		manifest.UUID = existing.UUID
		updateQuery := `UPDATE cargo_manifest SET mawb_number = ?, port_of_discharge = ?, flight_no = ?, freight_date = ?, shipper = ?, consignee = ?, total_ctn = ?, transshipment = ?, status = ?, updated_at = ? WHERE uuid = ?`
		_, err := tx.Exec(updateQuery, manifest.MAWBNumber, manifest.PortOfDischarge, manifest.FlightNo, manifest.FreightDate, manifest.Shipper, manifest.Consignee, manifest.TotalCtn, manifest.Transshipment, "Draft", time.Now(), manifest.UUID)
		if err != nil {
			return nil, err
		}

		deleteItemsQuery := `DELETE FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?`
		_, err = tx.Exec(deleteItemsQuery, manifest.UUID)
		if err != nil {
			return nil, err
		}
	} else {
		manifest.UUID = uuid.New().String()
		insertQuery := `INSERT INTO cargo_manifest (uuid, mawb_info_uuid, mawb_number, port_of_discharge, flight_no, freight_date, shipper, consignee, total_ctn, transshipment, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err := tx.Exec(insertQuery, manifest.UUID, manifest.MAWBInfoUUID, manifest.MAWBNumber, manifest.PortOfDischarge, manifest.FlightNo, manifest.FreightDate, manifest.Shipper, manifest.Consignee, manifest.TotalCtn, manifest.Transshipment, "Draft", time.Now(), time.Now())
		if err != nil {
			return nil, err
		}
	}

	if len(manifest.Items) > 0 {
		_, err := tx.Model(&manifest.Items).Insert()
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
}
