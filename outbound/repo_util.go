package outbound

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// qer สามารถเป็นได้ทั้ง *pg.DB และ *pg.Tx
type qer interface {
	Model(...interface{}) *orm.Query
	Query(model, query interface{}, params ...interface{}) (orm.Result, error)
	QueryOne(model, query interface{}, params ...interface{}) (orm.Result, error)
	Exec(query interface{}, params ...interface{}) (orm.Result, error)
	ExecOne(query interface{}, params ...interface{}) (orm.Result, error)
}

func getQer(ctx context.Context) (qer, error) {
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
