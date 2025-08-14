package server

import (
	"context"
	"fmt"
	"hpc-express-service/outbound"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"hpc-express-service/outbound/mawbinfo"
)

type mawbInfoHandler struct {
	s                mawbinfo.Service
	cargoManifestSvc outbound.CargoManifestService
	draftMAWBSvc     outbound.DraftMAWBService
}

func (h *mawbInfoHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createMawbInfo)
	r.Get("/", h.getAllMawbInfo)

	// Draft MAWB List Route (without uuid parameter)
	r.Get("/draft-mawb", h.getAllDraftMAWB)

	// Draft MAWB Detail Route by draft UUID (not mawb_info_uuid)
	r.Get("/draft-mawb/{draft_uuid}", h.getDraftMAWBByUUID)

	r.Route("/{uuid}", func(r chi.Router) {
		r.Get("/", h.getMawbInfo)
		r.Put("/", h.updateMawbInfo)
		r.Delete("/", h.deleteMawbInfo)
		r.Delete("/attachments", h.deleteMawbInfoAttachment)

		// Cargo Manifest Routes
		r.Get("/cargo-manifest", h.getCargoManifest)
		r.Post("/cargo-manifest", h.createCargoManifest)
		r.Put("/cargo-manifest", h.updateCargoManifest)
		r.Post("/cargo-manifest/confirm", h.confirmCargoManifest)
		r.Post("/cargo-manifest/reject", h.rejectCargoManifest)
		r.Get("/cargo-manifest/print", h.printCargoManifest)

		// Draft MAWB Routes
		r.Get("/draft-mawb", h.getDraftMAWB)
		r.Post("/draft-mawb", h.createOrUpdateDraftMAWB)
		r.Patch("/draft-mawb", h.updateDraftMAWB)
		r.Post("/draft-mawb/confirm", h.confirmDraftMAWB)
		r.Post("/draft-mawb/reject", h.rejectDraftMAWB)
		r.Get("/draft-mawb/print", h.printDraftMAWB)

		// MAWB management routes
		r.Get("/cancel", h.cancelMAWB)
		r.Get("/undo_cancel", h.undoCancelMAWB)
	})

	return r
}

// Cargo Manifest Handlers

func (h *mawbInfoHandler) getCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	manifest, err := h.cargoManifestSvc.GetCargoManifestByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if manifest == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Cargo Manifest not found for this MAWB"})
		return
	}

	render.Respond(w, r, SuccessResponse(manifest, "Success"))
}

func (h *mawbInfoHandler) createCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	data := &outbound.CargoManifest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	data.MAWBInfoUUID = mawbUUID

	result, err := h.cargoManifestSvc.CreateCargoManifest(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "Cargo Manifest created successfully"))
}

func (h *mawbInfoHandler) updateCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	data := &outbound.CargoManifest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	data.MAWBInfoUUID = mawbUUID

	result, err := h.cargoManifestSvc.UpdateCargoManifest(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "Cargo Manifest updated successfully"))
}

func (h *mawbInfoHandler) confirmCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	err := h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, "Confirmed")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Cargo Manifest confirmed successfully"))
}

func (h *mawbInfoHandler) rejectCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	err := h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, "Rejected")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Cargo Manifest rejected successfully"))
}

func (h *mawbInfoHandler) printCargoManifest(w http.ResponseWriter, r *http.Request) {
	// PDF generation logic goes here
	render.Respond(w, r, SuccessResponse(nil, "Print endpoint not implemented yet"))
}

// Draft MAWB Handlers

func (h *mawbInfoHandler) getDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	draft, err := h.draftMAWBSvc.GetDraftMAWBByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if draft == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Draft MAWB not found for this MAWB"})
		return
	}

	render.Respond(w, r, SuccessResponse(draft, "Success"))
}

func (h *mawbInfoHandler) createOrUpdateDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	inputData := &outbound.DraftMAWBInput{}
	if err := render.Bind(r, inputData); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Convert input to DraftMAWB
	data := inputData.ToDraftMAWB()
	data.MAWBInfoUUID = mawbUUID

	// Check if draft MAWB already exists for this MAWB UUID
	existing, _ := h.draftMAWBSvc.GetDraftMAWBByMAWBUUID(r.Context(), mawbUUID)

	var result *outbound.DraftMAWB
	var err error
	if existing != nil {
		// Update existing draft MAWB
		data.UUID = existing.UUID
		result, err = h.draftMAWBSvc.UpdateDraftMAWB(r.Context(), data, inputData.Items, inputData.Charges)
	} else {
		// Create new draft MAWB
		result, err = h.draftMAWBSvc.CreateDraftMAWB(r.Context(), data, inputData.Items, inputData.Charges)
	}
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(map[string]string{"uuid": result.UUID}, "Draft MAWB created successfully"))
}

