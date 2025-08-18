package dropdown

import (
	"context"
	"time"
)

type Service interface {
	GetServiceTypes(ctx context.Context) ([]DropdownItem, error)
	GetShippingTypes(ctx context.Context) ([]DropdownItem, error)
	GetAirlineLogos(ctx context.Context) ([]AirlineLogoItem, error)
	GetMasterStatusesByType(ctx context.Context, statusType string) ([]DropdownItem, error)
}

type DropdownItem struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

type AirlineLogoItem struct {
	Value string `json:"value"`
	Image string `json:"image"`
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

func (s *service) GetAirlineLogos(ctx context.Context) ([]AirlineLogoItem, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Get data from database
	airlines, err := s.repo.GetAirlineLogos(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	airlineLogos := make([]AirlineLogoItem, len(airlines))
	for i, airline := range airlines {
		airlineLogos[i] = AirlineLogoItem{
			Value: airline.UUID,
			Image: airline.LogoURL,
			Text:  airline.Name,
		}
	}

	return airlineLogos, nil
}

func (s *service) GetMasterStatusesByType(ctx context.Context, statusType string) ([]DropdownItem, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	statuses, err := s.repo.GetMasterStatusesByType(ctx, statusType)
	if err != nil {
		return nil, err
	}

	dropdownItems := make([]DropdownItem, len(statuses))
	for i, status := range statuses {
		dropdownItems[i] = DropdownItem{
			Value: status.UUID,
			Text:  status.Name,
		}
	}

	return dropdownItems, nil
}
