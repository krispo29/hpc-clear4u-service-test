package mawbinfo

import (
	"context"
	"errors"
	"fmt"
	"hpc-express-service/utils"
	"math"
	"os"
	"path/filepath"
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
	GetCargoManifest(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	CreateOrUpdateCargoManifest(ctx context.Context, mawbUUID string, data *CargoManifest) (*CargoManifest, error)
}

type service struct {
	selfRepo       Repository
	contextTimeout time.Duration
}

func NewService(
	selfRepo Repository,
	timeout time.Duration,
) Service {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
	}
}

func (s *service) CreateMawbInfo(ctx context.Context, data *CreateMawbInfoRequest) (*MawbInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate input data
	if err := s.validateInput(data); err != nil {
		return nil, err
	}

	exists, err := s.selfRepo.IsMawbExists(ctx, data.Mawb, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("mawb already exists")
	}

	// Convert chargeableWeight string to float64 with 2 decimal places
	chargeableWeight, err := s.convertChargeableWeight(data.ChargeableWeight)
	if err != nil {
		return nil, err
	}

	// Validate date format
	if err := s.validateDateFormat(data.Date); err != nil {
		return nil, err
	}

	// Call repository to create MAWB info
	result, err := s.selfRepo.CreateMawbInfo(ctx, data, chargeableWeight)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// validateInput validates all required fields
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

func (s *service) IsMawbExists(ctx context.Context, mawb string, uuid string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	return s.selfRepo.IsMawbExists(ctx, mawb, uuid)
}

// convertChargeableWeight converts string to float64 with exactly 2 decimal places
func (s *service) convertChargeableWeight(weightStr string) (float64, error) {
	// Trim whitespace
	weightStr = strings.TrimSpace(weightStr)

	if weightStr == "" {
		return 0, errors.New("chargeableWeight cannot be empty")
	}

	// Parse string to float64
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid chargeableWeight format: %s", weightStr)
	}

	// Check for negative values
	if weight < 0 {
		return 0, errors.New("chargeableWeight cannot be negative")
	}

	// Round to 2 decimal places
	weight = math.Round(weight*100) / 100

	return weight, nil
}

// validateDateFormat validates date format YYYY-MM-DD
func (s *service) validateDateFormat(dateStr string) error {
	dateStr = strings.TrimSpace(dateStr)

	if dateStr == "" {
		return errors.New("date cannot be empty")
	}

	// Parse date in YYYY-MM-DD format
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", dateStr)
	}

	return nil
}
func (s *service) GetMawbInfo(ctx context.Context, uuid string) (*MawbInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}

	result, err := s.selfRepo.GetMawbInfo(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetAllMawbInfo(ctx context.Context, startDate, endDate string) ([]*MawbInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate date formats if provided
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

	result, err := s.selfRepo.GetAllMawbInfo(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (s *service) UpdateMawbInfo(ctx context.Context, uuid string, data *UpdateMawbInfoRequest) (*MawbInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}

	// Validate input data
	if err := s.validateUpdateInput(data); err != nil {
		return nil, err
	}

	exists, err := s.selfRepo.IsMawbExists(ctx, data.Mawb, uuid)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("mawb already exists")
	}

	// Convert chargeableWeight string to float64 with 2 decimal places
	chargeableWeight, err := s.convertChargeableWeight(data.ChargeableWeight)
	if err != nil {
		return nil, err
	}

	// Validate date format
	if err := s.validateDateFormat(data.Date); err != nil {
		return nil, err
	}

	// Handle file attachments if present
	var attachmentInfos []AttachmentInfo
	if len(data.Attachments) > 0 {
		// Create upload directory based on MAWB and date
		uploadPath := filepath.Join("uploads", "mawb", data.Mawb, data.Date)

		// Upload files
		fileInfos, err := utils.UploadDocumentsToLocal(uploadPath, data.Attachments)
		if err != nil {
			return nil, fmt.Errorf("failed to upload attachments: %v", err)
		}

		// Convert to AttachmentInfo
		for _, fileInfo := range fileInfos {
			attachmentInfo := AttachmentInfo{
				FileName: fileInfo["fileName"].(string),
				FileURL:  fileInfo["filePath"].(string),
				FileSize: fileInfo["fileSize"].(int64),
			}
			attachmentInfos = append(attachmentInfos, attachmentInfo)
		}
	}

	// Call repository to update MAWB info
	result, err := s.selfRepo.UpdateMawbInfo(ctx, uuid, data, chargeableWeight, attachmentInfos)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) DeleteMawbInfo(ctx context.Context, uuid string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return errors.New("uuid is required")
	}

	err := s.selfRepo.DeleteMawbInfo(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteMawbInfoAttachment(ctx context.Context, uuid string, fileName string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return errors.New("uuid is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("fileName is required")
	}

	// Delete from repository, which returns the file path
	filePath, err := s.selfRepo.DeleteMawbInfoAttachment(ctx, uuid, fileName)
	if err != nil {
		return err // Repository error (e.g., not found)
	}

	// If file path is not empty, delete the file from local storage
	if filePath != "" {
		// It's good practice to ensure the path is within an expected directory
		// to prevent path traversal attacks, though for this implementation we'll assume it's safe.
		if err := os.Remove(filePath); err != nil {
			// Log the error but don't fail the whole operation,
			// as the DB record is already deleted.
			// Consider a more robust error handling strategy for production.
			fmt.Printf("warning: failed to delete attachment file '%s': %v\n", filePath, err)
		}
	}

	return nil
}

// validateUpdateInput validates all required fields for update
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

func (s *service) GetCargoManifest(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbUUID) == "" {
		return nil, errors.New("mawb uuid is required")
	}

	return s.selfRepo.GetCargoManifest(ctx, mawbUUID)
}

func (s *service) CreateOrUpdateCargoManifest(ctx context.Context, mawbUUID string, data *CargoManifest) (*CargoManifest, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbUUID) == "" {
		return nil, errors.New("mawb uuid is required")
	}

	if data == nil {
		return nil, errors.New("manifest data is required")
	}

	if strings.TrimSpace(data.MAWBNumber) == "" {
		return nil, errors.New("mawbNumber is required")
	}

	return s.selfRepo.CreateOrUpdateCargoManifest(ctx, mawbUUID, data)
}
