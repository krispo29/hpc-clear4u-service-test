package dropdown

import (
	"context"
	"time"
)

type Service interface {
	GetServiceTypes(ctx context.Context) ([]DropdownItem, error)
	GetShippingTypes(ctx context.Context) ([]DropdownItem, error)
}

type DropdownItem struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

type service struct {
	repo           Repository
	contextTimeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:           repo,
		contextTimeout: timeout,
	}
}

func (s *service) GetServiceTypes(ctx context.Context) ([]DropdownItem, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Static data for service types
	serviceTypes := []DropdownItem{
		{Value: "cargo", Text: "Cargo"},
		{Value: "transit", Text: "Transit"},
	}

	return serviceTypes, nil
}

func (s *service) GetShippingTypes(ctx context.Context) ([]DropdownItem, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Static data for shipping types
	shippingTypes := []DropdownItem{
		{Value: "sea", Text: "Sea"},
		{Value: "air", Text: "Air"},
	}

	return shippingTypes, nil
}
