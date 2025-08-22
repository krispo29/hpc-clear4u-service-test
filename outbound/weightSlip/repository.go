package outbound

import (
	"context"
	"hpc-express-service/common"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

type WeightSlipRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*WeightSlip, error)
	GetByUUID(ctx context.Context, uuid string) (*WeightSlip, error)
	Create(ctx context.Context, ws *WeightSlip) (*WeightSlip, error)
	Update(ctx context.Context, ws *WeightSlip) (*WeightSlip, error)
}

type weightSlipRepository struct{}

func NewWeightSlipRepository() WeightSlipRepository {
	return &weightSlipRepository{}
}

func (r *weightSlipRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*WeightSlip, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}

	ws := &WeightSlip{}
	err = db.Model(ws).
		Column("weight_slip.*").
		ColumnExpr("ms.name AS status").
		Join("LEFT JOIN master_status AS ms ON ms.uuid = weight_slip.status_uuid").
		Where("weight_slip.mawb_info_uuid = ?", mawbUUID).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// load dimensions
	if err := db.Model(&ws.Dimensions).
		Where("weightslip_uuid = ?", ws.UUID).
		Select(); err != nil {
		return nil, err
	}

	// map nested structs
	ws.AfterSelect(nil)

	return ws, nil
}

func (r *weightSlipRepository) GetByUUID(ctx context.Context, uuid string) (*WeightSlip, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}

	ws := &WeightSlip{}
	err = db.Model(ws).
		Column("weight_slip.*").
		ColumnExpr("ms.name AS status").
		Join("LEFT JOIN master_status AS ms ON ms.uuid = weight_slip.status_uuid").
		Where("weight_slip.uuid = ?", uuid).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := db.Model(&ws.Dimensions).
		Where("weightslip_uuid = ?", ws.UUID).
		Select(); err != nil {
		return nil, err
	}

	ws.AfterSelect(nil)

	return ws, nil
}

func (r *weightSlipRepository) Create(ctx context.Context, ws *WeightSlip) (*WeightSlip, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	ws.UUID = uuid.New().String()
	ws.CreatedAt = now
	ws.UpdatedAt = now

	if _, err := db.Model(ws).Insert(); err != nil {
		return nil, err
	}

	for i := range ws.Dimensions {
		ws.Dimensions[i].WeightSlipUUID = ws.UUID
	}
	if len(ws.Dimensions) > 0 {
		if _, err := db.Model(&ws.Dimensions).Insert(); err != nil {
			return nil, err
		}
	}

	return r.GetByMAWBUUID(ctx, ws.MAWBInfoUUID)
}

func (r *weightSlipRepository) Update(ctx context.Context, ws *WeightSlip) (*WeightSlip, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	ws.UpdatedAt = now

	if _, err := db.Model(ws).WherePK().Update(); err != nil {
		return nil, err
	}

	if _, err := db.Model(&WeightSlipDimension{}).
		Where("weightslip_uuid = ?", ws.UUID).
		Delete(); err != nil {
		return nil, err
	}

	for i := range ws.Dimensions {
		ws.Dimensions[i].WeightSlipUUID = ws.UUID
	}
	if len(ws.Dimensions) > 0 {
		if _, err := db.Model(&ws.Dimensions).Insert(); err != nil {
			return nil, err
		}
	}

	return r.GetByMAWBUUID(ctx, ws.MAWBInfoUUID)
}
