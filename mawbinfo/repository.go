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
)

type Repository interface {
	CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest, chargeableWeight float64) (*MawbInfoResponse, error)
	GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error)
	GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error)
	UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest, chargeableWeight float64, attachments []AttachmentInfo) (*MawbInfoResponse, error)
	DeleteMawbInfo(ctx context.Context, uuid string) error
	DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) (string, error)
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

func (r repository) CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest, chargeableWeight float64) (*MawbInfoResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create table if not exists
	err := r.createTableIfNotExists(ctx, db)
	if err != nil {
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
			uuid, 
			chargeable_weight, 
			date, 
			mawb, 
			service_type, 
			shipping_type,
			to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&response.UUID,
		&response.ChargeableWeight,
		&response.Date,
		&response.Mawb,
		&response.ServiceType,
		&response.ShippingType,
		&response.CreatedAt,
	),
		chargeableWeight,
		data.Date,
		data.Mawb,
		data.ServiceType,
		data.ShippingType,
	)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	return &response, nil
}

func (r repository) createTableIfNotExists(ctx context.Context, db *pg.DB) error {
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

	_, err := db.ExecContext(ctx, sqlStr)
	if err != nil {
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

	_, err = db.ExecContext(ctx, alterSQL)
	return err
}

// Helper function to check if attachments column exists
func (r repository) hasAttachmentsColumn(ctx context.Context, db *pg.DB) bool {
	var count int
	_, err := db.QueryOneContext(ctx, pg.Scan(&count), `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = 'tbl_mawb_info' 
		AND column_name = 'attachments'
	`)
	return err == nil && count > 0
}
func (r repository) GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First ensure the table has the attachments column
	err := r.createTableIfNotExists(ctx, db)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	var response MawbInfoResponse
	var attachmentsStr string

	// Build query based on whether attachments column exists
	var sqlStr string
	if r.hasAttachmentsColumn(ctx, db) {
		sqlStr = `
			SELECT 
				uuid, 
				chargeable_weight, 
				date, 
				mawb, 
				service_type, 
				shipping_type,
				COALESCE(attachments::text, '[]') as attachments,
				to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
			FROM tbl_mawb_info 
			WHERE uuid = ?
		`
	} else {
		sqlStr = `
			SELECT 
				uuid, 
				chargeable_weight, 
				date, 
				mawb, 
				service_type, 
				shipping_type,
				'[]' as attachments,
				to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
			FROM tbl_mawb_info 
			WHERE uuid = ?
		`
	}

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&response.UUID,
		&response.ChargeableWeight,
		&response.Date,
		&response.Mawb,
		&response.ServiceType,
		&response.ShippingType,
		&attachmentsStr,
		&response.CreatedAt,
	), uuid)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Parse attachments JSON
	if attachmentsStr != "" && attachmentsStr != "[]" {
		err = json.Unmarshal([]byte(attachmentsStr), &response.Attachments)
		if err != nil {
			// If parsing fails, set empty array
			response.Attachments = []AttachmentInfo{}
		}
	}

	return &response, nil
}

