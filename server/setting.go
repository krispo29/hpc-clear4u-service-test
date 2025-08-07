package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"hpc-express-service/setting"
)

type settingHandler struct {
	s setting.Service
}

func (h *settingHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Route("/hscode", func(r chi.Router) {
		r.Post("/", h.createHsCode)
		r.Get("/", h.getAllHsCode)
		r.Get("/{uuid}", h.getOneHsCode)
		r.Put("/", h.updateHsCode)
		r.Patch("/{uuid}", h.updateStatusHsCode)
	})

	return r
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
