package mawbinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hpc-express-service/utils"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type Repository interface {
	CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest, chargeableWeight string) (*MawbInfoResponse, error)
	GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error)
	GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error)
	UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest, chargeableWeight string, attachments []AttachmentInfo) (*MawbInfoResponse, error)
	DeleteMawbInfo(ctx context.Context, uuid string) error
	DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) (*AttachmentInfo, error)
	IsMawbExists(ctx context.Context, mawb string, uuid string) (bool, error)
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

func (r repository) CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest, chargeableWeight string) (*MawbInfoResponse, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	if err := r.createTableIfNotExists(ctx, db); err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Insert MAWB info record
	var response MawbInfoResponse
	sqlStr := `
		INSERT INTO tbl_mawb_info 
			(chargeable_weight, date, mawb, service_type, shipping_type)
		VALUES 
			(?, ?, ?, ?, ?)
		RETURNING 
			uuid, chargeable_weight, date, mawb, service_type, shipping_type,
			to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
	`
	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	_, err = db.QueryOne(pg.Scan(
		&response.UUID, &response.ChargeableWeight, &response.Date, &response.Mawb,
		&response.ServiceType, &response.ShippingType, &response.CreatedAt,
	), chargeableWeight, data.Date, data.Mawb, data.ServiceType, data.ShippingType)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return &response, nil
}

func (r repository) createTableIfNotExists(ctx context.Context, db orm.DB) error {
	// First create the table if it doesn't exist
	sqlStr := `
		CREATE TABLE IF NOT EXISTS tbl_mawb_info (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			chargeable_weight DECIMAL(10,2) NOT NULL,
			date DATE NOT NULL,
			mawb VARCHAR(255) NOT NULL,
			service_type VARCHAR(100) NOT NULL,
			shipping_type VARCHAR(100) NOT NULL,
			attachments JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(sqlStr); err != nil {
		return err
	}

	// Add attachments column if it doesn't exist (for existing tables)
	alterSQL := `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'tbl_mawb_info'
				AND column_name = 'attachments'
			) THEN
				ALTER TABLE tbl_mawb_info ADD COLUMN attachments JSONB;
			END IF;
		END $$;
	`
	_, err := db.Exec(alterSQL)
	return err
}

func (r repository) GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	var response MawbInfoResponse
	_, err = db.QueryOne(&response, `
		SELECT
			uuid, chargeable_weight, date, mawb, service_type, shipping_type,
			COALESCE(attachments, '[]'::jsonb) as attachments,
			to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
		FROM tbl_mawb_info
		WHERE uuid = ?
	`, uuid)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	return &response, nil
}

func (r repository) GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*MawbInfoResponse
	sqlStr := `
		SELECT
			uuid, chargeable_weight, date, mawb, service_type, shipping_type,
			COALESCE(attachments, '[]'::jsonb) as attachments,
			to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
		FROM tbl_mawb_info
	`
	var whereConditions []string
	if startDate != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("date >= '%s'", startDate))
	}
	if endDate != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("date <= '%s'", endDate))
	}
	if len(whereConditions) > 0 {
		sqlStr += " WHERE " + strings.Join(whereConditions, " AND ")
	}
	sqlStr += " ORDER BY created_at DESC"

	_, err = db.Query(&responses, sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	return responses, nil
}

func (r repository) UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest, chargeableWeight string, attachments []AttachmentInfo) (*MawbInfoResponse, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	existingRecord, err := r.GetMawbInfo(ctx, uuid)
	if err != nil {
		existingRecord = &MawbInfoResponse{Attachments: []AttachmentInfo{}}
	}

	allAttachments := existingRecord.Attachments
	if len(attachments) > 0 {
		allAttachments = append(allAttachments, attachments...)
	}

	attachmentsJSON, err := json.Marshal(allAttachments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attachments: %v", err)
	}

	var response MawbInfoResponse
	_, err = db.QueryOne(&response, `
		UPDATE tbl_mawb_info SET
			chargeable_weight = ?, date = ?, mawb = ?, service_type = ?, shipping_type = ?,
			attachments = ?, updated_at = CURRENT_TIMESTAMP
		WHERE uuid = ?
		RETURNING 
			uuid, chargeable_weight, date, mawb, service_type, shipping_type,
			COALESCE(attachments, '[]'::jsonb) as attachments,
			to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
	`, chargeableWeight, data.Date, data.Mawb, data.ServiceType, data.ShippingType, string(attachmentsJSON), uuid)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	return &response, nil
}

func (r repository) DeleteMawbInfo(ctx context.Context, uuid string) error {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return err
	}

	res, err := db.Exec(`DELETE FROM tbl_mawb_info WHERE uuid = ?`, uuid)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("record not found")
	}
	return nil
}

func (r repository) DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) (*AttachmentInfo, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	var attachments []AttachmentInfo
	_, err = db.QueryOne(&attachments, `SELECT attachments FROM tbl_mawb_info WHERE uuid = ?`, uuid)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	var updatedAttachments []AttachmentInfo
	var deletedAttachment *AttachmentInfo
	found := false
	for _, attachment := range attachments {
		if attachment.FileName == fileName {
			deletedAttachment = &attachment
			found = true
		} else {
			updatedAttachments = append(updatedAttachments, attachment)
		}
	}

	if !found {
		return nil, errors.New("attachment not found")
	}

	updatedAttachmentsJSON, err := json.Marshal(updatedAttachments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated attachments: %v", err)
	}

	_, err = db.Exec(`UPDATE tbl_mawb_info SET attachments = ? WHERE uuid = ?`, string(updatedAttachmentsJSON), uuid)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return deletedAttachment, nil
}

func (r repository) IsMawbExists(ctx context.Context, mawb string, uuid string) (bool, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return false, err
	}

	q := `SELECT count(*) FROM tbl_mawb_info WHERE mawb = ?`
	params := []interface{}{mawb}
	if uuid != "" {
		q += " AND uuid != ?"
		params = append(params, uuid)
	}

	var count int
	_, err = db.QueryOne(pg.Scan(&count), q, params...)
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, utils.PostgresErrorTransform(err)
	}
	return count > 0, nil
}
