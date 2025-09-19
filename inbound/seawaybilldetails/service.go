package seawaybilldetails

import (
	"context"
	"errors"
	"fmt"
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
	"hpc-express-service/utils"
)

type Service interface {
	CreateSeaWaybillDetails(ctx context.Context, data *UpsertSeaWaybillDetailsRequest) (*SeaWaybillDetails, error)
	GetSeaWaybillDetails(ctx context.Context, uuid string) (*SeaWaybillDetails, error)
	UpdateSeaWaybillDetails(ctx context.Context, uuid string, data *UpsertSeaWaybillDetailsRequest) (*SeaWaybillDetails, error)
	DeleteSeaWaybillAttachment(ctx context.Context, uuid, fileName string) error
}

type service struct {
	repo           Repository
	contextTimeout time.Duration
	gcsClient      *gcs.Client
	conf           *config.Config
}

func NewService(repo Repository, timeout time.Duration, gcsClient *gcs.Client, conf *config.Config) Service {
	return &service{
		repo:           repo,
		contextTimeout: timeout,
		gcsClient:      gcsClient,
		conf:           conf,
	}
}

func (s *service) CreateSeaWaybillDetails(ctx context.Context, data *UpsertSeaWaybillDetailsRequest) (*SeaWaybillDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if data == nil {
		return nil, errors.New("request data cannot be nil")
	}

	payload, err := s.preparePayload(data, uuid.New().String())
	if err != nil {
		return nil, err
	}

	attachments, err := s.storeAttachments(ctx, payload.UUID, data.Attachments)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.CreateSeaWaybillDetails(ctx, payload, attachments)
	if err != nil {
		s.cleanupStoredAttachments(attachments)
		return nil, err
	}

	return result, nil
}

func (s *service) GetSeaWaybillDetails(ctx context.Context, uuid string) (*SeaWaybillDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}

	return s.repo.GetSeaWaybillDetails(ctx, uuid)
}

func (s *service) UpdateSeaWaybillDetails(ctx context.Context, uuid string, data *UpsertSeaWaybillDetailsRequest) (*SeaWaybillDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("uuid is required")
	}
	if data == nil {
		return nil, errors.New("request data cannot be nil")
	}

	payload, err := s.preparePayload(data, uuid)
	if err != nil {
		return nil, err
	}

	attachments, err := s.storeAttachments(ctx, uuid, data.Attachments)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.UpdateSeaWaybillDetails(ctx, uuid, payload, attachments)
	if err != nil {
		s.cleanupStoredAttachments(attachments)
		return nil, err
	}

	return result, nil
}

func (s *service) DeleteSeaWaybillAttachment(ctx context.Context, uuid, fileName string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(uuid) == "" {
		return errors.New("uuid is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("fileName is required")
	}

	fileURL, err := s.repo.DeleteSeaWaybillAttachment(ctx, uuid, fileName)
	if err != nil {
		return err
	}

	if fileURL == "" {
		return nil
	}

	if s.gcsClient != nil && strings.HasPrefix(fileURL, fmt.Sprintf("https://storage.googleapis.com/%s/", s.conf.GCSBucketName)) {
		objectName := strings.TrimPrefix(fileURL, fmt.Sprintf("https://storage.googleapis.com/%s/", s.conf.GCSBucketName))
		if err := s.gcsClient.DeleteImage(objectName); err != nil {
			fmt.Printf("warning: failed to delete attachment file '%s': %v\n", objectName, err)
		}
		return nil
	}

	if err := os.Remove(fileURL); err != nil && !os.IsNotExist(err) {
		fmt.Printf("warning: failed to delete attachment file '%s': %v\n", fileURL, err)
	}

	return nil
}

func (s *service) preparePayload(data *UpsertSeaWaybillDetailsRequest, uuid string) (*seaWaybillDetailsRecord, error) {
	grossWeight, err := s.parseDecimalString(data.GrossWeight, "grossWeight")
	if err != nil {
		return nil, err
	}

	volumeWeight, err := s.parseDecimalString(data.VolumeWeight, "volumeWeight")
	if err != nil {
		return nil, err
	}

	dutyTax, err := s.parseDecimalString(data.DutyTax, "dutyTax")
	if err != nil {
		return nil, err
	}

	return &seaWaybillDetailsRecord{
		UUID:         uuid,
		GrossWeight:  grossWeight,
		VolumeWeight: volumeWeight,
		DutyTax:      dutyTax,
	}, nil
}

func (s *service) parseDecimalString(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return "", fmt.Errorf("%s must be a valid number", field)
	}
	if parsed < 0 {
		return "", fmt.Errorf("%s must be greater than or equal to zero", field)
	}

	return fmt.Sprintf("%.2f", parsed), nil
}

func (s *service) storeAttachments(ctx context.Context, recordUUID string, files []*multipart.FileHeader) ([]AttachmentInfo, error) {
	if len(files) == 0 {
		return nil, nil
	}

	allowedContentTypes := map[string]bool{
		"application/pdf":    true,
		"image/jpeg":         true,
		"image/png":          true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}

	var attachments []AttachmentInfo

	if s.gcsClient == nil {
		basePath := filepath.Join("assets", "sea-waybill-details", recordUUID)
		fileInfos, err := utils.UploadDocumentsToLocal(basePath, files)
		if err != nil {
			return nil, err
		}
		for _, info := range fileInfos {
			fileURL, _ := info["filePath"].(string)
			fileURL = filepath.ToSlash(fileURL)
			attachments = append(attachments, AttachmentInfo{
				FileName: info["fileName"].(string),
				FileURL:  fileURL,
				FileSize: info["fileSize"].(int64),
			})
		}
		return attachments, nil
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open attachment %s: %v", fileHeader.Filename, err)
		}

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

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		switch ext {
		case ".doc":
			contentType = "application/msword"
		case ".docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case ".jpg":
			if contentType == "" {
				contentType = "image/jpeg"
			}
		}

		if !allowedContentTypes[contentType] {
			file.Close()
			return nil, fmt.Errorf("file type not allowed: %s", contentType)
		}

		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
		fullPath := path.Join("sea-waybill-details", recordUUID, newFileName)

		if _, _, err := s.gcsClient.UploadToGCS(ctx, file, fullPath, true, contentType); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to upload attachments: %v", err)
		}
		file.Close()

		fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.conf.GCSBucketName, fullPath)
		attachments = append(attachments, AttachmentInfo{
			FileName: newFileName,
			FileURL:  fileURL,
			FileSize: fileHeader.Size,
		})
	}

	return attachments, nil
}

func (s *service) cleanupStoredAttachments(attachments []AttachmentInfo) {
	if len(attachments) == 0 {
		return
	}

	if s.gcsClient != nil {
		for _, attachment := range attachments {
			prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.conf.GCSBucketName)
			if strings.HasPrefix(attachment.FileURL, prefix) {
				objectName := strings.TrimPrefix(attachment.FileURL, prefix)
				if err := s.gcsClient.DeleteImage(objectName); err != nil {
					fmt.Printf("warning: failed to cleanup attachment '%s': %v\n", objectName, err)
				}
			}
		}
		return
	}

	for _, attachment := range attachments {
		if err := os.Remove(attachment.FileURL); err != nil && !os.IsNotExist(err) {
			fmt.Printf("warning: failed to cleanup attachment '%s': %v\n", attachment.FileURL, err)
		}
	}
}
