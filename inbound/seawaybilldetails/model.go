package seawaybilldetails

import (
	"mime/multipart"
	"net/http"
)

// AttachmentInfo represents metadata for a stored attachment file.
type AttachmentInfo struct {
	FileName string `json:"fileName"`
	FileURL  string `json:"fileUrl"`
	FileSize int64  `json:"fileSize"`
}

// SeaWaybillDetails captures the persisted details and attachments for a sea waybill record.
type SeaWaybillDetails struct {
	UUID         string           `json:"uuid"`
	GrossWeight  float64          `json:"grossWeight"`
	VolumeWeight float64          `json:"volumeWeight"`
	DutyTax      float64          `json:"dutyTax"`
	Attachments  []AttachmentInfo `json:"attachments,omitempty"`
	CreatedAt    string           `json:"createdAt"`
	UpdatedAt    string           `json:"updatedAt"`
}

// UpsertSeaWaybillDetailsRequest represents the multipart form payload for creating or updating a record.
type UpsertSeaWaybillDetailsRequest struct {
	GrossWeight  string                  `form:"grossWeight" validate:"required"`
	VolumeWeight string                  `form:"volumeWeight" validate:"required"`
	DutyTax      string                  `form:"dutyTax" validate:"required"`
	Attachments  []*multipart.FileHeader `form:"attachments"`
}

// Bind satisfies the render.Binder interface.
func (r *UpsertSeaWaybillDetailsRequest) Bind(req *http.Request) error {
	return nil
}

// DeleteAttachmentRequest represents the body for deleting a specific attachment.
type DeleteAttachmentRequest struct {
	FileName string `json:"fileName" validate:"required"`
}

// Bind satisfies the render.Binder interface.
func (r *DeleteAttachmentRequest) Bind(req *http.Request) error {
	return nil
}
