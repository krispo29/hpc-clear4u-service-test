package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	seawaybill "hpc-express-service/inbound/seawaybill"
)

type seaWaybillHandler struct {
	s seawaybill.Service
}

func (h *seaWaybillHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.getAllSeaWaybillDetails)
	r.Post("/", h.createSeaWaybillDetail)
	r.Get("/{uuid}", h.getSeaWaybillDetail)
	r.Put("/{uuid}", h.updateSeaWaybillDetail)
	r.Delete("/{uuid}/attachments", h.deleteAttachment)

	return r
}

func (h *seaWaybillHandler) createSeaWaybillDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %w", err)))
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	data := &seawaybill.CreateSeaWaybillDetailRequest{
		GrossWeight:  r.FormValue("grossWeight"),
		VolumeWeight: r.FormValue("volumeWeight"),
		DutyTax:      r.FormValue("dutyTax"),
	}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, ok := r.MultipartForm.File["attachments"]; ok {
			data.Attachments = files
		}
	}

	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	result, err := h.s.CreateSeaWaybillDetail(ctx, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillHandler) getSeaWaybillDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	result, err := h.s.GetSeaWaybillDetail(ctx, uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillHandler) getAllSeaWaybillDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := h.s.GetSeaWaybillDetails(ctx)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillHandler) updateSeaWaybillDetail(w http.ResponseWriter, r *http.Request) {
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
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %w", err)))
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	data := &seawaybill.UpdateSeaWaybillDetailRequest{
		GrossWeight:  r.FormValue("grossWeight"),
		VolumeWeight: r.FormValue("volumeWeight"),
		DutyTax:      r.FormValue("dutyTax"),
	}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, ok := r.MultipartForm.File["attachments"]; ok {
			data.Attachments = files
		}
	}

	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	result, err := h.s.UpdateSeaWaybillDetail(ctx, uuid, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *seaWaybillHandler) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	request := &seawaybill.DeleteAttachmentRequest{}
	if err := render.Bind(r, request); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := h.s.DeleteAttachment(ctx, uuid, request.FileName); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}
