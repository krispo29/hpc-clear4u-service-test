package outbound

import (
	"context"
	"fmt"
	"math"
	"strconv"
)

type DraftMAWBService interface {
	GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
	CreateOrUpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error)
	UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, status string) error
}

type draftMAWBService struct {
	repo DraftMAWBRepository
}

func NewDraftMAWBService(repo DraftMAWBRepository) DraftMAWBService {
	return &draftMAWBService{repo: repo}
}

func (s *draftMAWBService) GetDraftMAWBByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	return s.repo.GetByMAWBUUID(ctx, mawbUUID)
}

func (s *draftMAWBService) CreateOrUpdateDraftMAWB(ctx context.Context, draftMAWB *DraftMAWB) (*DraftMAWB, error) {
	// Perform calculations before saving
	for i := range draftMAWB.Items {
		totalVolume, chargeableWeight := s.calculateVolumeAndChargeableWeight(draftMAWB.Items[i].Dims, draftMAWB.Items[i].GrossWeight, draftMAWB.Items[i].KgLb)
		draftMAWB.Items[i].TotalVolume = totalVolume
		draftMAWB.Items[i].ChargeableWeight = chargeableWeight
	}
	s.calculateTotals(draftMAWB)

	return s.repo.CreateOrUpdate(ctx, draftMAWB)
}

func (s *draftMAWBService) UpdateDraftMAWBStatus(ctx context.Context, mawbUUID, status string) error {
	draft, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return err
	}
	if draft == nil {
		return nil // Or a not found error
	}
	return s.repo.UpdateStatus(ctx, draft.UUID, status)
}

func (s *draftMAWBService) calculateVolumeAndChargeableWeight(dims []DraftMAWBItemDim, grossWeight string, kgLb string) (string, string) {
	var totalVolume float64

	for _, dim := range dims {
		length, _ := strconv.ParseFloat(dim.Length, 64)
		width, _ := strconv.ParseFloat(dim.Width, 64)
		height, _ := strconv.ParseFloat(dim.Height, 64)
		count, _ := strconv.ParseFloat(dim.Count, 64)

		if length > 0 && width > 0 && height > 0 && count > 0 {
			volume := (length * width * height) / 1000000 // cm³ to m³
			totalVolume += volume * count
		}
	}

	volumetricWeight := totalVolume * 166.67
	weight, _ := strconv.ParseFloat(grossWeight, 64)

	if kgLb == "lb" {
		weight = weight * 0.453592 // Convert lb to kg
	}

	chargeableWeight := math.Max(weight, volumetricWeight)

	return fmt.Sprintf("%.3f", totalVolume), fmt.Sprintf("%.2f", chargeableWeight)
}

func (s *draftMAWBService) calculateTotals(draftMAWB *DraftMAWB) {
	var totalCharges float64
	for _, charge := range draftMAWB.Charges {
		totalCharges += charge.Value
	}

	var totalItemCharges float64
	for _, item := range draftMAWB.Items {
		chargeableWeight, _ := strconv.ParseFloat(item.ChargeableWeight, 64)
		item.Total = item.RateCharge * chargeableWeight
		totalItemCharges += item.Total
	}

	draftMAWB.TotalOtherChargesDueCarrier = totalCharges // Assuming item charges are separate

	draftMAWB.TotalPrepaid = draftMAWB.Prepaid +
		draftMAWB.ValuationCharge +
		draftMAWB.Tax +
		draftMAWB.TotalOtherChargesDueAgent +
		draftMAWB.TotalOtherChargesDueCarrier +
		totalItemCharges
}
