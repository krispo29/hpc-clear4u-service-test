package draft_mawb

import (
	"context"
	"errors"
	"fmt"
	"hpc-express-service/common"
	"hpc-express-service/utils"
	"math"
	"strconv"
	"strings"
	"time"

	customerrors "hpc-express-service/errors"
)

// PDFGenerator interface for generating PDF documents
type PDFGenerator interface {
	GenerateDraftMAWBPDF(draftMAWB *DraftMAWB) ([]byte, error)
}

// NewPDFGenerator creates a new PDF generator instance
var NewPDFGenerator func() (PDFGenerator, error)

// Custom error types for better error handling
var (
	ErrMAWBInfoNotFound      = errors.New("MAWB Info not found")
	ErrDraftMAWBNotFound     = errors.New("draft MAWB not found")
	ErrInvalidMAWBUUID       = errors.New("invalid MAWB Info UUID")
	ErrInvalidRequestData    = errors.New("invalid request data")
	ErrBusinessRuleViolation = errors.New("business rule violation")
	ErrPDFGenerationFailed   = errors.New("PDF generation failed")
	ErrCalculationFailed     = errors.New("calculation failed")
)

// Service interface defines the contract for draft MAWB business logic
type Service interface {
	GetDraftMAWB(ctx context.Context, mawbUUID string) (*DraftMAWBResponse, error)
	CreateOrUpdateDraftMAWB(ctx context.Context, mawbUUID string, req *DraftMAWBRequest) (*DraftMAWBResponse, error)
	ConfirmDraftMAWB(ctx context.Context, mawbUUID string) error
	RejectDraftMAWB(ctx context.Context, mawbUUID string) error
	GenerateDraftMAWBPDF(ctx context.Context, mawbUUID string) ([]byte, error)
}

type service struct {
	repo             Repository
	contextTimeout   time.Duration
	calculationCache *common.CalculationCache
}

// NewService creates a new draft MAWB service instance with caching
func NewService(repo Repository, timeout time.Duration) Service {
	// Create cache for calculations
	memoryCache := common.NewMemoryCache(500)         // Cache up to 500 calculation results
	memoryCache.StartCleanupRoutine(10 * time.Minute) // Cleanup every 10 minutes

	calculationCache := common.NewCalculationCache(memoryCache, 15*time.Minute) // 15 min TTL for calculations
	return &service{
		repo:             repo,
		contextTimeout:   timeout,
		calculationCache: calculationCache,
	}
}

// GetDraftMAWB retrieves a draft MAWB by MAWB Info UUID
func (s *service) GetDraftMAWB(ctx context.Context, mawbUUID string) (*DraftMAWBResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate input
	if strings.TrimSpace(mawbUUID) == "" {
		return nil, fmt.Errorf("%w: MAWB Info UUID is required", ErrInvalidMAWBUUID)
	}

	// Validate MAWB Info exists
	if err := s.repo.ValidateMAWBExists(ctx, mawbUUID); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMAWBInfoNotFound, mawbUUID)
	}

	// Get draft MAWB from repository
	draftMAWB, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("%w for MAWB: %s", ErrDraftMAWBNotFound, mawbUUID)
		}
		return nil, fmt.Errorf("failed to retrieve draft MAWB: %w", err)
	}

	// Convert to response model
	response := s.convertToResponse(draftMAWB)
	return response, nil
}

