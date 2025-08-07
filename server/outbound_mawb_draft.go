package server

import (
	"context"
	"errors"
	outbound "hpc-express-service/outbound/mawb"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

func (h *outboundMawbHandler) getAllMawbDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	var start, end string
	if r.URL.Query().Get("start") == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("required start date")))
		return
	}

	if r.URL.Query().Get("end") == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("required end date")))
		return
	}

	start = r.URL.Query().Get("start")
	end = r.URL.Query().Get("end")

	result, err := h.s.GetAllMawbDraft(r.Context(), start, end)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *outboundMawbHandler) getOneMawbDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")

	result, err := h.s.GetOneMawbDraft(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *outboundMawbHandler) PrintMawbDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uuid := chi.URLParam(r, "uuid")

	buf, err := h.s.PrintMawbDraft(r.Context(), uuid)
	if err == nil {
		w.Header().Set("Content-Type", "application/pdf")
		buf.WriteTo(w)
	} else {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	return
}

func (h *outboundMawbHandler) previewMawbDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	data := &outbound.RequestDraftModel{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	buf, err := h.s.PreviewDraftMawb(r.Context(), data)
	if err == nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/pdf")
		buf.WriteTo(w)
	} else {
		log.Println(err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	return
}

func (h *outboundMawbHandler) saveMawbDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	data := &outbound.RequestDraftModel{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate Data
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	buf, err := h.s.SaveDraftMawb(r.Context(), data)
	if err == nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/pdf")
		buf.WriteTo(w)
	} else {
		log.Println(err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	return
}

func (h *outboundMawbHandler) updateMawbDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	data := &outbound.RequestUpdateMawbDraftModel{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate Data
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	buf, err := h.s.UpdateDraftMawb(r.Context(), data)
	if err == nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/pdf")
		buf.WriteTo(w)
	} else {
		log.Println(err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	return
}