func (h *mawbInfoHandler) updateDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	inputData := &outbound.DraftMAWBInput{}
	if err := render.Bind(r, inputData); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// For PATCH operation, we need to find the existing draft MAWB by mawb_info_uuid first
	existing, err := h.draftMAWBSvc.GetDraftMAWBByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if existing == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Draft MAWB not found for this MAWB"})
		return
	}

	// Convert input to DraftMAWB and update the existing draft MAWB using its UUID
	data := inputData.ToDraftMAWB()
	data.MAWBInfoUUID = mawbUUID
	data.UUID = existing.UUID // Set the existing UUID for update
	result, err := h.draftMAWBSvc.UpdateDraftMAWB(r.Context(), data, inputData.Items, inputData.Charges)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(map[string]string{"uuid": result.UUID}, "Draft MAWB updated successfully"))
}

func (h *mawbInfoHandler) confirmDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	err := h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, "Confirmed")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Draft MAWB confirmed successfully"))
}

func (h *mawbInfoHandler) rejectDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	err := h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, "Rejected")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Draft MAWB rejected successfully"))
}

func (h *mawbInfoHandler) printDraftMAWB(w http.ResponseWriter, r *http.Request) {
	// PDF generation logic goes here
	render.Respond(w, r, SuccessResponse(nil, "Print endpoint not implemented yet"))
}

func (h *mawbInfoHandler) getAllDraftMAWB(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	drafts, err := h.draftMAWBSvc.GetAllDraftMAWB(r.Context(), startDate, endDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(drafts, "Success"))
}

func (h *mawbInfoHandler) getDraftMAWBByUUID(w http.ResponseWriter, r *http.Request) {
	draftUUID := chi.URLParam(r, "draft_uuid")
	if draftUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("draft_uuid parameter is required")))
		return
	}

	draft, err := h.draftMAWBSvc.GetDraftMAWBWithRelations(r.Context(), draftUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if draft == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Draft MAWB not found"})
		return
	}

	// Convert to response without airline_logo and airline_name
	response := draft.ToDraftMAWBWithRelationsResponse()
	render.Respond(w, r, SuccessResponse(response, "Success"))
}

func (h *mawbInfoHandler) cancelMAWB(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	err := h.draftMAWBSvc.CancelDraftMAWB(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "MAWB cancelled successfully"))
}

func (h *mawbInfoHandler) undoCancelMAWB(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	err := h.draftMAWBSvc.UndoCancelDraftMAWB(r.Context(), uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "MAWB recovered successfully"))
}

func (h *mawbInfoHandler) createMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Bind request data
	data := &mawbinfo.CreateMawbInfoRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate request data using validator
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Call service to create MAWB info
	result, err := h.s.CreateMawbInfo(ctx, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}
func (h *mawbInfoHandler) getMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Call service to get MAWB info
	result, err := h.s.GetMawbInfo(ctx, uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *mawbInfoHandler) getAllMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get query parameters for date filtering
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	// Call service to get all MAWB info with date filters
	result, err := h.s.GetAllMawbInfo(ctx, startDate, endDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}
func (h *mawbInfoHandler) updateMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Parse multipart form data (max 32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %v", err)))
		return
	}

	// Extract form data
	data := &mawbinfo.UpdateMawbInfoRequest{
		ChargeableWeight: r.FormValue("chargeableWeight"),
		Date:             r.FormValue("date"),
		Mawb:             r.FormValue("mawb"),
		ServiceType:      r.FormValue("serviceType"),
		ShippingType:     r.FormValue("shippingType"),
	}

	// Get file attachments
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["attachments"]; exists {
			data.Attachments = files
		}
	}

	// Validate request data using validator
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Call service to update MAWB info
	result, err := h.s.UpdateMawbInfo(ctx, uuid, data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *mawbInfoHandler) deleteMawbInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Call service to delete MAWB info
	err := h.s.DeleteMawbInfo(ctx, uuid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "deleted successfully"))
}

func (h *mawbInfoHandler) deleteMawbInfoAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get UUID from URL parameter
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Bind request data for the file to be deleted
	data := &mawbinfo.DeleteAttachmentRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate request data
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Call service to delete the attachment
	err := h.s.DeleteMawbInfoAttachment(ctx, uuid, data.FileName)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Return success response
	render.Respond(w, r, SuccessResponse(nil, "attachment deleted successfully"))
}
