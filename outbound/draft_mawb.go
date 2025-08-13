package outbound

import (
	"net/http"
	"time"
)

type DraftMAWB struct {
	tableName                   struct{}  `pg:"public.draft_mawb"`
	UUID                        string    `json:"uuid" db:"uuid"`
	MAWBInfoUUID                string    `json:"mawb_info_uuid" db:"mawb_info_uuid"`
	CustomerUUID                string    `json:"customerUUID" db:"customer_uuid"`
	AirlineLogo                 string    `json:"airline_logo" db:"airline_logo"`
	AirlineName                 string    `json:"airline_name" db:"airline_name"`
	MAWB                        string    `json:"mawb" db:"mawb"`
	HAWB                        string    `json:"hawb" db:"hawb"`
	ShipperNameAndAddress       string    `json:"shipper_name_and_address" db:"shipper_name_and_address"`
	AWBIssuedBy                 string    `json:"awb_issued_by" db:"awb_issued_by"`
	ConsigneeNameAndAddress     string    `json:"consignee_name_and_address" db:"consignee_name_and_address"`
	IssuingCarrierAgentName     string    `json:"issuing_carrier_agent_name" db:"issuing_carrier_agent_name"`
	AccountingInfomation        string    `json:"accounting_infomation" db:"accounting_infomation"`
	AgentsIATACode              string    `json:"agents_iata_code" db:"agents_iata_code"`
	AccountNo                   string    `json:"account_no" db:"account_no"`
	AirportOfDeparture          string    `json:"airport_of_departure" db:"airport_of_departure"`
	ReferenceNumber             string    `json:"reference_number" db:"reference_number"`
	OptionalShippingInfo1       string    `json:"optional_shipping_info1" db:"optional_shipping_info1"`
	OptionalShippingInfo2       string    `json:"optional_shipping_info2" db:"optional_shipping_info2"`
	RoutingTo                   string    `json:"routing_to" db:"routing_to"`
	RoutingBy                   string    `json:"routing_by" db:"routing_by"`
	DestinationTo1              string    `json:"destination_to1" db:"destination_to1"`
	DestinationBy1              string    `json:"destination_by1" db:"destination_by1"`
	DestinationTo2              string    `json:"destination_to2" db:"destination_to2"`
	DestinationBy2              string    `json:"destination_by2" db:"destination_by2"`
	Currency                    string    `json:"currency" db:"currency"`
	ChgsCode                    string    `json:"chgs_code" db:"chgs_code"`
	WtValPpd                    string    `json:"wt_val_ppd" db:"wt_val_ppd"`
	WtValColl                   string    `json:"wt_val_coll" db:"wt_val_coll"`
	OtherPpd                    string    `json:"other_ppd" db:"other_ppd"`
	OtherColl                   string    `json:"other_coll" db:"other_coll"`
	DeclaredValCarriage         string    `json:"declared_val_carriage" db:"declared_val_carriage"`
	DeclaredValCustoms          string    `json:"declared_val_customs" db:"declared_val_customs"`
	AirportOfDestination        string    `json:"airport_of_destination" db:"airport_of_destination"`
	RequestedFlightDate1        string    `json:"requested_flight_date1" db:"requested_flight_date1"`
	RequestedFlightDate2        string    `json:"requested_flight_date2" db:"requested_flight_date2"`
	AmountOfInsurance           string    `json:"amount_of_insurance" db:"amount_of_insurance"`
	HandlingInfomation          string    `json:"handling_infomation" db:"handling_infomation"`
	SCI                         string    `json:"sci" db:"sci"`
	Prepaid                     float64   `json:"prepaid" db:"prepaid"`
	ValuationCharge             float64   `json:"valuation_charge" db:"valuation_charge"`
	Tax                         float64   `json:"tax" db:"tax"`
	TotalOtherChargesDueAgent   float64   `json:"total_other_charges_due_agent" db:"total_other_charges_due_agent"`
	TotalOtherChargesDueCarrier float64   `json:"total_other_charges_due_carrier" db:"total_other_charges_due_carrier"`
	TotalPrepaid                float64   `json:"total_prepaid" db:"total_prepaid"`
	CurrencyConversionRates     string    `json:"currency_conversion_rates" db:"currency_conversion_rates"`
	Signature1                  string    `json:"signature1" db:"signature1"`
	Signature2Date              string    `json:"signature2_date" db:"signature2_date"`
	Signature2Place             string    `json:"signature2_place" db:"signature2_place"`
	Signature2Issuing           string    `json:"signature2_issuing" db:"signature2_issuing"`
	ShippingMark                string    `json:"shipping_mark" db:"shipping_mark"`
	Status                      string    `json:"status" db:"status"`
	CreatedAt                   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at" db:"updated_at"`
}

