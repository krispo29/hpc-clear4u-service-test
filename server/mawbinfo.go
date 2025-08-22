package server

import (
	"bytes"
	"context"
	"fmt"
	cargoManifest "hpc-express-service/outbound/cargomanifest"
	draftMawb "hpc-express-service/outbound/draftmawb"
	weightslip "hpc-express-service/outbound/weightSlip"
	"hpc-express-service/setting"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jung-kurt/gofpdf"

	"hpc-express-service/outbound/mawbinfo"
)

type mawbInfoHandler struct {
	s                mawbinfo.Service
	cargoManifestSvc cargoManifest.CargoManifestService
	draftMAWBSvc     draftMawb.DraftMAWBService
	weightslipSvc    weightslip.WeightSlipService
	statusSvc        setting.MasterStatusService
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
		r.Post("/cargo-manifest/send-customer", h.sendCargoManifestToCustomer)
		r.Post("/cargo-manifest/customer-confirm", h.customerConfirmCargoManifest)
		r.Post("/cargo-manifest/customer-reject", h.customerRejectCargoManifest)
		r.Post("/cargo-manifest/confirm", h.confirmCargoManifest)
		r.Post("/cargo-manifest/reject", h.rejectCargoManifest)
		r.Get("/cargo-manifest/print", h.printCargoManifest)
		r.Post("/cargo-manifest/preview", h.previewCargoManifest)

		// Weight Slip Routes
		r.Get("/weight-slip", h.getWeightslip)
		r.Post("/weight-slip", h.createWeightslip)
		r.Put("/weight-slip", h.updateWeightslip)
		r.Post("/weight-slip/send-customer", h.sendWeightslipToCustomer)
		r.Post("/weight-slip/customer-confirm", h.customerConfirmWeightslip)
		r.Post("/weight-slip/customer-reject", h.customerRejectWeightslip)
		r.Post("/weight-slip/confirm", h.confirmWeightslip)
		r.Post("/weight-slip/reject", h.rejectWeightslip)
		r.Get("/weight-slip/print", h.printWeightslip)
		r.Post("/weight-slip/preview", h.previewWeightslip)

		// Draft MAWB Routes
		r.Get("/draft-mawb", h.getDraftMAWB)
		r.Post("/draft-mawb", h.createDraftMAWB)
		r.Put("/draft-mawb", h.updateDraftMAWB)
		r.Post("/draft-mawb/send-customer", h.sendDraftMAWBToCustomer)
		r.Post("/draft-mawb/customer-confirm", h.customerConfirmDraftMAWB)
		r.Post("/draft-mawb/customer-reject", h.customerRejectDraftMAWB)
		r.Post("/draft-mawb/confirm", h.confirmDraftMAWB)
		r.Post("/draft-mawb/reject", h.rejectDraftMAWB)
		r.Get("/draft-mawb/print", h.printDraftMAWB)
		r.Post("/draft-mawb/preview", h.previewDraftMAWB)

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

	data := &cargoManifest.CargoManifest{}
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

	data := &cargoManifest.CargoManifest{}
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
func (h *mawbInfoHandler) sendCargoManifestToCustomer(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	// เปลี่ยนเป็น "CM_AwaitingCustomer"
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CM_AwaitingCustomer", "cargo_manifest")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CM_AwaitingCustomer' not found")))
		return
	}
	err = h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Cargo Manifest sent to customer for confirmation"))
}
func (h *mawbInfoHandler) customerConfirmCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	// เปลี่ยนเป็น "CM_CustomerConfirmed"
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CM_CustomerConfirmed", "cargo_manifest")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CM_CustomerConfirmed' not found")))
		return
	}
	err = h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Cargo Manifest confirmed by customer"))
}

func (h *mawbInfoHandler) customerRejectCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	// เปลี่ยนเป็น "CM_CustomerRejected"
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CM_CustomerRejected", "cargo_manifest")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CM_CustomerRejected' not found")))
		return
	}
	err = h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Cargo Manifest rejected by customer"))
}

