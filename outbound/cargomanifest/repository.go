package cargomanifest

import (
	"context"
	"hpc-express-service/common"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

type CargoManifestRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	GetByUUID(ctx context.Context, uuid string) (*CargoManifest, error)
	Create(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	Update(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	GetAll(ctx context.Context, startDate, endDate string) ([]CargoManifestListItem, error)
}

type cargoManifestRepository struct{}

func NewCargoManifestRepository() CargoManifestRepository {
	return &cargoManifestRepository{}
}

func (r *cargoManifestRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}

	manifest := &CargoManifest{}
	err = db.Model(manifest).
		Column("cargo_manifest.*").
		ColumnExpr("ms.name AS status").
		Join("LEFT JOIN master_status AS ms ON ms.uuid = cargo_manifest.status_uuid").
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

func (r *cargoManifestRepository) GetByUUID(ctx context.Context, uuid string) (*CargoManifest, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}

	manifest := &CargoManifest{}
	err = db.Model(manifest).
		Column("cargo_manifest.*").
		ColumnExpr("ms.name AS status").
		Join("LEFT JOIN master_status AS ms ON ms.uuid = cargo_manifest.status_uuid").
		Where("cargo_manifest.uuid = ?", uuid).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := db.Model(&manifest.Items).
		Where("cargo_manifest_uuid = ?", manifest.UUID).
		Select(); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (r *cargoManifestRepository) Create(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
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

func (r *cargoManifestRepository) Update(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
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

func (r *cargoManifestRepository) GetAll(ctx context.Context, startDate, endDate string) ([]CargoManifestListItem, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}

	var items []CargoManifestListItem

	baseQuery := `
               SELECT
                       cm.uuid::text,
                       cm.mawb_info_uuid::text,
                       COALESCE(cm.mawb_number, '') AS mawb_number,
                       COALESCE(cm.flight_no, '') AS flight_no,
                       COALESCE(cm.shipper, '') AS shipper,
                       COALESCE(cm.consignee, '') AS consignee,
                       TO_CHAR(cm.created_at AT TIME ZONE 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') AS created_at,
                       COALESCE(ms.name, 'Draft') AS status
               FROM public.cargo_manifest cm
               LEFT JOIN public.master_status ms ON cm.status_uuid = ms.uuid
       `

	var whereConditions []string
	var args []interface{}
	if startDate != "" {
		whereConditions = append(whereConditions, "DATE(cm.created_at AT TIME ZONE 'Asia/Bangkok') >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		whereConditions = append(whereConditions, "DATE(cm.created_at AT TIME ZONE 'Asia/Bangkok') <= ?")
		args = append(args, endDate)
	}
	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}
	query += " ORDER BY cm.created_at DESC"

	_, err = db.Query(&items, query, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}
