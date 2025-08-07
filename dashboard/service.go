package dashboard

import (
	"context"
	"time"
)

type Service interface {
	GetDashboardV1(ctx context.Context) (*DashboardV2Model, error)
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

func (s *service) GetDashboardV1(ctx context.Context) (*DashboardV2Model, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if resp, err := s.selfRepo.GetDashboardV1(ctx); err != nil {
		return nil, err
	} else {
		return resp, nil
	}

}
