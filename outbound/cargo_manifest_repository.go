package outbound

import (
	"context"
	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
	"time"
)

// CargoManifestRepository defines the interface for cargo manifest database operations.
type CargoManifestRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	CreateOrUpdate(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
}

type cargoManifestRepository struct{}

// NewCargoManifestRepository creates a new cargo manifest repository.
func NewCargoManifestRepository() CargoManifestRepository {
	return &cargoManifestRepository{}
}

// GetByMAWBUUID retrieves a cargo manifest and its items by MAWB Info UUID.
func (r *cargoManifestRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	manifest := &CargoManifest{}
	err := db.Model(manifest).Where("mawb_info_uuid = ?", mawbUUID).Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil // Not found is not an error, it's a valid state.
		}
		return nil, err
	}

	// Eager load items associated with the manifest.
	err = db.Model(&manifest.Items).Where("cargo_manifest_uuid = ?", manifest.UUID).Select()
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

// CreateOrUpdate creates a new cargo manifest or updates an existing one.
func (r *cargoManifestRepository) CreateOrUpdate(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Use a transactional context for the check to ensure atomicity.
	txCtx := context.WithValue(context.Background(), "postgreSQLConn", tx)
	existing, _ := r.GetByMAWBUUID(txCtx, manifest.MAWBInfoUUID)

	if existing != nil {
		// Update existing manifest
		manifest.UUID = existing.UUID
		manifest.UpdatedAt = time.Now()
		_, err = tx.Model(manifest).WherePK().Update()
		if err != nil {
			return nil, err
		}

		// Delete old items to be replaced.
		_, err = tx.Model(&CargoManifestItem{}).Where("cargo_manifest_uuid = ?", manifest.UUID).Delete()
		if err != nil {
			return nil, err
		}
	} else {
		// Insert new manifest
		manifest.UUID = uuid.New().String()
		manifest.CreatedAt = time.Now()
		manifest.UpdatedAt = time.Now()
		_, err := tx.Model(manifest).Insert()
		if err != nil {
			return nil, err
		}
	}

	// Insert new items
	for i := range manifest.Items {
		manifest.Items[i].CargoManifestUUID = manifest.UUID
	}
	if len(manifest.Items) > 0 {
		_, err := tx.Model(&manifest.Items).Insert()
		if err != nil {
			return nil, err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return the full, updated object.
	return r.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
}
