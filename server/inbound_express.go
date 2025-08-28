package server

import (
	"context"
	"errors"
	inbound "hpc-express-service/inbound/express"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type inboundExpressHandler struct {
	s inbound.InboundExpressService
}

func (h *inboundExpressHandler) router() chi.Router {

	r := chi.NewRouter()

	r.Route("/mawb", func(r chi.Router) {
		r.Route("/download", func(r chi.Router) {
			r.Get("/pre-import/{headerUUID}", h.downloadPreImport)
			r.Get("/raw-pre-import/{headerUUID}", h.downloadRawPreImport)
		})
		r.Route("/upload", func(r chi.Router) {
			r.Post("/", h.uploadManifestDetails)
			r.Post("/update-raw-manifest", h.uploadUpdateRawPreImport)
		})
		r.Get("/", h.getAllMawb)
		r.Post("/", h.createMawb)
		r.Put("/", h.updateMawb)
		r.Get("/{headerUUID}/summary", h.getSummary)
		r.Get("/{headerUUID}", h.getManifest)
		// r.Put("/", h.updateMawb)
	})

	// r.Post("/upload-update-manifest", h.uploadUpdateManifest)
	// r.Route("/upload", func(r chi.Router) {
	// r.Post("/", h.uploadManifestDetails)
	// r.Get("/{uploadLoggingUUID}/summary", h.getSummary)
	// r.Get("/{uploadLoggingUUID}", h.getManifest)
	// })

	// r.Route("/download", func(r chi.Router) {
	// 	r.Get("/pre-import", h.downloadPreImport)
	// 	r.Get("/raw-pre-import", h.downloadRawPreImport)
	// })

	return r
}

func (h *inboundExpressHandler) getAllMawb(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := h.s.GetAllMawb(r.Context())
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *inboundExpressHandler) createMawb(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	data := &inbound.InsertPreImportHeaderManifestModel{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate Data
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	uuid, err := h.s.InsertPreImportManifestHeader(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(uuid, "success"))
}

func (h *inboundExpressHandler) updateMawb(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
		_ = ctx
	}

	data := &inbound.UpdatePreImportHeaderManifestModel{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate Data
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	err = h.s.UpdatePreImportManifestHeader(r.Context(), data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}

func (h *inboundExpressHandler) uploadManifestDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	r.ParseMultipartForm(10 << uint32(20)) // 10 * 2^20
	file, handler, err := r.FormFile("file")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	templateCode := r.FormValue("templateCode")
	headerUUID := r.FormValue("headerUUID")
	userUUID := GetUserUUIDFromContext(r)

	log.Println("#1 ", templateCode)

	// result, err := h.s.UploadManifest(ctx, userUUID, handler.Filename, fileBytes)
	err = h.s.UploadManifestDetails(ctx, userUUID, headerUUID, handler.Filename, templateCode, fileBytes)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}

func (h *inboundExpressHandler) downloadPreImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	headerUUID := chi.URLParam(r, "headerUUID")
	if len(headerUUID) == 0 {
		render.Render(w, r, ErrInvalidRequest(errors.New("required uuid")))
		return
	}

	fileName, zipBuf, err := h.s.DownloadPreImport(ctx, headerUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Send ZIP file as response
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`.zip"`)
	w.WriteHeader(http.StatusOK)
	w.Write(zipBuf.Bytes())
}

func (h *inboundExpressHandler) downloadRawPreImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	headerUUID := chi.URLParam(r, "headerUUID")
	if len(headerUUID) == 0 {
		render.Render(w, r, ErrInvalidRequest(errors.New("required uuid")))
		return
	}

	fileName, excelBuf, err := h.s.DownloadRawPreImport(ctx, headerUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Send ZIP file as response
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("File-Name", fileName)
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	w.Write(excelBuf.Bytes())
}

func (h *inboundExpressHandler) uploadUpdateRawPreImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	r.ParseMultipartForm(10 << uint32(20)) // 10 * 2^20
	file, handler, err := r.FormFile("file")
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	headerUUID := r.FormValue("headerUUID")
	userUUID := GetUserUUIDFromContext(r)

	err = h.s.UploadUpdateRawPreImport(ctx, userUUID, headerUUID, handler.Filename, fileBytes)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}

func (h *inboundExpressHandler) getManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	headerUUID := chi.URLParam(r, "headerUUID")

	result, err := h.s.GetOneByHeaderUUID(r.Context(), headerUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}

func (h *inboundExpressHandler) getSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	headerUUID := chi.URLParam(r, "headerUUID")

	result, err := h.s.GetSummaryByHeaderUUID(r.Context(), headerUUID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(result, "success"))
}
