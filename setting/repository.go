package setting

import (
	"context"
	"errors"
	"hpc-express-service/utils"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	// HS Code
	CreateHsCode(ctx context.Context, data *CreateHsCodeModel) (string, error)
	GetAllHsCode(ctx context.Context, orderby string) ([]*GetHsCodeModel, error)
	GetHsCodeByUUID(ctx context.Context, uuid string) (*GetHsCodeModel, error)
	UpdateHsCode(ctx context.Context, data *UpdateHsCodeModel) error
	UpdateStatusHsCode(ctx context.Context, uuid string) error
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

func (r repository) CreateHsCode(ctx context.Context, data *CreateHsCodeModel) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	var uuid string
	_, err := db.QueryOneContext(ctx, &uuid,
		`
		INSERT INTO public.master_hs_code_v2
			(
				goods_en, 
				goods_th, 
				hs_code, 
				tariff, 
				stat, 
				unit_code, 
				duty_rate, 
				remark, 
				air_service_charge, 
				sea_service_charge, 
				fob_price_control, 
				fob_price_control_origin_currency_code, 
				fob_price_control_origin_country_code, 
				weight_control, 
				weight_control_unit_code,
				cif_control, 
				cif_control_destination_currency_code, 
				cif_control_destination_country_code
			)
		VALUES
			(
				 ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			)
		RETURNING uuid
	`,
		utils.NewNullString(strings.TrimSpace(data.GoodsEN)),
		utils.NewNullString(strings.TrimSpace(data.GoodsTH)),
		utils.NewNullString(strings.TrimSpace(data.HsCode)),
		utils.NewNullString(strings.TrimSpace(data.Tariff)),
		utils.NewNullString(strings.TrimSpace(data.Stat)),
		utils.NewNullString(strings.TrimSpace(data.UnitCode)),
		data.DutyRate,
		utils.NewNullString(strings.TrimSpace(data.Remark)),
		data.AirServiceCharge,
		data.SeaServiceCharge,
		data.FobPriceControl,
		utils.NewNullString(strings.TrimSpace(data.FobPriceControlOriginCurrencyCode)),
		utils.NewNullString(strings.TrimSpace(data.FobPriceControlOriginCountryCode)),
		data.WeightControl,
		utils.NewNullString(strings.TrimSpace(data.WeightControlUnitCode)),
		data.CifControl,
		utils.NewNullString(strings.TrimSpace(data.CifControlDestinationCurrencyCode)),
		utils.NewNullString(strings.TrimSpace(data.CifControlDestinationCountryCode)),
	)

	if err != nil {
		return uuid, err
	}

	return uuid, nil
}

