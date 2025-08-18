package setting

import (
	"context"
	"time"
)

type MasterStatusService interface {
	CreateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error)
	GetAllMasterStatuses(ctx context.Context) ([]MasterStatus, error)
	GetMasterStatusesByType(ctx context.Context, statusType string) ([]MasterStatus, error)
	GetMasterStatusByUUID(ctx context.Context, uuid string) (*MasterStatus, error)
	UpdateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error)
	DeleteMasterStatus(ctx context.Context, uuid string) error
	GetStatusByNameAndType(ctx context.Context, name, statusType string) (*MasterStatus, error)
	GetDefaultStatusByType(ctx context.Context, statusType string) (*MasterStatus, error)
}

type masterStatusService struct {
	repo           MasterStatusRepository
	contextTimeout time.Duration
}

func (s *masterStatusService) GetDefaultStatusByType(ctx context.Context, statusType string) (*MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.GetDefaultStatusByType(ctx, statusType)
}

func NewMasterStatusService(repo MasterStatusRepository, timeout time.Duration) MasterStatusService {
	return &masterStatusService{
		repo:           repo,
		contextTimeout: timeout,
	}
}

func (s *masterStatusService) CreateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.CreateMasterStatus(ctx, status)
}

func (s *masterStatusService) GetAllMasterStatuses(ctx context.Context) ([]MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.GetAllMasterStatuses(ctx)
}

func (s *masterStatusService) GetMasterStatusesByType(ctx context.Context, statusType string) ([]MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.GetMasterStatusesByType(ctx, statusType)
}

func (s *masterStatusService) GetMasterStatusByUUID(ctx context.Context, uuid string) (*MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.GetMasterStatusByUUID(ctx, uuid)
}

func (s *masterStatusService) UpdateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.UpdateMasterStatus(ctx, status)
}

func (s *masterStatusService) DeleteMasterStatus(ctx context.Context, uuid string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.DeleteMasterStatus(ctx, uuid)
}

func (s *masterStatusService) GetStatusByNameAndType(ctx context.Context, name, statusType string) (*MasterStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	return s.repo.GetStatusByNameAndType(ctx, name, statusType)
}
