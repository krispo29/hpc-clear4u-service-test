package outbound

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

func (r repository) GetCargoManifestByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	var manifest CargoManifest

	// 1. Get the main manifest record using raw SQL
	queryManifest := `
		SELECT uuid, mawb_info_uuid, mawb_number, port_of_discharge, flight_no, freight_date,
		       shipper, consignee, total_ctn, transshipment, status, created_at, updated_at
		FROM cargo_manifest WHERE mawb_info_uuid = ?
	`
	_, err := db.QueryOneContext(ctx, pg.Scan(
		&manifest.UUID, &manifest.MAWBInfoUUID, &manifest.MAWBNumber, &manifest.PortOfDischarge,
		&manifest.FlightNo, &manifest.FreightDate, &manifest.Shipper, &manifest.Consignee,
		&manifest.TotalCtn, &manifest.Transshipment, &manifest.Status, &manifest.CreatedAt, &manifest.UpdatedAt,
	), queryManifest, mawbInfoUUID)

	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, nil // Not found is not an error here, return nil
		}
		return nil, err
	}

	// 2. Get the items using raw SQL
	queryItems := `
		SELECT id, cargo_manifest_uuid, hawb_no, pkgs, gross_weight, destination,
		       commodity, shipper_name_address, consignee_name_address
		FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?
	`
	var items []CargoManifestItem
	_, err = db.QueryContext(ctx, &items, queryItems, manifest.UUID)
	if err != nil {
		return nil, err
	}
	manifest.Items = items

	return &manifest, nil
}

func (r repository) CreateOrUpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	existingManifest, err := r.GetCargoManifestByMAWBInfoUUID(ctx, manifest.MAWBInfoUUID)
	if err != nil && !errors.Is(err, pg.ErrNoRows) && err != nil {
        // if it's not a "not found" error, then it's a real error
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if existingManifest != nil {
		// UPDATE path
		manifest.UUID = existingManifest.UUID
		manifest.UpdatedAt = time.Now()

		updateQuery := `
			UPDATE cargo_manifest SET
				mawb_number = ?, port_of_discharge = ?, flight_no = ?, freight_date = ?,
				shipper = ?, consignee = ?, total_ctn = ?, transshipment = ?, status = ?, updated_at = ?
			WHERE uuid = ?
		`
		_, err = tx.ExecContext(ctx, updateQuery,
			manifest.MAWBNumber, manifest.PortOfDischarge, manifest.FlightNo, manifest.FreightDate,
			manifest.Shipper, manifest.Consignee, manifest.TotalCtn, manifest.Transshipment,
			manifest.Status, manifest.UpdatedAt, manifest.UUID,
		)
		if err != nil {
			return nil, err
		}

		// Delete old items
		deleteItemsQuery := `DELETE FROM cargo_manifest_items WHERE cargo_manifest_uuid = ?`
		_, err = tx.ExecContext(ctx, deleteItemsQuery, manifest.UUID)
		if err != nil {
			return nil, err
		}
	} else {
		// INSERT path
		manifest.UUID = uuid.New().String()
		manifest.CreatedAt = time.Now()
		manifest.UpdatedAt = time.Now()
		if manifest.Status == "" {
			manifest.Status = "Draft"
		}

		insertQuery := `
			INSERT INTO cargo_manifest (
				uuid, mawb_info_uuid, mawb_number, port_of_discharge, flight_no, freight_date,
				shipper, consignee, total_ctn, transshipment, status, created_by, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err = tx.ExecContext(ctx, insertQuery,
			manifest.UUID, manifest.MAWBInfoUUID, manifest.MAWBNumber, manifest.PortOfDischarge,
			manifest.FlightNo, manifest.FreightDate, manifest.Shipper, manifest.Consignee,
			manifest.TotalCtn, manifest.Transshipment, manifest.Status, "admin", manifest.CreatedAt, manifest.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	// Insert new items
	if len(manifest.Items) > 0 {
		// Build bulk insert query
		var values []interface{}
		var query strings.Builder
		query.WriteString(`
			INSERT INTO cargo_manifest_items (
				cargo_manifest_uuid, hawb_no, pkgs, gross_weight, destination,
				commodity, shipper_name_address, consignee_name_address
			) VALUES `)

		for i, item := range manifest.Items {
			query.WriteString("(?, ?, ?, ?, ?, ?, ?, ?)")
			if i < len(manifest.Items)-1 {
				query.WriteString(", ")
			}
			values = append(values, manifest.UUID, item.HAWBNo, item.Pkgs, item.GrossWeight,
				item.Destination, item.Commodity, item.ShipperNameAndAddress, item.ConsigneeNameAndAddress)
		}

		_, err = tx.ExecContext(ctx, query.String(), values...)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (r repository) UpdateCargoManifestStatus(ctx context.Context, mawbInfoUUID, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	query := `UPDATE cargo_manifest SET status = ?, updated_at = ? WHERE mawb_info_uuid = ?`
	res, err := db.ExecContext(ctx, query, status, time.Now(), mawbInfoUUID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Indicate that no record was updated
	}

	return nil
}