func (h *mawbInfoHandler) confirmCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	// เปลี่ยนเป็น "CM_Confirmed"
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CM_Confirmed", "cargo_manifest")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CM_Confirmed' not found")))
		return
	}
	err = h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, status.UUID)
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
	// เปลี่ยนเป็น "CM_Rejected"
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CM_Rejected", "cargo_manifest")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CM_Rejected' not found")))
		return
	}
	err = h.cargoManifestSvc.UpdateCargoManifestStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Cargo Manifest rejected successfully"))
}
func (h *mawbInfoHandler) printCargoManifest(w http.ResponseWriter, r *http.Request) {
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

	pdfBuffer, err := h.generateCargoManifestPDF(manifest)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=cargo_manifest.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
	w.Write(pdfBuffer.Bytes())
}

func (h *mawbInfoHandler) previewCargoManifest(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	manifest := &cargoManifest.CargoManifest{}
	if err := render.Bind(r, manifest); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	manifest.MAWBInfoUUID = mawbUUID

	pdfBuffer, err := h.generateCargoManifestPDF(manifest)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=cargo_manifest_preview.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
	w.Write(pdfBuffer.Bytes())
}

// Weight Slip Handlers

func (h *mawbInfoHandler) getWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	ws, err := h.weightslipSvc.GetWeightSlipByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if ws == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Weight Slip not found for this MAWB"})
		return
	}

	render.Respond(w, r, SuccessResponse(ws, "Success"))
}

func (h *mawbInfoHandler) createWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	data := &weightslip.WeightSlip{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	data.MAWBInfoUUID = mawbUUID

	result, err := h.weightslipSvc.CreateWeightSlip(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "Weight Slip created successfully"))
}

func (h *mawbInfoHandler) updateWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	data := &weightslip.WeightSlip{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	data.MAWBInfoUUID = mawbUUID

	result, err := h.weightslipSvc.UpdateWeightSlip(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "Weight Slip updated successfully"))
}

func (h *mawbInfoHandler) sendWeightslipToCustomer(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "WS_AwaitingCustomer", "weight_slip")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'WS_AwaitingCustomer' not found")))
		return
	}
	err = h.weightslipSvc.UpdateWeightSlipStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Weight Slip sent to customer for confirmation"))
}

func (h *mawbInfoHandler) customerConfirmWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "WS_CustomerConfirmed", "weight_slip")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'WS_CustomerConfirmed' not found")))
		return
	}
	err = h.weightslipSvc.UpdateWeightSlipStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Weight Slip confirmed by customer"))
}

func (h *mawbInfoHandler) customerRejectWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "WS_CustomerRejected", "weight_slip")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'WS_CustomerRejected' not found")))
		return
	}
	err = h.weightslipSvc.UpdateWeightSlipStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Weight Slip rejected by customer"))
}

func (h *mawbInfoHandler) confirmWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "WS_Confirmed", "weight_slip")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'WS_Confirmed' not found")))
		return
	}
	err = h.weightslipSvc.UpdateWeightSlipStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Weight Slip confirmed successfully"))
}

func (h *mawbInfoHandler) rejectWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "WS_Rejected", "weight_slip")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'WS_Rejected' not found")))
		return
	}
	err = h.weightslipSvc.UpdateWeightSlipStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Weight Slip rejected successfully"))
}

func (h *mawbInfoHandler) printWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	ws, err := h.weightslipSvc.GetWeightSlipByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if ws == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Weight Slip not found for this MAWB"})
		return
	}

	pdfBuffer, err := h.generateWeightSlipPDF(ws)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=weight_slip.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
	w.Write(pdfBuffer.Bytes())
}

func (h *mawbInfoHandler) previewWeightslip(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	ws := &weightslip.WeightSlip{}
	if err := render.Bind(r, ws); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	ws.MAWBInfoUUID = mawbUUID

	pdfBuffer, err := h.generateWeightSlipPDF(ws)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=weight_slip_preview.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
	w.Write(pdfBuffer.Bytes())
}

// Draft MAWB Handlers

func (h *mawbInfoHandler) getDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	draft, err := h.draftMAWBSvc.GetDraftMAWBWithRelationsByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if draft == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Draft MAWB not found for this MAWB"})
		return
	}

	// Convert to response format
	response := draft.ToDraftMAWBWithRelationsResponse()
	render.Respond(w, r, SuccessResponse(response, "Success"))
}

