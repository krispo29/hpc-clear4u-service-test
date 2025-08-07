package server

import (
	"context"
	"hpc-express-service/common"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type commonHandler struct {
	s common.Service
}

func (h *commonHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Get("/exchange_rates", h.getAllExchangeRates)
	r.Get("/convert_templates", h.getAllConvertTemplates)

	return r
}

func (h *commonHandler) getAllExchangeRates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	result, err := h.s.GetAllExchangeRates(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *commonHandler) getAllConvertTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	category := r.URL.Query().Get("category")
	result, err := h.s.GetAllConvertTemplates(r.Context(), category)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
