package draft_mawb

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

func (m *MockRepository) GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error) {
	args := m.Called(ctx, mawbUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DraftMAWB), args.Error(1)
}

func (m *MockRepository) GetByUUID(ctx context.Context, uuid string) (*DraftMAWB, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DraftMAWB), args.Error(1)
}

func (m *MockRepository) CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) error {
	args := m.Called(ctx, draftMAWB)
	return args.Error(0)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, uuid, status string) error {
	args := m.Called(ctx, uuid, status)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockRepository) ValidateMAWBExists(ctx context.Context, mawbUUID string) error {
	args := m.Called(ctx, mawbUUID)
	return args.Error(0)
}

func (m *MockRepository) ValidateUUIDExists(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockRepository) GetMultipleByMAWBUUIDs(ctx context.Context, mawbUUIDs []string) ([]*DraftMAWB, error) {
	args := m.Called(ctx, mawbUUIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*DraftMAWB), args.Error(1)
}

func (m *MockRepository) BatchUpdateStatus(ctx context.Context, uuids []string, status string) error {
	args := m.Called(ctx, uuids, status)
	return args.Error(0)
}

func (m *MockRepository) GetWithFilters(ctx context.Context, filters DraftMAWBFilters) ([]*DraftMAWB, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*DraftMAWB), args.Error(1)
}

func (m *MockRepository) GetItemsByDraftMAWBUUID(ctx context.Context, draftMAWBUUID string) ([]DraftMAWBItem, error) {
	args := m.Called(ctx, draftMAWBUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]DraftMAWBItem), args.Error(1)
}

func (m *MockRepository) GetChargesByDraftMAWBUUID(ctx context.Context, draftMAWBUUID string) ([]DraftMAWBCharge, error) {
	args := m.Called(ctx, draftMAWBUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]DraftMAWBCharge), args.Error(1)
}

// MockPDFGenerator is a mock implementation of the PDFGenerator interface
type MockPDFGenerator struct {
	mock.Mock
}

