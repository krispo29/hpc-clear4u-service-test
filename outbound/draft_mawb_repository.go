package outbound

import (
	"context"
	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
	"time"
)

// DraftMAWBRepository defines the interface for draft MAWB database operations.
type DraftMAWBRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	UpdateStatus(ctx context.Context, uuid, status string) error
}

type draftMAWBRepository struct{}

// NewDraftMAWBRepository creates a new draft MAWB repository.
func NewDraftMAWBRepository() DraftMAWBRepository {
	return &draftMAWBRepository{}
}

// GetByMAWBUUID retrieves a draft MAWB and its related entities by MAWB Info UUID.
func (r *draftMAWBRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	draft := &DraftMAWB{}
	err := db.Model(draft).
		Relation("Items.Dims").
		Relation("Charges").
		Where("mawb_info_uuid = ?", mawbUUID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return draft, nil
}

// CreateOrUpdate creates a new draft MAWB or updates an existing one.
func (r *draftMAWBRepository) CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txCtx := context.WithValue(context.Background(), "postgreSQLConn", tx)
	existing, _ := r.GetByMAWBUUID(txCtx, draftMAWB.MAWBInfoUUID)

	if existing != nil {
		draftMAWB.UUID = existing.UUID
		draftMAWB.UpdatedAt = time.Now()
		_, err = tx.Model(draftMAWB).WherePK().Update()
		if err != nil {
			return nil, err
		}

		// Delete old children
		if len(existing.Items) > 0 {
			var itemIDs []int
			for _, item := range existing.Items {
				itemIDs = append(itemIDs, item.ID)
			}
			_, err = tx.Model((*DraftMAWBItemDim)(nil)).
				Where("draft_mawb_item_id IN (?)", pg.In(itemIDs)).
				Delete()
			if err != nil {
				return nil, err
			}
			_, err = tx.Model((*DraftMAWBItem)(nil)).Where("draft_mawb_uuid = ?", draftMAWB.UUID).Delete()
			if err != nil {
				return nil, err
			}
		}
		_, err = tx.Model((*DraftMAWBCharge)(nil)).Where("draft_mawb_uuid = ?", draftMAWB.UUID).Delete()
		if err != nil {
			return nil, err
		}
	} else {
		draftMAWB.UUID = uuid.New().String()
		draftMAWB.CreatedAt = time.Now()
		draftMAWB.UpdatedAt = time.Now()
		_, err = tx.Model(draftMAWB).Insert()
		if err != nil {
			return nil, err
		}
	}

	// Insert Items, Dims, Charges
	for i := range draftMAWB.Items {
		draftMAWB.Items[i].DraftMAWBUUID = draftMAWB.UUID
		// Insert item to get its ID
		_, err := tx.Model(&draftMAWB.Items[i]).Insert()
		if err != nil {
			return nil, err
		}
		// Set item ID for all its dims and insert them
		for j := range draftMAWB.Items[i].Dims {
			draftMAWB.Items[i].Dims[j].DraftMAWBItemID = draftMAWB.Items[i].ID
		}
		if len(draftMAWB.Items[i].Dims) > 0 {
			_, err = tx.Model(&draftMAWB.Items[i].Dims).Insert()
			if err != nil {
				return nil, err
			}
		}
	}

	if len(draftMAWB.Charges) > 0 {
		for i := range draftMAWB.Charges {
			draftMAWB.Charges[i].DraftMAWBUUID = draftMAWB.UUID
		}
		_, err := tx.Model(&draftMAWB.Charges).Insert()
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
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
