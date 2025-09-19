package seawaybill

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-pg/pg/v9"

	"hpc-express-service/utils"
)

// Repository defines database operations for sea waybill details.
type Repository interface {
	CreateSeaWaybillDetail(ctx context.Context, data *seaWaybillDetailData) (*SeaWaybillDetailResponse, error)
	ListSeaWaybillDetails(ctx context.Context) ([]*SeaWaybillDetailResponse, error)
	GetSeaWaybillDetail(ctx context.Context, uuid string) (*SeaWaybillDetailResponse, error)
	UpdateSeaWaybillDetail(ctx context.Context, data *seaWaybillDetailData) (*SeaWaybillDetailResponse, error)
	DeleteAttachment(ctx context.Context, uuid, fileName string) (string, error)
}

type repository struct {
	contextTimeout time.Duration
}

// NewRepository creates a new repository instance.
func NewRepository(timeout time.Duration) Repository {
	return &repository{contextTimeout: timeout}
}

func (r repository) CreateSeaWaybillDetail(ctx context.Context, data *seaWaybillDetailData) (*SeaWaybillDetailResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	attachmentsJSON := "[]"
	if len(data.Attachments) > 0 {
		b, err := json.Marshal(data.Attachments)
		if err != nil {
			return nil, err
		}
		attachmentsJSON = string(b)
	}

	sqlStr := `
        INSERT INTO public.tbl_sea_waybill_details
            (uuid, gross_weight, volume_weight, duty_tax, attachments)
        VALUES
            (?, ?, ?, ?, ?::jsonb)
        RETURNING
            uuid,
            gross_weight,
            volume_weight,
            duty_tax,
            COALESCE(attachments::text, '[]') as attachments,
            to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at,
            to_char(updated_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as updated_at
    `

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	var (
		result          SeaWaybillDetailResponse
		attachmentsText string
	)

	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&result.UUID,
		&result.GrossWeight,
		&result.VolumeWeight,
		&result.DutyTax,
		&attachmentsText,
		&result.CreatedAt,
		&result.UpdatedAt,
	),
		data.UUID,
		data.GrossWeight,
		data.VolumeWeight,
		data.DutyTax,
		attachmentsJSON,
	)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	if attachmentsText != "" && attachmentsText != "[]" {
		if err := json.Unmarshal([]byte(attachmentsText), &result.Attachments); err != nil {
			return nil, fmt.Errorf("failed to parse attachments: %w", err)
		}
	}

	return &result, nil
}

func (r repository) ListSeaWaybillDetails(ctx context.Context) ([]*SeaWaybillDetailResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlStr := `
        SELECT
            uuid,
            gross_weight,
            volume_weight,
            duty_tax,
            COALESCE(attachments::text, '[]') as attachments,
            to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS.US') as created_at,
            to_char(updated_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS.US') as updated_at
        FROM public.tbl_sea_waybill_details
        ORDER BY created_at DESC
    `

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	type seaWaybillDetailRow struct {
		UUID         string
		GrossWeight  float64
		VolumeWeight float64
		DutyTax      float64
		Attachments  string
		CreatedAt    string
		UpdatedAt    string
	}

	rows := []seaWaybillDetailRow{}
	_, err = stmt.QueryContext(ctx, &rows)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	results := make([]*SeaWaybillDetailResponse, 0, len(rows))
	for _, row := range rows {
		item := &SeaWaybillDetailResponse{
			UUID:         row.UUID,
			GrossWeight:  row.GrossWeight,
			VolumeWeight: row.VolumeWeight,
			DutyTax:      row.DutyTax,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}

		if row.Attachments != "" && row.Attachments != "[]" {
			if err := json.Unmarshal([]byte(row.Attachments), &item.Attachments); err != nil {
				return nil, fmt.Errorf("failed to parse attachments: %w", err)
			}
		}

		results = append(results, item)
	}

	return results, nil
}

func (r repository) GetSeaWaybillDetail(ctx context.Context, uuid string) (*SeaWaybillDetailResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlStr := `
        SELECT
            uuid,
            gross_weight,
            volume_weight,
            duty_tax,
            COALESCE(attachments::text, '[]') as attachments,
            to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS.US') as created_at,
            to_char(updated_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS.US') as updated_at
        FROM public.tbl_sea_waybill_details
        WHERE uuid = ?
    `

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	var (
		result          SeaWaybillDetailResponse
		attachmentsText string
	)

	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&result.UUID,
		&result.GrossWeight,
		&result.VolumeWeight,
		&result.DutyTax,
		&attachmentsText,
		&result.CreatedAt,
		&result.UpdatedAt,
	), uuid)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	if attachmentsText != "" && attachmentsText != "[]" {
		if err := json.Unmarshal([]byte(attachmentsText), &result.Attachments); err != nil {
			return nil, fmt.Errorf("failed to parse attachments: %w", err)
		}
	}

	return &result, nil
}

