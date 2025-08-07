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

func (s *loggingService) UploadManifest(ctx context.Context, userUUID, originName, templateCode string, fileBytes []byte) (err error) {
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
	return s.next.UploadManifest(ctx, userUUID, originName, templateCode, fileBytes)
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

func (s *loggingService) UploadUpdateRawPreImport(ctx context.Context, userUUID, originName string, fileBytes []byte) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "download_pre_import",
			"user_uuid", userUUID,
			"origin_name", originName,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.UploadUpdateRawPreImport(ctx, userUUID, originName, fileBytes)
}

func (s *loggingService) GetOneByUploaddingUUID(ctx context.Context, uploadLoggingUUID string) (result *GetPreImportManifestModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_one_by_upload_uuid",
			"upload_uuid", uploadLoggingUUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetOneByUploaddingUUID(ctx, uploadLoggingUUID)
}

func (s *loggingService) GetSummaryByUploaddingUUID(ctx context.Context, uploadLoggingUUID string) (result *UploadSummaryModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_summary",
			"upload_uuid", uploadLoggingUUID,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetSummaryByUploaddingUUID(ctx, uploadLoggingUUID)
}
