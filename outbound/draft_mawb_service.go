package outbound

import (
	"context"
)

type DraftMAWBService interface {
	GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	CreateOrUpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, status string) error
	GetAllDraftMAWB(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error)
}

type draftMAWBService struct {
	repo DraftMAWBRepository
}

func NewDraftMAWBService(repo DraftMAWBRepository) DraftMAWBService {
	return &draftMAWBService{repo: repo}
}

func (s *draftMAWBService) GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	return s.repo.GetByMAWBUUID(ctx, mawbUUID)
}

func (s *draftMAWBService) CreateOrUpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
	return s.repo.CreateOrUpdate(ctx, draftMAWB)
}
func (s *draftMAWBService) UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, status string) error {
	draft, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return err
	}
	if draft == nil {
		return nil // Or a not found error
	}
	return s.repo.UpdateStatus(ctx, draft.UUID, status)
}

func (s *draftMAWBService) GetAllDraftMAWB(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error) {
	return s.repo.GetAll(ctx, startDate, endDate)
}

// Calculation methods removed for now - will be implemented when Items and Charges are added back
