package outbound

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

func getQer(ctx context.Context) (orm.DB, error) {
	v := ctx.Value("postgreSQLConn")
	switch db := v.(type) {
	case *pg.DB:
		return db, nil
	case *pg.Tx:
		return db, nil
	default:
		return nil, fmt.Errorf("db not found in context")
	}
}

func BeginTx(ctx context.Context) (*pg.Tx, context.Context, error) {
	db, ok := ctx.Value("postgreSQLConn").(*pg.DB)
	if !ok {
		return nil, nil, fmt.Errorf("could not get postgres DB from context for transaction")
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, err
	}
	txCtx := context.WithValue(ctx, "postgreSQLConn", tx)
	return tx, txCtx, nil
}
