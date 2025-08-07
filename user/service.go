package user

import (
	"context"
	"time"
)

type Service interface {
	Get(ctx context.Context, uuid string) (*GetModel, error)
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

func (s *service) Get(ctx context.Context, uuid string) (*GetModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if info, err := s.selfRepo.Get(ctx, uuid); err != nil {
		return nil, err
	} else {
		return info, nil
	}

}
