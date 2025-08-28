package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"hpc-express-service/setting"
)

type settingHandler struct {
	s         setting.Service
	statusSvc setting.MasterStatusService
}

func (h *settingHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Route("/hscode", func(r chi.Router) {
		r.Post("/", h.createHsCode)
		r.Get("/", h.getAllHsCode)
		r.Get("/export", h.exportHsCode)
		r.Get("/{uuid}", h.getOneHsCode)
		r.Put("/", h.updateHsCode)
		r.Patch("/{uuid}", h.updateStatusHsCode)
	})

	r.Route("/master-status", func(r chi.Router) {
		r.Post("/", h.createMasterStatus)
		r.Get("/", h.getAllMasterStatuses)
		r.Get("/{uuid}", h.getOneMasterStatus)
		r.Put("/", h.updateMasterStatus)
		r.Delete("/{uuid}", h.deleteMasterStatus)
	})

	return r
}

func (h *settingHandler) createMasterStatus(w http.ResponseWriter, r *http.Request) {
	data := &setting.MasterStatus{}
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

	createdStatus, err := h.statusSvc.CreateMasterStatus(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(createdStatus, "success"))
}

func (h *settingHandler) getAllMasterStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.statusSvc.GetAllMasterStatuses(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(statuses, "success"))
}

func (h *settingHandler) getOneMasterStatus(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	status, err := h.statusSvc.GetMasterStatusByUUID(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(status, "success"))
}

func (h *settingHandler) updateMasterStatus(w http.ResponseWriter, r *http.Request) {
	data := &setting.MasterStatus{}
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

	updatedStatus, err := h.statusSvc.UpdateMasterStatus(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(updatedStatus, "success"))
}

func (h *settingHandler) deleteMasterStatus(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	err := h.statusSvc.DeleteMasterStatus(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}

func (h *settingHandler) createHsCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	data := &setting.CreateHsCodeModel{}
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

	uuid, err := h.s.CreateHsCode(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(uuid, "success"))

}

func (h *settingHandler) getAllHsCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	result, err := h.s.GetAllHsCode(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))

}

func (h *settingHandler) getOneHsCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	uuid := chi.URLParam(r, "uuid")

	result, err := h.s.GetHsCodeByUUID(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))

}

func (h *settingHandler) updateHsCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	data := &setting.UpdateHsCodeModel{}
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

	err = h.s.UpdateHsCode(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))

}

func (h *settingHandler) updateStatusHsCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	uuid := chi.URLParam(r, "uuid")

	err := h.s.UpdateStatusHsCode(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))

}

func (h *settingHandler) exportHsCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	excelBuf, err := h.s.ExportHsCode(ctx)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	fileName := fmt.Sprintf("master_hs_code_%v.xlsx", time.Now().Format("20060102"))

	// Send ZIP file as response
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("File-Name", fileName)
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	w.Write(excelBuf.Bytes())
}
