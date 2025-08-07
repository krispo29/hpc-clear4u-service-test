package server

import (
	"context"
	"errors"
	outbound "hpc-express-service/outbound/express"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type outboundExpressHandler struct {
	s outbound.OutboundExpressService
}

func (h *outboundExpressHandler) router() chi.Router {

	r := chi.NewRouter()

	r.Post("/upload", h.uploadManifest)
	r.Get("/download/pre-export", h.downloadPreExport)

	return r
}

func (h *outboundExpressHandler) uploadManifest(w http.ResponseWriter, r *http.Request) {
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
	userUUID := GetUserUUIDFromContext(r)

	log.Println("#1 ", templateCode)

	// result, err := h.s.UploadManifest(ctx, userUUID, handler.Filename, fileBytes)
	err = h.s.UploadManifest(ctx, userUUID, handler.Filename, templateCode, fileBytes)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Respond(w, r, SuccessResponse(nil, "success"))
}

func (h *outboundExpressHandler) downloadPreExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	uploadLoggingUUID := r.URL.Query().Get("uploadLoggingUUID")
	if len(uploadLoggingUUID) == 0 {
		render.Render(w, r, ErrInvalidRequest(errors.New("required uuid")))
		return
	}

	fileName, zipBuf, err := h.s.DownloadPreExport(ctx, uploadLoggingUUID)
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