func (m *MockPDFGenerator) GenerateDraftMAWBPDF(draftMAWB *DraftMAWB) ([]byte, error) {
	args := m.Called(draftMAWB)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Test helper functions
func createTestDraftMAWB() *DraftMAWB {
	flightDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	executedDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	return &DraftMAWB{
		UUID:                    "test-uuid",
		MAWBInfoUUID:            "mawb-uuid",
		CustomerUUID:            "customer-uuid",
		AirlineLogo:             "logo.png",
		AirlineName:             "Test Airlines",
		MAWB:                    "123456789",
		HAWB:                    "H123456",
		ShipperNameAndAddress:   "Test Shipper Address",
		ConsigneeNameAndAddress: "Test Consignee Address",
		FlightNo:                "TG123",
		FlightDate:              &flightDate,
		ExecutedOnDate:          &executedDate,
		InsuranceAmount:         1000.0,
		TotalNoOfPieces:         10,
		TotalGrossWeight:        500.5,
		TotalChargeableWeight:   600.0,
		TotalRateCharge:         2000.0,
		TotalAmount:             2500.0,
		Status:                  StatusDraft,
		Items: []DraftMAWBItem{
			{
				ID:                1,
				DraftMAWBUUID:     "test-uuid",
				PiecesRCP:         "5",
				GrossWeight:       "250.5",
				KgLb:              "kg",
				RateClass:         "N",
				TotalVolume:       0.5,
				ChargeableWeight:  300.0,
				RateCharge:        1000.0,
				Total:             1000.0,
				NatureAndQuantity: "Electronics",
				Dims: []DraftMAWBItemDim{
					{
						ID:              1,
						DraftMAWBItemID: 1,
						Length:          "100",
						Width:           "50",
						Height:          "30",
						Count:           "2",
						CreatedAt:       time.Now(),
					},
				},
				CreatedAt: time.Now(),
			},
		},
		Charges: []DraftMAWBCharge{
			{
				ID:            1,
				DraftMAWBUUID: "test-uuid",
				Key:           "fuel_surcharge",
				Value:         500.0,
				CreatedAt:     time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestDraftMAWBRequest() *DraftMAWBRequest {
	return &DraftMAWBRequest{
		CustomerUUID:            "customer-uuid",
		AirlineLogo:             "logo.png",
		AirlineName:             "Test Airlines",
		MAWB:                    "123456789",
		HAWB:                    "H123456",
		ShipperNameAndAddress:   "Test Shipper Address",
		ConsigneeNameAndAddress: "Test Consignee Address",
		FlightNo:                "TG123",
		FlightDate:              "2024-01-01",
		ExecutedOnDate:          "2024-01-02",
		InsuranceAmount:         1000.0,
		Items: []DraftMAWBItemRequest{
			{
				PiecesRCP:         "5",
				GrossWeight:       "250.5",
				KgLb:              "kg",
				RateClass:         "N",
				RateCharge:        1000.0,
				NatureAndQuantity: "Electronics",
				Dims: []DraftMAWBItemDimRequest{
					{
						Length: "100",
						Width:  "50",
						Height: "30",
						Count:  "2",
					},
				},
			},
		},
		Charges: []DraftMAWBChargeRequest{
			{
				Key:   "fuel_surcharge",
				Value: 500.0,
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

func TestService_GetDraftMAWB_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	expectedDraftMAWB := createTestDraftMAWB()

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(expectedDraftMAWB, nil)

	result, err := service.GetDraftMAWB(ctx, mawbUUID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDraftMAWB.UUID, result.UUID)
	assert.Equal(t, expectedDraftMAWB.MAWB, result.MAWB)
	assert.Equal(t, len(expectedDraftMAWB.Items), len(result.Items))
	assert.Equal(t, len(expectedDraftMAWB.Charges), len(result.Charges))
	mockRepo.AssertExpectations(t)
}

func TestService_GetDraftMAWB_EmptyMAWBUUID(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()

	result, err := service.GetDraftMAWB(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MAWB Info UUID is required")
	assert.ErrorIs(t, err, ErrInvalidMAWBUUID)
}

func TestService_GetDraftMAWB_MAWBNotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "non-existent-uuid"

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(errors.New("MAWB Info not found"))

	result, err := service.GetDraftMAWB(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MAWB Info not found")
	assert.ErrorIs(t, err, ErrMAWBInfoNotFound)
	mockRepo.AssertExpectations(t)
}

func TestService_GetDraftMAWB_DraftMAWBNotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil, utils.ErrRecordNotFound)

	result, err := service.GetDraftMAWB(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "draft MAWB not found")
	assert.ErrorIs(t, err, ErrDraftMAWBNotFound)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateOrUpdateDraftMAWB_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	request := createTestDraftMAWBRequest()
	expectedDraftMAWB := createTestDraftMAWB()

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("CreateOrUpdate", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*draft_mawb.DraftMAWB")).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(expectedDraftMAWB, nil)

	result, err := service.CreateOrUpdateDraftMAWB(ctx, mawbUUID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDraftMAWB.UUID, result.UUID)
	assert.Equal(t, expectedDraftMAWB.MAWB, result.MAWB)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateOrUpdateDraftMAWB_EmptyMAWBUUID(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	request := createTestDraftMAWBRequest()

	result, err := service.CreateOrUpdateDraftMAWB(ctx, "", request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MAWB Info UUID is required")
	assert.ErrorIs(t, err, ErrInvalidMAWBUUID)
}

func TestService_CreateOrUpdateDraftMAWB_NilRequest(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	result, err := service.CreateOrUpdateDraftMAWB(ctx, mawbUUID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "request data cannot be nil")
	assert.ErrorIs(t, err, ErrInvalidRequestData)
}

func TestService_CreateOrUpdateDraftMAWB_ValidationError(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"

	// Create invalid request (missing required MAWB number)
	request := &DraftMAWBRequest{
		Items: []DraftMAWBItemRequest{
			{
				PiecesRCP:   "5",
				GrossWeight: "250.5",
			},
		},
	}

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)

	result, err := service.CreateOrUpdateDraftMAWB(ctx, mawbUUID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation errors")
	assert.ErrorIs(t, err, ErrInvalidRequestData)
	mockRepo.AssertExpectations(t)
}

func TestService_ConfirmDraftMAWB_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	draftMAWB := createTestDraftMAWB()
	draftMAWB.Status = StatusDraft

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(draftMAWB, nil)
	mockRepo.On("UpdateStatus", mock.AnythingOfType("*context.timerCtx"), draftMAWB.UUID, StatusConfirmed).Return(nil)

	err := service.ConfirmDraftMAWB(ctx, mawbUUID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_ConfirmDraftMAWB_AlreadyConfirmed(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	draftMAWB := createTestDraftMAWB()
	draftMAWB.Status = StatusConfirmed

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(draftMAWB, nil)

	err := service.ConfirmDraftMAWB(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already confirmed")
	assert.ErrorIs(t, err, ErrBusinessRuleViolation)
	mockRepo.AssertExpectations(t)
}

func TestService_RejectDraftMAWB_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	draftMAWB := createTestDraftMAWB()
	draftMAWB.Status = StatusDraft

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(draftMAWB, nil)
	mockRepo.On("UpdateStatus", mock.AnythingOfType("*context.timerCtx"), draftMAWB.UUID, StatusRejected).Return(nil)

	err := service.RejectDraftMAWB(ctx, mawbUUID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_RejectDraftMAWB_AlreadyRejected(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo, 30*time.Second)
	ctx := context.Background()
	mawbUUID := "test-mawb-uuid"
	draftMAWB := createTestDraftMAWB()
	draftMAWB.Status = StatusRejected

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(draftMAWB, nil)

	err := service.RejectDraftMAWB(ctx, mawbUUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already rejected")
	assert.ErrorIs(t, err, ErrBusinessRuleViolation)
	mockRepo.AssertExpectations(t)
}

func TestService_GenerateDraftMAWBPDF_Success(t *testing.T) {
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
	draftMAWB := createTestDraftMAWB()
	expectedPDF := []byte("mock pdf content")

	mockRepo.On("ValidateMAWBExists", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(nil)
	mockRepo.On("GetByMAWBUUID", mock.AnythingOfType("*context.timerCtx"), mawbUUID).Return(draftMAWB, nil)
	mockPDFGen.On("GenerateDraftMAWBPDF", draftMAWB).Return(expectedPDF, nil)

	result, err := service.GenerateDraftMAWBPDF(ctx, mawbUUID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPDF, result)
	mockRepo.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

// Test calculation methods
func TestService_calculateVolumetricWeight(t *testing.T) {
	service := &service{}

	tests := []struct {
		name           string
		dims           []DraftMAWBItemDim
		expectedVolume float64
		expectError    bool
	}{
		{
			name: "Valid dimensions",
			dims: []DraftMAWBItemDim{
				{Length: "100", Width: "50", Height: "30", Count: "2"},
			},
			expectedVolume: 0.3, // (100*50*30/1000000)*2 = 0.3
			expectError:    false,
		},
		{
			name: "Multiple dimensions",
			dims: []DraftMAWBItemDim{
				{Length: "100", Width: "50", Height: "30", Count: "1"},
				{Length: "200", Width: "100", Height: "50", Count: "1"},
			},
			expectedVolume: 1.15, // 0.15 + 1.0 = 1.15
			expectError:    false,
		},
		{
			name: "Invalid length",
			dims: []DraftMAWBItemDim{
				{Length: "invalid", Width: "50", Height: "30", Count: "2"},
			},
			expectedVolume: 0,
			expectError:    true,
		},
		{
			name: "Zero dimensions",
			dims: []DraftMAWBItemDim{
				{Length: "0", Width: "50", Height: "30", Count: "2"},
			},
			expectedVolume: 0,
			expectError:    true,
		},
		{
			name: "Invalid count",
			dims: []DraftMAWBItemDim{
				{Length: "100", Width: "50", Height: "30", Count: "invalid"},
			},
			expectedVolume: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.calculateVolumetricWeight(tt.dims)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedVolume, result, 0.01)
			}
		})
	}
}

func TestService_calculateChargeableWeight(t *testing.T) {
	service := &service{}

	tests := []struct {
		name                     string
		actualWeight             float64
		volumetricWeight         float64
		expectedChargeableWeight float64
	}{
		{
			name:                     "Actual weight higher",
			actualWeight:             500.0,
			volumetricWeight:         1.0,
			expectedChargeableWeight: 500.0, // max(500, 1*166.67) = 500
		},
		{
			name:                     "Volumetric weight higher",
			actualWeight:             100.0,
			volumetricWeight:         1.0,
			expectedChargeableWeight: 166.67, // max(100, 1*166.67) = 166.67
		},
		{
			name:                     "Equal weights",
			actualWeight:             166.67,
			volumetricWeight:         1.0,
			expectedChargeableWeight: 166.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateChargeableWeight(tt.actualWeight, tt.volumetricWeight)
			assert.InDelta(t, tt.expectedChargeableWeight, result, 0.01)
		})
	}
}

func TestService_parseWeight(t *testing.T) {
	service := &service{}

	tests := []struct {
		name           string
		weightStr      string
		unit           string
		expectedWeight float64
		expectError    bool
	}{
		{
			name:           "Valid kg weight",
			weightStr:      "100.5",
			unit:           "kg",
			expectedWeight: 100.5,
			expectError:    false,
		},
		{
			name:           "Valid lb weight",
			weightStr:      "100",
			unit:           "lb",
			expectedWeight: 45.3592, // 100 * 0.453592
			expectError:    false,
		},
		{
			name:           "Empty weight",
			weightStr:      "",
			unit:           "kg",
			expectedWeight: 0,
			expectError:    false,
		},
		{
			name:           "Invalid weight format",
			weightStr:      "invalid",
			unit:           "kg",
			expectedWeight: 0,
			expectError:    true,
		},
		{
			name:           "Case insensitive lb",
			weightStr:      "100",
			unit:           "LB",
			expectedWeight: 45.3592,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parseWeight(tt.weightStr, tt.unit)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedWeight, result, 0.01)
			}
		})
	}
}

func TestService_parsePieces(t *testing.T) {
	service := &service{}

	tests := []struct {
		name           string
		piecesStr      string
		expectedPieces int
		expectError    bool
	}{
		{
			name:           "Valid pieces",
			piecesStr:      "10",
			expectedPieces: 10,
			expectError:    false,
		},
		{
			name:           "Empty pieces",
			piecesStr:      "",
			expectedPieces: 0,
			expectError:    false,
		},
		{
			name:           "Invalid pieces format",
			piecesStr:      "invalid",
			expectedPieces: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parsePieces(tt.piecesStr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPieces, result)
			}
		})
	}
}

func TestService_calculateChargesTotal(t *testing.T) {
	service := &service{}

	charges := []DraftMAWBCharge{
		{Value: 100.0},
		{Value: 200.5},
		{Value: 50.25},
	}

	result := service.calculateChargesTotal(charges)
	assert.InDelta(t, 350.75, result, 0.01)
}

func TestService_performCalculations(t *testing.T) {
	service := &service{}

	draftMAWB := &DraftMAWB{
		Items: []DraftMAWBItem{
			{
				PiecesRCP:   "5",
				GrossWeight: "100",
				KgLb:        "kg",
				RateCharge:  10.0,
				Dims: []DraftMAWBItemDim{
					{Length: "100", Width: "50", Height: "30", Count: "2"},
				},
			},
			{
				PiecesRCP:   "3",
				GrossWeight: "50",
				KgLb:        "kg",
				RateCharge:  15.0,
				Dims: []DraftMAWBItemDim{
					{Length: "80", Width: "40", Height: "20", Count: "1"},
				},
			},
		},
		Charges: []DraftMAWBCharge{
			{Value: 100.0},
			{Value: 50.0},
		},
	}

	err := service.performCalculations(draftMAWB)

	assert.NoError(t, err)
	assert.Equal(t, 8, draftMAWB.TotalNoOfPieces)              // 5 + 3
	assert.InDelta(t, 150.0, draftMAWB.TotalGrossWeight, 0.01) // 100 + 50
	assert.InDelta(t, 25.0, draftMAWB.TotalRateCharge, 0.01)   // 10 + 15

	// Total amount should include item totals + charges
	expectedItemTotals := (10.0 * draftMAWB.Items[0].ChargeableWeight) + (15.0 * draftMAWB.Items[1].ChargeableWeight)
	expectedTotal := expectedItemTotals + 150.0 // charges total
	assert.InDelta(t, expectedTotal, draftMAWB.TotalAmount, 0.01)
}

// Test business rule validation methods
func TestService_validateBusinessRules(t *testing.T) {
	service := &service{}

	tests := []struct {
		name        string
		request     *DraftMAWBRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &DraftMAWBRequest{
				MAWB:            "123456789",
				InsuranceAmount: 1000.0,
				Items: []DraftMAWBItemRequest{
					{
						GrossWeight: "100",
						RateCharge:  10.0,
					},
				},
				Charges: []DraftMAWBChargeRequest{
					{
						Key:   "fuel_surcharge",
						Value: 100.0,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Too many items",
			request: &DraftMAWBRequest{
				MAWB:  "123456789",
				Items: make([]DraftMAWBItemRequest, 51), // Exceeds max of 50
			},
			expectError: true,
			errorMsg:    "cannot have more than 50 items",
		},
		{
			name: "Too many charges",
			request: &DraftMAWBRequest{
				MAWB:    "123456789",
				Charges: make([]DraftMAWBChargeRequest, 21), // Exceeds max of 20
			},
			expectError: true,
			errorMsg:    "cannot have more than 20 charges",
		},
		{
			name: "Short MAWB number",
			request: &DraftMAWBRequest{
				MAWB: "12",
			},
			expectError: true,
			errorMsg:    "must be at least 3 characters long",
		},
		{
			name: "Negative insurance amount",
			request: &DraftMAWBRequest{
				MAWB:            "123456789",
				InsuranceAmount: -100.0,
			},
			expectError: true,
			errorMsg:    "cannot be negative",
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
		item        *DraftMAWBItemRequest
		index       int
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid item",
			item: &DraftMAWBItemRequest{
				GrossWeight: "100",
				RateCharge:  10.0,
				Dims: []DraftMAWBItemDimRequest{
					{Length: "100", Width: "50", Height: "30", Count: "2"},
				},
			},
			index:       0,
			expectError: false,
		},
		{
			name: "Negative rate charge",
			item: &DraftMAWBItemRequest{
				GrossWeight: "100",
				RateCharge:  -10.0,
			},
			index:       0,
			expectError: true,
			errorMsg:    "rate charge cannot be negative",
		},
		{
			name: "Too many dimensions",
			item: &DraftMAWBItemRequest{
				GrossWeight: "100",
				RateCharge:  10.0,
				Dims:        make([]DraftMAWBItemDimRequest, 11), // Exceeds max of 10
			},
			index:       0,
			expectError: true,
			errorMsg:    "cannot have more than 10 dimensions",
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

func TestService_validateDimensionBusinessRules(t *testing.T) {
	service := &service{}

	tests := []struct {
		name        string
		dim         *DraftMAWBItemDimRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid dimension",
			dim: &DraftMAWBItemDimRequest{
				Length: "100",
				Width:  "50",
				Height: "30",
				Count:  "2",
			},
			expectError: false,
		},
		{
			name: "Excessive length",
			dim: &DraftMAWBItemDimRequest{
				Length: "20000", // Exceeds max of 10000
				Width:  "50",
				Height: "30",
				Count:  "2",
			},
			expectError: true,
			errorMsg:    "length cannot exceed 10000 cm",
		},
		{
			name: "Excessive count",
			dim: &DraftMAWBItemDimRequest{
				Length: "100",
				Width:  "50",
				Height: "30",
				Count:  "20000", // Exceeds max of 10000
			},
			expectError: true,
			errorMsg:    "count cannot exceed 10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateDimensionBusinessRules(tt.dim, 0, 0)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_validateChargeBusinessRules(t *testing.T) {
	service := &service{}

	tests := []struct {
		name        string
		charge      *DraftMAWBChargeRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid charge",
			charge: &DraftMAWBChargeRequest{
				Key:   "fuel_surcharge",
				Value: 100.0,
			},
			expectError: false,
		},
		{
			name: "Excessive value",
			charge: &DraftMAWBChargeRequest{
				Key:   "fuel_surcharge",
				Value: 2000000.0, // Exceeds max of 1,000,000
			},
			expectError: true,
			errorMsg:    "value cannot exceed 1,000,000",
		},
		{
			name: "Long key",
			charge: &DraftMAWBChargeRequest{
				Key:   string(make([]byte, 101)), // Exceeds max of 100 characters
				Value: 100.0,
			},
			expectError: true,
			errorMsg:    "key cannot exceed 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateChargeBusinessRules(tt.charge, 0)
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
	request := createTestDraftMAWBRequest()

	result, err := service.convertRequestToModel(mawbUUID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mawbUUID, result.MAWBInfoUUID)
	assert.Equal(t, request.MAWB, result.MAWB)
	assert.Equal(t, request.CustomerUUID, result.CustomerUUID)
	assert.Equal(t, len(request.Items), len(result.Items))
	assert.Equal(t, len(request.Charges), len(result.Charges))

	if len(result.Items) > 0 {
		assert.Equal(t, request.Items[0].PiecesRCP, result.Items[0].PiecesRCP)
		assert.Equal(t, request.Items[0].GrossWeight, result.Items[0].GrossWeight)
		assert.Equal(t, len(request.Items[0].Dims), len(result.Items[0].Dims))
	}

	if len(result.Charges) > 0 {
		assert.Equal(t, request.Charges[0].Key, result.Charges[0].Key)
		assert.Equal(t, request.Charges[0].Value, result.Charges[0].Value)
	}
}

func TestService_convertToResponse(t *testing.T) {
	service := &service{}
	draftMAWB := createTestDraftMAWB()

	result := service.convertToResponse(draftMAWB)

	assert.NotNil(t, result)
	assert.Equal(t, draftMAWB.UUID, result.UUID)
	assert.Equal(t, draftMAWB.MAWBInfoUUID, result.MAWBInfoUUID)
	assert.Equal(t, draftMAWB.MAWB, result.MAWB)
	assert.Equal(t, draftMAWB.Status, result.Status)
	assert.Equal(t, len(draftMAWB.Items), len(result.Items))
	assert.Equal(t, len(draftMAWB.Charges), len(result.Charges))
	assert.NotEmpty(t, result.CreatedAt)
	assert.NotEmpty(t, result.UpdatedAt)

	if draftMAWB.FlightDate != nil {
		assert.Equal(t, "2024-01-01", result.FlightDate)
	}

	if draftMAWB.ExecutedOnDate != nil {
		assert.Equal(t, "2024-01-02", result.ExecutedOnDate)
	}
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
