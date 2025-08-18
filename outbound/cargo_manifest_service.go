package outbound

import (
	"context"
	"fmt"
	"hpc-express-service/setting"
)

type CargoManifestService interface {
	GetCargoManifestByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
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

func (s *cargoManifestService) CreateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	// Add validation or transformation logic here.
	// For example, ensuring MAWBInfoUUID is present and valid.

	// Check if cargo manifest already exists for this MAWB
	existing, err := s.repo.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("cargo manifest already exists for this MAWB")
	}

	defaultStatus, err := s.statusSvc.GetDefaultStatusByType(ctx, "cargo_manifest")
	if err != nil {
		return nil, fmt.Errorf("error getting default status: %w", err)
	}
	if defaultStatus == nil {
		return nil, fmt.Errorf("no default status found for cargo_manifest")
	}
	manifest.StatusUUID = defaultStatus.UUID

	return s.repo.Create(ctx, manifest)
}

func (s *cargoManifestService) UpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	// Add validation or transformation logic here.
	// For example, ensuring MAWBInfoUUID is present and valid.

	// Check if cargo manifest exists for this MAWB
	existing, err := s.repo.GetByMAWBUUID(ctx, manifest.MAWBInfoUUID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("cargo manifest not found for this MAWB")
	}

	// Set the UUID from existing record for update
	manifest.UUID = existing.UUID

	// Per user request, always set status to default 'Draft' on update
	defaultStatus, err := s.statusSvc.GetDefaultStatusByType(ctx, "cargo_manifest")
	if err != nil {
		return nil, fmt.Errorf("error getting default status: %w", err)
	}
	if defaultStatus == nil {
		return nil, fmt.Errorf("no default status found for cargo_manifest")
	}
	manifest.StatusUUID = defaultStatus.UUID

	return s.repo.Update(ctx, manifest)
}

func (s *cargoManifestService) UpdateCargoManifestStatus(ctx context.Context, mawbUUID, statusUUID string) error {
	// Get the existing manifest
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return err
	}
	if manifest == nil {
		return fmt.Errorf("cargo manifest not found for this MAWB")
	}

	// Update the status
	manifest.StatusUUID = statusUUID
	_, err = s.repo.Update(ctx, manifest)
	return err
}
