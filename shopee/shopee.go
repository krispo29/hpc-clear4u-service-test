package shopee

import (
	"fmt"
	"hpc-express-service/utils"
	"strings"
)

type UploadManifestModel struct {
	OutboundTime            string
	LMTracking              string
	ShopeeTracking          string
	OrderSN                 string
	InvoiceDate             string
	CartonNo                string
	UnitCode                string
	DispatchNumber          string
	CartonSize              string
	ParcelWeight            float64
	ParcelSize              string
	CartonWeight            float64
	CartonVolume            string
	ParcelVolume            int64
	Transportation          string
	Channel                 string
	ServiceCode             string
	Country                 string
	DestinationCode         string
	ReceiverName            string
	ReceiverProvinceOrState string
	ReceiverCity            string
	PostalCode              string
	ReceiverTelephone       string
	ReceiverAddress         string
	SenderName              string
	SenderCountry           string
	SenderProvince          string
	SenderCity              string
	SenderAddress           string
	SenderTelephone         string
	SellerTaxNumber         string
	BrInvoiceNumber         string
	Declareusername         string
	Declareusertelephone    string
	DeclareuserID           string
	KYCPopulationID         string
	IncludingShoes          string
	FootwearQuantity        string
	FootwearDeclareValue    string
	PackageQTY              string

	// DeclaredName1       string
	// HSCode1             string
	// ProductName1        string
	// DeclaredValue1      float64
	// DeclaredQTY1        int64
	// DeclaredCategoryid1 string
	// DeclaredNameLocal1  string

	// DeclaredName2       string
	// HSCode2             string
	// ProductName2        string
	// DeclaredValue2      float64
	// DeclaredQTY2        int64
	// DeclaredCategoryid2 string
	// DeclaredNameLocal2  string

	// DeclaredName3       string
	// HSCode3             string
	// ProductName3        string
	// DeclaredValue3      float64
	// DeclaredQTY3        int64
	// DeclaredCategoryid3 string
	// DeclaredNameLocal3  string

	// DeclaredName4       string
	// HSCode4             string
	// ProductName4        string
	// DeclaredValue4      float64
	// DeclaredQTY4        int64
	// DeclaredCategoryid4 string
	// DeclaredNameLocal4  string

	// DeclaredName5       string
	// HSCode5             string
	// ProductName5        string
	// DeclaredValue5      float64
	// DeclaredQTY5        int64
	// DeclaredCategoryid5 string
	// DeclaredNameLocal5  string

	// DeclaredName6       string
	// HSCode6             string
	// ProductName6        string
	// DeclaredValue6      float64
	// DeclaredQTY6        int64
	// DeclaredCategoryid6 string
	// DeclaredNameLocal6  string

	// DeclaredName7       string
	// HSCode7             string
	// ProductName7        string
	// DeclaredValue7      float64
	// DeclaredQTY7        int64
	// DeclaredCategoryid7 string
	// DeclaredNameLocal7  string

	// DeclaredName8       string
	// HSCode8             string
	// ProductName8        string
	// DeclaredValue8      float64
	// DeclaredQTY8        int64
	// DeclaredCategoryid8 string
	// DeclaredNameLocal8  string

	// DeclaredName9       string
	// HSCode9             string
	// ProductName9        string
	// DeclaredValue9      float64
	// DeclaredQTY9        int64
	// DeclaredCategoryid9 string
	// DeclaredNameLocal9  string

	// DeclaredName10       string
	// HSCode10             string
	// ProductName10        string
	// DeclaredValue10      float64
	// DeclaredQTY10        int64
	// DeclaredCategoryid10 string
	// DeclaredNameLocal10  string

	// DeclaredName11       string
	// HSCode11             string
	// ProductName11        string
	// DeclaredValue11      float64
	// DeclaredQTY11        int64
	// DeclaredCategoryid11 string
	// DeclaredNameLocal11  string

	// DeclaredName12       string
	// HSCode12             string
	// ProductName12        string
	// DeclaredValue12      float64
	// DeclaredQTY12        int64
	// DeclaredCategoryid12 string
	// DeclaredNameLocal12  string

	// DeclaredName13       string
	// HSCode13             string
	// ProductName13        string
	// DeclaredValue13      float64
	// DeclaredQTY13        int64
	// DeclaredCategoryid13 string
	// DeclaredNameLocal13  string

	// DeclaredName14       string
	// HSCode14             string
	// ProductName14        string
	// DeclaredValue14      float64
	// DeclaredQTY14        int64
	// DeclaredCategoryid14 string
	// DeclaredNameLocal14  string

	// DeclaredName15       string
	// HSCode15             string
	// ProductName15        string
	// DeclaredValue15      float64
	// DeclaredQTY15        int64
	// DeclaredCategoryid15 string
	// DeclaredNameLocal15  string

	// DeclaredName16       string
	// HSCode16             string
	// ProductName16        string
	// DeclaredValue16      float64
	// DeclaredQTY16        int64
	// DeclaredCategoryid16 string
	// DeclaredNameLocal16  string

	// DeclaredName17       string
	// HSCode17             string
	// ProductName17        string
	// DeclaredValue17      float64
	// DeclaredQTY17        int64
	// DeclaredCategoryid17 string
	// DeclaredNameLocal17  string

	// DeclaredName18       string
	// HSCode18             string
	// ProductName18        string
	// DeclaredValue18      float64
	// DeclaredQTY18        int64
	// DeclaredCategoryid18 string
	// DeclaredNameLocal18  string

	// DeclaredName19       string
	// HSCode19             string
	// ProductName19        string
	// DeclaredValue19      float64
	// DeclaredQTY19        int64
	// DeclaredCategoryid19 string
	// DeclaredNameLocal19  string

	// DeclaredName20       string
	// HSCode20             string
	// ProductName20        string
	// DeclaredValue20      float64
	// DeclaredQTY20        int64
	// DeclaredCategoryid20 string
	// DeclaredNameLocal20  string

	// DeclaredName21       string
	// HSCode21             string
	// ProductName21        string
	// DeclaredValue21      float64
	// DeclaredQTY21        int64
	// DeclaredCategoryid21 string
	// DeclaredNameLocal21  string

	// DeclaredName22       string
	// HSCode22             string
	// ProductName22        string
	// DeclaredValue22      float64
	// DeclaredQTY22        int64
	// DeclaredCategoryid22 string
	// DeclaredNameLocal22  string

	// DeclaredName23       string
	// HSCode23             string
	// ProductName23        string
	// DeclaredValue23      float64
	// DeclaredQTY23        int64
	// DeclaredCategoryid23 string
	// DeclaredNameLocal23  string

	// DeclaredName24       string
	// HSCode24             string
	// ProductName24        string
	// DeclaredValue24      float64
	// DeclaredQTY24        int64
	// DeclaredCategoryid24 string
	// DeclaredNameLocal24  string

	// DeclaredName25       string
	// HSCode25             string
	// ProductName25        string
	// DeclaredValue25      float64
	// DeclaredQTY25        int64
	// DeclaredCategoryid25 string
	// DeclaredNameLocal25  string

	// DeclaredName26       string
	// HSCode26             string
	// ProductName26        string
	// DeclaredValue26      float64
	// DeclaredQTY26        int64
	// DeclaredCategoryid26 string
	// DeclaredNameLocal26  string

	// DeclaredName27       string
	// HSCode27             string
	// ProductName27        string
	// DeclaredValue27      float64
	// DeclaredQTY27        int64
	// DeclaredCategoryid27 string
	// DeclaredNameLocal27  string

	// DeclaredName28       string
	// HSCode28             string
	// ProductName28        string
	// DeclaredValue28      float64
	// DeclaredQTY28        int64
	// DeclaredCategoryid28 string
	// DeclaredNameLocal28  string

	// DeclaredName29       string
	// HSCode29             string
	// ProductName29        string
	// DeclaredValue29      float64
	// DeclaredQTY29        int64
	// DeclaredCategoryid29 string
	// DeclaredNameLocal29  string

	// DeclaredName30       string
	// HSCode30             string
	// ProductName30        string
	// DeclaredValue30      float64
	// DeclaredQTY30        int64
	// DeclaredCategoryid30 string
	// DeclaredNameLocal30  string

	// DeclaredName31       string
	// HSCode31             string
	// ProductName31        string
	// DeclaredValue31      float64
	// DeclaredQTY31        int64
	// DeclaredCategoryid31 string
	// DeclaredNameLocal31  string

	// DeclaredName32       string
	// HSCode32             string
	// ProductName32        string
	// DeclaredValue32      float64
	// DeclaredQTY32        int64
	// DeclaredCategoryid32 string
	// DeclaredNameLocal32  string

	// DeclaredName33       string
	// HSCode33             string
	// ProductName33        string
	// DeclaredValue33      float64
	// DeclaredQTY33        int64
	// DeclaredCategoryid33 string
	// DeclaredNameLocal33  string

	// DeclaredName34       string
	// HSCode34             string
	// ProductName34        string
	// DeclaredValue34      float64
	// DeclaredQTY34        int64
	// DeclaredCategoryid34 string
	// DeclaredNameLocal34  string

	// DeclaredName35       string
	// HSCode35             string
	// ProductName35        string
	// DeclaredValue35      float64
	// DeclaredQTY35        int64
	// DeclaredCategoryid35 string
	// DeclaredNameLocal35  string

	// DeclaredName36       string
	// HSCode36             string
	// ProductName36        string
	// DeclaredValue36      float64
	// DeclaredQTY36        int64
	// DeclaredCategoryid36 string
	// DeclaredNameLocal36  string

	// DeclaredName37       string
	// HSCode37             string
	// ProductName37        string
	// DeclaredValue37      float64
	// DeclaredQTY37        int64
	// DeclaredCategoryid37 string
	// DeclaredNameLocal37  string

	// DeclaredName38       string
	// HSCode38             string
	// ProductName38        string
	// DeclaredValue38      float64
	// DeclaredQTY38        int64
	// DeclaredCategoryid38 string
	// DeclaredNameLocal38  string

	// DeclaredName39       string
	// HSCode39             string
	// ProductName39        string
	// DeclaredValue39      float64
	// DeclaredQTY39        int64
	// DeclaredCategoryid39 string
	// DeclaredNameLocal39  string

	// DeclaredName40       string
	// HSCode40             string
	// ProductName40        string
	// DeclaredValue40      float64
	// DeclaredQTY40        int64
	// DeclaredCategoryid40 string
	// DeclaredNameLocal40  string
	DeclaredDetails []*DeclaredDetailModel
}

