package outbound

import (
	"bytes"
	"context"
	"errors"
)

func (s *service) GetCargoManifestByMAWBInfoUUID(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error) {
	return s.selfRepo.GetCargoManifestByMAWBInfoUUID(ctx, mawbInfoUUID)
}

func (s *service) CreateOrUpdateCargoManifest(ctx context.Context, manifest *CargoManifest) (*CargoManifest, error) {
	// In a real application, you might have more complex business logic here,
	// such as validation, authorization checks, etc.
	// For now, we just pass through to the repository.
	manifest.MAWBInfoUUID = ctx.Value("mawb_info_uuid").(string)
	return s.selfRepo.CreateOrUpdateCargoManifest(ctx, manifest)
}

func (s *service) ConfirmCargoManifest(ctx context.Context, mawbInfoUUID string) error {
	return s.selfRepo.UpdateCargoManifestStatus(ctx, mawbInfoUUID, "Confirmed")
}

func (s *service) RejectCargoManifest(ctx context.Context, mawbInfoUUID string) error {
	return s.selfRepo.UpdateCargoManifestStatus(ctx, mawbInfoUUID, "Rejected")
}

func (s *service) PrintCargoManifest(ctx context.Context, mawbInfoUUID string) (bytes.Buffer, error) {
	// PDF generation logic would go here.
	// For now, returning an error to indicate it's not implemented.
	var buffer bytes.Buffer
	return buffer, errors.New("PDF printing for Cargo Manifest is not implemented yet")
}
