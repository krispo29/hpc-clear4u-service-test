package seawaybilldetails

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-pg/pg/v9"

	"hpc-express-service/utils"
)

type Repository interface {
	CreateSeaWaybillDetails(ctx context.Context, data *seaWaybillDetailsRecord, attachments []AttachmentInfo) (*SeaWaybillDetails, error)
	GetSeaWaybillDetails(ctx context.Context, uuid string) (*SeaWaybillDetails, error)
	UpdateSeaWaybillDetails(ctx context.Context, uuid string, data *seaWaybillDetailsRecord, attachments []AttachmentInfo) (*SeaWaybillDetails, error)
	DeleteSeaWaybillAttachment(ctx context.Context, uuid, fileName string) (string, error)
}

type repository struct {
	contextTimeout time.Duration
}

type seaWaybillDetailsRecord struct {
	UUID         string
	GrossWeight  string
	VolumeWeight string
	DutyTax      string
}

func NewRepository(timeout time.Duration) Repository {
	return &repository{contextTimeout: timeout}
}

func (r repository) ensureTable(ctx context.Context, db *pg.DB) error {
	sqlStr := `
        CREATE TABLE IF NOT EXISTS tbl_sea_waybill_details (
            id SERIAL PRIMARY KEY,
            uuid UUID NOT NULL UNIQUE,
            gross_weight NUMERIC(10,2) NOT NULL,
            volume_weight NUMERIC(10,2) NOT NULL,
            duty_tax NUMERIC(10,2) NOT NULL,
            attachments JSONB,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `
	if _, err := db.ExecContext(ctx, sqlStr); err != nil {
		return err
	}

	alterSQL := `
        DO $$
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_name = 'tbl_sea_waybill_details'
                AND column_name = 'attachments'
            ) THEN
                ALTER TABLE tbl_sea_waybill_details ADD COLUMN attachments JSONB;
            END IF;
        END $$;
    `
	if _, err := db.ExecContext(ctx, alterSQL); err != nil {
		return err
	}

	return nil
}

func (r repository) CreateSeaWaybillDetails(ctx context.Context, data *seaWaybillDetailsRecord, attachments []AttachmentInfo) (*SeaWaybillDetails, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.ensureTable(ctx, db); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	attachmentsJSON := "[]"
	if len(attachments) > 0 {
		bytes, err := json.Marshal(attachments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attachments: %v", err)
		}
		attachmentsJSON = string(bytes)
	}

	sqlStr := `
        INSERT INTO tbl_sea_waybill_details
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
		response       SeaWaybillDetails
		attachmentsStr string
	)
	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&response.UUID,
		&response.GrossWeight,
		&response.VolumeWeight,
		&response.DutyTax,
		&attachmentsStr,
		&response.CreatedAt,
		&response.UpdatedAt,
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

	if attachmentsStr != "" && attachmentsStr != "[]" {
		if err := json.Unmarshal([]byte(attachmentsStr), &response.Attachments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attachments: %v", err)
		}
	}

	return &response, nil
}

func (r repository) GetSeaWaybillDetails(ctx context.Context, uuid string) (*SeaWaybillDetails, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.ensureTable(ctx, db); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	sqlStr := `
        SELECT
            uuid,
            gross_weight,
            volume_weight,
            duty_tax,
            COALESCE(attachments::text, '[]') as attachments,
            to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at,
            to_char(updated_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as updated_at
        FROM tbl_sea_waybill_details
        WHERE uuid = ?
    `
	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	var (
		response       SeaWaybillDetails
		attachmentsStr string
	)
	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&response.UUID,
		&response.GrossWeight,
		&response.VolumeWeight,
		&response.DutyTax,
		&attachmentsStr,
		&response.CreatedAt,
		&response.UpdatedAt,
	), uuid)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	if attachmentsStr != "" && attachmentsStr != "[]" {
		if err := json.Unmarshal([]byte(attachmentsStr), &response.Attachments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attachments: %v", err)
		}
	}

	return &response, nil
}

func (r repository) UpdateSeaWaybillDetails(ctx context.Context, uuid string, data *seaWaybillDetailsRecord, attachments []AttachmentInfo) (*SeaWaybillDetails, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.ensureTable(ctx, db); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	var existingAttachmentsStr string
	_, err := db.QueryOneContext(ctx, pg.Scan(&existingAttachmentsStr), utils.ReplaceSQL(`
        SELECT COALESCE(attachments::text, '[]')
        FROM tbl_sea_waybill_details
        WHERE uuid = ?
    `, "?"), uuid)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	var existingAttachments []AttachmentInfo
	if existingAttachmentsStr != "" && existingAttachmentsStr != "[]" {
		if err := json.Unmarshal([]byte(existingAttachmentsStr), &existingAttachments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing attachments: %v", err)
		}
	}

	allAttachments := existingAttachments
	if len(attachments) > 0 {
		allAttachments = append(allAttachments, attachments...)
	}

	attachmentsJSON := "[]"
	if len(allAttachments) > 0 {
		bytes, err := json.Marshal(allAttachments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attachments: %v", err)
		}
		attachmentsJSON = string(bytes)
	}

	sqlStr := `
        UPDATE tbl_sea_waybill_details
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
		response       SeaWaybillDetails
		attachmentsStr string
	)
	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&response.UUID,
		&response.GrossWeight,
		&response.VolumeWeight,
		&response.DutyTax,
		&attachmentsStr,
		&response.CreatedAt,
		&response.UpdatedAt,
	),
		data.GrossWeight,
		data.VolumeWeight,
		data.DutyTax,
		attachmentsJSON,
		uuid,
	)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	if attachmentsStr != "" && attachmentsStr != "[]" {
		if err := json.Unmarshal([]byte(attachmentsStr), &response.Attachments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attachments: %v", err)
		}
	}

	return &response, nil
}

