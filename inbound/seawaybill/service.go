package seawaybill

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"hpc-express-service/config"
	"hpc-express-service/gcs"
)

// Service exposes business logic for sea waybill details.
type Service interface {
	CreateSeaWaybillDetail(ctx context.Context, data *CreateSeaWaybillDetailRequest) (*SeaWaybillDetailResponse, error)
	GetSeaWaybillDetail(ctx context.Context, uuid string) (*SeaWaybillDetailResponse, error)
	UpdateSeaWaybillDetail(ctx context.Context, uuid string, data *UpdateSeaWaybillDetailRequest) (*SeaWaybillDetailResponse, error)
	DeleteAttachment(ctx context.Context, uuid, fileName string) error
}

type service struct {
	selfRepo       Repository
	contextTimeout time.Duration
	gcsClient      *gcs.Client
	conf           *config.Config
}

// NewService creates a new Service instance.
func NewService(repo Repository, timeout time.Duration, gcsClient *gcs.Client, conf *config.Config) Service {
	return &service{
		selfRepo:       repo,
		contextTimeout: timeout,
		gcsClient:      gcsClient,
		conf:           conf,
	}
}

func (s *service) CreateSeaWaybillDetail(ctx context.Context, data *CreateSeaWaybillDetailRequest) (*SeaWaybillDetailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if data == nil {
		return nil, errors.New("request data cannot be nil")
	}

	grossWeight, err := parseDecimal(data.GrossWeight)
	if err != nil {
		return nil, err
	}

	volumeWeight, err := parseDecimal(data.VolumeWeight)
	if err != nil {
		return nil, err
	}

	dutyTax, err := parseDecimal(data.DutyTax)
	if err != nil {
		return nil, err
	}

	recordUUID := uuid.New().String()

	attachmentInfos, err := s.storeAttachments(ctx, recordUUID, data.Attachments)
	if err != nil {
		return nil, err
	}

	repoData := &seaWaybillDetailData{
		UUID:         recordUUID,
		GrossWeight:  grossWeight,
		VolumeWeight: volumeWeight,
		DutyTax:      dutyTax,
		Attachments:  attachmentInfos,
	}

	result, err := s.selfRepo.CreateSeaWaybillDetail(ctx, repoData)
	if err != nil {
		s.cleanupAttachments(attachmentInfos)
		return nil, err
	}

	return result, nil
}

func (s *service) GetSeaWaybillDetail(ctx context.Context, uuid string) (*SeaWaybillDetailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}

	return s.selfRepo.GetSeaWaybillDetail(ctx, uuid)
}

func (s *service) UpdateSeaWaybillDetail(ctx context.Context, uuid string, data *UpdateSeaWaybillDetailRequest) (*SeaWaybillDetailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}

	if data == nil {
		return nil, errors.New("request data cannot be nil")
	}

	grossWeight, err := parseDecimal(data.GrossWeight)
	if err != nil {
		return nil, err
	}

	volumeWeight, err := parseDecimal(data.VolumeWeight)
	if err != nil {
		return nil, err
	}

	dutyTax, err := parseDecimal(data.DutyTax)
	if err != nil {
		return nil, err
	}

	attachmentInfos, err := s.storeAttachments(ctx, uuid, data.Attachments)
	if err != nil {
		return nil, err
	}

	repoData := &seaWaybillDetailData{
		UUID:         uuid,
		GrossWeight:  grossWeight,
		VolumeWeight: volumeWeight,
		DutyTax:      dutyTax,
		Attachments:  attachmentInfos,
	}

	result, err := s.selfRepo.UpdateSeaWaybillDetail(ctx, repoData)
	if err != nil {
		s.cleanupAttachments(attachmentInfos)
		return nil, err
	}

	return result, nil
}

func (s *service) DeleteAttachment(ctx context.Context, uuid, fileName string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return errors.New("uuid is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("fileName is required")
	}

	fileURL, err := s.selfRepo.DeleteAttachment(ctx, uuid, fileName)
	if err != nil {
		return err
	}

	if fileURL != "" {
		s.removeStoredFile(fileURL)
	}

	return nil
}

func parseDecimal(value string) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, errors.New("value is required")
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", value)
	}

	return parsed, nil
}

func (s *service) storeAttachments(ctx context.Context, recordUUID string, files []*multipart.FileHeader) ([]AttachmentInfo, error) {
	if len(files) == 0 {
		return nil, nil
	}

	const maxFileSize = 5 * 1024 * 1024 // 5MB
	allowedContentTypes := map[string]bool{
		"application/pdf":    true,
		"image/jpeg":         true,
		"image/png":          true,
		"image/jpg":          true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}

	attachments := make([]AttachmentInfo, 0, len(files))

	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return nil, fmt.Errorf("file %s is larger than 5MB", fileHeader.Filename)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			buffer := make([]byte, 512)
			n, _ := file.Read(buffer)
			contentType = http.DetectContentType(buffer[:n])
			if seeker, ok := file.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				file.Close()
				file, err = fileHeader.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to reopen file %s: %w", fileHeader.Filename, err)
				}
			}
		}

		if !allowedContentTypes[contentType] {
			lowerExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if !(lowerExt == ".doc" || lowerExt == ".docx") {
				file.Close()
				return nil, fmt.Errorf("file type %s is not allowed", contentType)
			}
		}

		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFileName(fileHeader.Filename))

		var fileURL string
		if s.gcsClient != nil {
			if seeker, ok := file.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				file.Close()
				file, err = fileHeader.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to reopen file %s: %w", fileHeader.Filename, err)
				}
			}

			objectPath := path.Join("sea-waybill-details", recordUUID, newFileName)
			if _, _, err := s.gcsClient.UploadToGCS(ctx, file, objectPath, true, contentType); err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to upload file %s: %w", fileHeader.Filename, err)
			}
			fileURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.conf.GCSBucketName, objectPath)
		} else {
			basePath := filepath.Join("assets", "sea-waybill-details", recordUUID)
			if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}

			destinationPath := filepath.Join(basePath, newFileName)
			destFile, err := os.Create(destinationPath)
			if err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to create file: %w", err)
			}

			if seeker, ok := file.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			}

			if _, err := io.Copy(destFile, file); err != nil {
				destFile.Close()
				file.Close()
				return nil, fmt.Errorf("failed to save file: %w", err)
			}
			destFile.Close()
			fileURL = filepath.ToSlash(destinationPath)
		}

		file.Close()

		attachments = append(attachments, AttachmentInfo{
			FileName: newFileName,
			FileURL:  fileURL,
			FileSize: fileHeader.Size,
		})
	}

	return attachments, nil
}

func (s *service) cleanupAttachments(attachments []AttachmentInfo) {
	for _, attachment := range attachments {
		s.removeStoredFile(attachment.FileURL)
	}
}

func (s *service) removeStoredFile(fileURL string) {
	if fileURL == "" {
		return
	}

	if s.gcsClient != nil && strings.HasPrefix(fileURL, "https://storage.googleapis.com/") {
		prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.conf.GCSBucketName)
		if strings.HasPrefix(fileURL, prefix) {
			objectName := strings.TrimPrefix(fileURL, prefix)
			if err := s.gcsClient.DeleteImage(objectName); err != nil {
				fmt.Printf("warning: failed to delete GCS object %s: %v\n", objectName, err)
			}
			return
		}
	}

	if err := os.Remove(fileURL); err != nil && !os.IsNotExist(err) {
		fmt.Printf("warning: failed to delete file %s: %v\n", fileURL, err)
	}
}

func sanitizeFileName(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return '-'
		default:
			return r
		}
	}, name)
	return cleaned
}