func (r repository) GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var responses []*MawbInfoResponse

	// First ensure the table has the attachments column
	err := r.createTableIfNotExists(ctx, db)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Build base query based on whether attachments column exists
	var sqlStr string
	if r.hasAttachmentsColumn(ctx, db) {
		sqlStr = `
			SELECT 
				uuid, 
				chargeable_weight, 
				date, 
				mawb, 
				service_type, 
				shipping_type,
				COALESCE(attachments::text, '[]') as attachments,
				to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
			FROM tbl_mawb_info`
	} else {
		sqlStr = `
			SELECT 
				uuid, 
				chargeable_weight, 
				date, 
				mawb, 
				service_type, 
				shipping_type,
				'[]' as attachments,
				to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
			FROM tbl_mawb_info`
	}

	var whereConditions []string

	// Add date filtering conditions (dates are already validated in service layer)
	if startDate != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("date >= '%s'", startDate))
	}

	if endDate != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("date <= '%s'", endDate))
	}

	// Add WHERE clause if there are conditions
	if len(whereConditions) > 0 {
		sqlStr += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	sqlStr += " ORDER BY created_at DESC"

	// Execute query with manual scanning to handle attachments JSON
	type tempResponse struct {
		UUID             string  `pg:"uuid"`
		ChargeableWeight float64 `pg:"chargeable_weight"`
		Date             string  `pg:"date"`
		Mawb             string  `pg:"mawb"`
		ServiceType      string  `pg:"service_type"`
		ShippingType     string  `pg:"shipping_type"`
		AttachmentsStr   string  `pg:"attachments"`
		CreatedAt        string  `pg:"created_at"`
	}

	var tempResponses []tempResponse
	_, err = db.QueryContext(ctx, &tempResponses, sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Convert temp responses to actual responses with parsed attachments
	for _, temp := range tempResponses {
		response := &MawbInfoResponse{
			UUID:             temp.UUID,
			ChargeableWeight: temp.ChargeableWeight,
			Date:             temp.Date,
			Mawb:             temp.Mawb,
			ServiceType:      temp.ServiceType,
			ShippingType:     temp.ShippingType,
			CreatedAt:        temp.CreatedAt,
		}

		// Parse attachments JSON
		if temp.AttachmentsStr != "" && temp.AttachmentsStr != "[]" {
			err = json.Unmarshal([]byte(temp.AttachmentsStr), &response.Attachments)
			if err != nil {
				// If parsing fails, set empty array
				response.Attachments = []AttachmentInfo{}
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
}
func (r repository) UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest, chargeableWeight float64, attachments []AttachmentInfo) (*MawbInfoResponse, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get existing attachments first with a fresh context
	getCtx := context.WithValue(context.Background(), "postgreSQLConn", db)
	existingRecord, err := r.GetMawbInfo(getCtx, uuid)
	if err != nil {
		// If record not found, continue with empty attachments
		// This handles the case where the record might not exist yet
		existingRecord = &MawbInfoResponse{
			Attachments: []AttachmentInfo{},
		}
	}

	// Combine existing attachments with new ones
	var allAttachments []AttachmentInfo
	if existingRecord.Attachments != nil {
		allAttachments = append(allAttachments, existingRecord.Attachments...)
	}
	if len(attachments) > 0 {
		allAttachments = append(allAttachments, attachments...)
	}

	// Convert combined attachments to JSON
	var attachmentsJSON []byte
	if len(allAttachments) > 0 {
		attachmentsJSON, err = json.Marshal(allAttachments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attachments: %v", err)
		}
	}

	var response MawbInfoResponse
	var attachmentsStr string
	sqlStr := `
		UPDATE tbl_mawb_info 
		SET 
			chargeable_weight = ?, 
			date = ?, 
			mawb = ?, 
			service_type = ?, 
			shipping_type = ?,
			attachments = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE uuid = ?
		RETURNING 
			uuid, 
			chargeable_weight, 
			date, 
			mawb, 
			service_type, 
			shipping_type,
			COALESCE(attachments::text, '[]') as attachments,
			to_char(created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
	`

	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	_, err = stmt.QueryOneContext(ctx, pg.Scan(
		&response.UUID,
		&response.ChargeableWeight,
		&response.Date,
		&response.Mawb,
		&response.ServiceType,
		&response.ShippingType,
		&attachmentsStr,
		&response.CreatedAt,
	),
		chargeableWeight,
		data.Date,
		data.Mawb,
		data.ServiceType,
		data.ShippingType,
		string(attachmentsJSON),
		uuid,
	)

	if err != nil {
		return nil, utils.PostgresErrorTransform(err)
	}

	// Parse attachments JSON
	if attachmentsStr != "" && attachmentsStr != "[]" {
		err = json.Unmarshal([]byte(attachmentsStr), &response.Attachments)
		if err != nil {
			// If parsing fails, set empty array
			response.Attachments = []AttachmentInfo{}
		}
	}

	return &response, nil
}

func (r repository) DeleteMawbInfo(ctx context.Context, uuid string) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlStr := `DELETE FROM tbl_mawb_info WHERE uuid = ?`
	sqlStr = utils.ReplaceSQL(sqlStr, "?")

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, uuid)
	if err != nil {
		return utils.PostgresErrorTransform(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

func (r repository) DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.Begin()
	if err != nil {
		return "", utils.PostgresErrorTransform(err)
	}
	defer tx.Rollback()

	// 1. Get the current attachments
	var attachmentsStr string
	_, err = tx.QueryOneContext(ctx, pg.Scan(&attachmentsStr), `
        SELECT COALESCE(attachments::text, '[]') 
        FROM tbl_mawb_info 
        WHERE uuid = ?
    `, uuid)
	if err != nil {
		if err == pg.ErrNoRows {
			return "", errors.New("mawb info not found")
		}
		return "", utils.PostgresErrorTransform(err)
	}

	// 2. Unmarshal into a slice of AttachmentInfo
	var attachments []AttachmentInfo
	if err := json.Unmarshal([]byte(attachmentsStr), &attachments); err != nil {
		return "", fmt.Errorf("failed to unmarshal attachments: %v", err)
	}

	// 3. Find the attachment to delete and create a new slice
	var updatedAttachments []AttachmentInfo
	var deletedFileUrl string
	found := false
	for _, attachment := range attachments {
		if attachment.FileName == fileName {
			deletedFileUrl = attachment.FileURL
			found = true
		} else {
			updatedAttachments = append(updatedAttachments, attachment)
		}
	}

	if !found {
		return "", errors.New("attachment not found")
	}

	// 4. Marshal the updated slice back to JSON
	updatedAttachmentsJSON, err := json.Marshal(updatedAttachments)
	if err != nil {
		return "", fmt.Errorf("failed to marshal updated attachments: %v", err)
	}

	// 5. Update the database
	sqlStr := `
        UPDATE tbl_mawb_info 
        SET attachments = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE uuid = ?
    `
	sqlStr = utils.ReplaceSQL(sqlStr, "?")
	stmt, err := tx.Prepare(sqlStr)
	if err != nil {
		return "", utils.PostgresErrorTransform(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, string(updatedAttachmentsJSON), uuid)
	if err != nil {
		return "", utils.PostgresErrorTransform(err)
	}

	if result.RowsAffected() == 0 {
		return "", errors.New("mawb info not found during update")
	}

	// 6. Commit the transaction
	if err := tx.Commit(); err != nil {
		return "", utils.PostgresErrorTransform(err)
	}

	return deletedFileUrl, nil
}
