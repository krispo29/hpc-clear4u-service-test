package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/dashboard"
)

type dashboardHandler struct {
	s dashboard.Service
}

func (h *dashboardHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Get("/v1", h.getDashboardV1)

	return r
}

func (h *dashboardHandler) getDashboardV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := h.s.GetDashboardV1(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
