package user

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	Get(ctx context.Context, uuid string) (*GetModel, error)
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

func (r repository) Get(ctx context.Context, uuid string) (*GetModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := GetModel{}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.UUID,
		&x.Username,
	), `
			SELECT "uuid", username
			FROM public.tbl_users
			where uuid = ? AND deleted_at is null
	 `, uuid)

	if err != nil {
		return nil, err
	}

	return &x, nil
}
