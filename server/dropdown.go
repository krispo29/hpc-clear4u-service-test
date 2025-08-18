package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"hpc-express-service/dropdown"
)

type dropdownHandler struct {
	dropdownSvc dropdown.Service
}

func (h *dropdownHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Get("/service-type", h.getServiceTypes)
	r.Get("/shipping-type", h.getShippingTypes)
	r.Get("/airline-logo", h.getAirlineLogos)
	r.Get("/master-statuses", h.getMasterStatuses)

	return r
}

func (h *dropdownHandler) getMasterStatuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	statusType := r.URL.Query().Get("type")
	if statusType == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("type query parameter is required")))
		return
	}

	statuses, err := h.dropdownSvc.GetMasterStatusesByType(ctx, statusType)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(statuses, "Master statuses retrieved successfully"))
}

func (h *dropdownHandler) getServiceTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	serviceTypes, err := h.dropdownSvc.GetServiceTypes(ctx)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(serviceTypes, "Service types retrieved successfully"))
}

func (h *dropdownHandler) getShippingTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shippingTypes, err := h.dropdownSvc.GetShippingTypes(ctx)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(shippingTypes, "Shipping types retrieved successfully"))
}

func (h *dropdownHandler) getAirlineLogos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	airlineLogos, err := h.dropdownSvc.GetAirlineLogos(ctx)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(airlineLogos, "Airline logos retrieved successfully"))
}
