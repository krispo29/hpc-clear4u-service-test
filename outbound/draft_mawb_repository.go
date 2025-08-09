package outbound

import (
	"context"
	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
	"time"
)

type DraftMAWBRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	UpdateStatus(ctx context.Context, uuid, status string) error
}

type draftMAWBRepository struct{}

func NewDraftMAWBRepository() DraftMAWBRepository {
	return &draftMAWBRepository{}
}

func (r *draftMAWBRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	draft := &DraftMAWB{}
	err := db.Model(draft).Where("mawb_info_uuid = ?", mawbUUID).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}

	// Get Items and their Dims
	for i := range draft.Items {
		err := db.Model(&draft.Items[i].Dims).Where("draft_mawb_item_id = ?", draft.Items[i].ID).Select()
		if err != nil {
			return nil, err
		}
	}

	// Get Charges
	err = db.Model(&draft.Charges).Where("draft_mawb_uuid = ?", draft.UUID).Select()
	if err != nil {
		return nil, err
	}

	return draft, nil
}

func (r *draftMAWBRepository) CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing := &DraftMAWB{}
	err = db.Model(existing).Where("mawb_info_uuid = ?", draftMAWB.MAWBInfoUUID).Select()

	if err == nil && existing.UUID != "" {
		draftMAWB.UUID = existing.UUID
		draftMAWB.UpdatedAt = time.Now()
		_, err = tx.Model(draftMAWB).WherePK().Update()
		if err != nil {
			return nil, err
		}

		// Delete old children
		_, err = tx.Exec(`DELETE FROM draft_mawb_item_dims WHERE draft_mawb_item_id IN (SELECT id FROM draft_mawb_items WHERE draft_mawb_uuid = ?)`, draftMAWB.UUID)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(`DELETE FROM draft_mawb_items WHERE draft_mawb_uuid = ?`, draftMAWB.UUID)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(`DELETE FROM draft_mawb_charges WHERE draft_mawb_uuid = ?`, draftMAWB.UUID)
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
		_, err := tx.Model(&draftMAWB.Items[i]).Insert()
		if err != nil {
			return nil, err
		}
		for j := range draftMAWB.Items[i].Dims {
			draftMAWB.Items[i].Dims[j].DraftMAWBItemID = draftMAWB.Items[i].ID
			_, err := tx.Model(&draftMAWB.Items[i].Dims[j]).Insert()
			if err != nil {
				return nil, err
			}
		}
	}

	for i := range draftMAWB.Charges {
		draftMAWB.Charges[i].DraftMAWBUUID = draftMAWB.UUID
		_, err := tx.Model(&draftMAWB.Charges[i]).Insert()
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByMAWBUUID(ctx, draftMAWB.MAWBInfoUUID)
}

func (r *draftMAWBRepository) UpdateStatus(ctx context.Context, uuid, status string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	_, err := db.Model(&DraftMAWB{}).
		Set("status = ?, updated_at = ?", status, time.Now()).
		Where("uuid = ?", uuid).
		Update()
	return err
}
