package cargo_manifest

import (
	"context"
	"errors"
	"testing"
	"time"

	customerrors "hpc-express-service/errors"
	"hpc-express-service/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error) {
	args := m.Called(ctx, mawbUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CargoManifest), args.Error(1)
}

func (m *MockRepository) CreateOrUpdate(ctx context.Context, manifest *CargoManifest) error {
	args := m.Called(ctx, manifest)
	return args.Error(0)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, uuid, status string) error {
	args := m.Called(ctx, uuid, status)
	return args.Error(0)
}

func (m *MockRepository) ValidateMAWBExists(ctx context.Context, mawbUUID string) error {
	args := m.Called(ctx, mawbUUID)
	return args.Error(0)
}

// MockPDFGenerator is a mock implementation of the PDFGenerator interface
type MockPDFGenerator struct {
	mock.Mock
}

func (m *MockPDFGenerator) GenerateCargoManifestPDF(manifest *CargoManifest) ([]byte, error) {
	args := m.Called(manifest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Test helper functions
func createTestCargoManifest() *CargoManifest {
	return &CargoManifest{
		UUID:            "test-uuid",
		MAWBInfoUUID:    "mawb-uuid",
		MAWBNumber:      "123456789",
		PortOfDischarge: "BKK",
		FlightNo:        "TG123",
		FreightDate:     "2024-01-01",
		Shipper:         "Test Shipper",
		Consignee:       "Test Consignee",
		TotalCtn:        "10",
		Transshipment:   "No",
		Status:          StatusDraft,
		Items: []CargoManifestItem{
			{
				ID:                      1,
				CargoManifestUUID:       "test-uuid",
				HAWBNo:                  "H123",
				Pkgs:                    "5",
				GrossWeight:             "100.5",
				Destination:             "NYC",
				Commodity:               "Electronics",
				ShipperNameAndAddress:   "Shipper Address",
				ConsigneeNameAndAddress: "Consignee Address",
				CreatedAt:               time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestCargoManifestRequest() *CargoManifestRequest {
	return &CargoManifestRequest{
		MAWBNumber:      "123456789",
		PortOfDischarge: "BKK",
		FlightNo:        "TG123",
		FreightDate:     "2024-01-01",
		Shipper:         "Test Shipper",
		Consignee:       "Test Consignee",
		TotalCtn:        "10",
		Transshipment:   "No",
		Items: []CargoManifestItemRequest{
			{
				HAWBNo:                  "H123",
				Pkgs:                    "5",
				GrossWeight:             "100.5",
				Destination:             "NYC",
				Commodity:               "Electronics",
				ShipperNameAndAddress:   "Shipper Address",
				ConsigneeNameAndAddress: "Consignee Address",
			},
		},
	}
}

func TestNewService(t *testing.T) {
	mockRepo := &MockRepository{}
	timeout := 30 * time.Second

	service := NewService(mockRepo, timeout)

	assert.NotNil(t, service)
	assert.IsType(t, &service{}, service)
}

func TestService_GetCargoManifest_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	expectedManifest := createTestCargoManifest()

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(expectedManifest, nil)

	result, err := service.GetCargoManifest(ctx, mawbUUID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedManifest.UUID, result.UUID)
	assert.Equal(t, expectedManifest.MAWBNumber, result.MAWBNumber)
	assert.Equal(t, len(expectedManifest.Items), len(result.Items))
	mockRepo.AssertExpectations(t)
}

func TestService_GetCargoManifest_EmptyMAWBUUID(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()

	result, err := service.GetCargoManifest(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MAWB Info UUID is required")
	assert.ErrorIs(t, err, ErrInvalidMAWBUUID)
}

func TestService_GetCargoManifest_MAWBNotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "non-existent-uuid"

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(errors.New("MAWB Info not found"))

	result, err := service.GetCargoManifest(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MAWB Info not found")
	assert.ErrorIs(t, err, ErrMAWBInfoNotFound)
	mockRepo.AssertExpectations(t)
}

func TestService_GetCargoManifest_CargoManifestNotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil, utils.ErrRecordNotFound)

	result, err := service.GetCargoManifest(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cargo manifest not found")
	assert.ErrorIs(t, err, ErrCargoManifestNotFound)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateOrUpdateCargoManifest_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	request := createTestCargoManifestRequest()
	expectedManifest := createTestCargoManifest()

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("CreateOrUpdate", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*cargo_manifest.CargoManifest")).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(expectedManifest, nil)

	result, err := service.CreateOrUpdateCargoManifest(ctx, mawbUUID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedManifest.UUID, result.UUID)
	assert.Equal(t, expectedManifest.MAWBNumber, result.MAWBNumber)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateOrUpdateCargoManifest_EmptyMAWBUUID(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	request := createTestCargoManifestRequest()

	result, err := service.CreateOrUpdateCargoManifest(ctx, "", request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MAWB Info UUID is required")
	assert.ErrorIs(t, err, ErrInvalidMAWBUUID)
}

func TestService_CreateOrUpdateCargoManifest_NilRequest(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	result, err := service.CreateOrUpdateCargoManifest(ctx, mawbUUID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "request data cannot be nil")
	assert.ErrorIs(t, err, ErrInvalidRequestData)
}

func TestService_CreateOrUpdateCargoManifest_ValidationError(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	// Create invalid request (missing required MAWB number)
	request := &CargoManifestRequest{
		Items: []CargoManifestItemRequest{
			{
				HAWBNo:      "H123",
				GrossWeight: "100.5",
			},
		},
	}

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)

	result, err := service.CreateOrUpdateCargoManifest(ctx, mawbUUID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation errors")
	assert.ErrorIs(t, err, ErrInvalidRequestData)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateOrUpdateCargoManifest_BusinessRuleViolation(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	// Create request with too many items
	request := &CargoManifestRequest{
		MAWBNumber: "123456789",
		Items:      make([]CargoManifestItemRequest, 101), // Exceeds max of 100
	}

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)

	result, err := service.CreateOrUpdateCargoManifest(ctx, mawbUUID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot have more than 100 items")
	assert.ErrorIs(t, err, ErrBusinessRuleViolation)
	mockRepo.AssertExpectations(t)
}

func TestService_ConfirmCargoManifest_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	manifest := createTestCargoManifest()
	manifest.Status = StatusDraft

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(manifest, nil)
	mockRepo.On("UpdateStatus", mock.AnythingOfType("*context.timerCtx"), manifest.UUID, StatusConfirmed).Return(nil)

	err := service.ConfirmCargoManifest(ctx, mawbUUID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_ConfirmCargoManifest_AlreadyConfirmed(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	manifest := createTestCargoManifest()
	manifest.Status = StatusConfirmed

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(manifest, nil)

	err := service.ConfirmCargoManifest(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already confirmed")
	assert.ErrorIs(t, err, ErrBusinessRuleViolation)
	mockRepo.AssertExpectations(t)
}

func TestService_RejectCargoManifest_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	manifest := createTestCargoManifest()
	manifest.Status = StatusDraft

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(manifest, nil)
	mockRepo.On("UpdateStatus", mock.AnythingOfType("*context.timerCtx"), manifest.UUID, StatusRejected).Return(nil)

	err := service.RejectCargoManifest(ctx, mawbUUID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_RejectCargoManifest_AlreadyRejected(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	manifest := createTestCargoManifest()
	manifest.Status = StatusRejected

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(manifest, nil)

	err := service.RejectCargoManifest(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already rejected")
	assert.ErrorIs(t, err, ErrBusinessRuleViolation)
	mockRepo.AssertExpectations(t)
}

func TestService_GenerateCargoManifestPDF_Success(t *testing.T) {
	// Setup mock PDF generator
	originalNewPDFGenerator := NewPDFGenerator
	defer func() { NewPDFGenerator = originalNewPDFGenerator }()

	mockPDFGen := &MockPDFGenerator{}
	NewPDFGenerator = func() (PDFGenerator, error) {
		return mockPDFGen, nil
	}

	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	manifest := createTestCargoManifest()
	expectedPDF := []byte("mock pdf content")

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(manifest, nil)
	mockPDFGen.On("GenerateCargoManifestPDF", manifest).Return(expectedPDF, nil)

	result, err := service.GenerateCargoManifestPDF(ctx, mawbUUID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPDF, result)
	mockRepo.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

func TestService_GenerateCargoManifestPDF_PDFGenerationFailed(t *testing.T) {
	// Setup mock PDF generator
	originalNewPDFGenerator := NewPDFGenerator
	defer func() { NewPDFGenerator = originalNewPDFGenerator }()

	mockPDFGen := &MockPDFGenerator{}
	NewPDFGenerator = func() (PDFGenerator, error) {
		return mockPDFGen, nil
	}

	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	manifest := createTestCargoManifest()

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(manifest, nil)
	mockPDFGen.On("GenerateCargoManifestPDF", manifest).Return(nil, errors.New("PDF generation error"))

	result, err := service.GenerateCargoManifestPDF(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "PDF generation failed")
	assert.ErrorIs(t, err, ErrPDFGenerationFailed)
	mockRepo.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

// Test business rule validation methods
func TestService_validateBusinessRules(t *testing.T) {
	service := &service{}

	tests := []struct {
		name        string
		request     *CargoManifestRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &CargoManifestRequest{
				MAWBNumber: "123456789",
				Items: []CargoManifestItemRequest{
					{
						HAWBNo:      "H123",
						Pkgs:        "5",
						GrossWeight: "100.5",
						Commodity:   "Electronics",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Too many items",
			request: &CargoManifestRequest{
				MAWBNumber: "123456789",
				Items:      make([]CargoManifestItemRequest, 101),
			},
			expectError: true,
			errorMsg:    "cannot have more than 100 items",
		},
		{
			name: "Short MAWB number",
			request: &CargoManifestRequest{
				MAWBNumber: "12",
				Items: []CargoManifestItemRequest{
					{
						HAWBNo:      "H123",
						Pkgs:        "5",
						GrossWeight: "100.5",
						Commodity:   "Electronics",
					},
				},
			},
			expectError: true,
			errorMsg:    "must be at least 3 characters long",
		},
		{
			name: "No items",
			request: &CargoManifestRequest{
				MAWBNumber: "123456789",
				Items:      []CargoManifestItemRequest{},
			},
			expectError: true,
			errorMsg:    "must contain at least one item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateBusinessRules(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_validateItemBusinessRules(t *testing.T) {
	service := &service{}

	tests := []struct {
		name        string
		item        *CargoManifestItemRequest
		index       int
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid item",
			item: &CargoManifestItemRequest{
				HAWBNo:      "H123456",
				Pkgs:        "5",
				GrossWeight: "100.5",
				Commodity:   "Electronics",
			},
			index:       0,
			expectError: false,
		},
		{
			name: "Short HAWB number",
			item: &CargoManifestItemRequest{
				HAWBNo:      "H1",
				Pkgs:        "5",
				GrossWeight: "100.5",
				Commodity:   "Electronics",
			},
			index:       0,
			expectError: true,
			errorMsg:    "HAWB number must be at least 3 characters long",
		},
		{
			name: "Empty gross weight",
			item: &CargoManifestItemRequest{
				HAWBNo:      "H123456",
				Pkgs:        "5",
				GrossWeight: "",
				Commodity:   "Electronics",
			},
			index:       0,
			expectError: true,
			errorMsg:    "gross weight cannot be empty",
		},
		{
			name: "Empty packages",
			item: &CargoManifestItemRequest{
				HAWBNo:      "H123456",
				Pkgs:        "",
				GrossWeight: "100.5",
				Commodity:   "Electronics",
			},
			index:       0,
			expectError: true,
			errorMsg:    "packages field cannot be empty",
		},
		{
			name: "Empty commodity",
			item: &CargoManifestItemRequest{
				HAWBNo:      "H123456",
				Pkgs:        "5",
				GrossWeight: "100.5",
				Commodity:   "",
			},
			index:       0,
			expectError: true,
			errorMsg:    "commodity field cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateItemBusinessRules(tt.item, tt.index)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_convertRequestToModel(t *testing.T) {
	service := &service{}
	mawbUUID := "test-mawb-uuid"
	request := createTestCargoManifestRequest()

	result := service.convertRequestToModel(mawbUUID, request)

	assert.NotNil(t, result)
	assert.Equal(t, mawbUUID, result.MAWBInfoUUID)
	assert.Equal(t, request.MAWBNumber, result.MAWBNumber)
	assert.Equal(t, request.PortOfDischarge, result.PortOfDischarge)
	assert.Equal(t, len(request.Items), len(result.Items))

	if len(result.Items) > 0 {
		assert.Equal(t, request.Items[0].HAWBNo, result.Items[0].HAWBNo)
		assert.Equal(t, request.Items[0].Pkgs, result.Items[0].Pkgs)
		assert.Equal(t, request.Items[0].GrossWeight, result.Items[0].GrossWeight)
	}
}

func TestService_convertToResponse(t *testing.T) {
	service := &service{}
	manifest := createTestCargoManifest()

	result := service.convertToResponse(manifest)

	assert.NotNil(t, result)
	assert.Equal(t, manifest.UUID, result.UUID)
	assert.Equal(t, manifest.MAWBInfoUUID, result.MAWBInfoUUID)
	assert.Equal(t, manifest.MAWBNumber, result.MAWBNumber)
	assert.Equal(t, manifest.Status, result.Status)
	assert.Equal(t, len(manifest.Items), len(result.Items))
	assert.NotEmpty(t, result.CreatedAt)
	assert.NotEmpty(t, result.UpdatedAt)
}

func TestService_formatValidationErrors(t *testing.T) {
	service := &service{}
	validationErrors := []customerrors.ValidationError{
		customerrors.NewValidationError("field1", "error message 1"),
		customerrors.NewValidationError("field2", "error message 2"),
	}

	result := service.formatValidationErrors(validationErrors)

	assert.Error(t, result)
	assert.Contains(t, result.Error(), "field1: error message 1")
	assert.Contains(t, result.Error(), "field2: error message 2")
	assert.Contains(t, result.Error(), "validation errors:")
}