type DeclaredDetailModel struct {
	DeclaredName       string
	HSCode             string
	ProductName        string
	DeclaredValue      float64
	DeclaredQTY        int64
	DeclaredCategoryid string
	DeclaredNameLocal  string
}

type ResponseUploadManifestModel struct {
	Mawb   string
	Amount int64
}

func (d *UploadManifestModel) ConvertToManifest() *utils.InsertPreExportDetailManifestModel {
	// var thaiDescriptionOfGoods string
	var englishDescriptionOfGoods []string
	var category int64
	var quantity int64
	var fobValueBaht float64
	for _, v := range d.DeclaredDetails {
		englishDescriptionOfGoods = append(englishDescriptionOfGoods, v.DeclaredName)
		quantity += v.DeclaredQTY
		fobValueBaht += v.DeclaredValue
	}

	if fobValueBaht > 1500 {
		category = 3
	} else {
		category = 2
	}

	return &utils.InsertPreExportDetailManifestModel{
		HeaderUUID:                  "",
		MasterAirWaybill:            "",
		HouseAirWaybill:             d.ShopeeTracking,
		Category:                    category,
		ConsignorCompanyTaxNumber:   "01234567890123",
		ConsignorCompanyBranch:      "",
		ConsignorName:               "TEST CONSIGNEE NAME",
		ConsignorStreetAndAddress:   "161 SUKSAWAD ROAD",
		ConsignorDistrict:           "CSN_DISTRICT",
		ConsignorSubProvince:        "CSN_SUBPROVINCE",
		ConsignorProvince:           "CSN_PROVINCE",
		ConsignorPostcode:           "10140",
		ConsignorEmail:              "SHP_EMAIL",
		ConsigneeName:               d.ReceiverName,
		ConsigneeStreetAndAddress:   d.ReceiverAddress + " " + d.ReceiverTelephone,
		ConsigneeDistrict:           d.ReceiverCity,
		ConsigneeSubProvince:        "",
		ConsigneeProvince:           d.ReceiverProvinceOrState,
		ConsigneePostcode:           d.PostalCode,
		ConsigneeCountryCode:        d.DestinationCode,
		ConsigneeEmail:              "",
		PurchaseCountryCode:         d.SenderCountry,
		DestinationCountryCode:      d.DestinationCode,
		ThaiDescriptionOfGoods:      "",
		EnglishDescriptionOfGoods:   strings.Join(englishDescriptionOfGoods, ", "),
		Quantity:                    quantity,
		QuantityUnitCode:            "C62",
		NetWeight:                   d.ParcelWeight,
		NetWeightUnitCode:           "KGM",
		GrossWeight:                 d.ParcelWeight,
		GrossWeightUnitCode:         "KGM",
		PackageAmount:               d.ParcelVolume,
		PackageUnitCode:             "PK",
		Remark:                      "REMARK",
		FobValueBaht:                fobValueBaht, // Declared Value 1 (AS) to 40
		FobValueForeign:             0,            // X
		CurrencyCode:                "THB",
		ExchangeRate:                1,
		FreightAmount:               1,
		FreightAmountCurrencyCode:   "THB",
		InsuranceAmount:             1,
		InsuranceAmountCurrencyCode: "THB",
		TariffCode:                  "000049111090",
		StatCode:                    "000",
		TariffSequence:              "50001",
	}
}