func (r repository) GetAllHsCode(ctx context.Context, orderby string) ([]*GetHsCodeModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	strQuery := `
			SELECT
				"uuid",
				goods_en,
				goods_th,
				hs_code,
				tariff,
				stat,
				unit_code,
				duty_rate,
				remark,
				air_service_charge,
				sea_service_charge,
				fob_price_control,
				fob_price_control_origin_currency_code,
				weight_control,
				weight_control_unit_code,
				cif_control,
				cif_control_destination_currency_code,
				fob_price_control_origin_country_code,
				cif_control_destination_country_code,
				created_at,
				updated_at,
				deleted_at,
				CASE
						WHEN deleted_at is null THEN false
						ELSE true
				END as is_deleted
			FROM public.master_hs_code_v2
	`

	if len(orderby) == 0 {
		strQuery += " ORDER BY id DESC"
	} else {
		strQuery += " " + orderby
	}

	var list []*GetHsCodeModel
	_, err := db.QueryContext(ctx, &list,
		strQuery)

	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) GetHsCodeByUUID(ctx context.Context, uuid string) (*GetHsCodeModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := GetHsCodeModel{HsCodeBaseModel: &HsCodeBaseModel{}}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.UUID,
		&x.GoodsEN,
		&x.GoodsTH,
		&x.HsCode,
		&x.Tariff,
		&x.Stat,
		&x.UnitCode,
		&x.DutyRate,
		&x.Remark,
		&x.AirServiceCharge,
		&x.SeaServiceCharge,
		&x.FobPriceControl,
		&x.FobPriceControlOriginCurrencyCode,
		&x.FobPriceControlOriginCountryCode,
		&x.WeightControl,
		&x.WeightControlUnitCode,
		&x.CifControl,
		&x.CifControlDestinationCurrencyCode,
		&x.CifControlDestinationCountryCode,
		&x.CreatedAt,
		&x.UpdatedAt,
		&x.DeletedAt,
		&x.IsDeleted,
	), `
			SELECT
				"uuid",
				goods_en,
				goods_th,
				hs_code,
				tariff,
				stat,
				unit_code,
				duty_rate,
				remark,
				air_service_charge,
				sea_service_charge,
				fob_price_control,
				fob_price_control_origin_currency_code,
				fob_price_control_origin_country_code,
				weight_control,
				weight_control_unit_code,
				cif_control,
				cif_control_destination_currency_code,
				cif_control_destination_country_code,
				created_at,
				updated_at,
				deleted_at,
				CASE
						WHEN deleted_at is null THEN false
						ELSE true
				END as is_deleted
			FROM public.master_hs_code_v2
			where uuid = ?
	 `, uuid)

	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (r repository) UpdateHsCode(ctx context.Context, data *UpdateHsCodeModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	result, err := db.ExecOneContext(ctx,
		`
			UPDATE public.master_hs_code_v2
				SET  
				goods_en=?1,
				goods_th=?2,
				hs_code=?3,
				tariff=?4,
				stat=?5,
				unit_code=?6,
				duty_rate=?7,
				remark=?8,
				air_service_charge=?9,
				sea_service_charge=?10,
				fob_price_control=?11,
				fob_price_control_origin_currency_code=?12,
				fob_price_control_origin_country_code=?13,
				weight_control=?14,
				weight_control_unit_code=?15,
				cif_control=?16,
				cif_control_destination_currency_code=?17,
				cif_control_destination_country_code=?18,
				updated_at=NOW()
			WHERE "uuid" = ?0;
		`,
		data.UUID,
		utils.NewNullString(strings.TrimSpace(data.GoodsEN)),
		utils.NewNullString(strings.TrimSpace(data.GoodsTH)),
		utils.NewNullString(strings.TrimSpace(data.HsCode)),
		utils.NewNullString(strings.TrimSpace(data.Tariff)),
		utils.NewNullString(strings.TrimSpace(data.Stat)),
		utils.NewNullString(strings.TrimSpace(data.UnitCode)),
		data.DutyRate,
		utils.NewNullString(strings.TrimSpace(data.Remark)),
		data.AirServiceCharge,
		data.SeaServiceCharge,
		data.FobPriceControl,
		utils.NewNullString(strings.TrimSpace(data.FobPriceControlOriginCurrencyCode)),
		utils.NewNullString(strings.TrimSpace(data.FobPriceControlOriginCountryCode)),
		data.WeightControl,
		utils.NewNullString(strings.TrimSpace(data.WeightControlUnitCode)),
		data.CifControl,
		utils.NewNullString(strings.TrimSpace(data.CifControlDestinationCurrencyCode)),
		utils.NewNullString(strings.TrimSpace(data.CifControlDestinationCountryCode)),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}

	return nil
}

func (r repository) UpdateStatusHsCode(ctx context.Context, uuid string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	result, err := db.ExecOneContext(ctx,
		`
			UPDATE public.master_hs_code_v2
				SET  
					deleted_at =
					CASE 
							WHEN deleted_at IS NULL THEN NOW()
							ELSE NULL
					END
			WHERE "uuid" = ?0;
		`,
		uuid,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}

	return nil
}
