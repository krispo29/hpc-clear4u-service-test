package mawbinfo

import (
	"context"
	"database/sql"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type DraftMAWBRepository interface {
	GetByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*DraftMAWB, error)
	CreateOrUpdate(ctx context.Context, draft *DraftMAWB) (*DraftMAWB, error)
	UpdateStatus(ctx context.Context, mawbInfoUUID, status string) error
}

type draftMAWBRepository struct {
	contextTimeout time.Duration
}

func NewDraftMAWBRepository(timeout time.Duration) DraftMAWBRepository {
	return &draftMAWBRepository{
		contextTimeout: timeout,
	}
}

func (r *draftMAWBRepository) GetByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	if err := r.createTablesIfNotExists(ctx, db); err != nil {
		return nil, err
	}

	var draft DraftMAWB
	selectSQL := `SELECT * FROM draft_mawb WHERE mawb_info_uuid = ?`
	_, err := db.QueryOneContext(ctx, &draft, selectSQL, mawbInfoUUID)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, utils.PostgresErrorTransform(err)
	}

	// Load related entities
	// Load items
	selectItemsSQL := `SELECT * FROM draft_mawb_items WHERE draft_mawb_uuid = ?`
	if _, err := db.QueryContext(ctx, &draft.Items, selectItemsSQL, draft.UUID); err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	// For each item, load its dimensions
	selectDimsSQL := `SELECT * FROM draft_mawb_item_dims WHERE draft_mawb_item_id = ?`
	for i := range draft.Items {
		if _, err := db.QueryContext(ctx, &draft.Items[i].Dims, selectDimsSQL, draft.Items[i].ID); err != nil && err != pg.ErrNoRows {
			return nil, utils.PostgresErrorTransform(err)
		}
	}

	// Load charges
	selectChargesSQL := `SELECT * FROM draft_mawb_charges WHERE draft_mawb_uuid = ?`
	if _, err := db.QueryContext(ctx, &draft.Charges, selectChargesSQL, draft.UUID); err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	return &draft, nil
}

