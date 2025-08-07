package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/customer"
)

type customerHandler struct {
	s customer.Service
}

func (h *customerHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.getAll)
	r.Get("/dropdown", h.getDropdown)

	return r
}

func (h *customerHandler) getAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	result, err := h.s.GetAll(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *customerHandler) getDropdown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	result, err := h.s.GetAllDropdown(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