func (d *DraftMAWB) Bind(r *http.Request) error {
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	return nil
}

// DraftMAWBWithRelations includes the related items and charges
type DraftMAWBWithRelations struct {
	*DraftMAWB
	Items   []DraftMAWBItem   `json:"items,omitempty"`
	Charges []DraftMAWBCharge `json:"charges,omitempty"`
}

type DraftMAWBItem struct {
	tableName         struct{}           `pg:"public.draft_mawb_items"`
	ID                int                `json:"id" pg:"id"`
	DraftMAWBUUID     string             `json:"draft_mawb_uuid" pg:"draft_mawb_uuid"`
	PiecesRCP         string             `json:"pieces_rcp" pg:"pieces_rcp"`
	GrossWeight       string             `json:"gross_weight" pg:"gross_weight"`
	KgLb              string             `json:"kg_lb" pg:"kg_lb"`
	RateClass         string             `json:"rate_class" pg:"rate_class"`
	TotalVolume       float64            `json:"total_volume" pg:"total_volume"`
	ChargeableWeight  float64            `json:"chargeable_weight" pg:"chargeable_weight"`
	RateCharge        float64            `json:"rate_charge" pg:"rate_charge"`
	Total             float64            `json:"total" pg:"total"`
	NatureAndQuantity string             `json:"nature_and_quantity" pg:"nature_and_quantity"`
	Dims              []DraftMAWBItemDim `json:"dims,omitempty"`
}

type DraftMAWBItemDim struct {
	tableName       struct{} `pg:"public.draft_mawb_item_dims"`
	ID              int      `json:"id" pg:"id"`
	DraftMAWBItemID int      `json:"draft_mawb_item_id" pg:"draft_mawb_item_id"`
	Length          string   `json:"length" pg:"length"`
	Width           string   `json:"width" pg:"width"`
	Height          string   `json:"height" pg:"height"`
	Count           string   `json:"count" pg:"count"`
}

type DraftMAWBCharge struct {
	tableName     struct{} `pg:"public.draft_mawb_charges"`
	ID            int      `json:"id" pg:"id"`
	DraftMAWBUUID string   `json:"draft_mawb_uuid" pg:"draft_mawb_uuid"`
	Key           string   `json:"key" pg:"charge_key"`
	Value         float64  `json:"value" pg:"charge_value"`
}

