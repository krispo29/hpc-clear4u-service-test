package inbound

import (
	"fmt"
	"net/http"

	"github.com/shopspring/decimal"
)

type InsertPreImportHeaderManifestModel struct {
	Mawb               string `json:"mawb" validate:"required"`
	DischargePort      string `json:"dischargePort"`
	VasselName         string `json:"vasselName"`
	ArrivalDate        string `json:"arrivalDate"`
	CustomerName       string `json:"customerName"`
	FlightNo           string `json:"flightNo"`
	OriginCountryCode  string `json:"originCountryCode"`
	OriginCurrencyCode string `json:"originCurrencyCode"`
	IsEnableCustomsOT  bool   `json:"isEnableCustomsOT"`
}

func (o *InsertPreImportHeaderManifestModel) Bind(r *http.Request) error {
	return nil
}

func (o *InsertPreImportHeaderManifestModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type UpdatePreImportHeaderManifestModel struct {
	UUID               string `json:"uuid" validate:"required"`
	Mawb               string `json:"mawb" validate:"required"`
	DischargePort      string `json:"dischargePort"`
	VasselName         string `json:"vasselName"`
	ArrivalDate        string `json:"arrivalDate"`
	CustomerName       string `json:"customerName"`
	FlightNo           string `json:"flightNo"`
	OriginCountryCode  string `json:"originCountryCode"`
	OriginCurrencyCode string `json:"originCurrencyCode"`
	IsEnableCustomsOT  bool   `json:"isEnableCustomsOT"`
}

func (o *UpdatePreImportHeaderManifestModel) Bind(r *http.Request) error {
	return nil
}

func (o *UpdatePreImportHeaderManifestModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// type InsertPreImportDetailManifestModel struct {
// 	HeaderUUID               string
// 	MasterAirWaybill         string
// 	HouseAirWaybill          string
// 	Category                 string
// 	ConsigneeTax             string
// 	ConsigneeBranch          string
// 	ConsigneeName            string
// 	ConsigneeAddress         string
// 	ConsigneeDistrict        string
// 	ConsigneeSubprovince     string
// 	ConsigneeProvince        string
// 	ConsigneePostcode        string
// 	ConsigneeCountryCode     string
// 	ConsigneeEmail           string
// 	ConsigneePhoneNumber     string
// 	ShipperName              string
// 	ShipperAddress           string
// 	ShipperDistrict          string
// 	ShipperSubprovince       string
// 	ShipperProvince          string
// 	ShipperPostcode          string
// 	ShipperCountryCode       string
// 	ShipperEmail             string
// 	ShipperPhoneNumber       string
// 	TariffCode               string
// 	TariffSequence           string
// 	StatisticalCode          string
// 	EnglishDescriptionOfGood string
// 	ThaiDescriptionOfGood    string
// 	Quantity                 int64
// 	QuantityUnitCode         string
// 	NetWeight                float64
// 	NetWeightUnitCode        string
// 	GrossWeight              float64
// 	GrossWeightUnitCode      string
// 	Package                  string
// 	PackageUnitCode          string
// 	CifValueForeign          float64
// 	FobValueForeign          float64
// 	ExchangeRate             float64
// 	CurrencyCode             string
// 	ShippingMark             string
// 	ConsignmentCountry       string
// 	FreightValueForeign      float64
// 	FreightCurrencyCode      string
// 	InsuranceValueForeign    float64
// 	InsuranceCurrencyCode    string
// 	OtherChargeValueForeign  string
// 	OtherChargeCurrencyCode  string
// 	InvoiceNo                string
// 	InvoiceDate              string
// }

type GetPreImportManifestModel struct {
	UUID               string                            `json:"uuid"`
	Mawb               string                            `json:"mawb"`
	UploadLoggingUUID  string                            `json:"uploadLoggingUUID"`
	DischargePort      string                            `json:"dischargePort"`
	VasselName         string                            `json:"vasselName"`
	ArrivalDate        string                            `json:"arrivalDate"`
	CustomerName       string                            `json:"customerName"`
	FlightNo           string                            `json:"flightNo"`
	OriginCountryCode  string                            `json:"originCountryCode"`
	OriginCurrencyCode string                            `json:"originCurrencyCode"`
	IsEnableCustomsOT  bool                              `json:"isEnableCustomsOT"`
	CreatedAt          string                            `json:"createdAt"`
	UpdatedAt          string                            `json:"updatedAt"`
	Details            []*GetPreImportManifestDetilModel `json:"details,omitempty"`
	// Summary   string // TODO:
}

type GetPreImportManifestDetilModel struct {
	UUID                     string  `json:"uuid"`
	MasterAirWaybill         string  `json:"masterAirWaybill"`
	HouseAirWaybill          string  `json:"houseAirWaybill"`
	Category                 string  `json:"category"`
	ConsigneeTax             string  `json:"consigneeTax"`
	ConsigneeBranch          string  `json:"consigneeBranch"`
	ConsigneeName            string  `json:"consigneeName"`
	ConsigneeAddress         string  `json:"consigneeAddress"`
	ConsigneeDistrict        string  `json:"consigneeDistrict"`
	ConsigneeSubprovince     string  `json:"consigneeSubprovince"`
	ConsigneeProvince        string  `json:"consigneeProvince"`
	ConsigneePostcode        string  `json:"consigneePostcode"`
	ConsigneeCountryCode     string  `json:"consigneeCountryCode"`
	ConsigneeEmail           string  `json:"consigneeEmail"`
	ConsigneePhoneNumber     string  `json:"consigneePhoneNumber"`
	ShipperName              string  `json:"shipperName"`
	ShipperAddress           string  `json:"shipperAddress"`
	ShipperDistrict          string  `json:"shipperDistrict"`
	ShipperSubprovince       string  `json:"shipperSubprovince"`
	ShipperProvince          string  `json:"shipperProvince"`
	ShipperPostcode          string  `json:"shipperPostcode"`
	ShipperCountryCode       string  `json:"shipperCountryCode"`
	ShipperEmail             string  `json:"shipperEmail"`
	ShipperPhoneNumber       string  `json:"shipperPhoneNumber"`
	TariffCode               string  `json:"tariffCode"`
	TariffSequence           string  `json:"tariffSequence"`
	StatisticalCode          string  `json:"statisticalCode"`
	EnglishDescriptionOfGood string  `json:"englishDescriptionOfGood"`
	ThaiDescriptionOfGood    string  `json:"thaiDescriptionOfGood"`
	Quantity                 int64   `json:"quantity"`
	QuantityUnitCode         string  `json:"quantityUnitCode"`
	NetWeight                float64 `json:"netWeight"`
	NetWeightUnitCode        string  `json:"netWeightUnitCode"`
	GrossWeight              float64 `json:"grossWeight"`
	GrossWeightUnitCode      string  `json:"grossWeightUnitCode"`
	Package                  string  `json:"package"`
	PackageUnitCode          string  `json:"packageUnitCode"`
	CifValueForeign          float64 `json:"cifValueForeign"`
	FobValueForeign          float64 `json:"fobValueForeign"`
	ExchangeRate             float64 `json:"exchangeRate"`
	CurrencyCode             string  `json:"currencyCode"`
	ShippingMark             string  `json:"shippingMark"`
	ConsignmentCountry       string  `json:"consignmentCountry"`
	FreightValueForeign      float64 `json:"freightValueForeign"`
	FreightCurrencyCode      string  `json:"freightCurrencyCode"`
	InsuranceValueForeign    float64 `json:"insuranceValueForeign"`
	InsuranceCurrencyCode    string  `json:"insuranceCurrencyCode"`
	OtherChargeValueForeign  string  `json:"otherChargeValueForeign"`
	OtherChargeCurrencyCode  string  `json:"otherChargeCurrencyCode"`
	InvoiceNo                string  `json:"invoiceNo"`
	InvoiceDate              string  `json:"invoiceDate"`
	CreatedAt                string  `json:"createdAt"`
	UpdatedAt                string  `json:"updatedAt"`
	Vat                      float64 `json:"vat"`
	Duty                     float64 `json:"duty"`
	IsGoodsMatched           bool    `json:"isGoodsMatched"`
}

type UpdatePreImportManifestDetailModel struct {
	UUID                     string
	MasterAirWaybill         string
	HouseAirWaybill          string
	Category                 string
	ConsigneeTax             string
	ConsigneeBranch          string
	ConsigneeName            string
	ConsigneeAddress         string
	ConsigneeDistrict        string
	ConsigneeSubprovince     string
	ConsigneeProvince        string
	ConsigneePostcode        string
	ConsigneeCountryCode     string
	ConsigneeEmail           string
	ConsigneePhoneNumber     string
	ShipperName              string
	ShipperAddress           string
	ShipperDistrict          string
	ShipperSubprovince       string
	ShipperProvince          string
	ShipperPostcode          string
	ShipperCountryCode       string
	ShipperEmail             string
	ShipperPhoneNumber       string
	TariffCode               string
	TariffSequence           string
	StatisticalCode          string
	EnglishDescriptionOfGood string
	ThaiDescriptionOfGood    string
	Quantity                 int64
	QuantityUnitCode         string
	NetWeight                float64
	NetWeightUnitCode        string
	GrossWeight              float64
	GrossWeightUnitCode      string
	Package                  string
	PackageUnitCode          string
	CifValueForeign          float64
	FobValueForeign          float64
	ExchangeRate             float64
	CurrencyCode             string
	ShippingMark             string
	ConsignmentCountry       string
	FreightValueForeign      float64
	FreightCurrencyCode      string
	InsuranceValueForeign    float64
	InsuranceCurrencyCode    string
	OtherChargeValueForeign  string
	OtherChargeCurrencyCode  string
	InvoiceNo                string
	InvoiceDate              string
}

var expectedHeadersUploadUpdateRawPreImport = []string{
	"//1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "50", "51", "52",
}

func validateHeadersUploadUpdateRawPreImport(headers []string) error {
	for i, expected := range expectedHeadersUploadUpdateRawPreImport {
		if i >= len(headers) {
			return fmt.Errorf("missing header at column %d: expected '%s'", i+1, expected)
		}
		if headers[i] != expected {
			return fmt.Errorf("header mismatch at column %d: expected '%s', got '%s'", i+1, expected, headers[i])
		}
	}
	return nil
}

type GetSummaryModel struct {
	Hawb     string
	Category string
	Vat      float64
	Duty     float64
}

type UploadSummaryModel struct {
	CustomFee          *CustomFeeModel          `json:"customFee"`
	OTCustomFee        *OTCustomFeeModel        `json:"oTCustomFee"`
	BankFeeFee         *BankFeeFeeModel         `json:"bankFeeFee"`
	CargoPermitFee     *CargoPermitFeeModel     `json:"cargoPermitFee"`
	ExpressDeliveryFee *ExpressDeliveryFeeModel `json:"expressDeliveryFee"`
	Catogory2          *CatogorySummaryModel    `json:"catogory2"`
	Catogory3          *CatogorySummaryModel    `json:"catogory3"`
	OtherCatogory      *CatogorySummaryModel    `json:"otherCatogory"`
	TotalTax           float64                  `json:"totalTax"`
	TotalHawb          int64                    `json:"totalHawb"`
}

type CatogorySummaryModel struct {
	Category           string          `json:"category"`
	Total              int64           `json:"total"`
	Vat                float64         `json:"vat"`
	Duty               float64         `json:"duty"`
	DutyAndVat         float64         `json:"dutyAndVat"`
	CustomFee          decimal.Decimal `json:"customFee"`
	OTCustomsFee       decimal.Decimal `json:"otCustomsFee"`
	BankFee            decimal.Decimal `json:"bankFee"`
	CargoPermitFee     decimal.Decimal `json:"cargoPermitFee"`
	ExpressDeliveryFee decimal.Decimal `json:"expressDeliveryFee"`
}

type CustomFeeModel struct {
	TotalDeclaration int               `json:"totalDeclaration"`
	TotalFee         decimal.Decimal   `json:"totalFee"`
	FloorPerHawb     decimal.Decimal   `json:"floorPerHawb"`
	PerHawbFees      []decimal.Decimal `json:"-"`
}

type OTCustomFeeModel struct {
	TotalDeclaration int               `json:"totalDeclaration"`
	TotalFee         decimal.Decimal   `json:"totalFee"`
	FloorPerHawb     decimal.Decimal   `json:"floorPerHawb"`
	PerHawbFees      []decimal.Decimal `json:"-"`
}

type BankFeeFeeModel struct {
	TotalDeclaration int               `json:"totalDeclaration"`
	TotalFee         decimal.Decimal   `json:"totalFee"`
	FloorPerHawb     decimal.Decimal   `json:"floorPerHawb"`
	PerHawbFees      []decimal.Decimal `json:"-"`
}

type CargoPermitFeeModel struct {
	TotalDeclaration int               `json:"totalDeclaration"`
	TotalFee         decimal.Decimal   `json:"totalFee"`
	FloorPerHawb     decimal.Decimal   `json:"floorPerHawb"`
	PerHawbFees      []decimal.Decimal `json:"-"`
}

type ExpressDeliveryFeeModel struct {
	FloorPerHawb decimal.Decimal   `json:"floorPerHawb"`
	PerHawbFees  []decimal.Decimal `json:"-"`
}
