package common

import (
	"context"
	"time"
)

type Repository interface {
	GetAllExchangeRates(ctx context.Context) ([]*GetExchangeRateModel, error)
	GetAllConvertTemplates(ctx context.Context, category string) ([]*GetAllConvertTemplateModel, error)
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

func (r repository) GetAllExchangeRates(ctx context.Context) ([]*GetExchangeRateModel, error) {
	db, err := GetQer(ctx)
	if err != nil {
		return nil, err
	}

	sqlStr := `
		SELECT 
			cxr."id",
			cxr.item_no,
			cxr.use_for_country_code,
			cxr.country_code,
			cxr.country_name,
			cxr.currency_code,
			cxr.currency_name,
			cxr.import_exchange_rate / cxr.ratio as import_exchange_rate,
			cxr.export_exchange_rate / cxr.ratio as export_exchange_rate,
			cxr.is_enabled,
			to_char(cxr.created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') AS created_at,
			to_char(cxr.updated_at at time zone 'utc' at time zone 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') AS updated_at
		FROM ship2cu.customs_exchange_rate cxr
		WHERE is_enabled = true
		ORDER BY cxr.item_no;
	`

	var list []*GetExchangeRateModel
	_, err = db.Query(&list, sqlStr)

	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) GetAllConvertTemplates(ctx context.Context, category string) ([]*GetAllConvertTemplateModel, error) {
	db, err := GetQer(ctx)
	if err != nil {
		return nil, err
	}
	sqlStr := `select code, "name", "type" from master_convert_templates where deleted_at is null`

	values := []interface{}{}
	if len(category) > 0 {
		sqlStr += ` AND "type" = ?`
		values = append(
			values,
			category,
		)
	}

	var list []*GetAllConvertTemplateModel
	_, err = db.Query(&list, sqlStr, values...)

	if err != nil {
		return list, err
	}

	return list, nil
}
