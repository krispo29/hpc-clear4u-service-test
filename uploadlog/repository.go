package uploadlog

import (
	"context"
	"errors"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	Get(ctx context.Context, uuid string) (*GetUploadloggingModel, error)
	GetAllUploadloggingsByCategoryAndSubCategory(ctx context.Context, startDate, endDate, category, subCategory string) ([]*GetUploadloggingModel, error)
	Insert(ctx context.Context, data *InsertModel) (string, error)
	Update(ctx context.Context, data *UpdateModel) error
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

func (r repository) Get(ctx context.Context, uuid string) (*GetUploadloggingModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := GetUploadloggingModel{}
	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.UUID,
		&x.Mawb,
		&x.FileName,
		&x.FileURL,
		&x.TemplateCode,
		&x.Category,
		&x.Status,
		&x.Amount,
		&x.Remark,
		&x.Creator,
		&x.CreatedAt,
		&x.UpdatedAt,
	), `
		SELECT  
			ul."uuid",
			ul.mawb,
			ul.file_name,
			ul.file_url,
			ul.template_code,
			ul.category,
			ul.status,
			ul.amount,
			ul.remark,
			u.username as creator,
			TO_CHAR(ul.created_at at time zone 'utc' at time zone 'Asia/bangkok', 'DD-MM-YYYY HH24:MI:SS') AS created_at,
			TO_CHAR(ul.updated_at at time zone 'utc' at time zone 'Asia/bangkok', 'DD-MM-YYYY HH24:MI:SS') AS updated_at
		FROM public.tbl_upload_loggings ul
		left join tbl_users u on u.uuid = ul.creator_uuid
		WHERE ul.uuid = ?
		ORDER by ul.id DESC
	`, uuid)

	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (r repository) GetAllUploadloggingsByCategoryAndSubCategory(ctx context.Context, startDate, endDate, category, subCategory string) ([]*GetUploadloggingModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

	list := []*GetUploadloggingModel{}
	sqlStr := `
			SELECT 
				 tul.uuid,
				 tul.mawb,
				 tul.file_name,
				 tul.file_url,
				 tul.template_code,
				 tul.category,
				 tul.status,
				 tul.amount,
				 tul.remark,
				TO_CHAR(tul.created_at at time zone 'utc' at time zone 'Asia/bangkok', 'DD-MM-YYYY HH24:MI:SS') AS created_at,
				TO_CHAR(tul.updated_at at time zone 'utc' at time zone 'Asia/bangkok', 'DD-MM-YYYY HH24:MI:SS') AS updated_at
			FROM public.tbl_upload_loggings tul
			WHERE (tul.created_at)::date BETWEEN SYMMETRIC $1 AND $2
		`

	values := []interface{}{}
	values = append(
		values,
		startDate,
		endDate,
	)
	if len(category) > 0 {
		sqlStr += ` AND tul.category = $3`
		values = append(
			values,
			category,
		)
	}
	if len(category) > 0 {
		sqlStr += ` AND tul.sub_category = $4`
		values = append(
			values,
			subCategory,
		)
	}

	sqlStr += " ORDER by tul.id DESC"

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.QueryContext(ctx, &list, values...)
	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) Insert(ctx context.Context, data *InsertModel) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	var uuid string
	_, err := db.QueryOneContext(ctx, &uuid,
		`
		INSERT INTO public.tbl_upload_loggings
			(
				mawb, file_name, file_url, template_code, category, sub_category, creator_uuid, status, amount
			)
		VALUES
			(
				?, ?, ?, ?, ?, ?, ?, ?, ?
			)
		RETURNING uuid
	`,
		utils.NewNullString(data.Mawb),
		utils.NewNullString(data.FileName),
		utils.NewNullString(data.FileUrl),
		utils.NewNullString(data.TemplateCode),
		utils.NewNullString(data.Category),
		utils.NewNullString(data.SubCategory),
		utils.NewNullString(data.CreatorUUID),
		utils.NewNullString(data.Status),
		data.Amount,
	)

	if err != nil {
		return uuid, err
	}

	return uuid, nil
}

func (r repository) Update(ctx context.Context, data *UpdateModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	result, err := db.ExecOneContext(ctx,
		`
			UPDATE public.tbl_upload_loggings
				SET  mawb=?1, status=?2, amount=?3, remark=?4, updated_at=NOW()
			WHERE "uuid" = ?0;
		`,
		data.UUID,
		data.Mawb,
		data.Status,
		data.Amount,
		data.Remark,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("not found")
	}

	return nil

}
