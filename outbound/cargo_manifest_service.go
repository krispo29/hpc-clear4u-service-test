package outbound

import (
	"context"
	"fmt"
)

type CargoManifestService interface {
	GetCargoManifestByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	CreateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	UpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
	UpdateCargoManifestStatus(ctx context.Context, mawbUUID, status string) error
}

type cargoManifestService struct {
	repo CargoManifestRepository
}

func NewCargoManifestService(repo CargoManifestRepository) CargoManifestService {
	return &cargoManifestService{repo: repo}
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
	return s.repo.Update(ctx, manifest)
}

func (s *cargoManifestService) UpdateCargoManifestStatus(ctx context.Context, mawbUUID, status string) error {
	// Get the existing manifest
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return err
	}
	if manifest == nil {
		return fmt.Errorf("cargo manifest not found for this MAWB")
	}

	// Update the status
	manifest.Status = status
	_, err = s.repo.Update(ctx, manifest)
	return err
}
