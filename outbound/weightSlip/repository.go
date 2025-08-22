package outbound

import (
	"context"
	"hpc-express-service/common"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/google/uuid"
)

type WeightSlipRepository interface {
	GetByMAWBUUID(ctx context.Context, mawbUUID string) (*WeightSlip, error)
	GetByUUID(ctx context.Context, uuid string) (*WeightSlip, error)
	Create(ctx context.Context, ws *WeightSlip) (*WeightSlip, error)
	Update(ctx context.Context, ws *WeightSlip) (*WeightSlip, error)
	GetAll(ctx context.Context, startDate, endDate string) ([]WeightSlipListItem, error)
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

// GetAll retrieves all weight slip records with optional date filtering and customer information.
func (r *weightSlipRepository) GetAll(ctx context.Context, startDate, endDate string) ([]WeightSlipListItem, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}

	var items []WeightSlipListItem

	baseQuery := `
            SELECT
                    ws.uuid::text,
                    ws.mawb_info_uuid::text,
                    COALESCE(ws.slip_no, '') as slip_no,
                    COALESCE(ws.mawb, '') as mawb,
                    COALESCE(ws.hawb, '') as hawb,
                    COALESCE(c.name, '') as customer_name,
                    TO_CHAR(ws.created_at AT TIME ZONE 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') as created_at,
                    COALESCE(ms.name, 'Draft') as status,
                    CASE WHEN ms.name = 'Cancelled' THEN true ELSE false END as is_deleted
            FROM public.weight_slip ws
            LEFT JOIN public.tbl_mawb_info mi ON ws.mawb_info_uuid::text = mi.uuid::text
            LEFT JOIN public.tbl_customers c ON mi.customer_uuid::text = c.uuid::text
            LEFT JOIN public.master_status ms ON ws.status_uuid = ms.uuid
    `

	var whereConditions []string
	var args []interface{}

	if startDate != "" {
		whereConditions = append(whereConditions, "DATE(ws.created_at AT TIME ZONE 'Asia/Bangkok') >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		whereConditions = append(whereConditions, "DATE(ws.created_at AT TIME ZONE 'Asia/Bangkok') <= ?")
		args = append(args, endDate)
	}

	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}
	query += " ORDER BY ws.created_at DESC"

	_, err = db.Query(&items, query, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}
