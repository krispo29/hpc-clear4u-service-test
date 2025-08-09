package outbound

import (
	"bytes"
	"context"
	"log"
	"path/filepath"
	"strings"

	randomUUID "github.com/satori/go.uuid"
)

func (s *service) GetAllMawnInfo(ctx context.Context, start, end string) ([]*GetMawbInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if result, err := s.selfRepo.GetAll(ctx, start, end); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (s *service) CreateMawnInfo(ctx context.Context, data *CreateMawbInfo) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if uuid, err := s.selfRepo.Create(ctx, data); err != nil {
		return "", err
	} else {
		return uuid, nil
	}
}

func (s *service) GetOneMawnInfo(ctx context.Context, uuid string) (*GetMawbInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetOne(ctx, uuid)
	if err != nil {
		return nil, err
	}

	attchments, err := s.selfRepo.GetAttchments(ctx, result.UUID)
	if err != nil {
		return nil, err
	}

	result.Attchments = attchments
	return result, nil
}

func (s *service) UpdateMawnInfo(ctx context.Context, data *UpdateMawbInfoModel) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.Update(ctx, data); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteMawnInfo(ctx context.Context, uuid string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.Delete(ctx, uuid); err != nil {
		return err
	}

	return nil
}

func (s *service) UploadAttachment(ctx context.Context, uuid, fileOriginName string, fileBytes []byte) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	mawnInfo, err := s.selfRepo.GetOne(ctx, uuid)
	if err != nil {
		return err
	}

	// Upload File
	var attachmentFileUrl string
	if len(fileOriginName) > 0 {
		destinationPath := strings.TrimSpace(mawnInfo.Mawb) + "/"
		u2 := randomUUID.NewV4()

		var extension = filepath.Ext(fileOriginName)
		newFileName := u2.String() + extension
		fullPath := "mawb/" + destinationPath + newFileName

		// Determine content type
		contentType := "application/octet-stream" // default
		if extension == ".pdf" {
			contentType = "application/pdf"
		}

		if s.gcsClient == nil {
			return fmt.Errorf("GCS client is not initialized")
		}
		_, _, err := s.gcsClient.UploadToGCS(ctx, bytes.NewReader(fileBytes), fullPath, true, contentType)
		if err != nil {
			log.Println("err", err)
			return err
		}

		attachmentFileUrl = "https://storage.googleapis.com/" + s.conf.GCSBucketName + "/" + fullPath
	}

	if err := s.selfRepo.InsertAttchment(ctx, &InsertAttchmentModel{
		MawbUUID: uuid,
		FileName: fileOriginName,
		FileURL:  attachmentFileUrl,
	}); err != nil {
		return err
	}

	log.Println(attachmentFileUrl)
	return nil

}
