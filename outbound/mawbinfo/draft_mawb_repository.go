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
	err := db.Model(&draft).
		Where("mawb_info_uuid = ?", mawbInfoUUID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, utils.PostgresErrorTransform(err)
	}

	// Load related entities
	if err := r.loadRelated(db, &draft); err != nil {
		return nil, err
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
	existingDraft := &DraftMAWB{}
	err = tx.Model(existingDraft).Where("mawb_info_uuid = ?", draft.MAWBInfoUUID).Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, utils.PostgresErrorTransform(err)
	}

	if existingDraft.UUID != "" { // Update
		draft.UUID = existingDraft.UUID
		draft.CreatedAt = existingDraft.CreatedAt
		_, err = tx.Model(draft).WherePK().Update()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
		// Delete old children
		if err := r.deleteChildren(tx, draft.UUID); err != nil {
			return nil, err
		}
	} else { // Insert
		_, err = tx.Model(draft).Insert()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
	}

	// Insert new children
	for i := range draft.Items {
		draft.Items[i].DraftMAWBUUID = draft.UUID
		_, err := tx.Model(&draft.Items[i]).Insert()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
		}
		for j := range draft.Items[i].Dims {
			draft.Items[i].Dims[j].DraftMAWBItemID = draft.Items[i].ID
		}
		if len(draft.Items[i].Dims) > 0 {
			_, err = tx.Model(&draft.Items[i].Dims).Insert()
			if err != nil {
				return nil, utils.PostgresErrorTransform(err)
			}
		}
	}

	for i := range draft.Charges {
		draft.Charges[i].DraftMAWBUUID = draft.UUID
	}
	if len(draft.Charges) > 0 {
		_, err = tx.Model(&draft.Charges).Insert()
		if err != nil {
			return nil, utils.PostgresErrorTransform(err)
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

	res, err := db.Model(&DraftMAWB{}).
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

func (r *draftMAWBRepository) loadRelated(db orm.DB, draft *DraftMAWB) error {
	// Load items
	if err := db.Model(&draft.Items).Where("draft_mawb_uuid = ?", draft.UUID).Select(); err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	// For each item, load its dimensions
	for i := range draft.Items {
		if err := db.Model(&draft.Items[i].Dims).Where("draft_mawb_item_id = ?", draft.Items[i].ID).Select(); err != nil && err != pg.ErrNoRows {
			return utils.PostgresErrorTransform(err)
		}
	}

	// Load charges
	if err := db.Model(&draft.Charges).Where("draft_mawb_uuid = ?", draft.UUID).Select(); err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	return nil
}

func (r *draftMAWBRepository) deleteChildren(tx *pg.Tx, draftUUID string) error {
	// Need to get item IDs to delete dims first
	var items []DraftMAWBItem
	err := tx.Model(&items).Column("id").Where("draft_mawb_uuid = ?", draftUUID).Select()
	if err != nil && err != pg.ErrNoRows {
		return utils.PostgresErrorTransform(err)
	}

	if len(items) > 0 {
		var itemIDs []int
		for _, item := range items {
			itemIDs = append(itemIDs, item.ID)
		}
		_, err = tx.Model(&DraftMAWBItemDim{}).Where("draft_mawb_item_id IN (?)", pg.In(itemIDs)).Delete()
		if err != nil {
			return utils.PostgresErrorTransform(err)
		}
	}

	_, err = tx.Model(&DraftMAWBItem{}).Where("draft_mawb_uuid = ?", draftUUID).Delete()
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	_, err = tx.Model(&DraftMAWBCharge{}).Where("draft_mawb_uuid = ?", draftUUID).Delete()
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