func (r *draftMAWBRepository) CreateOrUpdate(ctx context.Context, draft *DraftMAWB) (*DraftMAWB, error) {
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

	// Upsert logic for the main draft_mawb table
	var existingUUID string
	checkSQL := `SELECT uuid FROM draft_mawb WHERE mawb_info_uuid = ?`
	_, err = tx.QueryOneContext(ctx, pg.Scan(&existingUUID), checkSQL, draft.MAWBInfoUUID)
	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	if existingUUID != "" { // Update
		draft.UUID = existingUUID
		// Using ORM for update is pragmatic due to number of fields, but with explicit table name
		_, err = tx.Model(draft).Table("draft_mawb").WherePK().Update()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
		// Delete old children
		if err := r.deleteChildren(ctx, tx, draft.UUID); err != nil {
			return nil, err
		}
	} else { // Insert
		// Using ORM for insert is also pragmatic due to number of fields
		_, err = tx.Model(draft).Table("draft_mawb").Insert()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
	}

	// Insert new children
	for i := range draft.Items {
		item := &draft.Items[i]
		item.DraftMAWBUUID = draft.UUID
		insertItemSQL := `INSERT INTO draft_mawb_items (draft_mawb_uuid, pieces_rcp, gross_weight, kg_lb, rate_class, total_volume, chargeable_weight, rate_charge, total, nature_and_quantity) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`
		_, err := tx.QueryOneContext(ctx, pg.Scan(&item.ID), insertItemSQL, item.DraftMAWBUUID, item.PiecesRCP, item.GrossWeight, item.KgLb, item.RateClass, item.TotalVolume, item.ChargeableWeight, item.RateCharge, item.Total, item.NatureAndQuantity)
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}

		if len(item.Dims) > 0 {
			insertDimSQL := `INSERT INTO draft_mawb_item_dims (draft_mawb_item_id, length, width, height, count) VALUES (?, ?, ?, ?, ?)`
			stmt, err := tx.Prepare(insertDimSQL) // Corrected
			if err != nil {
				return nil, utils.PostgresErrorTransform(err)
			}
			defer stmt.Close()
			for _, dim := range item.Dims {
				_, err = stmt.ExecContext(ctx, item.ID, dim.Length, dim.Width, dim.Height, dim.Count)
				if err != nil {
					return nil, utils.PostgresErrorTransform(err)
				}
			}
		}
	}

	if len(draft.Charges) > 0 {
		insertChargeSQL := `INSERT INTO draft_mawb_charges (draft_mawb_uuid, charge_key, charge_value) VALUES (?, ?, ?)`
		stmt, err := tx.Prepare(insertChargeSQL) // Corrected
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
		defer stmt.Close()
		for _, charge := range draft.Charges {
			_, err = stmt.ExecContext(ctx, draft.UUID, charge.Key, charge.Value)
			if err != nil {
				return nil, utils.PostgresErrorTransform(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return r.GetByMAWBInfoUUID(ctx, draft.MAWBInfoUUID)
}

func (r *draftMAWBRepository) UpdateStatus(ctx context.Context, mawbInfoUUID, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	if err := r.createTablesIfNotExists(ctx, db); err != nil {
		return err
	}

	updateSQL := `UPDATE draft_mawb SET status = ?, updated_at = ? WHERE mawb_info_uuid = ?`
	res, err := db.ExecContext(ctx, updateSQL, status, time.Now(), mawbInfoUUID)

	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	rowsAffected := res.RowsAffected() // Corrected
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *draftMAWBRepository) deleteChildren(ctx context.Context, tx *pg.Tx, draftUUID string) error {
	var itemIDs []int
	selectItemIDsSQL := `SELECT id FROM draft_mawb_items WHERE draft_mawb_uuid = ?`
	_, err := tx.QueryContext(ctx, pg.Scan(&itemIDs), selectItemIDsSQL, draftUUID)
	if err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	if len(itemIDs) > 0 {
		deleteDimsSQL := `DELETE FROM draft_mawb_item_dims WHERE draft_mawb_item_id IN (?)`
		_, err = tx.ExecContext(ctx, deleteDimsSQL, pg.In(itemIDs))
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}

	deleteItemsSQL := `DELETE FROM draft_mawb_items WHERE draft_mawb_uuid = ?`
	_, err = tx.ExecContext(ctx, deleteItemsSQL, draftUUID)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	deleteChargesSQL := `DELETE FROM draft_mawb_charges WHERE draft_mawb_uuid = ?`
	_, err = tx.ExecContext(ctx, deleteChargesSQL, draftUUID)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	return nil
}

func (r *draftMAWBRepository) createTablesIfNotExists(ctx context.Context, db orm.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS draft_mawb (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mawb_info_uuid UUID NOT NULL,
			customer_uuid UUID,
			airline_logo VARCHAR(500),
			airline_name VARCHAR(255),
			mawb VARCHAR(255),
			hawb VARCHAR(255),
			shipper_name_and_address TEXT,
			awb_issued_by VARCHAR(255),
			consignee_name_and_address TEXT,
			issuing_carrier_agent_name VARCHAR(255),
			accounting_infomation TEXT,
			agents_iata_code VARCHAR(50),
			account_no VARCHAR(100),
			airport_of_departure VARCHAR(100),
			reference_number VARCHAR(100),
			optional_shipping_info1 VARCHAR(255),
			optional_shipping_info2 VARCHAR(255),
			routing_to VARCHAR(100),
			routing_by VARCHAR(100),
			destination_to1 VARCHAR(100),
			destination_by1 VARCHAR(100),
			destination_to2 VARCHAR(100),
			destination_by2 VARCHAR(100),
			currency VARCHAR(10),
			chgs_code VARCHAR(10),
			wt_val_ppd VARCHAR(10),
			wt_val_coll VARCHAR(10),
			other_ppd VARCHAR(10),
			other_coll VARCHAR(10),
			declared_val_carriage VARCHAR(100),
			declared_val_customs VARCHAR(100),
			airport_of_destination VARCHAR(100),
			requested_flight_date1 VARCHAR(100),
			requested_flight_date2 VARCHAR(100),
			amount_of_insurance VARCHAR(100),
			handling_infomation TEXT,
			sci VARCHAR(255),
			prepaid DECIMAL(10,2) DEFAULT 0,
			valuation_charge DECIMAL(10,2) DEFAULT 0,
			tax DECIMAL(10,2) DEFAULT 0,
			total_other_charges_due_agent DECIMAL(10,2) DEFAULT 0,
			total_other_charges_due_carrier DECIMAL(10,2) DEFAULT 0,
			total_prepaid DECIMAL(10,2) DEFAULT 0,
			currency_conversion_rates VARCHAR(255),
			signature1 VARCHAR(255),
			signature2_date DATE,
			signature2_place VARCHAR(255),
			signature2_issuing VARCHAR(255),
			shipping_mark TEXT,
			status VARCHAR(50) DEFAULT 'Draft',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_mawb_info
				FOREIGN KEY(mawb_info_uuid)
				REFERENCES tbl_mawb_info(uuid)
				ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_mawb_info_uuid ON draft_mawb(mawb_info_uuid)`,
		`CREATE TABLE IF NOT EXISTS draft_mawb_items (
			id SERIAL PRIMARY KEY,
			draft_mawb_uuid UUID NOT NULL,
			pieces_rcp VARCHAR(100),
			gross_weight VARCHAR(100),
			kg_lb VARCHAR(10) DEFAULT 'kg',
			rate_class VARCHAR(255),
			total_volume VARCHAR(100),
			chargeable_weight VARCHAR(100),
			rate_charge DECIMAL(10,2) DEFAULT 0,
			total DECIMAL(10,2) DEFAULT 0,
			nature_and_quantity TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb
				FOREIGN KEY(draft_mawb_uuid)
				REFERENCES draft_mawb(uuid)
				ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_items_uuid ON draft_mawb_items(draft_mawb_uuid)`,
		`CREATE TABLE IF NOT EXISTS draft_mawb_item_dims (
			id SERIAL PRIMARY KEY,
			draft_mawb_item_id INT NOT NULL,
			length VARCHAR(50),
			width VARCHAR(50),
			height VARCHAR(50),
			count VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb_item
				FOREIGN KEY(draft_mawb_item_id)
				REFERENCES draft_mawb_items(id)
				ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_item_dims_item_id ON draft_mawb_item_dims(draft_mawb_item_id)`,
		`CREATE TABLE IF NOT EXISTS draft_mawb_charges (
			id SERIAL PRIMARY KEY,
			draft_mawb_uuid UUID NOT NULL,
			charge_key VARCHAR(255),
			charge_value DECIMAL(10,2) DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_draft_mawb
				FOREIGN KEY(draft_mawb_uuid)
				REFERENCES draft_mawb(uuid)
				ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_draft_mawb_charges_uuid ON draft_mawb_charges(draft_mawb_uuid)`,
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}
	return nil
}
