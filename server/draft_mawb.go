package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/common"
	"hpc-express-service/draft_mawb"
	"hpc-express-service/errors"
	"hpc-express-service/middleware"
)

type draftMAWBHandler struct {
	service draft_mawb.Service
}

func (h *draftMAWBHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.getDraftMAWB)
	r.Post("/", h.createOrUpdateDraftMAWB)
	r.Post("/confirm", h.confirmDraftMAWB)
	r.Post("/reject", h.rejectDraftMAWB)
	r.Get("/print", h.printDraftMAWB)
	return r
}

// getDraftMAWB handles GET requests to retrieve draft MAWB data
func (h *draftMAWBHandler) getDraftMAWB(w http.ResponseWriter, r *http.Request) {
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
	if err := common.CheckDraftMAWBPermission(ctx, "view"); err != nil {
		render.Render(w, r, ErrUnauthorized(err))
		return
	}

	// Get user ID for audit logging
	userID := GetUserUUIDFromContext(r)

	// Call service to get draft MAWB
	result, err := h.service.GetDraftMAWB(ctx, mawbUUID)
	if err != nil {
		// Audit log the failed access
		common.AuditDraftMAWBAccess(r, userID, mawbUUID, false, err)

		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to get draft MAWB", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Audit log the successful access
	common.AuditDraftMAWBAccess(r, userID, mawbUUID, true, nil)

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

// createOrUpdateDraftMAWB handles POST requests to create or update draft MAWB
func (h *draftMAWBHandler) createOrUpdateDraftMAWB(w http.ResponseWriter, r *http.Request) {
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

	// Bind request data
	data := &draft_mawb.DraftMAWBRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	// Sanitize input data first
	draft_mawb.SanitizeDraftMAWBRequest(data)

	// Validate request data using comprehensive validation
	if validationErrors := draft_mawb.ValidateDraftMAWBRequest(data); len(validationErrors) > 0 {
		// Create a detailed validation error response
		errorMessages := make([]string, len(validationErrors))
		for i, valErr := range validationErrors {
			errorMessages[i] = fmt.Sprintf("%s: %s", valErr.Field, valErr.Message)
		}
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))))
		return
	}

	// Call service to create or update draft MAWB
	result, err := h.service.CreateOrUpdateDraftMAWB(ctx, mawbUUID, data)
	if err != nil {
		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to create or update draft MAWB", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

// confirmDraftMAWB handles POST requests to confirm draft MAWB status
func (h *draftMAWBHandler) confirmDraftMAWB(w http.ResponseWriter, r *http.Request) {
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

	// Call service to confirm draft MAWB
	err := h.service.ConfirmDraftMAWB(ctx, mawbUUID)
	if err != nil {
		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to confirm draft MAWB", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "draft MAWB confirmed successfully"))
}

// rejectDraftMAWB handles POST requests to reject draft MAWB status
func (h *draftMAWBHandler) rejectDraftMAWB(w http.ResponseWriter, r *http.Request) {
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

	// Call service to reject draft MAWB
	err := h.service.RejectDraftMAWB(ctx, mawbUUID)
	if err != nil {
		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to reject draft MAWB", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "draft MAWB rejected successfully"))
}

// printDraftMAWB handles GET requests to generate and return PDF documents
func (h *draftMAWBHandler) printDraftMAWB(w http.ResponseWriter, r *http.Request) {
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

	// Call service to generate PDF
	pdfData, err := h.service.GenerateDraftMAWBPDF(ctx, mawbUUID)
	if err != nil {
		// Log error with context
		middleware.LogErrorWithContext(r, err, "Failed to generate draft MAWB PDF", map[string]interface{}{
			"mawbUUID": mawbUUID,
		})

		// Map service errors to appropriate HTTP responses
		httpErr := h.mapServiceErrorToHTTP(err)
		render.Render(w, r, httpErr)
		return
	}

	// Set proper headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=draft_mawb_%s.pdf", mawbUUID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))

	// Write PDF data to response
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}

// mapServiceErrorToHTTP maps service layer errors to appropriate HTTP error responses
func (h *draftMAWBHandler) mapServiceErrorToHTTP(err error) render.Renderer {
	// Use centralized error mapping
	return errors.MapErrorToHTTPResponse(err)
}
