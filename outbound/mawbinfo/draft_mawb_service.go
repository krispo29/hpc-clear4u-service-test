package mawbinfo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

type DraftMAWBService interface {
	GetDraftMAWB(ctx context.Context, mawbInfoUUID string) (*DraftMAWB, error)
	CreateOrUpdateDraftMAWB(ctx context.Context, mawbInfoUUID string, req *DraftMAWB) (*DraftMAWB, error)
	UpdateDraftMAWBStatus(ctx context.Context, mawbInfoUUID, status string) error
}

type draftMAWBService struct {
	mawbInfoRepo  Repository
	draftMAWBRepo DraftMAWBRepository
	contextTimeout time.Duration
	validate      *validator.Validate
}

func NewDraftMAWBService(
	mawbInfoRepo Repository,
	draftMAWBRepo DraftMAWBRepository,
	timeout time.Duration,
) DraftMAWBService {
	return &draftMAWBService{
		mawbInfoRepo:   mawbInfoRepo,
		draftMAWBRepo:  draftMAWBRepo,
		contextTimeout: timeout,
		validate:       validator.New(),
	}
}

func (s *draftMAWBService) GetDraftMAWB(ctx context.Context, mawbInfoUUID string) (*DraftMAWB, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbInfoUUID) == "" {
		return nil, errors.New("mawb info uuid is required")
	}

	draft, err := s.draftMAWBRepo.GetByMAWBInfoUUID(ctx, mawbInfoUUID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, sql.ErrNoRows
	}

	return draft, nil
}

func (s *draftMAWBService) CreateOrUpdateDraftMAWB(ctx context.Context, mawbInfoUUID string, req *DraftMAWB) (*DraftMAWB, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbInfoUUID) == "" {
		return nil, errors.New("mawb info uuid is required")
	}

	// 1. Validate that the MAWB Info record exists
	_, err := s.mawbInfoRepo.GetMawbInfo(ctx, mawbInfoUUID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows") {
			return nil, errors.New("mawb info not found")
		}
		return nil, err
	}

	// 2. Validate the request body
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	// 3. Perform calculations
	s.calculateAll(req)

	// 4. Set the UUID from the path and pass to repository
	req.MAWBInfoUUID = mawbInfoUUID
	req.Status = "Draft"

	result, err := s.draftMAWBRepo.CreateOrUpdate(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *draftMAWBService) UpdateDraftMAWBStatus(ctx context.Context, mawbInfoUUID, status string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if strings.TrimSpace(mawbInfoUUID) == "" {
		return errors.New("mawb info uuid is required")
	}

	validStatuses := map[string]bool{"Confirmed": true, "Rejected": true}
	if !validStatuses[status] {
		return errors.New("invalid status provided")
	}

	_, err := s.mawbInfoRepo.GetMawbInfo(ctx, mawbInfoUUID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows") {
			return errors.New("mawb info not found")
		}
		return err
	}

	err = s.draftMAWBRepo.UpdateStatus(ctx, mawbInfoUUID, status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("draft mawb not found for this mawb")
		}
		return err
	}

	return nil
}

func (s *draftMAWBService) calculateAll(draft *DraftMAWB) {
	for i := range draft.Items {
		item := &draft.Items[i]
		totalVolume, chargeableWeight := s.calculateVolumeAndChargeableWeight(item.Dims, item.GrossWeight, item.KgLb)
		item.TotalVolume = totalVolume
		item.ChargeableWeight = chargeableWeight

		// Calculate total for the item
		rateCharge, _ := item.RateCharge, 10
		chargeableWeightFloat, _ := strconv.ParseFloat(chargeableWeight, 64)
		item.Total = rateCharge * chargeableWeightFloat
	}

	s.calculateTotals(draft)
}

func (s *draftMAWBService) calculateVolumeAndChargeableWeight(dims []DraftMAWBItemDim, grossWeightStr string, kgLb string) (string, string) {
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
	weight, _ := strconv.ParseFloat(grossWeightStr, 64)

	if kgLb == "lb" {
		weight = weight * 0.453592 // Convert lb to kg
	}

	chargeableWeight := math.Max(weight, volumetricWeight)

	return fmt.Sprintf("%.3f", totalVolume), fmt.Sprintf("%.2f", chargeableWeight)
}

func (s *draftMAWBService) calculateTotals(draft *DraftMAWB) {
	var totalOtherChargesDueCarrier float64
	for _, charge := range draft.Charges {
		totalOtherChargesDueCarrier += charge.Value
	}

	var totalItemCharges float64
	for i := range draft.Items {
		totalItemCharges += draft.Items[i].Total
	}

	// According to requirement, this seems to be the sum of item totals and explicit charges
	draft.TotalOtherChargesDueCarrier = totalItemCharges + totalOtherChargesDueCarrier

	draft.TotalPrepaid = draft.Prepaid +
		draft.ValuationCharge +
		draft.Tax +
		draft.TotalOtherChargesDueAgent +
		draft.TotalOtherChargesDueCarrier
}
