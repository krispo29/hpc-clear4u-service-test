package draftmawb

import (
	"context"
	"fmt"
	"hpc-express-service/common"
	"hpc-express-service/setting"
)

type DraftMAWBService interface {
	GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	GetDraftMAWBByUUID(ctx context.Context, uuid string) (*DraftMAWB, error)
	CreateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error)
	UpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error)
	UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, statusUUID string) error
	GetAllDraftMAWB(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error)
	CancelDraftMAWB(ctx context.Context, mawbUUID string) error
	UndoCancelDraftMAWB(ctx context.Context, mawbUUID string) error
	GetDraftMAWBWithRelations(ctx context.Context, uuid string) (*DraftMAWBWithRelations, error)
	GetDraftMAWBWithRelationsByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWBWithRelations, error)
}

type draftMAWBService struct {
	repo      DraftMAWBRepository
	statusSvc setting.MasterStatusService
}

func NewDraftMAWBService(repo DraftMAWBRepository, statusSvc setting.MasterStatusService) DraftMAWBService {
	return &draftMAWBService{repo: repo, statusSvc: statusSvc}
}

func (s *draftMAWBService) GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	return s.repo.GetByMAWBUUID(ctx, mawbUUID)
}

func (s *draftMAWBService) GetDraftMAWBByUUID(ctx context.Context, uuid string) (*DraftMAWB, error) {
	return s.repo.GetByUUID(ctx, uuid)
}

// setDefaultStatus sets the status of the draft MAWB to the default 'Draft' status.
func (s *draftMAWBService) setDefaultStatus(ctx context.Context, draftMAWB *DraftMAWB) error {
	defaultStatus, err := s.statusSvc.GetDefaultStatusByType(ctx, "draft_mawb")
	if err != nil {
		return fmt.Errorf("error getting default status: %w", err)
	}
	if defaultStatus == nil {
		return fmt.Errorf("no default status found for draft_mawb")
	}
	draftMAWB.StatusUUID = defaultStatus.UUID
	return nil
}

func (s *draftMAWBService) CreateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error) {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.setDefaultStatus(txCtx, draftMAWB); err != nil {
		return nil, err
	}

	result, err := s.repo.CreateWithRelations(txCtx, draftMAWB, items, charges)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *draftMAWBService) UpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB, items []DraftMAWBItemInput, charges []DraftMAWBChargeInput) (*DraftMAWB, error) {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.setDefaultStatus(txCtx, draftMAWB); err != nil {
		return nil, err
	}

	result, err := s.repo.UpdateWithRelations(txCtx, draftMAWB, items, charges)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
func (s *draftMAWBService) UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, statusUUID string) error {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	draft, err := s.repo.GetByMAWBUUID(txCtx, mawbUUID)
	if err != nil {
		return err
	}
	if draft == nil {
		return nil // Or a not found error
	}

	if err := s.repo.UpdateStatus(txCtx, draft.UUID, statusUUID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *draftMAWBService) GetAllDraftMAWB(ctx context.Context, startDate, endDate string) ([]DraftMAWBListItem, error) {
	return s.repo.GetAll(ctx, startDate, endDate)
}

func (s *draftMAWBService) CancelDraftMAWB(ctx context.Context, mawbUUID string) error {
	cancelledStatus, err := s.statusSvc.GetStatusByNameAndType(ctx, "Cancelled", "draft_mawb")
	if err != nil {
		return err
	}
	if cancelledStatus == nil {
		return fmt.Errorf("status 'Cancelled' not found")
	}
	return s.UpdateDraftMAWBStatus(ctx, mawbUUID, cancelledStatus.UUID)
}

func (s *draftMAWBService) UndoCancelDraftMAWB(ctx context.Context, mawbUUID string) error {
	draftStatus, err := s.statusSvc.GetStatusByNameAndType(ctx, "Draft", "draft_mawb")
	if err != nil {
		return err
	}
	if draftStatus == nil {
		return fmt.Errorf("status 'Draft' not found")
	}
	return s.UpdateDraftMAWBStatus(ctx, mawbUUID, draftStatus.UUID)
}

func (s *draftMAWBService) GetDraftMAWBWithRelations(ctx context.Context, uuid string) (*DraftMAWBWithRelations, error) {
	return s.repo.GetWithRelations(ctx, uuid)
}
func (s *draftMAWBService) GetDraftMAWBWithRelationsByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWBWithRelations, error) {
	return s.repo.GetWithRelationsByMAWBUUID(ctx, mawbUUID)
}
