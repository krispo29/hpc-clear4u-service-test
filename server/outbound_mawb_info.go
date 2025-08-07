package server

import (
	"context"
	"errors"
	outbound "hpc-express-service/outbound/mawb"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

func (h *outboundMawbHandler) createMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	data := &outbound.CreateMawbInfo{}
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

	uuid, err := h.s.CreateMawnInfo(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(uuid, "success"))

}

func (h *outboundMawbHandler) getAllMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
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

	result, err := h.s.GetAllMawnInfo(r.Context(), start, end)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))

}

func (h *outboundMawbHandler) getOneMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	uuid := chi.URLParam(r, "uuid")

	result, err := h.s.GetOneMawnInfo(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))

}

func (h *outboundMawbHandler) updateMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	data := &outbound.UpdateMawbInfoModel{}
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

	err = h.s.UpdateMawnInfo(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))

}

func (h *outboundMawbHandler) deleteMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	uuid := chi.URLParam(r, "uuid")

	err := h.s.DeleteMawnInfo(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))

}

func (h *outboundMawbHandler) uploadAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	r.ParseMultipartForm(8 << 20) // 10 * 2^20

	attachmentFile, handler, err := r.FormFile("attachmentFile")

	var fileOriginName string
	var fileBytes []byte
	// var attachmentFileUrl string
	if err == nil {
		defer attachmentFile.Close()

		attachmentFileBytes, err := ioutil.ReadAll(attachmentFile)
		if err == nil {
			fileOriginName = handler.Filename
			fileBytes = attachmentFileBytes
		}
	} else {
		render.Render(w, r, ErrInvalidRequest(errors.New("file invalid")))
		return
	}

	uuid := chi.URLParam(r, "uuid")
	err = h.s.UploadAttachment(r.Context(), uuid, fileOriginName, fileBytes)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))

}
