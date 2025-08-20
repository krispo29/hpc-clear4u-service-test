package mawbinfo

import (
	"context"
	"errors"
	"fmt"
	"hpc-express-service/config"
	"hpc-express-service/gcs"
	"hpc-express-service/utils"
	"math"
	"net/http"
	"os"
	"path"
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
}

type service struct {
	selfRepo       Repository
	contextTimeout time.Duration
	gcsClient      *gcs.Client
	conf           *config.Config
}

func NewService(
	selfRepo Repository,
	timeout time.Duration,
	gcsClient *gcs.Client,
	conf *config.Config,
) Service {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
		gcsClient:      gcsClient,
		conf:           conf,
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
		if s.gcsClient == nil {
			basePath := filepath.Join("assets", "mawb-info", data.Mawb, data.Date)
			fileInfos, err := utils.UploadDocumentsToLocal(basePath, data.Attachments)
			if err != nil {
				return nil, fmt.Errorf("failed to upload attachments: %v", err)
			}
			for _, info := range fileInfos {
				fileURL, _ := info["filePath"].(string)
				fileURL = filepath.ToSlash(fileURL)
				attachmentInfos = append(attachmentInfos, AttachmentInfo{
					FileName: info["fileName"].(string),
					FileURL:  fileURL,
					FileSize: info["fileSize"].(int64),
				})
			}
		} else {
			for _, fileHeader := range data.Attachments {
				file, err := fileHeader.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open attachment %s: %v", fileHeader.Filename, err)
				}

				// Detect content type
				contentType := fileHeader.Header.Get("Content-Type")
				if contentType == "" {
					buff := make([]byte, 512)
					n, _ := file.Read(buff)
					contentType = http.DetectContentType(buff[:n])
					if seeker, ok := file.(interface {
						Seek(int64, int) (int64, error)
					}); ok {
						seeker.Seek(0, 0)
					}
				}
				// Generate unique filename and path
				newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
				fullPath := path.Join("mawb-info", data.Mawb, data.Date, newFileName)
				if _, _, err := s.gcsClient.UploadToGCS(ctx, file, fullPath, true, contentType); err != nil {
					file.Close()
					return nil, fmt.Errorf("failed to upload attachments: %v", err)
				}
				file.Close()

				fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.conf.GCSBucketName, fullPath)
				attachmentInfos = append(attachmentInfos, AttachmentInfo{
					FileName: newFileName,
					FileURL:  fileURL,
					FileSize: fileHeader.Size,
				})
			}
		}
	}

	fmt.Printf("DEBUG: Final attachmentInfos count: %d\n", len(attachmentInfos))

	// Call repository to update MAWB info
	result, err := s.selfRepo.UpdateMawbInfo(ctx, uuid, data, chargeableWeight, attachmentInfos)
	if err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: Repository result attachments count: %d\n", len(result.Attachments))

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
	fileURL, err := s.selfRepo.DeleteMawbInfoAttachment(ctx, uuid, fileName)
	if err != nil {
		return err // Repository error (e.g., not found)
	}

	// If file path is not empty, delete the file from storage
	if fileURL != "" {
		prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.conf.GCSBucketName)
		if s.gcsClient != nil && strings.HasPrefix(fileURL, prefix) {
			objectName := strings.TrimPrefix(fileURL, prefix)
			if err := s.gcsClient.DeleteImage(objectName); err != nil {
				fmt.Printf("warning: failed to delete attachment file '%s': %v\n", objectName, err)
			}
		} else {
			if err := os.Remove(fileURL); err != nil && !os.IsNotExist(err) {
				fmt.Printf("warning: failed to delete attachment file '%s': %v\n", fileURL, err)
			}
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
