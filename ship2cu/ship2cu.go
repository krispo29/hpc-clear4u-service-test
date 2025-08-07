package ship2cu

import (
	"fmt"
	"hpc-express-service/utils"
	"strings"
)

type UploadManifestModel struct {
	Mawb               string
	BagNo              string
	Hawb               string
	HsCode             string
	Origin             string
	ShipperName        string
	ConsigneeName      string
	WgtValue           float64
	WgtUnit            string
	Packaging          string
	ShipperAddress     string
	ConsigneeAddress   string
	Province           string
	District           string
	Postcode           string
	Pcs                string
	Qty                int64
	Goods              string
	GoodsEN            string
	Currency           string
	TotalPrice         float64
	FOB                float64
	Freight            float64
	Insurance          float64
	CIF                float64
	Cat                string
	Duty               float64
	Vat                float64
	Cost               float64
	LocalTrackingNo    string
	Reference1         string
	Reference2         string
	CustomerCode       string
	ShipperTel         string
	ConsigneeTel       string
	Dimension          string
	DimensionRepacking string
	Width              float64
	Length             float64
	Height             float64
	VolumeWeight       float64
	// FreightRate        float64
	// FreightZone        float64
}

type ResponseUploadManifestModel struct {
	Mawb   string
	Amount int64
}

func (d *UploadManifestModel) ConvertToManifest(shipperBrands []*GetShipperBrandModel, masterHsCodeData []*GetMasterHsCodeModel, freightConfig *GetFreightDataModel) *utils.InsertPreImportDetailManifestModel {
	//cif_value_foreign = (d.TotalPrice * exchange_rate) + (d.WgtValue + freight_rate) + ((d.TotalPrice * exchange_rate) * 0.01)

	foundShipperBrands := &GetShipperBrandModel{}
	foundHsCode := &GetMasterHsCodeModel{}

	for _, brand := range shipperBrands {
		if brand.ShipperCountryCode == d.Origin && brand.ShipperName == d.ShipperName {
			foundShipperBrands = brand
			break
		}
	}

	for _, row := range masterHsCodeData {
		if strings.ToUpper(row.GoodsEN) == strings.ToUpper(d.Goods) {
			foundHsCode = row
			break
		}
	}

	var category, tariffSequence string
	var fob, freight, insurance, cif float64
	// if cif == 0 {
	// fob + insurance + freight
	fob = d.TotalPrice * freightConfig.FreightRate
	freight = d.WgtValue * freightConfig.FreightZone
	insurance = fob * 0.01
	cif = fob + freight + insurance
	// } else {
	// 	fob = d.FOB
	// 	freight = d.Freight
	// 	insurance = d.Insurance
	// 	cif = d.CIF
	// }

	if cif <= 1500 {
		category = "2"
		tariffSequence = "68001"
	} else if cif >= 1500 {
		category = "3"
		tariffSequence = foundHsCode.TariffSequence
	} else {
		category = ""
		tariffSequence = ""
	}

	return &utils.InsertPreImportDetailManifestModel{
		HeaderUUID:               "",
		MasterAirWaybill:         d.Mawb,
		HouseAirWaybill:          d.Hawb,
		Category:                 category,
		ConsigneeTax:             "",
		ConsigneeBranch:          "",
		ConsigneeName:            d.ConsigneeName,
		ConsigneeAddress:         d.ConsigneeAddress,
		ConsigneeDistrict:        d.District,
		ConsigneeSubprovince:     "",
		ConsigneeProvince:        d.Province,
		ConsigneePostcode:        d.Postcode,
		ConsigneeCountryCode:     "TH",
		ConsigneeEmail:           "",
		ConsigneePhoneNumber:     d.ConsigneeTel,
		ShipperName:              d.ShipperName,
		ShipperAddress:           foundShipperBrands.ShipperAddress,
		ShipperDistrict:          foundShipperBrands.ShipperDistrict,    // FROM master_shipper_brands
		ShipperSubprovince:       foundShipperBrands.ShipperSubprovince, // FROM master_shipper_brands
		ShipperProvince:          foundShipperBrands.ShipperProvince,    // FROM master_shipper_brands
		ShipperPostcode:          foundShipperBrands.ShipperPostcode,    // FROM master_shipper_brands
		ShipperCountryCode:       d.Origin,
		ShipperEmail:             "",
		ShipperPhoneNumber:       d.ShipperTel,
		TariffCode:               foundHsCode.TariffCode,      // FROM master_hs_code
		TariffSequence:           tariffSequence,              // FROM master_hs_code,
		StatisticalCode:          foundHsCode.StatisticalCode, // FROM master_hs_code
		EnglishDescriptionOfGood: d.Goods,
		ThaiDescriptionOfGood:    d.GoodsEN,
		Quantity:                 d.Qty,
		QuantityUnitCode:         foundHsCode.QuantityUnitCode, // FROM master_hs_code => mhc.unit_code = 'KGM' THEN 'C6
		NetWeight:                d.WgtValue,
		NetWeightUnitCode:        "KGM",
		GrossWeight:              d.WgtValue, // d.VolumeWeight,
		GrossWeightUnitCode:      "KGM",
		Package:                  "1",
		PackageUnitCode:          "PK",
		CifValueForeign:          cif,
		FobValueForeign:          d.TotalPrice,
		ExchangeRate:             freightConfig.FreightRate, // FROM exchange_rate_cte.exchange_rate
		CurrencyCode:             d.Currency,
		ShippingMark:             d.Mawb,
		ConsignmentCountry:       "KR",
		FreightValueForeign:      freight, // x.weight * freight_rate_cte.freight_rate AS freight_value_foreign,
		FreightCurrencyCode:      "THB",
		InsuranceValueForeign:    insurance, // (x.total_price * exchange_rate_cte.exchange_rate) * 0.01 AS insurance_value_foreign,
		InsuranceCurrencyCode:    "THB",
		OtherChargeValueForeign:  "",
		OtherChargeCurrencyCode:  "",
		InvoiceNo:                d.Hawb,
		InvoiceDate:              "", // to_char(now() AT TIME ZONE 'utc' AT TIME ZONE 'Asia/Bangkok', 'DD/MM/YYYY') AS invoice_date
	}
}

