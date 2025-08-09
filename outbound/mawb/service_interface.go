package outbound

import (
	"bytes"
	"context"
	"time"

	"hpc-express-service/config"
	"hpc-express-service/gcs"
)

type OutboundMawbService interface {
	// Mawb Info
	GetAllMawnInfo(ctx context.Context, start, end string) ([]*GetMawbInfo, error)
	CreateMawnInfo(ctx context.Context, data *CreateMawbInfo) (string, error)
	GetOneMawnInfo(ctx context.Context, uuid string) (*GetMawbInfo, error)
	UpdateMawnInfo(ctx context.Context, data *UpdateMawbInfoModel) error
	DeleteMawnInfo(ctx context.Context, uuid string) error
	UploadAttachment(ctx context.Context, uuid, fileOriginName string, fileBytes []byte) error

	// Draft
	GetAllMawbDraft(ctx context.Context, start, end string) ([]*GetAllMawbDraftModel, error)
	GetOneMawbDraft(ctx context.Context, uuid string) (*GetMawbDraftModel, error)
	PrintMawbDraft(ctx context.Context, uuid string) (bytes.Buffer, error)
	PreviewDraftMawb(ctx context.Context, data *RequestDraftModel) (bytes.Buffer, error)
	SaveDraftMawb(ctx context.Context, data *RequestDraftModel) (bytes.Buffer, error)
	UpdateDraftMawb(ctx context.Context, data *RequestUpdateMawbDraftModel) (bytes.Buffer, error)
	generateDraftMawb(ctx context.Context, data *RequestDraftModel, isPreview bool) (bytes.Buffer, error)

	// Cargo Manifest
	GetCargoManifestByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error)
	CreateOrUpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	ConfirmCargoManifest(ctx context.Context, mawbInfoUUID string) error
	RejectCargoManifest(ctx context.Context, mawbInfoUUID string) error
	PrintCargoManifest(ctx context.Context, mawbInfoUUID string) (bytes.Buffer, error)

	// Draft V2
	GetDraftMAWBByMAWBInfoUUIDV2(ctx context.Context, mawbInfoUUID string) (*DraftMAWBV2, error)
	CreateOrUpdateDraftMAWBV2(ctx context.Context, draft *DraftMAWBV2) (*DraftMAWBV2, error)
	ConfirmDraftMAWBV2(ctx context.Context, mawbInfoUUID string) error
	RejectDraftMAWBV2(ctx context.Context, mawbInfoUUID string) error
	PrintDraftMAWBV2(ctx context.Context, mawbInfoUUID string) (bytes.Buffer, error)
}

type service struct {
	selfRepo       OutboundMawbRepository
	contextTimeout time.Duration
	gcsClient      *gcs.Client
	conf           *config.Config
}

func NewOutboundMawbService(
	selfRepo OutboundMawbRepository,
	timeout time.Duration,
	gcsClient *gcs.Client,
	conf *config.Config,
) OutboundMawbService {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
		gcsClient:      gcsClient,
		conf:           conf,
	}
}
