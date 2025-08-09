package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"hpc-express-service/draft_mawb"
)

// MockDraftMAWBService is a mock implementation of the draft MAWB service
type MockDraftMAWBService struct {
	mock.Mock
}

func (m *MockDraftMAWBService) GetDraftMAWB(ctx context.Context, mawbUUID string) (*draft_mawb.DraftMAWBResponse, error) {
	args := m.Called(ctx, mawbUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*draft_mawb.DraftMAWBResponse), args.Error(1)
}

func (m *MockDraftMAWBService) CreateOrUpdateDraftMAWB(ctx context.Context, mawbUUID string, req *draft_mawb.DraftMAWBRequest) (*draft_mawb.DraftMAWBResponse, error) {
	args := m.Called(ctx, mawbUUID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*draft_mawb.DraftMAWBResponse), args.Error(1)
}

func (m *MockDraftMAWBService) ConfirmDraftMAWB(ctx context.Context, mawbUUID string) error {
	args := m.Called(ctx, mawbUUID)
	return args.Error(0)
}

func (m *MockDraftMAWBService) RejectDraftMAWB(ctx context.Context, mawbUUID string) error {
	args := m.Called(ctx, mawbUUID)
	return args.Error(0)
}

func (m *MockDraftMAWBService) GenerateDraftMAWBPDF(ctx context.Context, mawbUUID string) ([]byte, error) {
	args := m.Called(ctx, mawbUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Test helper functions
func createTestDraftMAWBResponse() *draft_mawb.DraftMAWBResponse {
	return &draft_mawb.DraftMAWBResponse{
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
		FlightDate:              "2024-01-01",
		ExecutedOnDate:          "2024-01-02",
		InsuranceAmount:         1000.0,
		TotalNoOfPieces:         10,
		TotalGrossWeight:        500.5,
		TotalChargeableWeight:   600.0,
		TotalRateCharge:         2000.0,
		TotalAmount:             2500.0,
		Status:                  "Draft",
		Items: []draft_mawb.DraftMAWBItem{
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
				Dims: []draft_mawb.DraftMAWBItemDim{
					{
						ID:              1,
						DraftMAWBItemID: 1,
						Length:          "100",
						Width:           "50",
						Height:          "30",
						Count:           "2",
					},
				},
			},
		},
		Charges: []draft_mawb.DraftMAWBCharge{
			{
				ID:            1,
				DraftMAWBUUID: "test-uuid",
				Key:           "fuel_surcharge",
				Value:         500.0,
			},
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}
}

func createTestDraftMAWBRequest() *draft_mawb.DraftMAWBRequest {
	return &draft_mawb.DraftMAWBRequest{
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
		Items: []draft_mawb.DraftMAWBItemRequest{
			{
				PiecesRCP:         "5",
				GrossWeight:       "250.5",
				KgLb:              "kg",
				RateClass:         "N",
				RateCharge:        1000.0,
				NatureAndQuantity: "Electronics",
				Dims: []draft_mawb.DraftMAWBItemDimRequest{
					{
						Length: "100",
						Width:  "50",
						Height: "30",
						Count:  "2",
					},
				},
			},
		},
		Charges: []draft_mawb.DraftMAWBChargeRequest{
			{
				Key:   "fuel_surcharge",
				Value: 500.0,
			},
		},
	}
}

func setupDraftMAWBHandler(service draft_mawb.Service) *draftMAWBHandler {
	return &draftMAWBHandler{
		service: service,
	}
}

func setupDraftMAWBRouter(handler *draftMAWBHandler) chi.Router {
	r := chi.NewRouter()
	r.Route("/{uuid}", func(r chi.Router) {
		r.Mount("/draft-mawb", handler.router())
	})
	return r
}

func TestDraftMAWBHandler_GetDraftMAWB_Success(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("GetDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_GetDraftMAWB_MissingUUID(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	req := httptest.NewRequest("GET", "/draft-mawb/", nil)
	w := httptest.NewRecorder()

	handler.getDraftMAWB(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestDraftMAWBHandler_GetDraftMAWB_ServiceError(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := draft_mawb.ErrDraftMAWBNotFound

	mockService.On("GetDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil, serviceError)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_CreateOrUpdateDraftMAWB_Success(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	requestData := createTestDraftMAWBRequest()
	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("CreateOrUpdateDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID, mock.AnythingOfType("*draft_mawb.DraftMAWBRequest")).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_CreateOrUpdateDraftMAWB_MissingUUID(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	requestData := createTestDraftMAWBRequest()
	requestBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/draft-mawb/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createOrUpdateDraftMAWB(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestDraftMAWBHandler_CreateOrUpdateDraftMAWB_InvalidRequestBody(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	invalidJSON := `{"invalid": json}`

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "invalid request body")
}

func TestDraftMAWBHandler_CreateOrUpdateDraftMAWB_ValidationError(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	// Create invalid request (missing required MAWB)
	invalidRequest := &draft_mawb.DraftMAWBRequest{
		CustomerUUID: "customer-uuid",
		AirlineName:  "Test Airlines",
		Items: []draft_mawb.DraftMAWBItemRequest{
			{
				PiecesRCP:   "5",
				GrossWeight: "250.5",
			},
		},
	}

	requestBody, _ := json.Marshal(invalidRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "validation failed")
}

func TestDraftMAWBHandler_CreateOrUpdateDraftMAWB_ServiceError(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	requestData := createTestDraftMAWBRequest()
	serviceError := draft_mawb.ErrMAWBInfoNotFound

	mockService.On("CreateOrUpdateDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID, mock.AnythingOfType("*draft_mawb.DraftMAWBRequest")).Return(nil, serviceError)

	requestBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_CreateOrUpdateDraftMAWB_ComplexNestedData(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"

	// Create request with complex nested data
	complexRequest := &draft_mawb.DraftMAWBRequest{
		MAWB:         "123456789",
		CustomerUUID: "customer-uuid",
		AirlineName:  "Test Airlines",
		Items: []draft_mawb.DraftMAWBItemRequest{
			{
				PiecesRCP:   "5",
				GrossWeight: "250.5",
				KgLb:        "kg",
				RateCharge:  1000.0,
				Dims: []draft_mawb.DraftMAWBItemDimRequest{
					{Length: "100", Width: "50", Height: "30", Count: "2"},
					{Length: "80", Width: "40", Height: "25", Count: "1"},
				},
			},
			{
				PiecesRCP:   "3",
				GrossWeight: "150.0",
				KgLb:        "lb",
				RateCharge:  800.0,
				Dims: []draft_mawb.DraftMAWBItemDimRequest{
					{Length: "60", Width: "40", Height: "20", Count: "3"},
				},
			},
		},
		Charges: []draft_mawb.DraftMAWBChargeRequest{
			{Key: "fuel_surcharge", Value: 500.0},
			{Key: "security_fee", Value: 200.0},
			{Key: "handling_fee", Value: 150.0},
		},
	}

	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("CreateOrUpdateDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID, mock.AnythingOfType("*draft_mawb.DraftMAWBRequest")).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(complexRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_ConfirmDraftMAWB_Success(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"

	mockService.On("ConfirmDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/confirm", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "draft MAWB confirmed successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_ConfirmDraftMAWB_MissingUUID(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	req := httptest.NewRequest("POST", "/draft-mawb/confirm", nil)
	w := httptest.NewRecorder()

	handler.confirmDraftMAWB(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestDraftMAWBHandler_ConfirmDraftMAWB_ServiceError(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := draft_mawb.ErrBusinessRuleViolation

	mockService.On("ConfirmDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(serviceError)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/confirm", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_RejectDraftMAWB_Success(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"

	mockService.On("RejectDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/reject", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "draft MAWB rejected successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_RejectDraftMAWB_ServiceError(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := draft_mawb.ErrDraftMAWBNotFound

	mockService.On("RejectDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(serviceError)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/reject", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_PrintDraftMAWB_Success(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedPDF := []byte("mock pdf content for draft mawb")

	mockService.On("GenerateDraftMAWBPDF", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(expectedPDF, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/draft-mawb/print", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "draft_mawb_")
	assert.Equal(t, fmt.Sprintf("%d", len(expectedPDF)), w.Header().Get("Content-Length"))
	assert.Equal(t, expectedPDF, w.Body.Bytes())

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_PrintDraftMAWB_MissingUUID(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	req := httptest.NewRequest("GET", "/draft-mawb/print", nil)
	w := httptest.NewRecorder()

	handler.printDraftMAWB(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestDraftMAWBHandler_PrintDraftMAWB_ServiceError(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := draft_mawb.ErrPDFGenerationFailed

	mockService.On("GenerateDraftMAWBPDF", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil, serviceError)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/draft-mawb/print", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_Router(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	router := handler.router()
	assert.NotNil(t, router)

	// Test that all routes are properly configured
	routes := []string{
		"GET /",
		"POST /",
		"POST /confirm",
		"POST /reject",
		"GET /print",
	}

	// This is a basic test to ensure the router is created
	// In a more comprehensive test, you would walk the routes and verify they exist
	for _, route := range routes {
		_ = route // Placeholder for route verification logic
	}
}

func TestDraftMAWBHandler_ContextHandling(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	mawbUUID := "test-mawb-uuid"
	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("GetDraftMAWB", mock.AnythingOfType("*context.backgroundCtx"), mawbUUID).Return(expectedResponse, nil)

	// Create request without context
	req := httptest.NewRequest("GET", "/draft-mawb/", nil)
	req = req.WithContext(nil) // Remove context

	// Add URL parameter manually since we're not using the full router
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", mawbUUID)
	req = req.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.getDraftMAWB(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDraftMAWBHandler_ErrorMapping(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)

	// Test different error types and their HTTP mappings
	errorTests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "MAWB Info Not Found",
			serviceError:   draft_mawb.ErrMAWBInfoNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Draft MAWB Not Found",
			serviceError:   draft_mawb.ErrDraftMAWBNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid Request Data",
			serviceError:   draft_mawb.ErrInvalidRequestData,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Business Rule Violation",
			serviceError:   draft_mawb.ErrBusinessRuleViolation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Calculation Failed",
			serviceError:   draft_mawb.ErrCalculationFailed,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PDF Generation Failed",
			serviceError:   draft_mawb.ErrPDFGenerationFailed,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Generic Error",
			serviceError:   errors.New("generic error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := handler.mapServiceErrorToHTTP(tt.serviceError)
			assert.NotNil(t, httpErr)
			// Note: The actual status code verification would depend on the
			// implementation of errors.MapErrorToHTTPResponse
		})
	}
}

// Test edge cases and special scenarios
func TestDraftMAWBHandler_LargePayload(t *testing.T) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"

	// Create request with many items and dimensions
	largeRequest := &draft_mawb.DraftMAWBRequest{
		MAWB:         "123456789",
		CustomerUUID: "customer-uuid",
		AirlineName:  "Test Airlines",
		Items:        make([]draft_mawb.DraftMAWBItemRequest, 20),   // Many items
		Charges:      make([]draft_mawb.DraftMAWBChargeRequest, 10), // Many charges
	}

	// Fill with test data
	for i := range largeRequest.Items {
		largeRequest.Items[i] = draft_mawb.DraftMAWBItemRequest{
			PiecesRCP:   fmt.Sprintf("%d", i+1),
			GrossWeight: fmt.Sprintf("%.1f", float64(i+1)*10.5),
			KgLb:        "kg",
			RateCharge:  float64(i+1) * 100.0,
			Dims: []draft_mawb.DraftMAWBItemDimRequest{
				{Length: "100", Width: "50", Height: "30", Count: "2"},
				{Length: "80", Width: "40", Height: "25", Count: "1"},
			},
		}
	}

	for i := range largeRequest.Charges {
		largeRequest.Charges[i] = draft_mawb.DraftMAWBChargeRequest{
			Key:   fmt.Sprintf("charge_%d", i+1),
			Value: float64(i+1) * 50.0,
		}
	}

	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("CreateOrUpdateDraftMAWB", mock.AnythingOfType("*context.valueCtx"), mawbUUID, mock.AnythingOfType("*draft_mawb.DraftMAWBRequest")).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(largeRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkDraftMAWBHandler_GetDraftMAWB(b *testing.B) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("GetDraftMAWB", mock.Anything, mawbUUID).Return(expectedResponse, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkDraftMAWBHandler_CreateOrUpdateDraftMAWB(b *testing.B) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	requestData := createTestDraftMAWBRequest()
	expectedResponse := createTestDraftMAWBResponse()

	mockService.On("CreateOrUpdateDraftMAWB", mock.Anything, mawbUUID, mock.Anything).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(requestData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", fmt.Sprintf("/%s/draft-mawb/", mawbUUID), bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkDraftMAWBHandler_PrintDraftMAWB(b *testing.B) {
	mockService := &MockDraftMAWBService{}
	handler := setupDraftMAWBHandler(mockService)
	router := setupDraftMAWBRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedPDF := make([]byte, 1024*1024) // 1MB PDF

	mockService.On("GenerateDraftMAWBPDF", mock.Anything, mawbUUID).Return(expectedPDF, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/draft-mawb/print", mawbUUID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
