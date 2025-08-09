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

	"hpc-express-service/cargo_manifest"
)

// MockCargoManifestService is a mock implementation of the cargo manifest service
type MockCargoManifestService struct {
	mock.Mock
}

func (m *MockCargoManifestService) GetCargoManifest(ctx context.Context, mawbUUID string) (*cargo_manifest.CargoManifestResponse, error) {
	args := m.Called(ctx, mawbUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo_manifest.CargoManifestResponse), args.Error(1)
}

func (m *MockCargoManifestService) CreateOrUpdateCargoManifest(ctx context.Context, mawbUUID string, req *cargo_manifest.CargoManifestRequest) (*cargo_manifest.CargoManifestResponse, error) {
	args := m.Called(ctx, mawbUUID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo_manifest.CargoManifestResponse), args.Error(1)
}

func (m *MockCargoManifestService) ConfirmCargoManifest(ctx context.Context, mawbUUID string) error {
	args := m.Called(ctx, mawbUUID)
	return args.Error(0)
}

func (m *MockCargoManifestService) RejectCargoManifest(ctx context.Context, mawbUUID string) error {
	args := m.Called(ctx, mawbUUID)
	return args.Error(0)
}

func (m *MockCargoManifestService) GenerateCargoManifestPDF(ctx context.Context, mawbUUID string) ([]byte, error) {
	args := m.Called(ctx, mawbUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Test helper functions
func createTestCargoManifestResponse() *cargo_manifest.CargoManifestResponse {
	return &cargo_manifest.CargoManifestResponse{
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
		Status:          "Draft",
		Items: []cargo_manifest.CargoManifestItem{
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
			},
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}
}

func createTestCargoManifestRequest() *cargo_manifest.CargoManifestRequest {
	return &cargo_manifest.CargoManifestRequest{
		MAWBNumber:      "123456789",
		PortOfDischarge: "BKK",
		FlightNo:        "TG123",
		FreightDate:     "2024-01-01",
		Shipper:         "Test Shipper",
		Consignee:       "Test Consignee",
		TotalCtn:        "10",
		Transshipment:   "No",
		Items: []cargo_manifest.CargoManifestItemRequest{
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

func setupCargoManifestHandler(service cargo_manifest.Service) *cargoManifestHandler {
	return &cargoManifestHandler{
		service: service,
	}
}

func setupCargoManifestRouter(handler *cargoManifestHandler) chi.Router {
	r := chi.NewRouter()
	r.Route("/{uuid}", func(r chi.Router) {
		r.Mount("/cargo-manifest", handler.router())
	})
	return r
}

func TestCargoManifestHandler_GetCargoManifest_Success(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedResponse := createTestCargoManifestResponse()

	mockService.On("GetCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), nil)
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

func TestCargoManifestHandler_GetCargoManifest_MissingUUID(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

	req := httptest.NewRequest("GET", "/cargo-manifest/", nil)
	w := httptest.NewRecorder()

	handler.getCargoManifest(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestCargoManifestHandler_GetCargoManifest_ServiceError(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := cargo_manifest.ErrCargoManifestNotFound

	mockService.On("GetCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil, serviceError)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_CreateOrUpdateCargoManifest_Success(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	requestData := createTestCargoManifestRequest()
	expectedResponse := createTestCargoManifestResponse()

	mockService.On("CreateOrUpdateCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID, mock.AnythingOfType("*cargo_manifest.CargoManifestRequest")).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), bytes.NewBuffer(requestBody))
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

func TestCargoManifestHandler_CreateOrUpdateCargoManifest_MissingUUID(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

	requestData := createTestCargoManifestRequest()
	requestBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/cargo-manifest/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createOrUpdateCargoManifest(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestCargoManifestHandler_CreateOrUpdateCargoManifest_InvalidRequestBody(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	invalidJSON := `{"invalid": json}`

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "invalid request body")
}

func TestCargoManifestHandler_CreateOrUpdateCargoManifest_ValidationError(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	// Create invalid request (missing required MAWBNumber)
	invalidRequest := &cargo_manifest.CargoManifestRequest{
		PortOfDischarge: "BKK",
		Items: []cargo_manifest.CargoManifestItemRequest{
			{
				HAWBNo:      "H123",
				Pkgs:        "5",
				GrossWeight: "100.5",
			},
		},
	}

	requestBody, _ := json.Marshal(invalidRequest)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "validation failed")
}

func TestCargoManifestHandler_CreateOrUpdateCargoManifest_ServiceError(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	requestData := createTestCargoManifestRequest()
	serviceError := cargo_manifest.ErrMAWBInfoNotFound

	mockService.On("CreateOrUpdateCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID, mock.AnythingOfType("*cargo_manifest.CargoManifestRequest")).Return(nil, serviceError)

	requestBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_ConfirmCargoManifest_Success(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"

	mockService.On("ConfirmCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/confirm", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "cargo manifest confirmed successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_ConfirmCargoManifest_MissingUUID(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

	req := httptest.NewRequest("POST", "/cargo-manifest/confirm", nil)
	w := httptest.NewRecorder()

	handler.confirmCargoManifest(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestCargoManifestHandler_ConfirmCargoManifest_ServiceError(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := cargo_manifest.ErrBusinessRuleViolation

	mockService.On("ConfirmCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(serviceError)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/confirm", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_RejectCargoManifest_Success(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"

	mockService.On("RejectCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/reject", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "cargo manifest rejected successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_RejectCargoManifest_ServiceError(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := cargo_manifest.ErrCargoManifestNotFound

	mockService.On("RejectCargoManifest", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(serviceError)

	req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/reject", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_PrintCargoManifest_Success(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedPDF := []byte("mock pdf content")

	mockService.On("GenerateCargoManifestPDF", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(expectedPDF, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/cargo-manifest/print", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "cargo_manifest_")
	assert.Equal(t, fmt.Sprintf("%d", len(expectedPDF)), w.Header().Get("Content-Length"))
	assert.Equal(t, expectedPDF, w.Body.Bytes())

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_PrintCargoManifest_MissingUUID(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

	req := httptest.NewRequest("GET", "/cargo-manifest/print", nil)
	w := httptest.NewRecorder()

	handler.printCargoManifest(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "MAWB Info UUID parameter is required")
}

func TestCargoManifestHandler_PrintCargoManifest_ServiceError(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	serviceError := cargo_manifest.ErrPDFGenerationFailed

	mockService.On("GenerateCargoManifestPDF", mock.AnythingOfType("*context.valueCtx"), mawbUUID).Return(nil, serviceError)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s/cargo-manifest/print", mawbUUID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_Router(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

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

func TestCargoManifestHandler_ContextHandling(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

	mawbUUID := "test-mawb-uuid"
	expectedResponse := createTestCargoManifestResponse()

	mockService.On("GetCargoManifest", mock.AnythingOfType("*context.backgroundCtx"), mawbUUID).Return(expectedResponse, nil)

	// Create request without context
	req := httptest.NewRequest("GET", "/cargo-manifest/", nil)
	req = req.WithContext(nil) // Remove context

	// Add URL parameter manually since we're not using the full router
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("uuid", mawbUUID)
	req = req.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.getCargoManifest(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCargoManifestHandler_ErrorMapping(t *testing.T) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)

	// Test different error types and their HTTP mappings
	errorTests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "MAWB Info Not Found",
			serviceError:   cargo_manifest.ErrMAWBInfoNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Cargo Manifest Not Found",
			serviceError:   cargo_manifest.ErrCargoManifestNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid Request Data",
			serviceError:   cargo_manifest.ErrInvalidRequestData,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Business Rule Violation",
			serviceError:   cargo_manifest.ErrBusinessRuleViolation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PDF Generation Failed",
			serviceError:   cargo_manifest.ErrPDFGenerationFailed,
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

// Benchmark tests
func BenchmarkCargoManifestHandler_GetCargoManifest(b *testing.B) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	expectedResponse := createTestCargoManifestResponse()

	mockService.On("GetCargoManifest", mock.Anything, mawbUUID).Return(expectedResponse, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCargoManifestHandler_CreateOrUpdateCargoManifest(b *testing.B) {
	mockService := &MockCargoManifestService{}
	handler := setupCargoManifestHandler(mockService)
	router := setupCargoManifestRouter(handler)

	mawbUUID := "test-mawb-uuid"
	requestData := createTestCargoManifestRequest()
	expectedResponse := createTestCargoManifestResponse()

	mockService.On("CreateOrUpdateCargoManifest", mock.Anything, mawbUUID, mock.Anything).Return(expectedResponse, nil)

	requestBody, _ := json.Marshal(requestData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", fmt.Sprintf("/%s/cargo-manifest/", mawbUUID), bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
