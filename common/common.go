package common

type GetExchangeRateModel struct {
	Id                 int64   `json:"id"`
	ItemNo             int64   `json:"itemNo"`
	UseForCountryCode  string  `json:"useForCountryCode"`
	CountryCode        string  `json:"countryCode"`
	CountryName        string  `json:"countryName"`
	CurrencyCode       string  `json:"currencyCode"`
	CurrencyName       string  `json:"currencyName"`
	ImportExchangeRate float64 `json:"importExchangeRate"`
	ExportExchangeRate float64 `json:"exportExchangeRate"`
	IsEnabled          bool    `json:"isEnabled"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
}

type GetAllConvertTemplateModel struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// deleted_at
