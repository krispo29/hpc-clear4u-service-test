package outbound

import (
	"context"
)

type DraftMAWBService interface {
	GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	GetDraftMAWBByUUID(ctx context.Context, uuid string) (*DraftMAWB, error)
	CreateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error)
	UpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error)
	UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, status string) error
	GetAllDraftMAWB(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error)
	CancelDraftMAWB(ctx context.Context, mawbUUID string) error
	UndoCancelDraftMAWB(ctx context.Context, mawbUUID string) error
	GetDraftMAWBWithRelations(ctx context.Context, uuid string) (*DraftMAWBWithRelations, error)
	GetDraftMAWBWithRelationsByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWBWithRelations, error)
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

func (s *draftMAWBService) GetDraftMAWBByUUID(ctx context.Context, uuid string) (*DraftMAWB, error) {
	return s.repo.GetByUUID(ctx, uuid)
}

func (s *draftMAWBService) CreateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error) {
	draftMAWB.Status = "Draft"
	return s.repo.CreateWithRelations(ctx, draftMAWB, items, charges)
}

func (s *draftMAWBService) UpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error) {
	draftMAWB.Status = "Draft"
	return s.repo.UpdateWithRelations(ctx, draftMAWB, items, charges)
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

func (s *draftMAWBService) CancelDraftMAWB(ctx context.Context, mawbUUID string) error {
	return s.repo.CancelByMAWBUUID(ctx, mawbUUID)
}

func (s *draftMAWBService) UndoCancelDraftMAWB(ctx context.Context, mawbUUID string) error {
	return s.repo.UndoCancelByMAWBUUID(ctx, mawbUUID)
}

func (s *draftMAWBService) GetDraftMAWBWithRelations(ctx context.Context, uuid string) (*DraftMAWBWithRelations, error) {
	return s.repo.GetWithRelations(ctx, uuid)
}
func (s *draftMAWBService) GetDraftMAWBWithRelationsByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWBWithRelations, error) {
	return s.repo.GetWithRelationsByMAWBUUID(ctx, mawbUUID)
}
