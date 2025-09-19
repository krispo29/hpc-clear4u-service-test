package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"hpc-express-service/inbound/seawaybilldetails"
)

type seaWaybillDetailsHandler struct {
	s seawaybilldetails.Service
}

func (h *seaWaybillDetailsHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createSeaWaybillDetails)
	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.getSeaWaybillDetails)
		r.Put("/", h.updateSeaWaybillDetails)
		r.Delete("/attachments", h.deleteSeaWaybillAttachment)
	})
	return r
}

func (h *seaWaybillDetailsHandler) createSeaWaybillDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %v", err)))
		return
	}

	data := &seawaybilldetails.UpsertSeaWaybillDetailsRequest{
		GrossWeight:  r.FormValue("grossWeight"),
		VolumeWeight: r.FormValue("volumeWeight"),
		DutyTax:      r.FormValue("dutyTax"),
	}
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["attachments"]; exists {
			data.Attachments = files
		}
	}

	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	result, err := h.s.CreateSeaWaybillDetails(ctx, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillDetailsHandler) getSeaWaybillDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	result, err := h.s.GetSeaWaybillDetails(ctx, uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillDetailsHandler) updateSeaWaybillDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %v", err)))
		return
	}

	data := &seawaybilldetails.UpsertSeaWaybillDetailsRequest{
		GrossWeight:  r.FormValue("grossWeight"),
		VolumeWeight: r.FormValue("volumeWeight"),
		DutyTax:      r.FormValue("dutyTax"),
	}
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["attachments"]; exists {
			data.Attachments = files
		}
	}

	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	result, err := h.s.UpdateSeaWaybillDetails(ctx, uuid, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillDetailsHandler) deleteSeaWaybillAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	data := &seawaybilldetails.DeleteAttachmentRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := h.s.DeleteSeaWaybillAttachment(ctx, uuid, data.FileName); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}
