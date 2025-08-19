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
