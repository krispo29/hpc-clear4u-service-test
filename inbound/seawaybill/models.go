package seawaybill

import (
	"mime/multipart"
	"net/http"
)

// AttachmentInfo represents stored attachment metadata.
type AttachmentInfo struct {
	FileName string `json:"fileName"`
	FileURL  string `json:"fileUrl"`
	FileSize int64  `json:"fileSize"`
}

// SeaWaybillDetailResponse represents the response payload for sea waybill details.
type SeaWaybillDetailResponse struct {
	UUID         string           `json:"uuid"`
	GrossWeight  float64          `json:"grossWeight"`
	VolumeWeight float64          `json:"volumeWeight"`
	DutyTax      float64          `json:"dutyTax"`
	Attachments  []AttachmentInfo `json:"attachments"`
	CreatedAt    string           `json:"createdAt"`
	UpdatedAt    string           `json:"updatedAt"`
}

// CreateSeaWaybillDetailRequest holds form values for creating a record.
type CreateSeaWaybillDetailRequest struct {
	GrossWeight  string                  `form:"grossWeight" validate:"required"`
	VolumeWeight string                  `form:"volumeWeight" validate:"required"`
	DutyTax      string                  `form:"dutyTax" validate:"required"`
	Attachments  []*multipart.FileHeader `form:"attachments"`
}

// Bind implements render.Binder interface.
func (r *CreateSeaWaybillDetailRequest) Bind(req *http.Request) error {
	return nil
}

// UpdateSeaWaybillDetailRequest holds form values for updating a record.
type UpdateSeaWaybillDetailRequest struct {
	GrossWeight  string                  `form:"grossWeight" validate:"required"`
	VolumeWeight string                  `form:"volumeWeight" validate:"required"`
	DutyTax      string                  `form:"dutyTax" validate:"required"`
	Attachments  []*multipart.FileHeader `form:"attachments"`
}

// Bind implements render.Binder interface.
func (r *UpdateSeaWaybillDetailRequest) Bind(req *http.Request) error {
	return nil
}

// DeleteAttachmentRequest represents a request to remove an attachment.
type DeleteAttachmentRequest struct {
	FileName string `json:"fileName" validate:"required"`
}

// Bind implements render.Binder interface.
func (r *DeleteAttachmentRequest) Bind(req *http.Request) error {
	return nil
}

// seaWaybillDetailData represents sanitized data used by the repository layer.
type seaWaybillDetailData struct {
	UUID         string
	GrossWeight  float64
	VolumeWeight float64
	DutyTax      float64
	Attachments  []AttachmentInfo
}
