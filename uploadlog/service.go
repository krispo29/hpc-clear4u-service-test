package uploadlog

import (
	"bytes"
	"context"
	"fmt"
	"hpc-express-service/gcs"
	"path/filepath"
	"time"

	randomUUID "github.com/satori/go.uuid"
)

type Service interface {
	Get(ctx context.Context, uuid string) (*GetUploadloggingModel, error)
	GetAllUploadloggings(ctx context.Context, startDate, endDate, category, subCategory string) ([]*GetUploadloggingModel, error)
	UploadLogFile(ctx context.Context, data *UploadFileModel) (string, error)
	Update(ctx context.Context, data *UpdateModel) error
}

type service struct {
	selfRepo       Repository
	contextTimeout time.Duration
	gcsClient      *gcs.Client
}

func NewService(
	selfRepo Repository,
	timeout time.Duration,
	gcsClient *gcs.Client,
) Service {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
		gcsClient:      gcsClient,
	}
}

func (s *service) Get(ctx context.Context, uuid string) (*GetUploadloggingModel, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.selfRepo.Get(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetAllUploadloggings(ctx context.Context, startDate, endDate, category, subCategory string) ([]*GetUploadloggingModel, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.selfRepo.GetAllUploadloggingsByCategoryAndSubCategory(ctx, startDate, endDate, category, subCategory)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) UploadLogFile(ctx context.Context, data *UploadFileModel) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	u2 := randomUUID.NewV4()

	var extension = filepath.Ext(data.FileName)
	newFileName := u2.String() + extension
	fullPath := fmt.Sprintf("%s/%s/%s", data.Category, data.TemplateCode, newFileName)

	// Determine content type
	contentType := "application/octet-stream" // default

	_, objAttrs, err := s.gcsClient.UploadToGCS(ctx, bytes.NewReader(data.FileBytes), fullPath, true, contentType)
	if err != nil {
		return "", err
	}

	loggingUploadUUID, err := s.selfRepo.Insert(ctx, &InsertModel{
		Mawb:         data.Mawb,
		FileName:     data.FileName,
		FileUrl:      objAttrs.MediaLink,
		TemplateCode: data.TemplateCode,
		Category:     data.Category,
		SubCategory:  data.SubCategory,
		CreatorUUID:  data.UserUUID,
		Status:       "created",
		Amount:       data.Amount,
	})

	if err != nil {
		return "", err
	}

	return loggingUploadUUID, nil
}

func (s *service) Update(ctx context.Context, data *UpdateModel) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.Update(ctx, data); err != nil {
		return err
	}

	return nil
}
