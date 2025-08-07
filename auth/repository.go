package auth

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	Authentication(ctx context.Context, username string) (*GetSignInModel, error)
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

func (r repository) Authentication(ctx context.Context, username string) (*GetSignInModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	result := &GetSignInModel{}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&result.UUID,
		&result.HashedPassword,
	), `
		SELECT
			uuid, password
		FROM public.tbl_users
		WHERE username = ? AND deleted_at IS NULL
	 `, username)

	if err == pg.ErrNoRows {
		return nil, ErrUsernameOrPasswordIncorrect
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}
