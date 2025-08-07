package dropdown

import (
	"context"
	"testing"
	"time"
)

func TestService_GetServiceTypes(t *testing.T) {
	repo := NewRepository()
	svc := NewService(repo, time.Second*30)

	ctx := context.Background()
	result, err := svc.GetServiceTypes(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 service types, got %d", len(result))
	}

	// Check cargo
	if result[0].Value != "cargo" || result[0].Text != "Cargo" {
		t.Errorf("Expected cargo service type, got %+v", result[0])
	}

	// Check transit
	if result[1].Value != "transit" || result[1].Text != "Transit" {
		t.Errorf("Expected transit service type, got %+v", result[1])
	}
}

func TestService_GetShippingTypes(t *testing.T) {
	repo := NewRepository()
	svc := NewService(repo, time.Second*30)

	ctx := context.Background()
	result, err := svc.GetShippingTypes(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 shipping types, got %d", len(result))
	}

	// Check sea
	if result[0].Value != "sea" || result[0].Text != "Sea" {
		t.Errorf("Expected sea shipping type, got %+v", result[0])
	}

	// Check air
	if result[1].Value != "air" || result[1].Text != "Air" {
		t.Errorf("Expected air shipping type, got %+v", result[1])
	}
}
