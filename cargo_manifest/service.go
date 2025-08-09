package cargo_manifest

import (
	"context"
	"errors"
	"fmt"
	"hpc-express-service/utils"
	"strings"
	"time"

	customerrors "hpc-express-service/errors"
)

// PDFGenerator interface for generating PDF documents
type PDFGenerator interface {
	GenerateCargoManifestPDF(manifest *CargoManifest) ([]byte, error)
}

// NewPDFGenerator creates a new PDF generator instance
var NewPDFGenerator func() (PDFGenerator, error)

// Custom error types for better error handling
var (
	ErrMAWBInfoNotFound      = errors.New("MAWB Info not found")
	ErrCargoManifestNotFound = errors.New("cargo manifest not found")
	ErrInvalidMAWBUUID       = errors.New("invalid MAWB Info UUID")
	ErrInvalidRequestData    = errors.New("invalid request data")
	ErrBusinessRuleViolation = errors.New("business rule violation")
	ErrPDFGenerationFailed   = errors.New("PDF generation failed")
)

// Service interface defines the contract for cargo manifest business logic
type Service interface {
	GetCargoManifest(ctx context.Context, mawbUUID string) (*CargoManifestResponse, error)
	CreateOrUpdateCargoManifest(ctx context.Context, mawbUUID string, req *CargoManifestRequest) (*CargoManifestResponse, error)
	ConfirmCargoManifest(ctx context.Context, mawbUUID string) error
	RejectCargoManifest(ctx context.Context, mawbUUID string) error
	GenerateCargoManifestPDF(ctx context.Context, mawbUUID string) ([]byte, error)
}

type service struct {
	repo           Repository
	contextTimeout time.Duration
}

// NewService creates a new cargo manifest service instance
func NewService(repo Repository, timeout time.Duration) Service {
	return &service{
		repo:           repo,
		contextTimeout: timeout,
	}
}

// GetCargoManifest retrieves a cargo manifest by MAWB Info UUID
func (s *service) GetCargoManifest(ctx context.Context, mawbUUID string) (*CargoManifestResponse, error) {
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

	// Get cargo manifest from repository
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("%w for MAWB: %s", ErrCargoManifestNotFound, mawbUUID)
		}
		return nil, fmt.Errorf("failed to retrieve cargo manifest: %w", err)
	}

	// Convert to response model
	response := s.convertToResponse(manifest)
	return response, nil
}

// CreateOrUpdateCargoManifest creates a new cargo manifest or updates an existing one
func (s *service) CreateOrUpdateCargoManifest(ctx context.Context, mawbUUID string, req *CargoManifestRequest) (*CargoManifestResponse, error) {
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
	SanitizeCargoManifestRequest(req)

	// Validate request
	if validationErrors := ValidateCargoManifestRequest(req); len(validationErrors) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRequestData, s.formatValidationErrors(validationErrors).Error())
	}

	// Additional business rule validation
	if err := s.validateBusinessRules(req); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBusinessRuleViolation, err.Error())
	}

	// Convert request to domain model
	manifest := s.convertRequestToModel(mawbUUID, req)

	// Save to repository
	if err := s.repo.CreateOrUpdate(ctx, manifest); err != nil {
		return nil, fmt.Errorf("failed to save cargo manifest: %w", err)
	}

	// Get the saved manifest to return complete data
	savedManifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve saved cargo manifest: %w", err)
	}

	// Convert to response model
	response := s.convertToResponse(savedManifest)
	return response, nil
}

// ConfirmCargoManifest updates the cargo manifest status to confirmed
func (s *service) ConfirmCargoManifest(ctx context.Context, mawbUUID string) error {
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

	// Get existing cargo manifest to get its UUID
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return fmt.Errorf("%w for MAWB: %s", ErrCargoManifestNotFound, mawbUUID)
		}
		return fmt.Errorf("failed to retrieve cargo manifest: %w", err)
	}

	// Validate that the manifest can be confirmed (business rule)
	if manifest.Status == StatusConfirmed {
		return fmt.Errorf("%w: cargo manifest is already confirmed", ErrBusinessRuleViolation)
	}

	// Update status to confirmed
	if err := s.repo.UpdateStatus(ctx, manifest.UUID, StatusConfirmed); err != nil {
		return fmt.Errorf("failed to confirm cargo manifest: %w", err)
	}

	return nil
}

// RejectCargoManifest updates the cargo manifest status to rejected
func (s *service) RejectCargoManifest(ctx context.Context, mawbUUID string) error {
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

	// Get existing cargo manifest to get its UUID
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return fmt.Errorf("%w for MAWB: %s", ErrCargoManifestNotFound, mawbUUID)
		}
		return fmt.Errorf("failed to retrieve cargo manifest: %w", err)
	}

	// Validate that the manifest can be rejected (business rule)
	if manifest.Status == StatusRejected {
		return fmt.Errorf("%w: cargo manifest is already rejected", ErrBusinessRuleViolation)
	}

	// Update status to rejected
	if err := s.repo.UpdateStatus(ctx, manifest.UUID, StatusRejected); err != nil {
		return fmt.Errorf("failed to reject cargo manifest: %w", err)
	}

	return nil
}