func (r repository) UpdateSeaWaybillDetail(ctx context.Context, data *seaWaybillDetailData) (*SeaWaybillDetailResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingAttachmentsText string
	_, err := db.QueryOneContext(ctx, pg.Scan(&existingAttachmentsText), `
        SELECT COALESCE(attachments::text, '[]')
        FROM public.tbl_sea_waybill_details
        WHERE uuid = ?
    `, data.UUID)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	var combinedAttachments []AttachmentInfo
	if existingAttachmentsText != "" && existingAttachmentsText != "[]" {
		if err := json.Unmarshal([]byte(existingAttachmentsText), &combinedAttachments); err != nil {
			return nil, fmt.Errorf("failed to parse existing attachments: %w", err)
		}
	}

	if len(data.Attachments) > 0 {
		combinedAttachments = append(combinedAttachments, data.Attachments...)
	}

	attachmentsJSON := "[]"
	if len(combinedAttachments) > 0 {
		b, err := json.Marshal(combinedAttachments)
		if err != nil {
			return nil, err
		}
		attachmentsJSON = string(b)
	}

	sqlStr := `
        UPDATE public.tbl_sea_waybill_details
        SET
            gross_weight = ?,
            volume_weight = ?,
            duty_tax = ?,
            attachments = ?::jsonb,
            updated_at = CURRENT_TIMESTAMP
        WHERE uuid = ?
        RETURNING
            uuid,
            gross_weight,
            volume_weight,
            duty_tax,
            COALESCE(attachments::text, '[]') as attachments,
            to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS.US') as created_at,
            to_char(updated_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS.US') as updated_at
    `

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	var (
		result          SeaWaybillDetailResponse
		attachmentsText string
	)

	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&result.UUID,
		&result.GrossWeight,
		&result.VolumeWeight,
		&result.DutyTax,
		&attachmentsText,
		&result.CreatedAt,
		&result.UpdatedAt,
	),
		data.GrossWeight,
		data.VolumeWeight,
		data.DutyTax,
		attachmentsJSON,
		data.UUID,
	)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	if attachmentsText != "" && attachmentsText != "[]" {
		if err := json.Unmarshal([]byte(attachmentsText), &result.Attachments); err != nil {
			return nil, fmt.Errorf("failed to parse attachments: %w", err)
		}
	}

	return &result, nil
}

func (r repository) DeleteAttachment(ctx context.Context, uuid, fileName string) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Close()

	var attachmentsText string
	_, err = tx.QueryOneContext(ctx, pg.Scan(&attachmentsText), `
        SELECT COALESCE(attachments::text, '[]')
        FROM public.tbl_sea_waybill_details
        WHERE uuid = ?
        FOR UPDATE
    `, uuid)
	if err != nil {
		tx.Rollback()
		return "", utils.PostgresErrorTransform(err)
	}

	var attachments []AttachmentInfo
	if attachmentsText != "" && attachmentsText != "[]" {
		if err := json.Unmarshal([]byte(attachmentsText), &attachments); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to parse attachments: %w", err)
		}
	}

	updatedAttachments := make([]AttachmentInfo, 0, len(attachments))
	removedURL := ""
	for _, attachment := range attachments {
		if attachment.FileName == fileName {
			removedURL = attachment.FileURL
			continue
		}
		updatedAttachments = append(updatedAttachments, attachment)
	}

	if removedURL == "" {
		tx.Rollback()
		return "", utils.ErrRecordNotFound
	}

	attachmentsJSON := "[]"
	if len(updatedAttachments) > 0 {
		b, err := json.Marshal(updatedAttachments)
		if err != nil {
			tx.Rollback()
			return "", err
		}
		attachmentsJSON = string(b)
	}

	sqlStr := `
        UPDATE public.tbl_sea_waybill_details
        SET attachments = ?::jsonb, updated_at = CURRENT_TIMESTAMP
        WHERE uuid = ?
    `

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := tx.Prepare(sqlStr)
	if err != nil {
		tx.Rollback()
		return "", utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(attachmentsJSON, uuid)
	if err != nil {
		tx.Rollback()
		return "", utils.PostgresErrorTransform(err)
	}

	if err := tx.Commit(); err != nil {
		return "", utils.PostgresErrorTransform(err)
	}

	return removedURL, nil
}