// CreateOrUpdateDraftMAWB creates a new draft MAWB or updates an existing one
func (s *service) CreateOrUpdateDraftMAWB(ctx context.Context, mawbUUID string, req *DraftMAWBRequest) (*DraftMAWBResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate input
	if strings.TrimSpace(mawbUUID) == "" {
		return nil, fmt.Errorf("%w: MAWB Info UUID is required", ErrInvalidMAWBUUID)
	}

	if req == nil {
		return nil, fmt.Errorf("%w: request data cannot be nil", ErrInvalidRequestData)
	}

	// Validate MAWB Info exists
	if err := s.repo.ValidateMAWBExists(ctx, mawbUUID); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMAWBInfoNotFound, mawbUUID)
	}

	// Sanitize input
	SanitizeDraftMAWBRequest(req)

	// Validate request
	if validationErrors := ValidateDraftMAWBRequest(req); len(validationErrors) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRequestData, s.formatValidationErrors(validationErrors).Error())
	}

	// Additional business rule validation
	if err := s.validateBusinessRules(req); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBusinessRuleViolation, err.Error())
	}

	// Convert request to domain model with calculations
	draftMAWB, err := s.convertRequestToModel(mawbUUID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request to model: %w", err)
	}

	// Perform calculations
	if err := s.performCalculations(draftMAWB); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCalculationFailed, err.Error())
	}

	// Save to repository
	if err := s.repo.CreateOrUpdate(ctx, draftMAWB); err != nil {
		return nil, fmt.Errorf("failed to save draft MAWB: %w", err)
	}

	// Get the saved draft MAWB to return complete data
	savedDraftMAWB, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve saved draft MAWB: %w", err)
	}

	// Convert to response model
	response := s.convertToResponse(savedDraftMAWB)
	return response, nil
}

// ConfirmDraftMAWB updates the draft MAWB status to confirmed
func (s *service) ConfirmDraftMAWB(ctx context.Context, mawbUUID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate input
	if strings.TrimSpace(mawbUUID) == "" {
		return fmt.Errorf("%w: MAWB Info UUID is required", ErrInvalidMAWBUUID)
	}

	// Validate MAWB Info exists
	if err := s.repo.ValidateMAWBExists(ctx, mawbUUID); err != nil {
		return fmt.Errorf("%w: %s", ErrMAWBInfoNotFound, mawbUUID)
	}

	// Get existing draft MAWB to get its UUID
	draftMAWB, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return fmt.Errorf("%w for MAWB: %s", ErrDraftMAWBNotFound, mawbUUID)
		}
		return fmt.Errorf("failed to retrieve draft MAWB: %w", err)
	}

	// Validate that the draft MAWB can be confirmed (business rule)
	if draftMAWB.Status == StatusConfirmed {
		return fmt.Errorf("%w: draft MAWB is already confirmed", ErrBusinessRuleViolation)
	}

	// Update status to confirmed
	if err := s.repo.UpdateStatus(ctx, draftMAWB.UUID, StatusConfirmed); err != nil {
		return fmt.Errorf("failed to confirm draft MAWB: %w", err)
	}

	return nil
}

// RejectDraftMAWB updates the draft MAWB status to rejected
func (s *service) RejectDraftMAWB(ctx context.Context, mawbUUID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate input
	if strings.TrimSpace(mawbUUID) == "" {
		return fmt.Errorf("%w: MAWB Info UUID is required", ErrInvalidMAWBUUID)
	}

	// Validate MAWB Info exists
	if err := s.repo.ValidateMAWBExists(ctx, mawbUUID); err != nil {
		return fmt.Errorf("%w: %s", ErrMAWBInfoNotFound, mawbUUID)
	}

	// Get existing draft MAWB to get its UUID
	draftMAWB, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return fmt.Errorf("%w for MAWB: %s", ErrDraftMAWBNotFound, mawbUUID)
		}
		return fmt.Errorf("failed to retrieve draft MAWB: %w", err)
	}

	// Validate that the draft MAWB can be rejected (business rule)
	if draftMAWB.Status == StatusRejected {
		return fmt.Errorf("%w: draft MAWB is already rejected", ErrBusinessRuleViolation)
	}

	// Update status to rejected
	if err := s.repo.UpdateStatus(ctx, draftMAWB.UUID, StatusRejected); err != nil {
		return fmt.Errorf("failed to reject draft MAWB: %w", err)
	}

	return nil
}

