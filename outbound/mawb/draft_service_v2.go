package outbound

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// --- V2 Service Methods ---

func (s *service) GetDraftMAWBByMAWBInfoUUIDV2(ctx context.Context, mawbInfoUUID string) (*DraftMAWBV2, error) {
	return s.selfRepo.GetDraftMAWBByMAWBInfoUUIDV2(ctx, mawbInfoUUID)
}

func (s *service) CreateOrUpdateDraftMAWBV2(ctx context.Context, draft *DraftMAWBV2) (*DraftMAWBV2, error) {
	draft.MAWBInfoUUID = ctx.Value("mawb_info_uuid").(string)

	// Perform calculations as per requirements
	for _, item := range draft.Items {
		totalVolume, chargeableWeight := calculateVolumeAndChargeableWeight(item.Dims, item.GrossWeight, item.KgLb)
		item.TotalVolume = totalVolume
		item.ChargeableWeight = chargeableWeight
	}
	calculateTotals(draft)

	return s.selfRepo.CreateOrUpdateDraftMAWBV2(ctx, draft)
}

func (s *service) ConfirmDraftMAWBV2(ctx context.Context, mawbInfoUUID string) error {
	return s.selfRepo.UpdateDraftMAWBStatusV2(ctx, mawbInfoUUID, "Confirmed")
}

func (s *service) RejectDraftMAWBV2(ctx context.Context, mawbInfoUUID string) error {
	return s.selfRepo.UpdateDraftMAWBStatusV2(ctx, mawbInfoUUID, "Rejected")
}

func (s *service) PrintDraftMAWBV2(ctx context.Context, mawbInfoUUID string) (bytes.Buffer, error) {
	var buffer bytes.Buffer
	return buffer, errors.New("PDF printing for V2 Draft MAWB is not implemented yet")
}

// --- Calculation Functions from Requirements ---

func calculateVolumeAndChargeableWeight(dims []*DraftMAWBItemDimV2, grossWeight string, kgLb string) (string, string) {
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

func calculateTotals(draft *DraftMAWBV2) {
	var totalCharges float64
	for _, charge := range draft.Charges {
		totalCharges += charge.Value
	}

	var totalItemRateCharge float64
	for _, item := range draft.Items {
		totalItemRateCharge += item.RateCharge
	}
    // The requirement doc is a bit ambiguous here. It says "total other charges due carrier"
    // is the sum of charges and item.RateCharge. Let's assume that's correct.
	draft.TotalOtherChargesDueCarrier = totalCharges + totalItemRateCharge

	// Calculate total prepaid
	draft.TotalPrepaid = draft.Prepaid +
		draft.ValuationCharge +
		draft.Tax +
		draft.TotalOtherChargesDueAgent +
		draft.TotalOtherChargesDueCarrier
}
