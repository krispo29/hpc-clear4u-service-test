package customer

import (
	"context"
	"time"

	"hpc-express-service/constant"
)

type Service interface {
	GetAll(ctx context.Context) ([]*GetAllModel, error)
	GetAllDropdown(ctx context.Context) ([]*constant.DropdownModel, error)
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

func (s *service) GetAll(ctx context.Context) ([]*GetAllModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetAllDropdown(ctx context.Context) ([]*constant.DropdownModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetAllDropdown(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}