// GenerateDraftMAWBPDF generates a PDF document for the draft MAWB
func (s *service) GenerateDraftMAWBPDF(ctx context.Context, mawbUUID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Validate input
	if strings.TrimSpace(mawbUUID) == "" {
		return nil, fmt.Errorf("%w: MAWB Info UUID is required", ErrInvalidMAWBUUID)
	}

	// Validate MAWB Info exists
	if err := s.repo.ValidateMAWBExists(ctx, mawbUUID); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMAWBInfoNotFound, mawbUUID)
	}

	// Get draft MAWB from repository
	draftMAWB, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("%w for MAWB: %s", ErrDraftMAWBNotFound, mawbUUID)
		}
		return nil, fmt.Errorf("failed to retrieve draft MAWB: %w", err)
	}

	// Generate PDF using the draft MAWB data
	pdfContent, err := s.generatePDFContent(draftMAWB)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPDFGenerationFailed, err.Error())
	}

	return pdfContent, nil
}

// performCalculations performs all necessary calculations for the draft MAWB
func (s *service) performCalculations(draftMAWB *DraftMAWB) error {
	var totalPieces int
	var totalGrossWeight float64
	var totalChargeableWeight float64
	var totalRateCharge float64
	var totalAmount float64

	// Calculate values for each item
	for i := range draftMAWB.Items {
		item := &draftMAWB.Items[i]

		// Calculate volumetric weight for the item
		volumetricWeight, err := s.calculateVolumetricWeight(item.Dims)
		if err != nil {
			return fmt.Errorf("failed to calculate volumetric weight for item %d: %w", i+1, err)
		}
		item.TotalVolume = volumetricWeight

		// Parse gross weight
		grossWeight, err := s.parseWeight(item.GrossWeight, item.KgLb)
		if err != nil {
			return fmt.Errorf("failed to parse gross weight for item %d: %w", i+1, err)
		}

		// Calculate chargeable weight
		chargeableWeight := s.calculateChargeableWeight(grossWeight, volumetricWeight)
		item.ChargeableWeight = chargeableWeight

		// Calculate item total (rate × chargeable weight)
		itemTotal := item.RateCharge * chargeableWeight
		item.Total = itemTotal

		// Parse pieces count
		pieces, err := s.parsePieces(item.PiecesRCP)
		if err != nil {
			return fmt.Errorf("failed to parse pieces for item %d: %w", i+1, err)
		}

		// Accumulate totals
		totalPieces += pieces
		totalGrossWeight += grossWeight
		totalChargeableWeight += chargeableWeight
		totalRateCharge += item.RateCharge
		totalAmount += itemTotal
	}

	// Calculate financial totals from charges
	chargesTotal := s.calculateChargesTotal(draftMAWB.Charges)
	totalAmount += chargesTotal

	// Update draft MAWB totals
	draftMAWB.TotalNoOfPieces = totalPieces
	draftMAWB.TotalGrossWeight = totalGrossWeight
	draftMAWB.TotalChargeableWeight = totalChargeableWeight
	draftMAWB.TotalRateCharge = totalRateCharge
	draftMAWB.TotalAmount = totalAmount

	return nil
}

// calculateVolumetricWeight calculates volumetric weight using (L×W×H)/1,000,000 × count formula with caching
func (s *service) calculateVolumetricWeight(dims []DraftMAWBItemDim) (float64, error) {
	ctx := context.Background()

	// Try to get from cache first
	if cachedWeight, exists := s.calculationCache.GetVolumetricWeight(ctx, dims); exists {
		return cachedWeight, nil
	}

	var totalVolume float64

	for _, dim := range dims {
		// Parse dimensions
		length, err := strconv.ParseFloat(dim.Length, 64)
		if err != nil || length <= 0 {
			return 0, fmt.Errorf("invalid length: %s", dim.Length)
		}

		width, err := strconv.ParseFloat(dim.Width, 64)
		if err != nil || width <= 0 {
			return 0, fmt.Errorf("invalid width: %s", dim.Width)
		}

		height, err := strconv.ParseFloat(dim.Height, 64)
		if err != nil || height <= 0 {
			return 0, fmt.Errorf("invalid height: %s", dim.Height)
		}

		count, err := strconv.Atoi(dim.Count)
		if err != nil || count <= 0 {
			return 0, fmt.Errorf("invalid count: %s", dim.Count)
		}

		// Calculate volume for this dimension: (L×W×H)/1,000,000 × count
		volume := (length * width * height / 1000000) * float64(count)
		totalVolume += volume
	}

	// Cache the result
	if err := s.calculationCache.SetVolumetricWeight(ctx, dims, totalVolume); err != nil {
		// Log error but don't fail the calculation
		fmt.Printf("Warning: Failed to cache volumetric weight: %v\n", err)
	}

	return totalVolume, nil
}

