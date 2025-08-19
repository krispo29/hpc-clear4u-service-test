package mawbinfo

import (
	"context"
	"errors"
	"fmt"
	"hpc-express-service/gcs"
	"hpc-express-service/utils"
	"math"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest) (*MawbInfoResponse, error)
	GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error)
	GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error)
	UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest) (*MawbInfoResponse, error)
	DeleteMawbInfo(ctx context.Context, uuid string) error
	DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) error
	IsMawbExists(ctx context.Context, mawb string, uuid string) (bool, error)
}

type service struct {
	selfRepo       Repository
	gcsService     gcs.Service
	contextTimeout time.Duration
}

func NewService(
	selfRepo Repository,
	gcsService gcs.Service,
	timeout time.Duration,
) Service {
	return &service{
		selfRepo:       selfRepo,
		gcsService:     gcsService,
		contextTimeout: timeout,
	}
}

func (s *service) CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest) (*MawbInfoResponse, error) {
	tx, txCtx, err := utils.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.validateInput(data); err != nil {
		return nil, err
	}

	exists, err := s.selfRepo.IsMawbExists(txCtx, data.Mawb, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("mawb already exists")
	}

	chargeableWeight, err := s.convertChargeableWeight(data.ChargeableWeight)
	if err != nil {
		return nil, err
	}

	if err := s.validateDateFormat(data.Date); err != nil {
		return nil, err
	}

	result, err := s.selfRepo.CreateMawbInfo(txCtx, data, chargeableWeight)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest) (*MawbInfoResponse, error) {
	tx, txCtx, err := utils.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}

	if err := s.validateUpdateInput(data); err != nil {
		return nil, err
	}

	exists, err := s.selfRepo.IsMawbExists(txCtx, data.Mawb, uuid)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("mawb already exists")
	}

	chargeableWeight, err := s.convertChargeableWeight(data.ChargeableWeight)
	if err != nil {
		return nil, err
	}

	if err := s.validateDateFormat(data.Date); err != nil {
		return nil, err
	}

	var attachmentInfos []AttachmentInfo
	if len(data.Attachments) > 0 {
		for _, fileHeader := range data.Attachments {
			file, err := fileHeader.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open attachment file %s: %v", fileHeader.Filename, err)
			}
			defer file.Close()

			objectName := fmt.Sprintf("uploads/mawb/%s/%s/%d_%s", data.Mawb, data.Date, time.Now().UnixNano(), fileHeader.Filename)
			fileURL, err := s.gcsService.UploadToGCS(txCtx, file, objectName, true, fileHeader.Header.Get("Content-Type"))
			if err != nil {
				return nil, fmt.Errorf("failed to upload file %s to GCS: %v", fileHeader.Filename, err)
			}

			attachmentInfos = append(attachmentInfos, AttachmentInfo{
				FileName: fileHeader.Filename,
				FileURL:  fileURL,
				FileSize: fileHeader.Size,
			})
		}
	}

	result, err := s.selfRepo.UpdateMawbInfo(txCtx, uuid, data, chargeableWeight, attachmentInfos)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) error {
	tx, txCtx, err := utils.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if strings.TrimSpace(uuid) == "" {
		return errors.New("uuid is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("fileName is required")
	}

	deletedAttachment, err := s.selfRepo.DeleteMawbInfoAttachment(txCtx, uuid, fileName)
	if err != nil {
		return err
	}

	if deletedAttachment != nil && deletedAttachment.FileURL != "" {
		urlParts := strings.SplitN(deletedAttachment.FileURL, "/", 4)
		if len(urlParts) == 4 {
			objectName := urlParts[3]
			if err := s.gcsService.DeleteImage(objectName); err != nil {
				fmt.Printf("warning: failed to delete GCS object '%s': %v\n", objectName, err)
			}
		} else {
			fmt.Printf("warning: could not parse GCS object name from URL '%s'\n", deletedAttachment.FileURL)
		}
	}

	return tx.Commit()
}

// ... (rest of the helper functions like GetMawbInfo, GetAllMawbInfo, etc. remain the same)
func (s *service) GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}
	return s.selfRepo.GetMawbInfo(ctx, uuid)
}

func (s *service) GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	if startDate != "" {
		if err := s.validateDateFormat(startDate); err != nil {
			return nil, fmt.Errorf("invalid start date: %v", err)
		}
	}
	if endDate != "" {
		if err := s.validateDateFormat(endDate); err != nil {
			return nil, fmt.Errorf("invalid end date: %v", err)
		}
	}
	return s.selfRepo.GetAllMawbInfo(ctx, startDate, endDate)
}

func (s *service) DeleteMawbInfo(ctx context.Context, uuid string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	if strings.TrimSpace(uuid) == "" {
		return errors.New("uuid is required")
	}
	return s.selfRepo.DeleteMawbInfo(ctx, uuid)
}

func (s *service) IsMawbExists(ctx context.Context, mawb string, uuid string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.selfRepo.IsMawbExists(ctx, mawb, uuid)
}

func (s *service) validateInput(data *CreateMawbInfoRequest) error {
	if data == nil {
		return errors.New("request data cannot be nil")
	}
	if strings.TrimSpace(data.ChargeableWeight) == "" {
		return errors.New("chargeableWeight is required")
	}
	if strings.TrimSpace(data.Date) == "" {
		return errors.New("date is required")
	}
	if strings.TrimSpace(data.Mawb) == "" {
		return errors.New("mawb is required")
	}
	if strings.TrimSpace(data.ServiceType) == "" {
		return errors.New("serviceType is required")
	}
	if strings.TrimSpace(data.ShippingType) == "" {
		return errors.New("shippingType is required")
	}
	return nil
}

func (s *service) validateUpdateInput(data *UpdateMawbInfoRequest) error {
	if data == nil {
		return errors.New("request data cannot be nil")
	}
	if strings.TrimSpace(data.ChargeableWeight) == "" {
		return errors.New("chargeableWeight is required")
	}
	if strings.TrimSpace(data.Date) == "" {
		return errors.New("date is required")
	}
	if strings.TrimSpace(data.Mawb) == "" {
		return errors.New("mawb is required")
	}
	if strings.TrimSpace(data.ServiceType) == "" {
		return errors.New("serviceType is required")
	}
	if strings.TrimSpace(data.ShippingType) == "" {
		return errors.New("shippingType is required")
	}
	return nil
}

func (s *service) convertChargeableWeight(weightStr string) (string, error) {
	weightStr = strings.TrimSpace(weightStr)
	if weightStr == "" {
		return "", errors.New("chargeableWeight cannot be empty")
	}
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return "", fmt.Errorf("invalid chargeableWeight format: %s", weightStr)
	}
	if weight < 0 {
		return "", errors.New("chargeableWeight cannot be negative")
	}
	// return math.Round(weight*100) / 100, nil
	return fmt.Sprintf("%.2f", weight), nil
}

func (s *service) validateDateFormat(dateStr string) error {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return errors.New("date cannot be empty")
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", dateStr)
	}
	return nil
}
