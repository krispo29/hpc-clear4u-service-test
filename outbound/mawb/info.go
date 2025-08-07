package outbound

import "net/http"

type MawbInfoBaseModel struct {
	Mawb             string  `json:"mawb"`
	Date             string  `json:"date"`
	ServiceTypeCode  string  `json:"serviceTypeCode"`
	ShippingTypeCode string  `json:"shippingTypeCode"`
	ChargeableWeight float64 `json:"chargeableWeight"`
}

type CreateMawbInfo struct {
	*MawbInfoBaseModel
}

func (o *CreateMawbInfo) Bind(r *http.Request) error {
	return nil
}

func (o *CreateMawbInfo) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GetMawbInfo struct {
	UUID string `json:"uuid"`
	*MawbInfoBaseModel
	CreatedAt  string               `json:"createdAt"`
	UpdatedAt  string               `json:"updatedAt"`
	DeletedAt  string               `json:"deletedAt"`
	IsDeleted  bool                 `json:"isDeleted"`
	Attchments []*GetAttchmentModel `json:"attchments"`
}

type UpdateMawbInfoModel struct {
	UUID string `json:"uuid"`
	*MawbInfoBaseModel
}

func (o *UpdateMawbInfoModel) Bind(r *http.Request) error {
	return nil
}

func (o *UpdateMawbInfoModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type InsertAttchmentModel struct {
	MawbUUID string
	FileName string
	FileURL  string
}

type GetAttchmentModel struct {
	FileName string `json:"fileName"`
	FileURL  string `json:"fileURL"`
}
