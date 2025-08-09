package server

import (
	"context"
	"encoding/json"
	"net/http"

	outbound "hpc-express-service/outbound/mawb"

	"github.com/go-chi/chi/v5"
)

// --- MAWB Info-centric Handlers ---

// Middleware to extract mawb_info_uuid and put it into context
func mawbInfoContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mawbInfoUUID := chi.URLParam(r, "mawbInfoUUID")
		ctx := context.WithValue(r.Context(), "mawb_info_uuid", mawbInfoUUID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *outboundMawbHandler) mawbInfoRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(mawbInfoContext)

	// --- Cargo Manifest Routes ---
	r.Route("/cargo-manifest", func(r chi.Router) {
		r.Get("/", h.getCargoManifest)
		r.Post("/", h.createOrUpdateCargoManifest)
		r.Post("/confirm", h.confirmCargoManifest)
		r.Post("/reject", h.rejectCargoManifest)
		r.Get("/print", h.printCargoManifest)
	})

	// --- Draft MAWB V2 Routes ---
	r.Route("/draft-mawb", func(r chi.Router) {
		r.Get("/", h.getDraftMAWBV2)
		r.Post("/", h.createOrUpdateDraftMAWBV2)
		r.Post("/confirm", h.confirmDraftMAWBV2)
		r.Post("/reject", h.rejectDraftMAWBV2)
		r.Get("/print", h.printDraftMAWBV2)
	})

	return r
}

// --- Cargo Manifest Handlers ---

func (h *outboundMawbHandler) getCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbInfoUUID := r.Context().Value("mawb_info_uuid").(string)
	manifest, err := h.s.GetCargoManifestByMAWBInfoUUID(r.Context(), mawbInfoUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if manifest == nil {
		http.Error(w, "Cargo Manifest not found for this MAWB", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Success", Data: manifest})
}

func (h *outboundMawbHandler) createOrUpdateCargoManifest(w http.ResponseWriter, r *http.Request) {
	var req outbound.CargoManifest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	manifest, err := h.s.CreateOrUpdateCargoManifest(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Cargo Manifest created/updated successfully", Data: manifest})
}

func (h *outboundMawbHandler) confirmCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbInfoUUID := r.Context().Value("mawb_info_uuid").(string)
	if err := h.s.ConfirmCargoManifest(r.Context(), mawbInfoUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Cargo Manifest confirmed successfully"})
}

func (h *outboundMawbHandler) rejectCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbInfoUUID := r.Context().Value("mawb_info_uuid").(string)
	if err := h.s.RejectCargoManifest(r.Context(), mawbInfoUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Cargo Manifest rejected successfully"})
}

func (h *outboundMawbHandler) printCargoManifest(w http.ResponseWriter, r *http.Request) {
	// Implementation for PDF printing would go here
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("PDF printing for Cargo Manifest is not implemented yet"))
}


// --- Draft MAWB V2 Handlers ---

func (h *outboundMawbHandler) getDraftMAWBV2(w http.ResponseWriter, r *http.Request) {
	mawbInfoUUID := r.Context().Value("mawb_info_uuid").(string)
	draft, err := h.s.GetDraftMAWBByMAWBInfoUUIDV2(r.Context(), mawbInfoUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if draft == nil {
		http.Error(w, "Draft MAWB not found for this MAWB", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Success", Data: draft})
}

func (h *outboundMawbHandler) createOrUpdateDraftMAWBV2(w http.ResponseWriter, r *http.Request) {
	var req outbound.DraftMAWBV2
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	draft, err := h.s.CreateOrUpdateDraftMAWBV2(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Draft MAWB created/updated successfully", Data: draft})
}

func (h *outboundMawbHandler) confirmDraftMAWBV2(w http.ResponseWriter, r *http.Request) {
	mawbInfoUUID := r.Context().Value("mawb_info_uuid").(string)
	if err := h.s.ConfirmDraftMAWBV2(r.Context(), mawbInfoUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Draft MAWB confirmed successfully"})
}

func (h *outboundMawbHandler) rejectDraftMAWBV2(w http.ResponseWriter, r *http.Request) {
	mawbInfoUUID := r.Context().Value("mawb_info_uuid").(string)
	if err := h.s.RejectDraftMAWBV2(r.Context(), mawbInfoUUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Code: http.StatusOK, Message: "Draft MAWB rejected successfully"})
}

func (h *outboundMawbHandler) printDraftMAWBV2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("PDF printing for V2 Draft MAWB is not implemented yet"))
}

// Generic response struct for consistency
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
