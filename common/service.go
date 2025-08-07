package common

import (
	"context"
	"time"
)

type Service interface {
	GetAllExchangeRates(ctx context.Context) ([]*GetExchangeRateModel, error)
	GetAllConvertTemplates(ctx context.Context, category string) ([]*GetAllConvertTemplateModel, error)
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

func (s *service) GetAllExchangeRates(ctx context.Context) ([]*GetExchangeRateModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetAllExchangeRates(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetAllConvertTemplates(ctx context.Context, category string) ([]*GetAllConvertTemplateModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetAllConvertTemplates(ctx, category)
	if err != nil {
		return nil, err
	}

	return result, nil
}
