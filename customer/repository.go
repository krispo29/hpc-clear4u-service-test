package customer

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"

	"hpc-express-service/constant"
)

type Repository interface {
	GetAll(ctx context.Context) ([]*GetAllModel, error)
	GetAllDropdown(ctx context.Context) ([]*constant.DropdownModel, error)
}

type repository struct {
	contextTimeout time.Duration
}

func NewRepository(
	timeout time.Duration,
) Repository {
	return &repository{
		contextTimeout: timeout,
	}
}

func (r repository) GetAll(ctx context.Context) ([]*GetAllModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	sqlStr := `
		SELECT 
			"uuid",
			"name",
			to_char(created_at, 'DD-MM-YYYY') as created_date
		FROM public.tbl_customers
		WHERE deleted_at IS NULL
	`

	var list []*GetAllModel
	_, err := db.QueryContext(ctx, &list, sqlStr)

	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) GetAllDropdown(ctx context.Context) ([]*constant.DropdownModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	sqlStr := `
		SELECT "uuid" as value, "name" as text
		FROM public.tbl_customers
		WHERE deleted_at IS NULL
	`

	var list []*constant.DropdownModel
	_, err := db.QueryContext(ctx, &list, sqlStr)

	if err != nil {
		return list, err
	}

	return list, nil
}
