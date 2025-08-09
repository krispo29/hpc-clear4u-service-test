package outbound

import (
	"context"
	"time"
)

type OutboundMawbRepository interface {
	// Info
	GetAll(ctx context.Context, start, end string) ([]*GetMawbInfo, error)
	Create(ctx context.Context, data *CreateMawbInfo) (string, error)
	GetOne(ctx context.Context, uuid string) (*GetMawbInfo, error)
	Update(ctx context.Context, data *UpdateMawbInfoModel) error
	Delete(ctx context.Context, uuid string) error
	InsertAttchment(ctx context.Context, data *InsertAttchmentModel) error
	GetAttchments(ctx context.Context, uuid string) ([]*GetAttchmentModel, error)

	// Draft
	GetAllMawbDraft(ctx context.Context, start, end string) ([]*GetAllMawbDraftModel, error)
	GetOneMawbDraft(ctx context.Context, uuid string) (*GetMawbDraftModel, error)
	CreateMawbDraft(ctx context.Context, data *RequestDraftModel) error
	UpdateMawbDraft(ctx context.Context, data *RequestUpdateMawbDraftModel) error
}

type repository struct {
	contextTimeout time.Duration
}

func NewOutboundMawbRepository(
	timeout time.Duration,
) OutboundMawbRepository {
	return &repository{
		contextTimeout: timeout,
	}
}
