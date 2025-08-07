package utils

type InsertPreImportHeaderManifestModel struct {
	UploadLoggingUUID  string
	DischargePort      string
	VasselName         string
	ArrivalDate        string
	CustomerName       string
	OriginCountryCode  string
	OriginCurrencyCode string
	Details            []*InsertPreImportDetailManifestModel
}

type InsertPreImportDetailManifestModel struct {
	HeaderUUID               string
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
