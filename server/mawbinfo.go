package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"hpc-express-service/cargo_manifest"
	"hpc-express-service/draft_mawb"
	"hpc-express-service/outbound/mawbinfo"
)

type mawbInfoHandler struct {
	s                mawbinfo.Service
	cargoManifestSvc cargo_manifest.Service
	draftMAWBSvc     draft_mawb.Service
}

// newMawbInfoHandler creates a new mawbInfoHandler with optional services
func newMawbInfoHandler(mawbInfoSvc mawbinfo.Service, cargoManifestSvc cargo_manifest.Service, draftMAWBSvc draft_mawb.Service) *mawbInfoHandler {
	return &mawbInfoHandler{
		s:                mawbInfoSvc,
		cargoManifestSvc: cargoManifestSvc,
		draftMAWBSvc:     draftMAWBSvc,
	}
}

func (h *mawbInfoHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createMawbInfo)
	r.Get("/", h.getAllMawbInfo)
	r.Get("/{uuid}", h.getMawbInfo)
	r.Put("/{uuid}", h.updateMawbInfo)
	r.Delete("/{uuid}", h.deleteMawbInfo)
	r.Delete("/{uuid}/attachments", h.deleteMawbInfoAttachment)

	// Add sub-routes for cargo manifest and draft MAWB
	r.Route("/{uuid}", func(r chi.Router) {
		// Add UUID validation middleware
		r.Use(h.validateUUIDMiddleware)

		// Mount cargo manifest sub-routes
		if h.cargoManifestSvc != nil {
			cargoManifestHandler := &cargoManifestHandler{service: h.cargoManifestSvc}
			r.Mount("/cargo-manifest", cargoManifestHandler.router())
		}

		// Mount draft MAWB sub-routes
		if h.draftMAWBSvc != nil {
			draftMAWBHandler := &draftMAWBHandler{service: h.draftMAWBSvc}
			r.Mount("/draft-mawb", draftMAWBHandler.router())
		}
	})

	return r
}

func (h *mawbInfoHandler) createMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Bind request data
	data := &mawbinfo.CreateMawbInfoRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate request data using validator
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Call service to create MAWB info
	result, err := h.s.CreateMawbInfo(ctx, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}
func (h *mawbInfoHandler) getMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Call service to get MAWB info
	result, err := h.s.GetMawbInfo(ctx, uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *mawbInfoHandler) getAllMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get query parameters for date filtering
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	// Call service to get all MAWB info with date filters
	result, err := h.s.GetAllMawbInfo(ctx, startDate, endDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}
func (h *mawbInfoHandler) updateMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Parse multipart form data (max 32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %v", err)))
		return
	}

	// Extract form data
	data := &mawbinfo.UpdateMawbInfoRequest{
		ChargeableWeight: r.FormValue("chargeableWeight"),
		Date:             r.FormValue("date"),
		Mawb:             r.FormValue("mawb"),
		ServiceType:      r.FormValue("serviceType"),
		ShippingType:     r.FormValue("shippingType"),
	}

	// Get file attachments
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["attachments"]; exists {
			data.Attachments = files
		}
	}

	// Validate request data using validator
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Call service to update MAWB info
	result, err := h.s.UpdateMawbInfo(ctx, uuid, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *mawbInfoHandler) deleteMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Call service to delete MAWB info
	err := h.s.DeleteMawbInfo(ctx, uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "deleted successfully"))
}

func (h *mawbInfoHandler) deleteMawbInfoAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Bind request data for the file to be deleted
	data := &mawbinfo.DeleteAttachmentRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate request data
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Call service to delete the attachment
	err := h.s.DeleteMawbInfoAttachment(ctx, uuid, data.FileName)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "attachment deleted successfully"))
}

// validateUUIDMiddleware validates the UUID parameter format
func (h *mawbInfoHandler) validateUUIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			render.Render(w, r, ErrInvalidRequest(fmt.Errorf("UUID parameter is required")))
			return
		}

		// Basic UUID format validation (36 characters with hyphens)
		if len(uuid) != 36 {
			render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid UUID format")))
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}
