package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/cargo_manifest"
	"hpc-express-service/common"
	"hpc-express-service/errors"
	"hpc-express-service/middleware"
)

type cargoManifestHandler struct {
	service cargo_manifest.Service
}

func (h *cargoManifestHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.getCargoManifest)
	r.Post("/", h.createOrUpdateCargoManifest)
	r.Post("/confirm", h.confirmCargoManifest)
	r.Post("/reject", h.rejectCargoManifest)
	r.Get("/print", h.printCargoManifest)
	return r
}

// getCargoManifest handles GET requests to retrieve cargo manifest data
func (h *cargoManifestHandler) getCargoManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get MAWB UUID from URL parameter (passed from parent route)
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("MAWB Info UUID parameter is required")))
		return
	}

	// Check permissions
	if err := common.CheckCargoManifestPermission(ctx, "view"); err != nil {
		render.Render(w, r, ErrUnauthorized(err))
		return
	}

	// Get user ID for audit logging
	userID := GetUserUUIDFromContext(r)

	// Call service to get cargo manifest
	result, err := h.service.GetCargoManifest(ctx, mawbUUID)
	if err != nil {
		// Audit log the failed access
		common.AuditCargoManifestAccess(r, userID, mawbUUID, false, err)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to get cargo manifest", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Audit log the successful access
	common.AuditCargoManifestAccess(r, userID, mawbUUID, true, nil)

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

// createOrUpdateCargoManifest handles POST requests to create or update cargo manifest
func (h *cargoManifestHandler) createOrUpdateCargoManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get MAWB UUID from URL parameter (passed from parent route)
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("MAWB Info UUID parameter is required")))
		return
	}

	// Check permissions
	if err := common.CheckCargoManifestPermission(ctx, "update"); err != nil {
		render.Render(w, r, ErrUnauthorized(err))
		return
	}

	// Get user ID for audit logging
	userID := GetUserUUIDFromContext(r)

	// Bind request data
	data := &cargo_manifest.CargoManifestRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	// Sanitize input data first
	cargo_manifest.SanitizeCargoManifestRequest(data)

	// Validate request data using comprehensive validation
	if validationErrors := cargo_manifest.ValidateCargoManifestRequest(data); len(validationErrors) > 0 {
		// Create a detailed validation error response
		errorMessages := make([]string, len(validationErrors))
		for i, valErr := range validationErrors {
			errorMessages[i] = fmt.Sprintf("%s: %s", valErr.Field, valErr.Message)
		}
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))))
		return
	}

	// Prepare audit metadata
	changes := map[string]interface{}{
		"mawb_number":       data.MAWBNumber,
		"port_of_discharge": data.PortOfDischarge,
		"flight_no":         data.FlightNo,
		"items_count":       len(data.Items),
	}

	// Call service to create or update cargo manifest
	result, err := h.service.CreateOrUpdateCargoManifest(ctx, mawbUUID, data)
	if err != nil {
		// Audit log the failed modification
		common.AuditCargoManifestModification(r, userID, mawbUUID, "create_or_update", false, err, changes)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to create or update cargo manifest", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Audit log the successful modification
	common.AuditCargoManifestModification(r, userID, mawbUUID, "create_or_update", true, nil, changes)

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

// confirmCargoManifest handles POST requests to confirm cargo manifest status
func (h *cargoManifestHandler) confirmCargoManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get MAWB UUID from URL parameter (passed from parent route)
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("MAWB Info UUID parameter is required")))
		return
	}

	// Check permissions (requires supervisor or admin role)
	if err := common.CheckCargoManifestPermission(ctx, "confirm"); err != nil {
		render.Render(w, r, ErrUnauthorized(err))
		return
	}

	// Get user ID for audit logging
	userID := GetUserUUIDFromContext(r)

	// Call service to confirm cargo manifest
	err := h.service.ConfirmCargoManifest(ctx, mawbUUID)
	if err != nil {
		// Audit log the failed status change
		common.AuditCargoManifestStatusChange(r, userID, mawbUUID, "unknown", "confirmed", false, err)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to confirm cargo manifest", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Audit log the successful status change
	common.AuditCargoManifestStatusChange(r, userID, mawbUUID, "pending", "confirmed", true, nil)

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "cargo manifest confirmed successfully"))
}

// rejectCargoManifest handles POST requests to reject cargo manifest status
func (h *cargoManifestHandler) rejectCargoManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get MAWB UUID from URL parameter (passed from parent route)
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("MAWB Info UUID parameter is required")))
		return
	}

	// Check permissions (requires supervisor or admin role)
	if err := common.CheckCargoManifestPermission(ctx, "reject"); err != nil {
		render.Render(w, r, ErrUnauthorized(err))
		return
	}

	// Get user ID for audit logging
	userID := GetUserUUIDFromContext(r)

	// Call service to reject cargo manifest
	err := h.service.RejectCargoManifest(ctx, mawbUUID)
	if err != nil {
		// Audit log the failed status change
		common.AuditCargoManifestStatusChange(r, userID, mawbUUID, "unknown", "rejected", false, err)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to reject cargo manifest", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Audit log the successful status change
	common.AuditCargoManifestStatusChange(r, userID, mawbUUID, "pending", "rejected", true, nil)

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "cargo manifest rejected successfully"))
}

// printCargoManifest handles GET requests to generate and return PDF documents
func (h *cargoManifestHandler) printCargoManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get MAWB UUID from URL parameter (passed from parent route)
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("MAWB Info UUID parameter is required")))
		return
	}

	// Check permissions
	if err := common.CheckCargoManifestPermission(ctx, "print"); err != nil {
		render.Render(w, r, ErrUnauthorized(err))
		return
	}

	// Get user ID for audit logging
	userID := GetUserUUIDFromContext(r)

	// Get cargo manifest data first
	cargoManifest, err := h.service.GetCargoManifest(ctx, mawbUUID)
	if err != nil {
		// Audit log the failed PDF generation
		common.AuditPDFGeneration(r, userID, "cargo_manifest", mawbUUID, false, err, 0)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to get cargo manifest for PDF generation", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Generate secure PDF
	pdfData, err := common.GlobalSecurePDFGenerator.GenerateSecurePDF(ctx, cargoManifest, "cargo_manifest", userID)
	if err != nil {
		// Audit log the failed PDF generation
		common.AuditPDFGeneration(r, userID, "cargo_manifest", mawbUUID, false, err, 0)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to generate cargo manifest PDF", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Generate secure filename
	filename := fmt.Sprintf("cargo_manifest_%s.pdf", mawbUUID)

	// Send secure PDF response
	if err := common.SecurePDFResponse(w, r, pdfData, filename); err != nil {
		// Audit log the failed PDF generation
		common.AuditPDFGeneration(r, userID, "cargo_manifest", mawbUUID, false, err, int64(len(pdfData)))

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to send PDF response", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Audit log the successful PDF generation
	common.AuditPDFGeneration(r, userID, "cargo_manifest", mawbUUID, true, nil, int64(len(pdfData)))
}

// mapServiceErrorToHTTP maps service layer errors to appropriate HTTP error responses
func (h *cargoManifestHandler) mapServiceErrorToHTTP(err error) render.Renderer {
	// Use centralized error mapping
	return errors.MapErrorToHTTPResponse(err)
}