func (h *mawbInfoHandler) createDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	inputData := &draftMawb.DraftMAWBInput{}
	if err := render.Bind(r, inputData); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check if draft MAWB already exists for this MAWB UUID
	existing, _ := h.draftMAWBSvc.GetDraftMAWBByMAWBUUID(r.Context(), mawbUUID)
	if existing != nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusConflict, Message: "Draft MAWB already exists for this MAWB"})
		return
	}

	// Convert input to DraftMAWB
	data := inputData.ToDraftMAWB()
	data.MAWBInfoUUID = mawbUUID

	// Create new draft MAWB
	result, err := h.draftMAWBSvc.CreateDraftMAWB(r.Context(), data, inputData.Items, inputData.Charges)
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

	inputData := &draftMawb.DraftMAWBInput{}
	if err := render.Bind(r, inputData); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// For PUT operation, we need to find the existing draft MAWB by mawb_info_uuid first
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
func (h *mawbInfoHandler) sendDraftMAWBToCustomer(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "AwaitingCustomer", "draft_mawb")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'AwaitingCustomer' not found")))
		return
	}
	err = h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Draft MAWB sent to customer for confirmation"))
}

func (h *mawbInfoHandler) customerConfirmDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CustomerConfirmed", "draft_mawb")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CustomerConfirmed' not found")))
		return
	}
	err = h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Draft MAWB confirmed by customer"))
}

func (h *mawbInfoHandler) customerRejectDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "CustomerRejected", "draft_mawb")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'CustomerRejected' not found")))
		return
	}
	err = h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Draft MAWB rejected by customer"))
}
func (h *mawbInfoHandler) confirmDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "Confirmed", "draft_mawb")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'Confirmed' not found")))
		return
	}
	err = h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, status.UUID)
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
	status, err := h.statusSvc.GetStatusByNameAndType(r.Context(), "Rejected", "draft_mawb")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if status == nil {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("status 'Rejected' not found")))
		return
	}
	err = h.draftMAWBSvc.UpdateDraftMAWBStatus(r.Context(), mawbUUID, status.UUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, SuccessResponse(nil, "Draft MAWB rejected successfully"))
}

func (h *mawbInfoHandler) printDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	// Get the existing draft MAWB data from database
	draft, err := h.draftMAWBSvc.GetDraftMAWBWithRelationsByMAWBUUID(r.Context(), mawbUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if draft == nil {
		render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, Message: "Draft MAWB not found for this MAWB"})
		return
	}

	// Convert to input format for PDF generation
	inputData := draft.ToDraftMAWBInput()

	// Generate PDF for printing (without preview watermark)
	pdfBuffer, err := h.generateDraftMAWBPDF(inputData, false)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to generate PDF: %v", err)))
		return
	}

	// Set CORS headers for better browser compatibility
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Set headers for PDF response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=draft_mawb_print.pdf") // Changed to inline for iframe printing
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write PDF to response
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBuffer.Bytes())
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

func (h *mawbInfoHandler) previewDraftMAWB(w http.ResponseWriter, r *http.Request) {
	mawbUUID := chi.URLParam(r, "uuid")
	if mawbUUID == "" {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("uuid parameter is required")))
		return
	}

	inputData := &draftMawb.DraftMAWBInput{}
	if err := render.Bind(r, inputData); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Generate PDF preview
	pdfBuffer, err := h.generateDraftMAWBPDF(inputData, true)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Set headers for PDF response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=draft_mawb_preview.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	// Write PDF to response
	w.Write(pdfBuffer.Bytes())
}

// Helper functions to convert input types to response types
func convertItemInputsToItems(inputs []draftMawb.DraftMAWBItemInput) []draftMawb.DraftMAWBItem {
	items := make([]draftMawb.DraftMAWBItem, len(inputs))
	for i, input := range inputs {
		items[i] = draftMawb.DraftMAWBItem{
			ID:                input.ID,
			PiecesRCP:         input.PiecesRCP,
			GrossWeight:       fmt.Sprintf("%.2f", input.GrossWeight),
			KgLb:              input.KgLb,
			RateClass:         input.RateClass,
			TotalVolume:       input.TotalVolume,
			ChargeableWeight:  input.ChargeableWeight,
			RateCharge:        input.RateCharge,
			Total:             input.Total,
			NatureAndQuantity: input.NatureAndQuantity,
			Dims:              convertDimInputsToDims(input.Dims),
		}
	}
	return items
}

