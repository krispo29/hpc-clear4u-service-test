package server

import (
	outbound "hpc-express-service/outbound/mawb"

	"github.com/go-chi/chi/v5"
)

type outboundMawbHandler struct {
	s outbound.OutboundMawbService
}

func (h *outboundMawbHandler) router() chi.Router {

	r := chi.NewRouter()

	r.Route("/info", func(r chi.Router) {
		r.Post("/", h.createMawbInfo)
		r.Get("/", h.getAllMawbInfo)
		r.Get("/{uuid}", h.getOneMawbInfo)
		r.Put("/", h.updateMawbInfo)
		r.Delete("/{uuid}", h.deleteMawbInfo)
		r.Patch("/{uuid}/upload_attachment", h.uploadAttachment)
	})

	r.Route("/drafts", func(r chi.Router) {
		r.Get("/", h.getAllMawbDraft)
		r.Post("/", h.saveMawbDraft)
		r.Put("/", h.updateMawbDraft)
		r.Post("/preview", h.previewMawbDraft)
		r.Get("/print/{uuid}", h.PrintMawbDraft)
		r.Get("/{uuid}", h.getOneMawbDraft)
	})

	return r
}
