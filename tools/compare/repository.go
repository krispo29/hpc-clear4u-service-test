package compare

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-pg/pg/v9"
)

type ExcelRepositoryInterface interface {
	GetValuesFromDB(ctx context.Context, columnName string) ([]*DBDetails, error)
}

type excelRepository struct {
	contextTimeout time.Duration
}

func NewExcelRepository(timeout time.Duration) ExcelRepositoryInterface {
	return &excelRepository{
		contextTimeout: timeout,
	}
}

func (r *excelRepository) GetValuesFromDB(ctx context.Context, columnName string) ([]*DBDetails, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	if db == nil {
		return nil, fmt.Errorf("database connection not found in context")
	}

	ctxQuery, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	if columnName == "" {
		return nil, fmt.Errorf("columnName cannot be empty")
	}

	// ตรวจสอบว่าคอลัมน์มีอยู่ในตาราง
	// Column allowance validation moved to service layer.
	var exists bool
	_, err := db.QueryOneContext(ctxQuery, &exists, `
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'master_hs_code_v2' 
            AND column_name = ?
        )`, columnName)
	if err != nil {
		log.Printf("Failed to check column existence: %v", err)
		return nil, fmt.Errorf("failed to check column existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("column '%s' does not exist in table master_hs_code_v2", columnName)
	}

	query := fmt.Sprintf(`
        SELECT goods_en, goods_th, tariff, stat, unit_code, duty_rate, 
               created_at, updated_at, deleted_at, remark, hs_code 
        FROM public.master_hs_code_v2 
        WHERE %s IS NOT NULL AND %s != '' AND hs_code IS NOT NULL AND hs_code != ''`,
		pg.Ident(columnName), pg.Ident(columnName))
	log.Printf("Executing query: %s", query)

	var dbValues []*DBDetails
	_, err = db.WithContext(ctxQuery).Query(&dbValues, query)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return nil, fmt.Errorf("failed to query database for column %s: %w", columnName, err)
	}

	log.Printf("Retrieved %d rows", len(dbValues))
	return dbValues, nil
}