var expectedHeadersUploadPreImportManifests = []string{
	"Outbound Time", "LM Tracking", "Shopee Tracking", "Order SN", "Invoice Date", "Carton No", "Unit Code", "Dispatch Number", "Carton Size", "Parcel Weight(KG)", "Parcel Size", "Carton Weight(KG)", "Carton Volume", "Parcel Volume", "Transportation", "Channel", "Service Code", "Country", "Destination Code", "Receiver Name", "Receiver Province/State", "Receiver City", "Postal Code", "Receiver Telephone", "Receiver Address", "Sender Name", "Sender Country", "Sender Province", "Sender City", "Sender Address", "Sender Telephone", "Seller Tax Number", "Br Invoice Number", "Declare user name", "Declare user telephone", "Declare user ID", "KYC Population ID", "Including Shoes", "Footwear Quantity", "Footwear Declare Value", "Package QTY", "Declared Name 1", "HS Code 1", "Product Name 1", "Declared Value 1", "Declared QTY 1", "Declared Category id 1", "Declared Name Local 1", "Declared Name 2", "HS Code 2", "Product Name 2", "Declared Value 2", "Declared QTY 2", "Declared Category id 2", "Declared Name Local 2", "Declared Name 3", "HS Code 3", "Product Name 3", "Declared Value 3", "Declared QTY 3", "Declared Category id 3", "Declared Name Local 3", "Declared Name 4", "HS Code 4", "Product Name 4", "Declared Value 4", "Declared QTY 4", "Declared Category id 4", "Declared Name Local 4", "Declared Name 5", "HS Code 5", "Product Name 5", "Declared Value 5", "Declared QTY 5", "Declared Category id 5", "Declared Name Local 5", "Declared Name 6", "HS Code 6", "Product Name 6", "Declared Value 6", "Declared QTY 6", "Declared Category id 6", "Declared Name Local 6", "Declared Name 7", "HS Code 7", "Product Name 7", "Declared Value 7", "Declared QTY 7", "Declared Category id 7", "Declared Name Local 7", "Declared Name 8", "HS Code 8", "Product Name 8", "Declared Value 8", "Declared QTY 8", "Declared Category id 8", "Declared Name Local 8", "Declared Name 9", "HS Code 9", "Product Name 9", "Declared Value 9", "Declared QTY 9", "Declared Category id 9", "Declared Name Local 9", "Declared Name 10", "HS Code 10", "Product Name 10", "Declared Value 10", "Declared QTY 10", "Declared Category id 10", "Declared Name Local 10", "Declared Name 11", "HS Code 11", "Product Name 11", "Declared Value 11", "Declared QTY 11", "Declared Category id 11", "Declared Name Local 11", "Declared Name 12", "HS Code 12", "Product Name 12", "Declared Value 12", "Declared QTY 12", "Declared Category id 12", "Declared Name Local 12", "Declared Name 13", "HS Code 13", "Product Name 13", "Declared Value 13", "Declared QTY 13", "Declared Category id 13", "Declared Name Local 13", "Declared Name 14", "HS Code 14", "Product Name 14", "Declared Value 14", "Declared QTY 14", "Declared Category id 14", "Declared Name Local 14", "Declared Name 15", "HS Code 15", "Product Name 15", "Declared Value 15", "Declared QTY 15", "Declared Category id 15", "Declared Name Local 15", "Declared Name 16", "HS Code 16", "Product Name 16", "Declared Value 16", "Declared QTY 16", "Declared Category id 16", "Declared Name Local 16", "Declared Name 17", "HS Code 17", "Product Name 17", "Declared Value 17", "Declared QTY 17", "Declared Category id 17", "Declared Name Local 17", "Declared Name 18", "HS Code 18", "Product Name 18", "Declared Value 18", "Declared QTY 18", "Declared Category id 18", "Declared Name Local 18", "Declared Name 19", "HS Code 19", "Product Name 19", "Declared Value 19", "Declared QTY 19", "Declared Category id 19", "Declared Name Local 19", "Declared Name 20", "HS Code 20", "Product Name 20", "Declared Value 20", "Declared QTY 20", "Declared Category id 20", "Declared Name Local 20", "Declared Name 21", "HS Code 21", "Product Name 21", "Declared Value 21", "Declared QTY 21", "Declared Category id 21", "Declared Name Local 21", "Declared Name 22", "HS Code 22", "Product Name 22", "Declared Value 22", "Declared QTY 22", "Declared Category id 22", "Declared Name Local 22", "Declared Name 23", "HS Code 23", "Product Name 23", "Declared Value 23", "Declared QTY 23", "Declared Category id 23", "Declared Name Local 23", "Declared Name 24", "HS Code 24", "Product Name 24", "Declared Value 24", "Declared QTY 24", "Declared Category id 24", "Declared Name Local 24", "Declared Name 25", "HS Code 25", "Product Name 25", "Declared Value 25", "Declared QTY 25", "Declared Category id 25", "Declared Name Local 25", "Declared Name 26", "HS Code 26", "Product Name 26", "Declared Value 26", "Declared QTY 26", "Declared Category id 26", "Declared Name Local 26", "Declared Name 27", "HS Code 27", "Product Name 27", "Declared Value 27", "Declared QTY 27", "Declared Category id 27", "Declared Name Local 27", "Declared Name 28", "HS Code 28", "Product Name 28", "Declared Value 28", "Declared QTY 28", "Declared Category id 28", "Declared Name Local 28", "Declared Name 29", "HS Code 29", "Product Name 29", "Declared Value 29", "Declared QTY 29", "Declared Category id 29", "Declared Name Local 29", "Declared Name 30", "HS Code 30", "Product Name 30", "Declared Value 30", "Declared QTY 30", "Declared Category id 30", "Declared Name Local 30", "Declared Name 31", "HS Code 31", "Product Name 31", "Declared Value 31", "Declared QTY 31", "Declared Category id 31", "Declared Name Local 31", "Declared Name 32", "HS Code 32", "Product Name 32", "Declared Value 32", "Declared QTY 32", "Declared Category id 32", "Declared Name Local 32", "Declared Name 33", "HS Code 33", "Product Name 33", "Declared Value 33", "Declared QTY 33", "Declared Category id 33", "Declared Name Local 33", "Declared Name 34", "HS Code 34", "Product Name 34", "Declared Value 34", "Declared QTY 34", "Declared Category id 34", "Declared Name Local 34", "Declared Name 35", "HS Code 35", "Product Name 35", "Declared Value 35", "Declared QTY 35", "Declared Category id 35", "Declared Name Local 35", "Declared Name 36", "HS Code 36", "Product Name 36", "Declared Value 36", "Declared QTY 36", "Declared Category id 36", "Declared Name Local 36", "Declared Name 37", "HS Code 37", "Product Name 37", "Declared Value 37", "Declared QTY 37", "Declared Category id 37", "Declared Name Local 37", "Declared Name 38", "HS Code 38", "Product Name 38", "Declared Value 38", "Declared QTY 38", "Declared Category id 38", "Declared Name Local 38", "Declared Name 39", "HS Code 39", "Product Name 39", "Declared Value 39", "Declared QTY 39", "Declared Category id 39", "Declared Name Local 39", "Declared Name 40", "HS Code 40", "Product Name 40", "Declared Value 40", "Declared QTY 40", "Declared Category id 40", "Declared Name Local 40",
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
