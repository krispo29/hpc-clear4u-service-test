package outbound

import (
	"context"
	"fmt"
	"hpc-express-service/common"
	"hpc-express-service/setting"
)

type WeightSlipService interface {
	GetWeightSlipByMAWBUUID(ctx context.Context, mawbUUID string) (*WeightSlip, error)
	GetWeightSlipByUUID(ctx context.Context, wsUUID string) (*WeightSlip, error)
	CreateWeightSlip(ctx context.Context, ws *WeightSlip) (*WeightSlip, error)
	UpdateWeightSlip(ctx context.Context, ws *WeightSlip) (*WeightSlip, error)
	UpdateWeightSlipStatus(ctx context.Context, mawbUUID, statusUUID string) error
}

type weightSlipService struct {
	repo      WeightSlipRepository
	statusSvc setting.MasterStatusService
}

func NewWeightSlipService(repo WeightSlipRepository, statusSvc setting.MasterStatusService) WeightSlipService {
	return &weightSlipService{repo: repo, statusSvc: statusSvc}
}

func (s *weightSlipService) GetWeightSlipByMAWBUUID(ctx context.Context, mawbUUID string) (*WeightSlip, error) {
	return s.repo.GetByMAWBUUID(ctx, mawbUUID)
}

// GetWeightSlipByUUID retrieves a weight slip by its own UUID.
func (s *weightSlipService) GetWeightSlipByUUID(ctx context.Context, wsUUID string) (*WeightSlip, error) {
	return s.repo.GetByUUID(ctx, wsUUID)
}

func (s *weightSlipService) setDefaultStatus(ctx context.Context, ws *WeightSlip) error {
	status, err := s.statusSvc.GetDefaultStatusByType(ctx, "weight_slip")
	if err != nil {
		return fmt.Errorf("error getting default status: %w", err)
	}
	if status == nil {
		return fmt.Errorf("no default status found for weight_slip")
	}
	ws.StatusUUID = status.UUID
	return nil
}

func (s *weightSlipService) CreateWeightSlip(ctx context.Context, ws *WeightSlip) (*WeightSlip, error) {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := s.repo.GetByMAWBUUID(txCtx, ws.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("weight slip already exists for this MAWB")
	}

	if err := s.setDefaultStatus(txCtx, ws); err != nil {
		return nil, err
	}

	result, err := s.repo.Create(txCtx, ws)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *weightSlipService) UpdateWeightSlip(ctx context.Context, ws *WeightSlip) (*WeightSlip, error) {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := s.repo.GetByMAWBUUID(txCtx, ws.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("weight slip not found for this MAWB")
	}

	ws.UUID = existing.UUID

	if err := s.setDefaultStatus(txCtx, ws); err != nil {
		return nil, err
	}

	result, err := s.repo.Update(txCtx, ws)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *weightSlipService) UpdateWeightSlipStatus(ctx context.Context, mawbUUID, statusUUID string) error {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ws, err := s.repo.GetByMAWBUUID(txCtx, mawbUUID)
	if err != nil {
		return err
	}
	if ws == nil {
		return fmt.Errorf("weight slip not found for this MAWB")
	}

	ws.StatusUUID = statusUUID
	if _, err = s.repo.Update(txCtx, ws); err != nil {
		return err
	}

	return tx.Commit()
}