// calculateChargeableWeight calculates chargeable weight (max of actual weight and volumetric weight × 166.67) with caching
func (s *service) calculateChargeableWeight(actualWeight, volumetricWeight float64) float64 {
	ctx := context.Background()

	// Create cache key parameters
	params := map[string]float64{
		"actual_weight":     actualWeight,
		"volumetric_weight": volumetricWeight,
	}

	// Try to get from cache first
	if cachedWeight, exists := s.calculationCache.GetChargeableWeight(ctx, params); exists {
		return cachedWeight
	}

	// Convert volumetric weight to chargeable weight using factor 166.67
	volumetricChargeableWeight := volumetricWeight * 166.67

	// Return the maximum of actual weight and volumetric chargeable weight
	chargeableWeight := math.Max(actualWeight, volumetricChargeableWeight)

	// Cache the result
	if err := s.calculationCache.SetChargeableWeight(ctx, params, chargeableWeight); err != nil {
		// Log error but don't fail the calculation
		fmt.Printf("Warning: Failed to cache chargeable weight: %v\n", err)
	}

	return chargeableWeight
}

// parseWeight parses weight string and converts to kilograms if needed
func (s *service) parseWeight(weightStr, unit string) (float64, error) {
	if strings.TrimSpace(weightStr) == "" {
		return 0, nil
	}

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid weight format: %s", weightStr)
	}

	// Convert pounds to kilograms if needed
	if strings.ToLower(strings.TrimSpace(unit)) == "lb" {
		weight = weight * 0.453592 // Convert pounds to kilograms
	}

	return weight, nil
}

// parsePieces parses pieces string to integer
func (s *service) parsePieces(piecesStr string) (int, error) {
	if strings.TrimSpace(piecesStr) == "" {
		return 0, nil
	}

	pieces, err := strconv.Atoi(piecesStr)
	if err != nil {
		return 0, fmt.Errorf("invalid pieces format: %s", piecesStr)
	}

	return pieces, nil
}

// calculateChargesTotal calculates the total of all charges with caching
func (s *service) calculateChargesTotal(charges []DraftMAWBCharge) float64 {
	ctx := context.Background()

	// Try to get from cache first
	if cachedTotal, exists := s.calculationCache.GetFinancialTotals(ctx, charges); exists {
		return cachedTotal
	}

	var total float64
	for _, charge := range charges {
		total += charge.Value
	}

	// Cache the result
	if err := s.calculationCache.SetFinancialTotals(ctx, charges, total); err != nil {
		// Log error but don't fail the calculation
		fmt.Printf("Warning: Failed to cache financial totals: %v\n", err)
	}

	return total
}

