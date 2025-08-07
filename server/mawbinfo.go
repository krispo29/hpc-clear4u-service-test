package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"hpc-express-service/mawbinfo"
)

type mawbInfoHandler struct {
	s mawbinfo.Service
}

func (h *mawbInfoHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createMawbInfo)
	r.Get("/", h.getAllMawbInfo)
	r.Get("/{uuid}", h.getMawbInfo)
	r.Put("/{uuid}", h.updateMawbInfo)
	r.Delete("/{uuid}", h.deleteMawbInfo)
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
