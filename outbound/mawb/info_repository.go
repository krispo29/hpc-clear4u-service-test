package outbound

import (
	"context"
	"errors"
	"hpc-express-service/utils"
	"log"
	"time"

	"github.com/go-pg/pg/v9"
)

func (r repository) GetAll(ctx context.Context, start, end string) ([]*GetMawbInfo, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	var list []*GetMawbInfo
	_, err := db.QueryContext(ctx, &list,
		`
			SELECT
				"uuid",
				mawb,
				"date",
				service_type_code,
				shipping_type_code,
				chargeable_weight,
				created_at,
				updated_at,
				deleted_at,
				CASE
						WHEN deleted_at is null THEN false
						ELSE true
				END as is_deleted
			FROM public.tbl_pre_export_mawb_informations
			WHERE created_at::date BETWEEN SYMMETRIC ?0 AND ?1
			ORDER BY id DESC
	`, start, end)

	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) Create(ctx context.Context, data *CreateMawbInfo) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	var uuid string
	_, err := db.QueryOneContext(ctx, &uuid,
		`
		INSERT INTO public.tbl_pre_export_mawb_informations
			(
				 mawb, "date", service_type_code, shipping_type_code, chargeable_weight
			)
		VALUES
			(
				 ?, ?, ?, ?, ?
			)
		RETURNING uuid
	`,
		utils.NewNullString(data.Mawb),
		utils.NewNullString(data.Date),
		utils.NewNullString(data.ServiceTypeCode),
		utils.NewNullString(data.ShippingTypeCode),
		data.ChargeableWeight,
	)

	if err != nil {
		return uuid, err
	}

	return uuid, nil
}

func (r repository) GetOne(ctx context.Context, uuid string) (*GetMawbInfo, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := GetMawbInfo{MawbInfoBaseModel: &MawbInfoBaseModel{}}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.UUID,
		&x.Mawb,
		&x.Date,
		&x.ServiceTypeCode,
		&x.ShippingTypeCode,
		&x.ChargeableWeight,
		&x.CreatedAt,
		&x.UpdatedAt,
		&x.DeletedAt,
		&x.IsDeleted,
	), `
			SELECT
				"uuid",
				mawb,
				"date",
				service_type_code,
				shipping_type_code,
				chargeable_weight,
				created_at,
				updated_at,
				deleted_at,
				CASE
						WHEN deleted_at is null THEN false
						ELSE true
				END as is_deleted
			FROM public.tbl_pre_export_mawb_informations
			where uuid = ?
	 `, uuid)

	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (r repository) Update(ctx context.Context, data *UpdateMawbInfoModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err := db.ExecOneContext(ctx,
		`
			UPDATE public.tbl_pre_export_mawb_informations
				SET
				mawb=?1,
				"date"=?2,
				service_type_code=?3,
				shipping_type_code=?4,
				chargeable_weight=?5,
				updated_at=NOW()
			WHERE "uuid" = ?0 AND deleted_at is null
		`,
		data.UUID,
		utils.NewNullString(data.Mawb),
		utils.NewNullString(data.Date),
		utils.NewNullString(data.ServiceTypeCode),
		utils.NewNullString(data.ShippingTypeCode),
		data.ChargeableWeight,
	)

	if err != nil {
		if err == pg.ErrNoRows {
			return errors.New("not found")
		}
		return err
	}

	return nil
}

func (r repository) Delete(ctx context.Context, uuid string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err := db.ExecOneContext(ctx,
		`
			UPDATE public.tbl_pre_export_mawb_informations
				SET  
					deleted_at = NOW()
			WHERE "uuid" = ?0;
		`,
		uuid,
	)

	if err != nil {
		if err == pg.ErrNoRows {
			return errors.New("not found")
		}
		return err
	}

	return nil
}

func (r repository) InsertAttchment(ctx context.Context, data *InsertAttchmentModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err := db.ExecOneContext(ctx,
		`
		INSERT INTO public.tbl_pre_export_mawb_information_attchments
			(
				mawb_uuid, file_name, file_url
			)
		VALUES
			(
				 ?, ?, ?
			)
		RETURNING uuid
	`,
		utils.NewNullString(data.MawbUUID),
		utils.NewNullString(data.FileName),
		utils.NewNullString(data.FileURL),
	)

	if err != nil {
		return err
	}

	return nil
}

func (r repository) GetAttchments(ctx context.Context, uuid string) ([]*GetAttchmentModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	log.Println("uuid: ", uuid)
	var list []*GetAttchmentModel
	_, err := db.QueryContext(ctx, &list,
		`
			SELECT
				file_name,
				file_url
			FROM  public.tbl_pre_export_mawb_information_attchments
			WHERE mawb_uuid = ?0
			ORDER BY id DESC
	`, uuid)

	if err != nil {
		return list, err
	}

	return list, nil
}
