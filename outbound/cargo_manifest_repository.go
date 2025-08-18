package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/google/uuid"
)

type CargoManifestRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	Create(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	Update(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
}

type cargoManifestRepository struct{}

func NewCargoManifestRepository() CargoManifestRepository {
	return &cargoManifestRepository{}
}

// ใช้ orm.DB แทน *pg.DB เพื่อรองรับทั้ง *pg.DB และ *pg.Tx
func (r *cargoManifestRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	db := ctx.Value("postgreSQLConn").(orm.DB)

	manifest := &CargoManifest{}
	err := db.Model(manifest).
		Column("cargo_manifest.*", "ms.name as status").
		Join("JOIN master_status AS ms ON ms.uuid = cargo_manifest.status_uuid").
		Where("cargo_manifest.mawb_info_uuid = ?", mawbUUID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// eager load items
	if err := db.Model(&manifest.Items).
		Where("cargo_manifest_uuid = ?", manifest.UUID).
		Select(); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (r *cargoManifestRepository) Create(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	rootDB := ctx.Value("postgreSQLConn").(orm.DB)

	// ถ้ามี Tx อยู่แล้วให้ใช้ต่อ
	if tx, ok := rootDB.(*pg.Tx); ok {
		return r.createInTx(ctx, tx, manifest)
	}

	// ถ้าเป็น *pg.DB ให้เริ่ม Tx ใหม่
	if pgdb, ok := rootDB.(*pg.DB); ok {
		tx, err := pgdb.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Close() // จะ rollback อัตโนมัติถ้ายังไม่ commit

		txCtx := context.WithValue(ctx, "postgreSQLConn", tx)
		out, err := r.createInTx(txCtx, tx, manifest)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return out, nil
	}

	return nil, fmt.Errorf("invalid DB type in context")
}

func (r *cargoManifestRepository) Update(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	rootDB := ctx.Value("postgreSQLConn").(orm.DB)

	// ถ้ามี Tx อยู่แล้วให้ใช้ต่อ
	if tx, ok := rootDB.(*pg.Tx); ok {
		return r.updateInTx(ctx, tx, manifest)
	}

	// ถ้าเป็น *pg.DB ให้เริ่ม Tx ใหม่
	if pgdb, ok := rootDB.(*pg.DB); ok {
		tx, err := pgdb.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Close() // จะ rollback อัตโนมัติถ้ายังไม่ commit

		txCtx := context.WithValue(ctx, "postgreSQLConn", tx)
		out, err := r.updateInTx(txCtx, tx, manifest)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return out, nil
	}

	return nil, fmt.Errorf("invalid DB type in context")
}

// ฟังก์ชันสร้าง manifest ใหม่ภายใน Tx
func (r *cargoManifestRepository) createInTx(ctx context.Context, db orm.DB, manifest *CargoManifest) (*CargoManifest, error) {
	now := time.Now()

	// สร้าง manifest ใหม่
	manifest.UUID = uuid.New().String()
	manifest.CreatedAt = now
	manifest.UpdatedAt = now

	if _, err := db.Model(manifest).Insert(); err != nil {
		return nil, err
	}

	// ใส่ items ใหม่
	for i := range manifest.Items {
		manifest.Items[i].CargoManifestUUID = manifest.UUID
	}
	if len(manifest.Items) > 0 {
		if _, err := db.Model(&manifest.Items).Insert(); err != nil {
			return nil, err
		}
	}

	// return ตัวเต็มล่าสุด
	return r.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
}

// ฟังก์ชันอัปเดต manifest ภายใน Tx
func (r *cargoManifestRepository) updateInTx(ctx context.Context, db orm.DB, manifest *CargoManifest) (*CargoManifest, error) {
	now := time.Now()

	// อัปเดต manifest เดิม
	manifest.UpdatedAt = now

	if _, err := db.Model(manifest).WherePK().Update(); err != nil {
		return nil, err
	}

	// ลบ items เดิมเพื่อแทนที่
	if _, err := db.Model(&CargoManifestItem{}).
		Where("cargo_manifest_uuid = ?", manifest.UUID).
		Delete(); err != nil {
		return nil, err
	}

	// ใส่ items ใหม่
	for i := range manifest.Items {
		manifest.Items[i].CargoManifestUUID = manifest.UUID
	}
	if len(manifest.Items) > 0 {
		if _, err := db.Model(&manifest.Items).Insert(); err != nil {
			return nil, err
		}
	}

	// return ตัวเต็มล่าสุด
	return r.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
}