// convertRequestToModel converts a request model to domain model
func (s *service) convertRequestToModel(mawbUUID string, req *DraftMAWBRequest) (*DraftMAWB, error) {
	draftMAWB := &DraftMAWB{
		MAWBInfoUUID:                   mawbUUID,
		CustomerUUID:                   req.CustomerUUID,
		AirlineLogo:                    req.AirlineLogo,
		AirlineName:                    req.AirlineName,
		MAWB:                           req.MAWB,
		HAWB:                           req.HAWB,
		ShipperNameAndAddress:          req.ShipperNameAndAddress,
		ConsigneeNameAndAddress:        req.ConsigneeNameAndAddress,
		IssuingCarrierAgentNameAndCity: req.IssuingCarrierAgentNameAndCity,
		AccountingInformation:          req.AccountingInformation,
		AgentIATACode:                  req.AgentIATACode,
		AccountNo:                      req.AccountNo,
		AirportOfDeparture:             req.AirportOfDeparture,
		ReferenceNumber:                req.ReferenceNumber,
		To1:                            req.To1,
		ByFirstCarrier:                 req.ByFirstCarrier,
		To2:                            req.To2,
		By2:                            req.By2,
		To3:                            req.To3,
		By3:                            req.By3,
		Currency:                       req.Currency,
		ChgsCode:                       req.ChgsCode,
		WtValPPD:                       req.WtValPPD,
		WtValColl:                      req.WtValColl,
		OtherPPD:                       req.OtherPPD,
		OtherColl:                      req.OtherColl,
		DeclaredValueCarriage:          req.DeclaredValueCarriage,
		DeclaredValueCustoms:           req.DeclaredValueCustoms,
		AirportOfDestination:           req.AirportOfDestination,
		FlightNo:                       req.FlightNo,
		InsuranceAmount:                req.InsuranceAmount,
		HandlingInformation:            req.HandlingInformation,
		SCI:                            req.SCI,
		ShipperCertifiesText:           req.ShipperCertifiesText,
		ExecutedAtPlace:                req.ExecutedAtPlace,
		SignatureOfShipper:             req.SignatureOfShipper,
		SignatureOfIssuingCarrier:      req.SignatureOfIssuingCarrier,
		Items:                          make([]DraftMAWBItem, len(req.Items)),
		Charges:                        make([]DraftMAWBCharge, len(req.Charges)),
	}

	// Parse date fields
	if req.FlightDate != "" {
		if flightDate, err := time.Parse("2006-01-02", req.FlightDate); err == nil {
			draftMAWB.FlightDate = &flightDate
		}
	}

	if req.ExecutedOnDate != "" {
		if executedDate, err := time.Parse("2006-01-02", req.ExecutedOnDate); err == nil {
			draftMAWB.ExecutedOnDate = &executedDate
		}
	}

	// Convert items
	for i, itemReq := range req.Items {
		draftMAWB.Items[i] = DraftMAWBItem{
			PiecesRCP:         itemReq.PiecesRCP,
			GrossWeight:       itemReq.GrossWeight,
			KgLb:              itemReq.KgLb,
			RateClass:         itemReq.RateClass,
			RateCharge:        itemReq.RateCharge,
			NatureAndQuantity: itemReq.NatureAndQuantity,
			Dims:              make([]DraftMAWBItemDim, len(itemReq.Dims)),
		}

		// Convert dimensions
		for j, dimReq := range itemReq.Dims {
			draftMAWB.Items[i].Dims[j] = DraftMAWBItemDim{
				Length: dimReq.Length,
				Width:  dimReq.Width,
				Height: dimReq.Height,
				Count:  dimReq.Count,
			}
		}
	}

	// Convert charges
	for i, chargeReq := range req.Charges {
		draftMAWB.Charges[i] = DraftMAWBCharge{
			Key:   chargeReq.Key,
			Value: chargeReq.Value,
		}
	}

	return draftMAWB, nil
}