func (r repository) DeleteSeaWaybillAttachment(ctx context.Context, uuid, fileName string) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.Begin()
	if err != nil {
		return "", utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	var attachmentsStr string
	_, err = tx.QueryOneContext(ctx, pg.Scan(&attachmentsStr), utils.ReplaceSQL(`
        SELECT COALESCE(attachments::text, '[]')
        FROM tbl_sea_waybill_details
        WHERE uuid = ?
        FOR UPDATE
    `, "?"), uuid)
	if err != nil {
		tx.Rollback()
		return "", utils.PostgresErrorTransform(err)
	}

	var attachments []AttachmentInfo
	if attachmentsStr != "" && attachmentsStr != "[]" {
		if err := json.Unmarshal([]byte(attachmentsStr), &attachments); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to unmarshal attachments: %v", err)
		}
	}

	var (
		updatedAttachments []AttachmentInfo
		removedFileURL     string
		found              bool
	)

	for _, attachment := range attachments {
		if attachment.FileName == fileName {
			removedFileURL = attachment.FileURL
			found = true
			continue
		}
		updatedAttachments = append(updatedAttachments, attachment)
	}

	if !found {
		tx.Rollback()
		return "", utils.ErrRecordNotFound
	}

	attachmentsJSON := "[]"
	if len(updatedAttachments) > 0 {
		bytes, err := json.Marshal(updatedAttachments)
		if err != nil {
			tx.Rollback()
			return "", fmt.Errorf("failed to marshal updated attachments: %v", err)
		}
		attachmentsJSON = string(bytes)
	}

	stmt, err := tx.Prepare(utils.ReplaceSQL(`
        UPDATE tbl_sea_waybill_details
        SET attachments = ?::jsonb, updated_at = CURRENT_TIMESTAMP
        WHERE uuid = ?
    `, "?"))
	if err != nil {
		return "", utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, attachmentsJSON, uuid)
	if err != nil {
		return "", utils.PostgresErrorTransform(err)
	}

	if result.RowsAffected() == 0 {
		return "", utils.ErrRecordNotFound
	}

	if err := tx.Commit(); err != nil {
		return "", utils.PostgresErrorTransform(err)
	}

	return removedFileURL, nil
}
