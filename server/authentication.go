package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/auth"
)

type authHandler struct {
	s auth.Service
}

func (h *authHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Post("/signin", h.signIn)

	return r
}

func (h *authHandler) signIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	r.ParseMultipartForm(10 << 20) // 10 * 2^20
	username := r.FormValue("username")
	password := r.FormValue("password")

	result, err := h.s.SignIn(r.Context(), username, password)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if result == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("Incorrect username or password.")))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
