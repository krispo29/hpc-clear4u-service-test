package mawbinfo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

type CargoManifestService interface {
	GetCargoManifest(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error)
	CreateOrUpdateCargoManifest(ctx context.Context, mawbInfoUUID string, req *CargoManifest) (*CargoManifest, error)
	UpdateCargoManifestStatus(ctx context.Context, mawbInfoUUID, status string) error
}

type cargoManifestService struct {
	mawbInfoRepo      Repository
	cargoManifestRepo CargoManifestRepository
	contextTimeout    time.Duration
	validate          *validator.Validate
}

func NewCargoManifestService(
	mawbInfoRepo Repository,
	cargoManifestRepo CargoManifestRepository,
	timeout time.Duration,
) CargoManifestService {
	return &cargoManifestService{
		mawbInfoRepo:      mawbInfoRepo,
		cargoManifestRepo: cargoManifestRepo,
		contextTimeout:    timeout,
		validate:          validator.New(),
	}
}

func (s *cargoManifestService) GetCargoManifest(ctx context.Context, mawbInfoUUID string) (*CargoManifest, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbInfoUUID) == "" {
		return nil, errors.New("mawb info uuid is required")
	}

	manifest, err := s.cargoManifestRepo.GetByMAWBInfoUUID(ctx, mawbInfoUUID)
	if err != nil {
		return nil, err
	}
	if manifest == nil {
		return nil, sql.ErrNoRows // Use standard error for not found
	}

	return manifest, nil
}

func (s *cargoManifestService) CreateOrUpdateCargoManifest(ctx context.Context, mawbInfoUUID string, req *CargoManifest) (*CargoManifest, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbInfoUUID) == "" {
		return nil, errors.New("mawb info uuid is required")
	}

	// 1. Validate that the MAWB Info record exists
	_, err := s.mawbInfoRepo.GetMawbInfo(ctx, mawbInfoUUID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows") {
			return nil, errors.New("mawb info not found")
		}
		return nil, err
	}

	// 2. Validate the request body
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	// 3. Set the UUID from the path and pass to repository
	req.MAWBInfoUUID = mawbInfoUUID
	req.Status = "Draft" // Always starts as Draft

	result, err := s.cargoManifestRepo.CreateOrUpdate(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *cargoManifestService) UpdateCargoManifestStatus(ctx context.Context, mawbInfoUUID, status string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbInfoUUID) == "" {
		return errors.New("mawb info uuid is required")
	}

	// Validate status
	validStatuses := map[string]bool{"Confirmed": true, "Rejected": true}
	if !validStatuses[status] {
		return errors.New("invalid status provided")
	}

	// 1. Validate that the MAWB Info record exists
	_, err := s.mawbInfoRepo.GetMawbInfo(ctx, mawbInfoUUID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows") {
			return errors.New("mawb info not found")
		}
		return err
	}

	// 2. Call repository to update status
	err = s.cargoManifestRepo.UpdateStatus(ctx, mawbInfoUUID, status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("cargo manifest not found for this mawb")
		}
		return err
	}

	return nil
}