// (8200*0.0235)+(1.31*260)+((8200*0/0235)*0.01)

type GetShipperBrandModel struct {
	ShipperName        string
	ShipperAddress     string
	ShipperDistrict    string
	ShipperSubprovince string
	ShipperProvince    string
	ShipperPostcode    string
	ShipperCountryCode string
}

type GetMasterHsCodeModel struct {
	GoodsEN          string
	TariffCode       string
	TariffSequence   string
	StatisticalCode  string
	QuantityUnitCode string
}

var expectedHeadersUploadPreImportManifests = []string{
	"No.", "MAWB", "BAG No.", "AWB", "HS CODE", "Origin", "Shipper name", "Cnee name", "Wgt Value", "Wgt Unit", "Packaging Type", "Shpr add", "Cnee add", "Province", "District", "Postcode", "Pcs", "QTY", "Goods", "Goods(EN)", "Currency", "total price", "FOB(THB)", "Freight", "Insurance", "CIF", "CAT", "Duty", "Vat", "Cost (THB)", "Local Tracking No. / EMS no.", "Reference1", "Reference2", "Customer Code", "Shpr's tel", "Cnee's tel", "Dimension", "Dimension Repacking", "Width", "Length", "Height", "Volume Weight",
}

func validateHeadersUploadPreImportManifests(headers []string) error {
	for i, expected := range expectedHeadersUploadPreImportManifests {
		if i >= len(headers) {
			return fmt.Errorf("missing header at column %d: expected '%s'", i+1, expected)
		}
		if headers[i] != expected {
			return fmt.Errorf("header mismatch at column %d: expected '%s', got '%s'", i+1, expected, headers[i])
		}
	}
	return nil
}

type GetFreightDataModel struct {
	FreightRate float64
	FreightZone float64
}
