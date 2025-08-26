package outbound

import (
	"context"
	"fmt"
	"hpc-express-service/common"
	"hpc-express-service/setting"
)

type CargoManifestService interface {
	GetCargoManifestByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	GetCargoManifestByUUID(ctx context.Context, uuid string) (*CargoManifest, error)
	GetAllCargoManifest(ctx context.Context, startDate, endDate string) ([]CargoManifest, error)
	CreateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	UpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	UpdateCargoManifestStatus(ctx context.Context, mawbUUID, statusUUID string) error
}

type cargoManifestService struct {
	repo      CargoManifestRepository
	statusSvc setting.MasterStatusService
}

func NewCargoManifestService(repo CargoManifestRepository, statusSvc setting.MasterStatusService) CargoManifestService {
	return &cargoManifestService{repo: repo, statusSvc: statusSvc}
}

func (s *cargoManifestService) GetCargoManifestByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	// Add any business logic here if needed, e.g., permission checks.
	return s.repo.GetByMAWBUUID(ctx, mawbUUID)
}

func (s *cargoManifestService) GetCargoManifestByUUID(ctx context.Context, uuid string) (*CargoManifest, error) {
	return s.repo.GetByUUID(ctx, uuid)
}

func (s *cargoManifestService) GetAllCargoManifest(ctx context.Context, startDate, endDate string) ([]CargoManifest, error) {
	return s.repo.GetAll(ctx, startDate, endDate)
}

// setDefaultStatus sets the status of the manifest to the default 'Draft' status.
func (s *cargoManifestService) setDefaultStatus(ctx context.Context, manifest *CargoManifest) error {
	defaultStatus, err := s.statusSvc.GetDefaultStatusByType(ctx, "cargo_manifest")
	if err != nil {
		return fmt.Errorf("error getting default status: %w", err)
	}
	if defaultStatus == nil {
		return fmt.Errorf("no default status found for cargo_manifest")
	}
	manifest.StatusUUID = defaultStatus.UUID
	return nil
}

func (s *cargoManifestService) CreateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if cargo manifest already exists for this MAWB
	existing, err := s.repo.GetByMAWBUUID(txCtx, manifest.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("cargo manifest already exists for this MAWB")
	}

	if err := s.setDefaultStatus(txCtx, manifest); err != nil {
		return nil, err
	}

	result, err := s.repo.Create(txCtx, manifest)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *cargoManifestService) UpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if cargo manifest exists for this MAWB
	existing, err := s.repo.GetByMAWBUUID(txCtx, manifest.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("cargo manifest not found for this MAWB")
	}

	// Set the UUID from existing record for update
	manifest.UUID = existing.UUID

	// Always reset status to default (Draft) when updating
	if err := s.setDefaultStatus(txCtx, manifest); err != nil {
		return nil, err
	}
	result, err := s.repo.Update(txCtx, manifest)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *cargoManifestService) UpdateCargoManifestStatus(ctx context.Context, mawbUUID, statusUUID string) error {
	tx, txCtx, err := common.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the existing manifest
	manifest, err := s.repo.GetByMAWBUUID(txCtx, mawbUUID)
	if err != nil {
		return err
	}
	if manifest == nil {
		return fmt.Errorf("cargo manifest not found for this MAWB")
	}

	// Update the status
	manifest.StatusUUID = statusUUID
	if _, err = s.repo.Update(txCtx, manifest); err != nil {
		return err
	}

	return tx.Commit()
}
