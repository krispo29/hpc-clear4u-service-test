package outbound

import (
	"context"
)

type CargoManifestService interface {
	GetCargoManifestByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
	CreateOrUpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error)
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

func (s *cargoManifestService) CreateOrUpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	// Add validation or transformation logic here.
	// For example, ensuring MAWBInfoUUID is present and valid.
	return s.repo.CreateOrUpdate(ctx, manifest)
}

func (s *cargoManifestService) UpdateCargoManifestStatus(ctx context.Context, mawbUUID, status string) error {
	// This would typically involve getting the manifest, changing status, and saving.
	// For now, let's assume the repository will have a dedicated method for this.
	// As this method is not in the repo, I will add it later if needed.
	// For now, I will assume the POST /confirm and /reject endpoints will call CreateOrUpdate with a specific status.
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return err
	}
	if manifest == nil {
		return nil // Or an error indicating not found
	}
	manifest.Status = status
	_, err = s.repo.CreateOrUpdate(ctx, manifest)
	return err
}
