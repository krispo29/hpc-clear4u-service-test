package setting

import (
	"context"
	"time"
)

type Service interface {
	// HS Code
	CreateHsCode(ctx context.Context, data *CreateHsCodeModel) (string, error)
	GetAllHsCode(ctx context.Context) ([]*GetHsCodeModel, error)
	GetHsCodeByUUID(ctx context.Context, uuid string) (*GetHsCodeModel, error)
	UpdateHsCode(ctx context.Context, data *UpdateHsCodeModel) error
	UpdateStatusHsCode(ctx context.Context, uuid string) error
}

type service struct {
	selfRepo       Repository
	contextTimeout time.Duration
}

func NewService(
	selfRepo Repository,
	timeout time.Duration,
) Service {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
	}
}

func (s *service) CreateHsCode(ctx context.Context, data *CreateHsCodeModel) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if uuid, err := s.selfRepo.CreateHsCode(ctx, data); err != nil {
		return "", err
	} else {
		return uuid, nil
	}
}

func (s *service) GetAllHsCode(ctx context.Context) ([]*GetHsCodeModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if result, err := s.selfRepo.GetAllHsCode(ctx); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (s *service) GetHsCodeByUUID(ctx context.Context, uuid string) (*GetHsCodeModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if result, err := s.selfRepo.GetHsCodeByUUID(ctx, uuid); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (s *service) UpdateHsCode(ctx context.Context, data *UpdateHsCodeModel) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.UpdateHsCode(ctx, data); err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateStatusHsCode(ctx context.Context, uuid string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.UpdateStatusHsCode(ctx, uuid); err != nil {
		return err
	}

	return nil
}
