package mawbinfo

import (
	"mime/multipart"
	"net/http"
)

// CreateMawbInfoRequest represents the request payload for creating MAWB info
type CreateMawbInfoRequest struct {
	ChargeableWeight string `json:"chargeableWeight" validate:"required"`
	Date             string `json:"date" validate:"required"`
	Mawb             string `json:"mawb" validate:"required"`
	ServiceType      string `json:"serviceType" validate:"required"`
	ShippingType     string `json:"shippingType" validate:"required"`
}

// Bind implements the chi render.Binder interface for HTTP request binding
func (r *CreateMawbInfoRequest) Bind(req *http.Request) error {
	return nil
}

// MawbInfoResponse represents the response payload for MAWB info operations
type MawbInfoResponse struct {
	UUID             string           `json:"uuid"`
	ChargeableWeight float64          `json:"chargeableWeight"`
	Date             string           `json:"date"`
	Mawb             string           `json:"mawb"`
	ServiceType      string           `json:"serviceType"`
	ShippingType     string           `json:"shippingType"`
	CreatedAt        string           `json:"createdAt"`
	Attachments      []AttachmentInfo `json:"attachments,omitempty"`
}

// AttachmentInfo represents file attachment information
type AttachmentInfo struct {
	FileName string `json:"fileName"`
	FileURL  string `json:"fileUrl"`
	FileSize int64  `json:"fileSize"`
}

// UpdateMawbInfoRequest represents the request payload for updating MAWB info
type UpdateMawbInfoRequest struct {
	ChargeableWeight string                  `form:"chargeableWeight" validate:"required"`
	Date             string                  `form:"date" validate:"required"`
	Mawb             string                  `form:"mawb" validate:"required"`
	ServiceType      string                  `form:"serviceType" validate:"required"`
	ShippingType     string                  `form:"shippingType" validate:"required"`
	Attachments      []*multipart.FileHeader `form:"attachments"`
}

// Bind implements the chi render.Binder interface for HTTP request binding
func (r *UpdateMawbInfoRequest) Bind(req *http.Request) error {
	return nil
}

// Render implements the chi render.Renderer interface for HTTP response rendering
func (r *MawbInfoResponse) Render(w http.ResponseWriter, req *http.Request) error {
	return nil
}

// DeleteAttachmentRequest represents the request payload for deleting a MAWB attachment
type DeleteAttachmentRequest struct {
	FileName string `json:"fileName" validate:"required"`
}

// Bind implements the chi render.Binder interface for HTTP request binding
func (r *DeleteAttachmentRequest) Bind(req *http.Request) error {
	return nil
}
