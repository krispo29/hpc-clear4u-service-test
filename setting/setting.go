package setting

import "net/http"

type HsCodeBaseModel struct {
	GoodsEN                           string  `json:"goodsEN"`
	GoodsTH                           string  `json:"goodsTH"`
	HsCode                            string  `json:"hsCode"`
	Tariff                            string  `json:"tariff"`
	Stat                              string  `json:"stat"`
	UnitCode                          string  `json:"unitCode"`
	DutyRate                          float64 `json:"dutyRate"`
	Remark                            string  `json:"remark"`
	AirServiceCharge                  int64   `json:"airServiceCharge"`
	SeaServiceCharge                  int64   `json:"seaServiceCharge"`
	FobPriceControl                   int64   `json:"fobPriceControl"`
	FobPriceControlOriginCurrencyCode string  `json:"fobPriceControlOriginCurrencyCode"`
	FobPriceControlOriginCountryCode  string  `json:"fobPriceControlOriginCountryCode"`
	WeightControl                     int64   `json:"weightControl"`
	WeightControlUnitCode             string  `json:"weightControlUnitCode"`
	CifControl                        int64   `json:"cifControl"`
	CifControlDestinationCurrencyCode string  `json:"cifControlDestinationCurrencyCode"`
	CifControlDestinationCountryCode  string  `json:"cifControlDestinationCountryCode"`
}

type CreateHsCodeModel struct {
	*HsCodeBaseModel
}

func (o *CreateHsCodeModel) Bind(r *http.Request) error {
	return nil
}

func (o *CreateHsCodeModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GetHsCodeModel struct {
	UUID string `json:"uuid"`
	*HsCodeBaseModel
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	DeletedAt string `json:"deletedAt"`
	IsDeleted bool   `json:"isDeleted"`
}

type UpdateHsCodeModel struct {
	UUID string `json:"uuid"`
	*HsCodeBaseModel
}

func (o *UpdateHsCodeModel) Bind(r *http.Request) error {
	return nil
}

func (o *UpdateHsCodeModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
