package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/uploadlog"
)

type uploadLoggingHandler struct {
	s uploadlog.Service
}

func (h *uploadLoggingHandler) router() chi.Router {

	r := chi.NewRouter()

	r.Get("/", h.getAllUploadLoggings)

	return r
}

func (h *uploadLoggingHandler) getAllUploadLoggings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")
	category := r.URL.Query().Get("category")
	subCategory := r.URL.Query().Get("subCategory")
	if len(startDate) == 0 {
		render.Render(w, r, ErrInvalidRequest(errors.New("require start date")))
		return
	}
	if len(endDate) == 0 {
		render.Render(w, r, ErrInvalidRequest(errors.New("require end date")))
		return
	}

	result, err := h.s.GetAllUploadloggings(r.Context(), startDate, endDate, category, subCategory)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
