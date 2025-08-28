package inbound

import (
	"bytes"
	"context"
	"time"

	"github.com/go-kit/log"
)

type loggingService struct {
	logger log.Logger
	next   InboundExpressService
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s InboundExpressService) InboundExpressService {
	return &loggingService{logger, s}
}
func (s *loggingService) GetAllMawb(ctx context.Context) (result []*GetPreImportManifestModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_all_mawb",
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetAllMawb(ctx)
}

func (s *loggingService) InsertPreImportManifestHeader(ctx context.Context, data *InsertPreImportHeaderManifestModel) (uuid string, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "create_mawb",
			"uuid", uuid,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.InsertPreImportManifestHeader(ctx, data)
}

func (s *loggingService) UpdatePreImportManifestHeader(ctx context.Context, data *UpdatePreImportHeaderManifestModel) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "update_mawb",
			"uuid", data.UUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.UpdatePreImportManifestHeader(ctx, data)
}

func (s *loggingService) UploadManifestDetails(ctx context.Context, userUUID, headerUUID, originName, templateCode string, fileBytes []byte) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "upload_manifest",
			"userUUID", userUUID,
			"template_code", templateCode,
			"originName", originName,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.UploadManifestDetails(ctx, userUUID, headerUUID, originName, templateCode, fileBytes)
}

func (s *loggingService) DownloadPreImport(ctx context.Context, uploadLoggingUUID string) (fileName string, result *bytes.Buffer, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "download_pre_import",
			"upload_logging_uuid", uploadLoggingUUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.DownloadPreImport(ctx, uploadLoggingUUID)
}

func (s *loggingService) DownloadRawPreImport(ctx context.Context, uploadLoggingUUID string) (filename string, excelBuf *bytes.Buffer, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "download_raw_pre_import",
			"file_name", "filename",
			"upload_logging_uuid", uploadLoggingUUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.DownloadRawPreImport(ctx, uploadLoggingUUID)
}

func (s *loggingService) UploadUpdateRawPreImport(ctx context.Context, userUUID, headerUUID, originName string, fileBytes []byte) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "download_pre_import",
			"user_uuid", userUUID,
			"origin_name", originName,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.UploadUpdateRawPreImport(ctx, userUUID, headerUUID, originName, fileBytes)
}

func (s *loggingService) GetOneByHeaderUUID(ctx context.Context, headerUUID string) (result *GetPreImportManifestModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_one_by_header_uuid",
			"header_uuid", headerUUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetOneByHeaderUUID(ctx, headerUUID)
}

func (s *loggingService) GetSummaryByHeaderUUID(ctx context.Context, headerUUID string) (result *UploadSummaryModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_summary",
			"header_uuid", headerUUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetSummaryByHeaderUUID(ctx, headerUUID)
}