func convertDimInputsToDims(inputs []draftMawb.DraftMAWBItemDimInput) []draftMawb.DraftMAWBItemDim {
	dims := make([]draftMawb.DraftMAWBItemDim, len(inputs))
	for i, input := range inputs {
		dims[i] = draftMawb.DraftMAWBItemDim{
			ID:     input.ID,
			Length: fmt.Sprintf("%d", input.Length),
			Width:  fmt.Sprintf("%d", input.Width),
			Height: fmt.Sprintf("%d", input.Height),
			Count:  fmt.Sprintf("%d", input.Count),
		}
	}
	return dims
}

func convertChargeInputsToCharges(inputs []draftMawb.DraftMAWBChargeInput) []draftMawb.DraftMAWBCharge {
	charges := make([]draftMawb.DraftMAWBCharge, len(inputs))
	for i, input := range inputs {
		charges[i] = draftMawb.DraftMAWBCharge{
			ID:    input.ID,
			Key:   input.Key,
			Value: input.Value,
		}
	}
	return charges
}

func (h *mawbInfoHandler) generateCargoManifestPDF(manifest *cargoManifest.CargoManifest) (bytes.Buffer, error) {
	frontTHSarabunNew, err := os.ReadFile("assets/THSarabunNew.ttf")
	if err != nil {
		log.Println(err)
	}
	frontTHSarabunNewBold, err := os.ReadFile("assets/THSarabunNew Bold.ttf")
	if err != nil {
		log.Println(err)
	}

	var buf bytes.Buffer

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("THSarabunNew", "", frontTHSarabunNew)
	pdf.AddUTF8FontFromBytes("THSarabunNew Bold", "", frontTHSarabunNewBold)

	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// หัวกระดาษ
	pdf.SetFont("THSarabunNew Bold", "", 16)
	pdf.CellFormat(0, 7, "AIR CARGO MANIFEST", "0", 1, "C", false, 0, "")

	// helper สำหรับพิมพ์ label: value ในบรรทัดเดียว (label หนา value ปกติ)
	labelValue := func(label, value string, labelW, lineH float64) {
		x := pdf.GetX()
		y := pdf.GetY()
		pdf.SetFont("THSarabunNew Bold", "", 12) // label หนา
		pdf.CellFormat(labelW, lineH, label, "0", 0, "L", false, 0, "")
		pdf.SetFont("THSarabunNew", "", 12) // value ปกติ
		pdf.SetXY(x+labelW, y)
		pdf.CellFormat(0, lineH, value, "0", 1, "L", false, 0, "")
	}

	// helper สำหรับพิมพ์ label: value แบบหลายบรรทัด
	labelValueMulti := func(label, value string, labelW, lineH float64) {
		x := pdf.GetX()
		y := pdf.GetY()
		pdf.SetFont("THSarabunNew Bold", "", 12)
		pdf.CellFormat(labelW, lineH, label, "0", 0, "L", false, 0, "")
		pdf.SetFont("THSarabunNew", "", 12)
		pdf.SetXY(x+labelW, y)
		pdf.MultiCell(0, lineH, value, "", "L", false)
	}

	// บล็อกข้อมูลหัวก่อนตาราง (ย้ายมุมขวาเหมือนเดิม)
	pdf.SetFont("THSarabunNew", "", 12)
	pdf.SetX(200)

	labelW := 55.0
	lineH := 6.0

	labelValue("FLIGHT NO: ", manifest.FlightNo, labelW, lineH)
	labelValue("MAWB NO: ", manifest.MAWBNumber, labelW, lineH)
	labelValue("PORT OF DISCHARGE: ", manifest.PortOfDischarge, labelW, lineH)
	labelValue("FREIGHT DATE: ", manifest.FreightDate, labelW, lineH)

	// แปลง \n ที่มาจาก client ให้ขึ้นบรรทัดจริง
	shipper := strings.ReplaceAll(manifest.Shipper, "\\n", "\n")
	consignee := strings.ReplaceAll(manifest.Consignee, "\\n", "\n")

	labelValueMulti("SHIPPER: ", shipper, labelW, lineH)
	labelValueMulti("CONSIGNEE: ", consignee, labelW, lineH)

	labelValue("TOTAL CTN: ", manifest.TotalCtn, labelW, lineH)

	pdf.Ln(3)

	// ตารางรายการ
	headers := []string{"HAWB NO.", "CTNS", "WEIGHT(KG)", "ORIGIN", "DES", "SHIPPER NAME AND ADDRESS", "CONSIGNEE NAME AND ADDRESS", "NATURE OF GOODS"}
	colWidths := []float64{25, 15, 20, 20, 20, 40, 40, 30}

	pdf.SetFont("THSarabunNew Bold", "", 10)
	for i, htxt := range headers {
		pdf.CellFormat(colWidths[i], 7, htxt, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("THSarabunNew", "", 10)
	lineHeight := 5.0
	left, _, _, _ := pdf.GetMargins()

	for _, item := range manifest.Items {
		values := []string{
			item.HAWBNo,
			item.Pkgs,
			item.GrossWeight,
			manifest.PortOfDischarge,
			item.Destination,
			strings.ReplaceAll(item.ShipperNameAndAddress, "\\n", "\n"),
			strings.ReplaceAll(item.ConsigneeNameAndAddress, "\\n", "\n"),
			item.Commodity,
		}

		// คำนวณความสูงของแถวจากจำนวนบรรทัดมากสุดในแต่ละคอลัมน์
		maxLines := 1
		for i, val := range values {
			lines := pdf.SplitLines([]byte(val), colWidths[i])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}
		}
		rowHeight := float64(maxLines) * lineHeight

		x, y := left, pdf.GetY()
		for i, val := range values {
			pdf.Rect(x, y, colWidths[i], rowHeight, "D")
			pdf.SetXY(x, y)
			pdf.MultiCell(colWidths[i], lineHeight, val, "", "L", false)
			x += colWidths[i]
		}
		pdf.SetXY(left, y+rowHeight)
	}

	// Transshipment (ถ้ามี)
	if manifest.Transshipment != "" {
		pdf.Ln(5)
		pdf.SetFont("THSarabunNew Bold", "", 12)
		pdf.CellFormat(0, 6, strings.ToUpper(manifest.Transshipment), "0", 1, "C", false, 0, "")
	}

	err = pdf.Output(&buf)
	return buf, err
}

func (h *mawbInfoHandler) generateWeightSlipPDF(ws *weightslip.WeightSlip) (bytes.Buffer, error) {
	var buf bytes.Buffer
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Weight Slip")
	err := pdf.Output(&buf)
	return buf, err
}

func (h *mawbInfoHandler) generateDraftMAWBPDF(data *draftMawb.DraftMAWBInput, isPreview bool) (bytes.Buffer, error) {
	// Loading Font
	frontTHSarabunNew, err := os.ReadFile("assets/THSarabunNew.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewBold, err := os.ReadFile("assets/THSarabunNew Bold.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewBoldItalic, err := os.ReadFile("assets/THSarabunNew BoldItalic.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewItalic, err := os.ReadFile("assets/THSarabunNew Italic.ttf")
	if err != nil {
		log.Println(err)
	}

	var buf bytes.Buffer

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        gofpdf.PageSizeA4,
	})

	pdf.AddUTF8FontFromBytes("THSarabunNew", "", frontTHSarabunNew)
	pdf.AddUTF8FontFromBytes("THSarabunNew Bold", "", frontTHSarabunNewBold)
	pdf.AddUTF8FontFromBytes("THSarabunNew BoldItalic", "", frontTHSarabunNewBoldItalic)
	pdf.AddUTF8FontFromBytes("THSarabunNew Italic", "", frontTHSarabunNewItalic)

	pdf.SetFont("THSarabunNew Bold", "", 10)

	pdf.AddPage()
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(true, 0)
	width, height := pdf.GetPageSize()

	if isPreview {
		pdf.ImageOptions(
			"assets/bg-mawb.png", // path to the image
			0, 0,                 // x, y positions
			width, height, // width, height to fit full page
			false, // do not flow the image
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0, // link
			"",
		)
	}

	pdf.SetFont("THSarabunNew Bold", "", 22)
	pdf.SetXY(7, 4)
	pdf.MultiCell(51, 5, data.MAWB, "0", "L", false)

	pdf.SetFont("THSarabunNew Bold", "", 22)
	pdf.SetXY(143, 4)
	pdf.MultiCell(51, 5, data.HAWB, "0", "C", false)

	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(9, 19)
	pdf.MultiCell(89, 3.5, data.ShipperNameAndAddress, "0", "LT", false)

	pdf.SetFont("THSarabunNew Bold", "", 20)
	pdf.SetXY(118, 13)
	pdf.MultiCell(77, 20, data.AWBIssuedBy, "0", "C", false)

	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(9, 45)
	pdf.MultiCell(89, 3.5, data.ConsigneeNameAndAddress, "0", "LT", false)

	pdf.SetFont("THSarabunNew Bold", "", 16)
	pdf.SetXY(8, 65)
	pdf.CellFormat(89, 15, data.IssuingCarrierAgentName, "0", 0, "C", false, 0, "")

	pdf.SetFont("THSarabunNew Bold", "", 20)
	pdf.SetXY(98, 67)
	pdf.MultiCell(97, 6, data.AccountingInfomation, "0", "C", false)

	pdf.SetFont("THSarabunNew Bold", "", 15)
	pdf.SetXY(8, 84)
	pdf.CellFormat(44, 6, data.AgentsIATACode, "0", 0, "L", false, 0, "")

	pdf.SetFont("THSarabunNew Bold", "", 15)
	pdf.SetXY(53, 84)
	pdf.MultiCell(44, 6, data.AccountNo, "0", "L", false)

	pdf.SetXY(8, 92)
	pdf.MultiCell(89, 6, data.AirportOfDeparture, "0", "C", false)

	pdf.SetXY(97, 92)
	pdf.MultiCell(35, 6, data.ReferenceNumber, "0", "C", false)

	pdf.SetXY(132, 92)
	pdf.MultiCell(30, 6, data.OptionalShippingInfo1, "0", "C", false)

	pdf.SetXY(162, 92)
	pdf.MultiCell(32, 6, data.OptionalShippingInfo2, "0", "C", false)

	pdf.SetXY(8, 102)
	pdf.CellFormat(11, 6, data.RoutingTo, "0", 0, "C", false, 0, "")

	pdf.SetXY(20, 102)
	pdf.CellFormat(40, 6, data.RoutingBy, "0", 0, "C", false, 0, "")

	pdf.SetXY(62, 102)
	pdf.CellFormat(11, 6, data.DestinationTo1, "0", 0, "C", false, 0, "")

	pdf.SetXY(73, 102)
	pdf.CellFormat(8, 6, data.DestinationBy1, "0", 0, "C", false, 0, "")

	pdf.SetXY(81, 102)
	pdf.CellFormat(9, 6, data.DestinationTo2, "0", 0, "C", false, 0, "")

	pdf.SetXY(90, 102)
	pdf.CellFormat(8, 6, data.DestinationBy2, "0", 0, "C", false, 0, "")

	pdf.SetXY(98, 102)
	pdf.CellFormat(10, 6, data.Currency, "0", 0, "C", false, 0, "")

	pdf.SetXY(106, 102)
	pdf.CellFormat(8, 6, data.ChgsCode, "0", 0, "C", false, 0, "")

	pdf.SetXY(112, 102)
	pdf.CellFormat(8, 6, data.WtValPpd, "0", 0, "C", false, 0, "")

	pdf.SetXY(117, 102)
	pdf.CellFormat(8, 6, data.WtValColl, "0", 0, "C", false, 0, "")

	pdf.SetXY(122, 102)
	pdf.CellFormat(8, 6, data.OtherPpd, "0", 0, "C", false, 0, "")

	pdf.SetXY(127, 102)
	pdf.CellFormat(8, 6, data.OtherColl, "0", 0, "C", false, 0, "")

	pdf.SetXY(132, 102)
	pdf.CellFormat(31, 6, data.DeclaredValCarriage, "0", 0, "C", false, 0, "")

	pdf.SetXY(163, 102)
	pdf.CellFormat(31, 6, data.DeclaredValCustoms, "0", 0, "C", false, 0, "")

	pdf.SetXY(9, 111)
	pdf.CellFormat(45, 7, data.AirportOfDestination, "0", 0, "L", false, 0, "")

	pdf.SetXY(54, 111)
	pdf.MultiCell(22, 7, data.RequestedFlightDate1, "0", "C", false)

	pdf.SetXY(75, 111)
	pdf.MultiCell(22, 7, data.RequestedFlightDate2, "0", "C", false)

	pdf.SetXY(98, 111)
	pdf.MultiCell(28, 7, data.AmountOfInsurance, "0", "C", false)

	pdf.SetFont("THSarabunNew Bold", "", 16)
	pdf.SetXY(9, 119)
	pdf.MultiCell(156, 12, data.HandlingInfomation, "0", "L", false)

	pdf.SetXY(165, 125)
	pdf.MultiCell(30, 7, data.SCI, "0", "L", false)

	pdf.SetFont("THSarabunNew Bold", "", 15)
	dstartX := float64(8)
	dStartY := float64(143)
	pdf.SetY(dStartY)

	for _, v := range data.Items {
		currentY := pdf.GetY()
		pdf.SetX(dstartX)

		// No of Pieces RCP
		pdf.CellFormat(10, 7, v.PiecesRCP, "0", 0, "CM", false, 0, "")

		// Gross Weight
		pdf.CellFormat(19, 7, fmt.Sprintf("%.2f", v.GrossWeight), "0", 0, "CM", false, 0, "")

		// kg/lb column
		// pdf.CellFormat(8, 7, v.KgLb, "0", 0, "CM", false, 0, "")
		{
			// จุดเริ่มของคอลัมน์ Kg/Lb เดิม
			x := pdf.GetX()
			y := currentY

			nudge := -3.1 // ปรับระยะเลื่อนซ้าย (มม.) ตามต้องการ เช่น -2 หรือ -3
			colW := 8.0   // ความกว้างคอลัมน์ Kg/Lb

			pdf.SetXY(x+nudge, y)                                     // ขยับซ้ายชั่วคราว
			pdf.CellFormat(colW, 7, v.KgLb, "", 0, "L", false, 0, "") // ชิดซ้ายในคอลัมน์
			pdf.SetXY(x+colW, y)                                      // คืนตำแหน่ง X ไปท้ายคอลัมน์เดิม เพื่อไม่ให้คอลัมน์ถัดไปเลื่อนตาม
		}
		// Rate Class / Commodity Item No.
		pdf.CellFormat(17, 7, v.RateClass, "0", 0, "L", false, 0, "")

		// Chargeable Weight
		pdf.CellFormat(19, 7, fmt.Sprintf("%.2f", v.ChargeableWeight), "0", 0, "CM", false, 0, "")

		// Rate Charge
		pdf.CellFormat(22, 7, fmt.Sprintf("%.2f", v.RateCharge), "0", 0, "CM", false, 0, "")

		// Total
		pdf.CellFormat(35, 7, fmt.Sprintf("%.2f", v.Total), "0", 0, "CM", false, 0, "")

		// Nature and Quantity of Goods
		pdf.CellFormat(3, 7, "", "0", 0, "CM", false, 0, "")
		pdf.MultiCell(57, 4, v.NatureAndQuantity, "0", "L", false)
		if len(v.Dims) > 0 {
			pdf.SetFont("THSarabunNew Bold", "", 18) // เพิ่มขนาดฟอนต์เป็นสองเท่า (9 -> 18)

			// เริ่มพิมพ์ใต้แถวรายการ
			dimStartY := currentY + 7

			// Label "DIMS:" - ย้ายมาตรงกลาง
			pdf.SetXY(32, dimStartY)                                 // ย้ายจาก X=8 มาที่ X=32 เพื่อให้อยู่ตรงกลาง
			pdf.CellFormat(16, 6, "DIMS:", "", 0, "L", false, 0, "") // เพิ่มความสูงเป็น 6

			// สร้างข้อความมิติด้วยตัวเลข (%d) และเว้นวรรคคั่นแต่ละก้อน
			parts := make([]string, 0, len(v.Dims))
			for _, dim := range v.Dims {
				parts = append(parts, fmt.Sprintf("%dx%dx%d=%d CNS", dim.Length, dim.Width, dim.Height, dim.Count))
			}
			dimText := strings.Join(parts, " ")

			// ตัดบรรทัดตามความกว้าง 45 มม. และคง indent ที่ตรงกลาง
			lines := pdf.SplitText(dimText, 45)
			lineY := dimStartY
			for _, line := range lines {
				pdf.SetXY(48, lineY)                                  // ย้ายจาก X=16 มาที่ X=48 เพื่อให้อยู่ตรงกลาง
				pdf.CellFormat(45, 6, line, "", 0, "L", false, 0, "") // เพิ่มความสูงเป็น 6
				lineY += 6                                            // เพิ่มระยะห่างระหว่างบรรทัด
			}

			// พิมพ์ VOL: ใต้บรรทัดสุดท้ายของ DIMS - ย้ายมาตรงกลาง
			pdf.SetXY(32, lineY)                                                              // ย้ายจาก X=8 มาที่ X=32 เพื่อให้อยู่ตรงกลาง
			pdf.MultiCell(45, 6, fmt.Sprintf("VOL: %.3f CBM", v.TotalVolume), "", "L", false) // เพิ่มความสูงเป็น 6

			pdf.SetFont("THSarabunNew Bold", "", 15)
			pdf.SetY(pdf.GetY() + 4) // เพิ่มระยะห่างก่อนรายการถัดไป
		} else {
			pdf.Ln(2.5)
		}
	}

	pdf.SetXY(6, 204)
	pdf.MultiCell(33, 6, fmt.Sprintf("%.2f", data.Prepaid), "0", "C", false)

	pdf.SetXY(6, 212)
	pdf.MultiCell(33, 6, fmt.Sprintf("%.2f", data.ValuationCharge), "0", "C", false)

	pdf.SetXY(6, 221)
	pdf.MultiCell(33, 6, fmt.Sprintf("%.2f", data.Tax), "0", "C", false)

	pdf.SetXY(6, 230)
	pdf.MultiCell(33, 6, fmt.Sprintf("%.2f", data.TotalOtherChargesDueAgent), "0", "C", false)

	pdf.SetXY(6, 239)
	pdf.MultiCell(33, 6, fmt.Sprintf("%.2f", data.TotalOtherChargesDueCarrier), "0", "C", false)

	pdf.SetXY(6, 257)
	pdf.MultiCell(33, 6, fmt.Sprintf("%.2f", data.TotalPrepaid), "0", "C", false)

	pdf.SetXY(6, 266)
	pdf.MultiCell(33, 6, data.CurrencyConversionRates, "0", "C", false)

	// Other Charges
	pdf.SetY(208)
	for _, v := range data.Charges {
		pdf.SetX(83)
		pdf.CellFormat(45, 3.6, v.Key, "0", 0, "L", false, 0, "")
		pdf.MultiCell(20, 3.6, fmt.Sprintf("%.2f", v.Value), "0", "L", false)
		pdf.Ln(1)
	}

	pdf.SetFont("THSarabunNew Bold", "", 16)
	pdf.SetXY(82, 247)
	pdf.MultiCell(120, 5, data.Signature1, "0", "C", false)

	pdf.SetXY(82, 265)
	pdf.CellFormat(39, 5, data.Signature2Date, "0", 0, "C", false, 0, "")
	pdf.CellFormat(39, 5, data.Signature2Place, "0", 0, "C", false, 0, "")
	pdf.CellFormat(39, 5, data.Signature2Issuing, "0", 0, "C", false, 0, "")

	pdf.SetFont("THSarabunNew Bold", "", 22)
	pdf.SetXY(136, 280)
	pdf.MultiCell(51, 3, data.MAWB, "0", "L", false)

	if isPreview {
		// Add watermark
		pdf.SetFont("THSarabunNew Bold", "", 85)
		pdf.SetTextColor(200, 200, 200) // Light grey
		pdf.TransformBegin()
		pdf.TransformRotate(45, width/2, height/2) // Rotate 45 degrees
		pdf.Text(width/2-40, height/2, "P R E V I E W")
		pdf.TransformEnd()
	}

	err = pdf.Output(&buf)
	pdf.Close()

	if err == nil {
		return buf, nil
	}

	return bytes.Buffer{}, err
}
