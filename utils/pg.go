package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/lib/pq"
)

var (
	// ErrRecordNotFound record not found error
	ErrRecordNotFound = errors.New("record not found")
	// ErrInvalidTransaction invalid transaction when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = errors.New("invalid transaction")
	// ErrNotImplemented not implemented
	ErrNotImplemented = errors.New("not implemented")
	// ErrMissingWhereClause missing where clause
	ErrMissingWhereClause = errors.New("WHERE conditions required")
	// ErrUnsupportedRelation unsupported relations
	ErrUnsupportedRelation = errors.New("unsupported relations")
	// ErrPrimaryKeyRequired primary keys required
	ErrPrimaryKeyRequired = errors.New("primary key required")
	// ErrModelValueRequired model value required
	ErrModelValueRequired = errors.New("model value required")
	// ErrModelAccessibleFieldsRequired model accessible fields required
	ErrModelAccessibleFieldsRequired = errors.New("model accessible fields required")
	// ErrSubQueryRequired sub query required
	ErrSubQueryRequired = errors.New("sub query required")
	// ErrInvalidData unsupported data
	ErrInvalidData = errors.New("unsupported data")
	// ErrUnsupportedDriver unsupported driver
	ErrUnsupportedDriver = errors.New("unsupported driver")
	// ErrRegistered registered
	ErrRegistered = errors.New("registered")
	// ErrInvalidField invalid field
	ErrInvalidField = errors.New("invalid field")
	// ErrEmptySlice empty slice found
	ErrEmptySlice = errors.New("empty slice found")
	// ErrDryRunModeUnsupported dry run mode unsupported
	ErrDryRunModeUnsupported = errors.New("dry run mode unsupported")
	// ErrInvalidDB invalid db
	ErrInvalidDB = errors.New("invalid db")
	// ErrInvalidValue invalid value
	ErrInvalidValue = errors.New("invalid value, should be pointer to struct or slice")
	// ErrInvalidValueOfLength invalid values do not match length
	ErrInvalidValueOfLength = errors.New("invalid association values, length doesn't match")
	// ErrPreloadNotAllowed preload is not allowed when count is used
	ErrPreloadNotAllowed = errors.New("preload is not allowed when count is used")
	// ErrDuplicatedKey occurs when there is a unique key constraint violation
	ErrDuplicatedKey = errors.New("duplicated key not allowed")
	// ErrForeignKeyViolated occurs when there is a foreign key constraint violation
	ErrForeignKeyViolated = errors.New("violates foreign key constraint")
	// ErrCheckConstraintViolated occurs when there is a check constraint violation
	ErrCheckConstraintViolated = errors.New("violates check constraint")
)

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func NewNullInt(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}

func PrepareSQL(sqlStr string) string {

	// remove last comma,
	sqlStr = sqlStr[0 : len(sqlStr)-1]

	// Convert symbol ? to $
	sqlStr = ReplaceSQL(sqlStr, "?")

	return sqlStr
}

func PostgresErrorTransform(err error) error {
	if err == nil {
		return nil
	}

	pgErr, ok := err.(*pq.Error)
	if ok {
		if pgErr.Code == "23505" {
			return ErrDuplicatedKey
		}
	}

	if err == pg.ErrNoRows {
		return ErrRecordNotFound
	}

	return err
}

func GetQuerier(ctx context.Context) (orm.DB, error) {
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