// convertToResponse converts a domain model to response model
func (s *service) convertToResponse(draftMAWB *DraftMAWB) *DraftMAWBResponse {
	response := &DraftMAWBResponse{
		UUID:                           draftMAWB.UUID,
		MAWBInfoUUID:                   draftMAWB.MAWBInfoUUID,
		CustomerUUID:                   draftMAWB.CustomerUUID,
		AirlineLogo:                    draftMAWB.AirlineLogo,
		AirlineName:                    draftMAWB.AirlineName,
		MAWB:                           draftMAWB.MAWB,
		HAWB:                           draftMAWB.HAWB,
		ShipperNameAndAddress:          draftMAWB.ShipperNameAndAddress,
		ConsigneeNameAndAddress:        draftMAWB.ConsigneeNameAndAddress,
		IssuingCarrierAgentNameAndCity: draftMAWB.IssuingCarrierAgentNameAndCity,
		AccountingInformation:          draftMAWB.AccountingInformation,
		AgentIATACode:                  draftMAWB.AgentIATACode,
		AccountNo:                      draftMAWB.AccountNo,
		AirportOfDeparture:             draftMAWB.AirportOfDeparture,
		ReferenceNumber:                draftMAWB.ReferenceNumber,
		To1:                            draftMAWB.To1,
		ByFirstCarrier:                 draftMAWB.ByFirstCarrier,
		To2:                            draftMAWB.To2,
		By2:                            draftMAWB.By2,
		To3:                            draftMAWB.To3,
		By3:                            draftMAWB.By3,
		Currency:                       draftMAWB.Currency,
		ChgsCode:                       draftMAWB.ChgsCode,
		WtValPPD:                       draftMAWB.WtValPPD,
		WtValColl:                      draftMAWB.WtValColl,
		OtherPPD:                       draftMAWB.OtherPPD,
		OtherColl:                      draftMAWB.OtherColl,
		DeclaredValueCarriage:          draftMAWB.DeclaredValueCarriage,
		DeclaredValueCustoms:           draftMAWB.DeclaredValueCustoms,
		AirportOfDestination:           draftMAWB.AirportOfDestination,
		FlightNo:                       draftMAWB.FlightNo,
		InsuranceAmount:                draftMAWB.InsuranceAmount,
		HandlingInformation:            draftMAWB.HandlingInformation,
		SCI:                            draftMAWB.SCI,
		TotalNoOfPieces:                draftMAWB.TotalNoOfPieces,
		TotalGrossWeight:               draftMAWB.TotalGrossWeight,
		TotalKgLb:                      draftMAWB.TotalKgLb,
		TotalRateClass:                 draftMAWB.TotalRateClass,
		TotalChargeableWeight:          draftMAWB.TotalChargeableWeight,
		TotalRateCharge:                draftMAWB.TotalRateCharge,
		TotalAmount:                    draftMAWB.TotalAmount,
		ShipperCertifiesText:           draftMAWB.ShipperCertifiesText,
		ExecutedAtPlace:                draftMAWB.ExecutedAtPlace,
		SignatureOfShipper:             draftMAWB.SignatureOfShipper,
		SignatureOfIssuingCarrier:      draftMAWB.SignatureOfIssuingCarrier,
		Status:                         draftMAWB.Status,
		Items:                          draftMAWB.Items,
		Charges:                        draftMAWB.Charges,
		CreatedAt:                      draftMAWB.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                      draftMAWB.UpdatedAt.Format(time.RFC3339),
	}

	// Format date fields
	if draftMAWB.FlightDate != nil {
		response.FlightDate = draftMAWB.FlightDate.Format("2006-01-02")
	}

	if draftMAWB.ExecutedOnDate != nil {
		response.ExecutedOnDate = draftMAWB.ExecutedOnDate.Format("2006-01-02")
	}

	return response
}

// validateBusinessRules performs additional business logic validation
func (s *service) validateBusinessRules(req *DraftMAWBRequest) error {
	// Validate maximum number of items
	const maxItems = 50
	if len(req.Items) > maxItems {
		return fmt.Errorf("draft MAWB cannot have more than %d items", maxItems)
	}

	// Validate maximum number of charges
	const maxCharges = 20
	if len(req.Charges) > maxCharges {
		return fmt.Errorf("draft MAWB cannot have more than %d charges", maxCharges)
	}

	// Validate MAWB number format (basic validation)
	if len(req.MAWB) < 3 {
		return errors.New("MAWB number must be at least 3 characters long")
	}

	// Validate insurance amount
	if req.InsuranceAmount < 0 {
		return errors.New("insurance amount cannot be negative")
	}

	// Validate item-specific business rules
	for i, item := range req.Items {
		if err := s.validateItemBusinessRules(&item, i); err != nil {
			return err
		}
	}

	// Validate charge-specific business rules
	for i, charge := range req.Charges {
		if err := s.validateChargeBusinessRules(&charge, i); err != nil {
			return err
		}
	}

	return nil
}

