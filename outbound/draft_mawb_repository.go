package outbound

import (
	"context"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

// DraftMAWBRepository defines the interface for draft MAWB database operations.
type DraftMAWBRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	UpdateStatus(ctx context.Context, uuid, status string) error
	GetAll(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error)
}

type draftMAWBRepository struct{}

// NewDraftMAWBRepository creates a new draft MAWB repository.
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

// CreateOrUpdate creates a new draft MAWB or updates an existing one.
func (r *draftMAWBRepository) CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
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

	// Check if draft MAWB already exists
	existing, _ := r.GetByMAWBUUID(ctx, draftMAWB.MAWBInfoUUID)

	if existing != nil {
		// Update existing record
		draftMAWB.UUID = existing.UUID
		draftMAWB.CreatedAt = existing.CreatedAt // Keep original created_at
		draftMAWB.UpdatedAt = time.Now()
		_, err = db.Model(draftMAWB).WherePK().Update()
		if err != nil {
			return nil, err
		}
	} else {
		// Create new record
		draftMAWB.UUID = uuid.New().String()
		draftMAWB.CreatedAt = time.Now()
		draftMAWB.UpdatedAt = time.Now()
		_, err = db.Model(draftMAWB).Insert()
		if err != nil {
			return nil, err
		}
	}

	return r.GetByMAWBUUID(ctx, draftMAWB.MAWBInfoUUID)
}

// UpdateStatus updates the status of a draft MAWB.
func (r *draftMAWBRepository) UpdateStatus(ctx context.Context, uuid, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	_, err := db.Model(&DraftMAWB{}).
		Set("status = ?, updated_at = ?", status, time.Now()).
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
			COALESCE(dm.mawb, '') as mawb,
			COALESCE(dm.hawb, '') as hawb,
			COALESCE(dm.airline_name, '') as airline,
			COALESCE(dm.shipper_name_and_address, '') as shipper_name_and_address,
			COALESCE(dm.consignee_name_and_address, '') as consignee_name_and_address,
			COALESCE(c.name, '') as customer_name,
			TO_CHAR(dm.created_at AT TIME ZONE 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') as created_at
		FROM public.draft_mawb dm
		LEFT JOIN public.tbl_mawb_info mi ON dm.mawb_info_uuid::text = mi.uuid::text
		LEFT JOIN public.tbl_customers c ON dm.customer_uuid::text = c.uuid::text
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