// DraftMAWBListItem represents a draft MAWB item in the list view
type DraftMAWBListItem struct {
	UUID                    string `json:"uuid"`
	MAWBInfoUUID            string `json:"mawb_info_uuid"`
	MAWB                    string `json:"mawb"`
	HAWB                    string `json:"hawb"`
	Airline                 string `json:"airline"`
	ShipperNameAndAddress   string `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress string `json:"consigneeNameAndAddress"`
	CustomerName            string `json:"customerName"`
	CreatedAt               string `json:"createdAt"`
	Status                  string `json:"status"`
	IsDeleted               bool   `json:"isDeleted"`
}

// DraftMAWBInput is used for API input that includes items and charges
type DraftMAWBInput struct {
	UUID                        string                 `json:"uuid,omitempty"`
	MAWBInfoUUID                string                 `json:"mawb_info_uuid,omitempty"`
	CustomerUUID                string                 `json:"customerUUID"`
	AirlineLogo                 string                 `json:"airlineLogo"`
	AirlineName                 string                 `json:"airlineName"`
	MAWB                        string                 `json:"mawb"`
	HAWB                        string                 `json:"hawb"`
	ShipperNameAndAddress       string                 `json:"shipperNameAndAddress"`
	AWBIssuedBy                 string                 `json:"awbIssuedBy"`
	ConsigneeNameAndAddress     string                 `json:"consigneeNameAndAddress"`
	IssuingCarrierAgentName     string                 `json:"issuingCarrierAgentName"`
	AccountingInfomation        string                 `json:"accountingInfomation"`
	AgentsIATACode              string                 `json:"agentsIATACode"`
	AccountNo                   string                 `json:"accountNo"`
	AirportOfDeparture          string                 `json:"airportOfDeparture"`
	ReferenceNumber             string                 `json:"referenceNumber"`
	OptionalShippingInfo1       string                 `json:"optionalShippingInfo1"`
	OptionalShippingInfo2       string                 `json:"optionalShippingInfo2"`
	RoutingTo                   string                 `json:"routingTo"`
	RoutingBy                   string                 `json:"routingBy"`
	DestinationTo1              string                 `json:"destinationTo1"`
	DestinationBy1              string                 `json:"destinationBy1"`
	DestinationTo2              string                 `json:"destinationTo2"`
	DestinationBy2              string                 `json:"destinationBy2"`
	Currency                    string                 `json:"currency"`
	ChgsCode                    string                 `json:"chgsCode"`
	WtValPpd                    string                 `json:"wtValPpd"`
	WtValColl                   string                 `json:"wtValColl"`
	OtherPpd                    string                 `json:"otherPpd"`
	OtherColl                   string                 `json:"otherColl"`
	DeclaredValCarriage         string                 `json:"declaredValCarriage"`
	DeclaredValCustoms          string                 `json:"declaredValCustoms"`
	AirportOfDestination        string                 `json:"airportOfDestination"`
	RequestedFlightDate1        string                 `json:"requestedFlightDate1"`
	RequestedFlightDate2        string                 `json:"requestedFlightDate2"`
	AmountOfInsurance           string                 `json:"amountOfInsurance"`
	HandlingInfomation          string                 `json:"handlingInfomation"`
	SCI                         string                 `json:"sci"`
	Prepaid                     float64                `json:"prepaid"`
	ValuationCharge             float64                `json:"valuationCharge"`
	Tax                         float64                `json:"tax"`
	TotalOtherChargesDueAgent   float64                `json:"totalOtherChargesDueAgent"`
	TotalOtherChargesDueCarrier float64                `json:"totalOtherChargesDueCarrier"`
	TotalPrepaid                float64                `json:"totalPrepaid"`
	CurrencyConversionRates     string                 `json:"currencyConversionRates"`
	Signature1                  string                 `json:"signature1"`
	Signature2Date              string                 `json:"signature2Date"`
	Signature2Place             string                 `json:"signature2Place"`
	Signature2Issuing           string                 `json:"signature2Issuing"`
	ShippingMark                string                 `json:"shippingMark"`
	AirlineUUID                 string                 `json:"airlineUUID"`
	Items                       []DraftMAWBItemInput   `json:"items,omitempty"`
	Charges                     []DraftMAWBChargeInput `json:"charges,omitempty"`
}

type DraftMAWBItemInput struct {
	ID                int                     `json:"id,omitempty"`
	PiecesRCP         string                  `json:"piecesRCP"`
	GrossWeight       float64                 `json:"grossWeight"`
	KgLb              string                  `json:"kgLb"`
	RateClass         string                  `json:"rateClass"`
	TotalVolume       float64                 `json:"totalVolume"`
	ChargeableWeight  float64                 `json:"chargeableWeight"`
	RateCharge        float64                 `json:"rateCharge"`
	Total             float64                 `json:"total"`
	NatureAndQuantity string                  `json:"natureAndQuantity"`
	Dims              []DraftMAWBItemDimInput `json:"dims,omitempty"`
}

type DraftMAWBItemDimInput struct {
	ID     int `json:"id,omitempty"`
	Length int `json:"length"`
	Width  int `json:"width"`
	Height int `json:"height"`
	Count  int `json:"count"`
}

type DraftMAWBChargeInput struct {
	ID    int     `json:"id,omitempty"`
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

func (d *DraftMAWBInput) Bind(r *http.Request) error {
	return nil
}

// ToDraftMAWB converts DraftMAWBInput to DraftMAWB
func (d *DraftMAWBInput) ToDraftMAWB() *DraftMAWB {
	now := time.Now()
	return &DraftMAWB{
		UUID:                        d.UUID,
		MAWBInfoUUID:                d.MAWBInfoUUID,
		CustomerUUID:                d.CustomerUUID,
		AirlineLogo:                 d.AirlineLogo,
		AirlineName:                 d.AirlineName,
		MAWB:                        d.MAWB,
		HAWB:                        d.HAWB,
		ShipperNameAndAddress:       d.ShipperNameAndAddress,
		AWBIssuedBy:                 d.AWBIssuedBy,
		ConsigneeNameAndAddress:     d.ConsigneeNameAndAddress,
		IssuingCarrierAgentName:     d.IssuingCarrierAgentName,
		AccountingInfomation:        d.AccountingInfomation,
		AgentsIATACode:              d.AgentsIATACode,
		AccountNo:                   d.AccountNo,
		AirportOfDeparture:          d.AirportOfDeparture,
		ReferenceNumber:             d.ReferenceNumber,
		OptionalShippingInfo1:       d.OptionalShippingInfo1,
		OptionalShippingInfo2:       d.OptionalShippingInfo2,
		RoutingTo:                   d.RoutingTo,
		RoutingBy:                   d.RoutingBy,
		DestinationTo1:              d.DestinationTo1,
		DestinationBy1:              d.DestinationBy1,
		DestinationTo2:              d.DestinationTo2,
		DestinationBy2:              d.DestinationBy2,
		Currency:                    d.Currency,
		ChgsCode:                    d.ChgsCode,
		WtValPpd:                    d.WtValPpd,
		WtValColl:                   d.WtValColl,
		OtherPpd:                    d.OtherPpd,
		OtherColl:                   d.OtherColl,
		DeclaredValCarriage:         d.DeclaredValCarriage,
		DeclaredValCustoms:          d.DeclaredValCustoms,
		AirportOfDestination:        d.AirportOfDestination,
		RequestedFlightDate1:        d.RequestedFlightDate1,
		RequestedFlightDate2:        d.RequestedFlightDate2,
		AmountOfInsurance:           d.AmountOfInsurance,
		HandlingInfomation:          d.HandlingInfomation,
		SCI:                         d.SCI,
		Prepaid:                     d.Prepaid,
		ValuationCharge:             d.ValuationCharge,
		Tax:                         d.Tax,
		TotalOtherChargesDueAgent:   d.TotalOtherChargesDueAgent,
		TotalOtherChargesDueCarrier: d.TotalOtherChargesDueCarrier,
		TotalPrepaid:                d.TotalPrepaid,
		CurrencyConversionRates:     d.CurrencyConversionRates,
		Signature1:                  d.Signature1,
		Signature2Date:              d.Signature2Date,
		Signature2Place:             d.Signature2Place,
		Signature2Issuing:           d.Signature2Issuing,
		ShippingMark:                d.ShippingMark,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}
}
