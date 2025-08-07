package server

import (
	"context"
	"hpc-express-service/user"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type userHandler struct {
	s user.Service
}

func (h *userHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.get)
	return r
}

func (h *userHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	userUUID := GetUserUUIDFromContext(r)

	result, err := h.s.Get(r.Context(), userUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