// convertRequestToModel converts a request model to domain model
func (s *service) convertRequestToModel(mawbUUID string, req *CargoManifestRequest) *CargoManifest {
	manifest := &CargoManifest{
		MAWBInfoUUID:    mawbUUID,
		MAWBNumber:      req.MAWBNumber,
		PortOfDischarge: req.PortOfDischarge,
		FlightNo:        req.FlightNo,
		FreightDate:     req.FreightDate,
		Shipper:         req.Shipper,
		Consignee:       req.Consignee,
		TotalCtn:        req.TotalCtn,
		Transshipment:   req.Transshipment,
		Items:           make([]CargoManifestItem, len(req.Items)),
	}

	// Convert items
	for i, itemReq := range req.Items {
		manifest.Items[i] = CargoManifestItem{
			HAWBNo:                  itemReq.HAWBNo,
			Pkgs:                    itemReq.Pkgs,
			GrossWeight:             itemReq.GrossWeight,
			Destination:             itemReq.Destination,
			Commodity:               itemReq.Commodity,
			ShipperNameAndAddress:   itemReq.ShipperNameAndAddress,
			ConsigneeNameAndAddress: itemReq.ConsigneeNameAndAddress,
		}
	}

	return manifest
}

// convertToResponse converts a domain model to response model
func (s *service) convertToResponse(manifest *CargoManifest) *CargoManifestResponse {
	return &CargoManifestResponse{
		UUID:            manifest.UUID,
		MAWBInfoUUID:    manifest.MAWBInfoUUID,
		MAWBNumber:      manifest.MAWBNumber,
		PortOfDischarge: manifest.PortOfDischarge,
		FlightNo:        manifest.FlightNo,
		FreightDate:     manifest.FreightDate,
		Shipper:         manifest.Shipper,
		Consignee:       manifest.Consignee,
		TotalCtn:        manifest.TotalCtn,
		Transshipment:   manifest.Transshipment,
		Status:          manifest.Status,
		Items:           manifest.Items,
		CreatedAt:       manifest.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       manifest.UpdatedAt.Format(time.RFC3339),
	}
}

// GenerateCargoManifestPDF generates a PDF document for the cargo manifest
func (s *service) GenerateCargoManifestPDF(ctx context.Context, mawbUUID string) ([]byte, error) {
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

	// Get cargo manifest from repository
	manifest, err := s.repo.GetByMAWBUUID(ctx, mawbUUID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("%w for MAWB: %s", ErrCargoManifestNotFound, mawbUUID)
		}
		return nil, fmt.Errorf("failed to retrieve cargo manifest: %w", err)
	}

	// Generate PDF using the PDF generator service
	pdfContent, err := s.generatePDFContent(manifest)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPDFGenerationFailed, err.Error())
	}

	return pdfContent, nil
}

// validateBusinessRules performs additional business logic validation
func (s *service) validateBusinessRules(req *CargoManifestRequest) error {
	// Validate maximum number of items
	const maxItems = 100
	if len(req.Items) > maxItems {
		return fmt.Errorf("cargo manifest cannot have more than %d items", maxItems)
	}

	// Validate MAWB number format (basic validation)
	if len(req.MAWBNumber) < 3 {
		return errors.New("MAWB number must be at least 3 characters long")
	}

	// Validate that at least one item is provided for non-empty manifests
	if len(req.Items) == 0 {
		return errors.New("cargo manifest must contain at least one item")
	}

	// Validate item-specific business rules
	for i, item := range req.Items {
		if err := s.validateItemBusinessRules(&item, i); err != nil {
			return err
		}
	}

	return nil
}

// validateItemBusinessRules validates business rules for individual items
func (s *service) validateItemBusinessRules(item *CargoManifestItemRequest, index int) error {
	// Validate HAWB number format
	if len(item.HAWBNo) < 3 {
		return fmt.Errorf("item %d: HAWB number must be at least 3 characters long", index+1)
	}

	// Validate gross weight is not empty and contains valid characters
	if strings.TrimSpace(item.GrossWeight) == "" {
		return fmt.Errorf("item %d: gross weight cannot be empty", index+1)
	}

	// Validate packages field
	if strings.TrimSpace(item.Pkgs) == "" {
		return fmt.Errorf("item %d: packages field cannot be empty", index+1)
	}

	// Validate commodity field
	if strings.TrimSpace(item.Commodity) == "" {
		return fmt.Errorf("item %d: commodity field cannot be empty", index+1)
	}

	return nil
}

// generatePDFContent creates PDF content from cargo manifest data using PDF generator
func (s *service) generatePDFContent(manifest *CargoManifest) ([]byte, error) {
	// Create PDF generator instance
	generator, err := s.createPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %w", err)
	}

	// Generate PDF using the generator
	pdfBytes, err := generator.GenerateCargoManifestPDF(manifest)
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