// validateItemBusinessRules validates business rules for individual items
func (s *service) validateItemBusinessRules(item *DraftMAWBItemRequest, index int) error {
	// Validate rate charge is not negative
	if item.RateCharge < 0 {
		return fmt.Errorf("item %d: rate charge cannot be negative", index+1)
	}

	// Validate dimensions if provided
	const maxDimensions = 10
	if len(item.Dims) > maxDimensions {
		return fmt.Errorf("item %d: cannot have more than %d dimensions", index+1, maxDimensions)
	}

	// Validate each dimension
	for j, dim := range item.Dims {
		if err := s.validateDimensionBusinessRules(&dim, index, j); err != nil {
			return err
		}
	}

	return nil
}

// validateDimensionBusinessRules validates business rules for individual dimensions
func (s *service) validateDimensionBusinessRules(dim *DraftMAWBItemDimRequest, itemIndex, dimIndex int) error {
	// Validate dimension values are reasonable
	if dim.Length != "" {
		if length, err := strconv.ParseFloat(dim.Length, 64); err == nil {
			if length > 10000 { // 10 meters max
				return fmt.Errorf("item %d, dimension %d: length cannot exceed 10000 cm", itemIndex+1, dimIndex+1)
			}
		}
	}

	if dim.Width != "" {
		if width, err := strconv.ParseFloat(dim.Width, 64); err == nil {
			if width > 10000 { // 10 meters max
				return fmt.Errorf("item %d, dimension %d: width cannot exceed 10000 cm", itemIndex+1, dimIndex+1)
			}
		}
	}

	if dim.Height != "" {
		if height, err := strconv.ParseFloat(dim.Height, 64); err == nil {
			if height > 10000 { // 10 meters max
				return fmt.Errorf("item %d, dimension %d: height cannot exceed 10000 cm", itemIndex+1, dimIndex+1)
			}
		}
	}

	if dim.Count != "" {
		if count, err := strconv.Atoi(dim.Count); err == nil {
			if count > 10000 { // Max 10000 pieces
				return fmt.Errorf("item %d, dimension %d: count cannot exceed 10000", itemIndex+1, dimIndex+1)
			}
		}
	}

	return nil
}

// validateChargeBusinessRules validates business rules for individual charges
func (s *service) validateChargeBusinessRules(charge *DraftMAWBChargeRequest, index int) error {
	// Validate charge value is reasonable
	if charge.Value > 1000000 { // Max 1 million
		return fmt.Errorf("charge %d: value cannot exceed 1,000,000", index+1)
	}

	// Validate charge key length
	if len(charge.Key) > 100 {
		return fmt.Errorf("charge %d: key cannot exceed 100 characters", index+1)
	}

	return nil
}

// generatePDFContent creates PDF content from draft MAWB data using PDF generator
func (s *service) generatePDFContent(draftMAWB *DraftMAWB) ([]byte, error) {
	// Create PDF generator instance
	generator, err := s.createPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %w", err)
	}

	// Generate PDF using the generator
	pdfBytes, err := generator.GenerateDraftMAWBPDF(draftMAWB)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBytes, nil
}

// createPDFGenerator creates a new PDF generator instance with error handling and fallback
func (s *service) createPDFGenerator() (PDFGenerator, error) {
	generator, err := NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PDF generator: %w", err)
	}
	return generator, nil
}

// formatValidationErrors formats validation errors into a single error message
func (s *service) formatValidationErrors(validationErrors []customerrors.ValidationError) error {
	var messages []string
	for _, err := range validationErrors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return fmt.Errorf("validation errors: %s", strings.Join(messages, "; "))
}
