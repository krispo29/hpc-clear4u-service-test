package outbound

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

type DraftMAWBRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	GetByUUID(ctx context.Context, uuid string) (*DraftMAWB, error)
	Create(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	Update(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	UpdateStatus(ctx context.Context, uuid, statusUUID string) error
	GetAll(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error)
	CreateWithRelations(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error)
	UpdateWithRelations(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error)
	GetWithRelations(ctx context.Context, uuid string) (*DraftMAWBWithRelations, error)
	GetWithRelationsByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWBWithRelations, error)
}

type draftMAWBRepository struct{}

func NewDraftMAWBRepository() DraftMAWBRepository {
	return &draftMAWBRepository{}
}

func (r *draftMAWBRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	q, err := getQer(ctx)
	if err != nil {
		return nil, err
	}

	draft := &DraftMAWB{}
	err = q.Model(draft).
		Column("draft_mawb.*").
		ColumnExpr("ms.name AS status").
		Join("LEFT JOIN master_status AS ms ON ms.uuid = draft_mawb.status_uuid").
		Where("mawb_info_uuid = ?", mawbUUID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return draft, nil
}

func (r *draftMAWBRepository) GetByUUID(ctx context.Context, uuid string) (*DraftMAWB, error) {
	q, err := getQer(ctx)
	if err != nil {
		return nil, err
	}

	draft := &DraftMAWB{}
	err = q.Model(draft).
		Column("draft_mawb.*").
		ColumnExpr("ms.name AS status").
		Join("LEFT JOIN master_status AS ms ON ms.uuid = draft_mawb.status_uuid").
		Where("uuid = ?", uuid).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return draft, nil
}

// Create creates a new draft MAWB
func (r *draftMAWBRepository) Create(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	// First check if MAWB Info exists
	var mawbInfoExists bool
	_, err := db.QueryOne(pg.Scan(&mawbInfoExists),
		"SELECT EXISTS(SELECT 1 FROM public.tbl_mawb_info WHERE uuid = ?)",
		draftMAWB.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}

	if !mawbInfoExists {
		// If MAWB Info doesn't exist, create a basic one
		_, err = db.Exec(`
			INSERT INTO public.tbl_mawb_info (uuid, chargeable_weight, date, mawb, service_type, shipping_type, created_at, updated_at) 
			VALUES (?, 0, CURRENT_DATE, 'AUTO-GENERATED', 'cargo', 'air', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (uuid) DO NOTHING`,
			draftMAWB.MAWBInfoUUID)
		if err != nil {
			return nil, err
		}
	}

	// Create new record
	draftMAWB.UUID = uuid.New().String()
	draftMAWB.CreatedAt = time.Now()
	draftMAWB.UpdatedAt = time.Now()
	_, err = db.Model(draftMAWB).Insert()
	if err != nil {
		return nil, err
	}

	return r.GetByUUID(ctx, draftMAWB.UUID)
}

// Update updates an existing draft MAWB
func (r *draftMAWBRepository) Update(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	// Get existing record to preserve created_at
	existing, err := r.GetByUUID(ctx, draftMAWB.UUID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, pg.ErrNoRows
	}

	// Preserve created_at and update updated_at
	draftMAWB.CreatedAt = existing.CreatedAt
	draftMAWB.UpdatedAt = time.Now()

	_, err = db.Model(draftMAWB).WherePK().Update()
	if err != nil {
		return nil, err
	}

	return r.GetByUUID(ctx, draftMAWB.UUID)
}

// UpdateStatus updates the status of a draft MAWB.
func (r *draftMAWBRepository) UpdateStatus(ctx context.Context, uuid, statusUUID string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	_, err := db.Model(&DraftMAWB{}).
		Set("status_uuid = ?, updated_at = ?", statusUUID, time.Now()).
		Where("uuid = ?", uuid).
		Update()
	return err
}

// GetAll retrieves all draft MAWB records with customer information and date filtering
func (r *draftMAWBRepository) GetAll(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	var items []DraftMAWBListItem

	baseQuery := `
		SELECT 
			dm.uuid::text,
			dm.mawb_info_uuid::text,
			COALESCE(dm.mawb, '') as mawb,
			COALESCE(dm.hawb, '') as hawb,
			COALESCE(dm.airline_name, '') as airline,
			COALESCE(dm.shipper_name_and_address, '') as shipper_name_and_address,
			COALESCE(dm.consignee_name_and_address, '') as consignee_name_and_address,
			COALESCE(c.name, '') as customer_name,
			TO_CHAR(dm.created_at AT TIME ZONE 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') as created_at,
			COALESCE(ms.name, 'Draft') as status,
			CASE WHEN ms.name = 'Cancelled' THEN true ELSE false END as is_deleted
		FROM public.draft_mawb dm
		LEFT JOIN public.tbl_mawb_info mi ON dm.mawb_info_uuid::text = mi.uuid::text
		LEFT JOIN public.tbl_customers c ON dm.customer_uuid::text = c.uuid::text
		LEFT JOIN public.master_status ms ON dm.status_uuid = ms.uuid
	`

	var whereConditions []string
	var args []interface{}

	// Add date filtering if provided
	if startDate != "" {
		whereConditions = append(whereConditions, "DATE(dm.created_at AT TIME ZONE 'Asia/Bangkok') >= ?")
		args = append(args, startDate)
	}

	if endDate != "" {
		whereConditions = append(whereConditions, "DATE(dm.created_at AT TIME ZONE 'Asia/Bangkok') <= ?")
		args = append(args, endDate)
	}

	// Build final query
	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}
	query += " ORDER BY dm.created_at DESC"

	_, err := db.Query(&items, query, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// CreateWithRelations creates a new draft MAWB with its items and charges
func (r *draftMAWBRepository) CreateWithRelations(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// First check if MAWB Info exists
	var mawbInfoExists bool
	_, err = tx.QueryOne(pg.Scan(&mawbInfoExists),
		"SELECT EXISTS(SELECT 1 FROM public.tbl_mawb_info WHERE uuid = ?)",
		draftMAWB.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}

	if !mawbInfoExists {
		// If MAWB Info doesn't exist, create a basic one
		_, err = tx.Exec(`
			INSERT INTO public.tbl_mawb_info (uuid, chargeable_weight, date, mawb, service_type, shipping_type, created_at, updated_at) 
			VALUES (?, 0, CURRENT_DATE, 'AUTO-GENERATED', 'cargo', 'air', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (uuid) DO NOTHING`,
			draftMAWB.MAWBInfoUUID)
		if err != nil {
			return nil, err
		}
	}

	// Create new record
	draftMAWB.UUID = uuid.New().String()
	draftMAWB.CreatedAt = time.Now()
	draftMAWB.UpdatedAt = time.Now()
	_, err = tx.Model(draftMAWB).Insert()
	if err != nil {
		return nil, err
	}

	// No need to delete existing items and charges for create operation

	// Insert new items
	for _, itemInput := range items {
		item := &DraftMAWBItem{
			DraftMAWBUUID:     draftMAWB.UUID,
			PiecesRCP:         itemInput.PiecesRCP,
			GrossWeight:       fmt.Sprintf("%.2f", itemInput.GrossWeight),
			KgLb:              itemInput.KgLb,
			RateClass:         itemInput.RateClass,
			TotalVolume:       itemInput.TotalVolume,
			ChargeableWeight:  itemInput.ChargeableWeight,
			RateCharge:        itemInput.RateCharge,
			Total:             itemInput.Total,
			NatureAndQuantity: itemInput.NatureAndQuantity,
		}
		_, err = tx.Model(item).Insert()
		if err != nil {
			return nil, err
		}

		// Insert dimensions for this item
		for _, dimInput := range itemInput.Dims {
			dim := &DraftMAWBItemDim{
				DraftMAWBItemID: item.ID,
				Length:          fmt.Sprintf("%d", dimInput.Length),
				Width:           fmt.Sprintf("%d", dimInput.Width),
				Height:          fmt.Sprintf("%d", dimInput.Height),
				Count:           fmt.Sprintf("%d", dimInput.Count),
			}
			_, err = tx.Model(dim).Insert()
			if err != nil {
				return nil, err
			}
		}
	}

	// Insert new charges
	for _, chargeInput := range charges {
		charge := &DraftMAWBCharge{
			DraftMAWBUUID: draftMAWB.UUID,
			Key:           chargeInput.Key,
			Value:         chargeInput.Value,
		}
		_, err = tx.Model(charge).Insert()
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.GetByUUID(ctx, draftMAWB.UUID)
}

// UpdateWithRelations updates an existing draft MAWB with its items and charges
func (r *draftMAWBRepository) UpdateWithRelations(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get existing record to preserve created_at
	existing, err := r.GetByUUID(ctx, draftMAWB.UUID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, pg.ErrNoRows
	}

	// Update existing record
	draftMAWB.CreatedAt = existing.CreatedAt // Keep original created_at
	draftMAWB.UpdatedAt = time.Now()
	_, err = tx.Model(draftMAWB).WherePK().Update()
	if err != nil {
		return nil, err
	}

	// Delete existing items and charges
	_, err = tx.Model((*DraftMAWBItem)(nil)).Where("draft_mawb_uuid = ?", draftMAWB.UUID).Delete()
	if err != nil {
		return nil, err
	}
	_, err = tx.Model((*DraftMAWBCharge)(nil)).Where("draft_mawb_uuid = ?", draftMAWB.UUID).Delete()
	if err != nil {
		return nil, err
	}

	// Insert new items
	for _, itemInput := range items {
		item := &DraftMAWBItem{
			DraftMAWBUUID:     draftMAWB.UUID,
			PiecesRCP:         itemInput.PiecesRCP,
			GrossWeight:       fmt.Sprintf("%.2f", itemInput.GrossWeight),
			KgLb:              itemInput.KgLb,
			RateClass:         itemInput.RateClass,
			TotalVolume:       itemInput.TotalVolume,
			ChargeableWeight:  itemInput.ChargeableWeight,
			RateCharge:        itemInput.RateCharge,
			Total:             itemInput.Total,
			NatureAndQuantity: itemInput.NatureAndQuantity,
		}
		_, err = tx.Model(item).Insert()
		if err != nil {
			return nil, err
		}

		// Insert dimensions for this item
		for _, dimInput := range itemInput.Dims {
			dim := &DraftMAWBItemDim{
				DraftMAWBItemID: item.ID,
				Length:          fmt.Sprintf("%d", dimInput.Length),
				Width:           fmt.Sprintf("%d", dimInput.Width),
				Height:          fmt.Sprintf("%d", dimInput.Height),
				Count:           fmt.Sprintf("%d", dimInput.Count),
			}
			_, err = tx.Model(dim).Insert()
			if err != nil {
				return nil, err
			}
		}
	}

	// Insert new charges
	for _, chargeInput := range charges {
		charge := &DraftMAWBCharge{
			DraftMAWBUUID: draftMAWB.UUID,
			Key:           chargeInput.Key,
			Value:         chargeInput.Value,
		}
		_, err = tx.Model(charge).Insert()
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.GetByUUID(ctx, draftMAWB.UUID)
}

// GetWithRelations retrieves a draft MAWB with its items and charges
func (r *draftMAWBRepository) GetWithRelations(ctx context.Context, uuid string) (*DraftMAWBWithRelations, error) {
	// Get the main draft MAWB
	draft, err := r.GetByUUID(ctx, uuid)
	if err != nil || draft == nil {
		return nil, err
	}

	db := ctx.Value("postgreSQLConn").(*pg.DB)

	// Get items
	var items []DraftMAWBItem
	err = db.Model(&items).Where("draft_mawb_uuid = ?", uuid).Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}

	// Get dimensions for each item
	for i := range items {
		var dims []DraftMAWBItemDim
		err = db.Model(&dims).Where("draft_mawb_item_id = ?", items[i].ID).Select()
		if err != nil && err != pg.ErrNoRows {
			return nil, err
		}
		items[i].Dims = dims
	}

	// Get charges
	var charges []DraftMAWBCharge
	err = db.Model(&charges).Where("draft_mawb_uuid = ?", uuid).Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}

	return &DraftMAWBWithRelations{
		DraftMAWB: draft,
		Items:     items,
		Charges:   charges,
	}, nil
}

// GetWithRelationsByMAWBUUID retrieves a draft MAWB with its items and charges by MAWB UUID
func (r *draftMAWBRepository) GetWithRelationsByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWBWithRelations, error) {
	// Get the main draft MAWB by MAWB UUID
	draft, err := r.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil || draft == nil {
		return nil, err
	}

	db := ctx.Value("postgreSQLConn").(*pg.DB)

	// Get items
	var items []DraftMAWBItem
	err = db.Model(&items).Where("draft_mawb_uuid = ?", draft.UUID).Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}

	// Get dimensions for each item
	for i := range items {
		var dims []DraftMAWBItemDim
		err = db.Model(&dims).Where("draft_mawb_item_id = ?", items[i].ID).Select()
		if err != nil && err != pg.ErrNoRows {
			return nil, err
		}
		items[i].Dims = dims
	}

	// Get charges
	var charges []DraftMAWBCharge
	err = db.Model(&charges).Where("draft_mawb_uuid = ?", draft.UUID).Select()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}

	return &DraftMAWBWithRelations{
		DraftMAWB: draft,
		Items:     items,
		Charges:   charges,
	}, nil
}
