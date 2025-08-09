package mawbinfo

import (
	"net/http"
	"time"
)

type DraftMAWB struct {
	UUID                        string            `json:"uuid" db:"uuid"`
	MAWBInfoUUID                string            `json:"mawb_info_uuid" db:"mawb_info_uuid"`
	CustomerUUID                string            `json:"customerUUID" db:"customer_uuid"`
	AirlineLogo                 string            `json:"airlineLogo" db:"airline_logo"`
	AirlineName                 string            `json:"airlineName" db:"airline_name"`
	MAWB                        string            `json:"mawb" db:"mawb"`
	HAWB                        string            `json:"hawb" db:"hawb"`
	ShipperNameAndAddress       string            `json:"shipperNameAndAddress" db:"shipper_name_and_address"`
	AWBIssuedBy                 string            `json:"awbIssuedBy" db:"awb_issued_by"`
	ConsigneeNameAndAddress     string            `json:"consigneeNameAndAddress" db:"consignee_name_and_address"`
	IssuingCarrierAgentName     string            `json:"issuingCarrierAgentName" db:"issuing_carrier_agent_name"`
	AccountingInfomation        string            `json:"accountingInfomation" db:"accounting_infomation"`
	AgentsIATACode              string            `json:"agentsIATACode" db:"agents_iata_code"`
	AccountNo                   string            `json:"accountNo" db:"account_no"`
	AirportOfDeparture          string            `json:"airportOfDeparture" db:"airport_of_departure"`
	ReferenceNumber             string            `json:"referenceNumber" db:"reference_number"`
	OptionalShippingInfo1       string            `json:"optionalShippingInfo1" db:"optional_shipping_info1"`
	OptionalShippingInfo2       string            `json:"optionalShippingInfo2" db:"optional_shipping_info2"`
	RoutingTo                   string            `json:"routingTo" db:"routing_to"`
	RoutingBy                   string            `json:"routingBy" db:"routing_by"`
	DestinationTo1              string            `json:"destinationTo1" db:"destination_to1"`
	DestinationBy1              string            `json:"destinationBy1" db:"destination_by1"`
	DestinationTo2              string            `json:"destinationTo2" db:"destination_to2"`
	DestinationBy2              string            `json:"destinationBy2" db:"destination_by2"`
	Currency                    string            `json:"currency" db:"currency"`
	ChgsCode                    string            `json:"chgsCode" db:"chgs_code"`
	WtValPpd                    string            `json:"wtValPpd" db:"wt_val_ppd"`
	WtValColl                   string            `json:"wtValColl" db:"wt_val_coll"`
	OtherPpd                    string            `json:"otherPpd" db:"other_ppd"`
	OtherColl                   string            `json:"otherColl" db:"other_coll"`
	DeclaredValCarriage         string            `json:"declaredValCarriage" db:"declared_val_carriage"`
	DeclaredValCustoms          string            `json:"declaredValCustoms" db:"declared_val_customs"`
	AirportOfDestination        string            `json:"airportOfDestination" db:"airport_of_destination"`
	RequestedFlightDate1        string            `json:"requestedFlightDate1" db:"requested_flight_date1"`
	RequestedFlightDate2        string            `json:"requestedFlightDate2" db:"requested_flight_date2"`
	AmountOfInsurance           string            `json:"amountOfInsurance" db:"amount_of_insurance"`
	HandlingInfomation          string            `json:"handlingInfomation" db:"handling_infomation"`
	SCI                         string            `json:"sci" db:"sci"`
	Prepaid                     float64           `json:"prepaid" db:"prepaid"`
	ValuationCharge             float64           `json:"valuationCharge" db:"valuation_charge"`
	Tax                         float64           `json:"tax" db:"tax"`
	TotalOtherChargesDueAgent   float64           `json:"totalOtherChargesDueAgent" db:"total_other_charges_due_agent"`
	TotalOtherChargesDueCarrier float64           `json:"totalOtherChargesDueCarrier" db:"total_other_charges_due_carrier"`
	TotalPrepaid                float64           `json:"totalPrepaid" db:"total_prepaid"`
	CurrencyConversionRates     string            `json:"currencyConversionRates" db:"currency_conversion_rates"`
	Signature1                  string            `json:"signature1" db:"signature1"`
	Signature2Date              *CustomDate       `json:"signature2Date" db:"signature2_date"`
	Signature2Place             string            `json:"signature2Place" db:"signature2_place"`
	Signature2Issuing           string            `json:"signature2Issuing" db:"signature2_issuing"`
	ShippingMark                string            `json:"shippingMark" db:"shipping_mark"`
	Status                      string            `json:"status" db:"status"`
	Items                       []DraftMAWBItem   `json:"items" validate:"required,dive"`
	Charges                     []DraftMAWBCharge `json:"charges" validate:"dive"`
	CreatedAt                   time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt                   time.Time         `json:"updatedAt" db:"updated_at"`
}

type DraftMAWBItem struct {
	ID                int                `json:"id" db:"id"`
	DraftMAWBUUID     string             `json:"draft_mawb_uuid,omitempty" db:"draft_mawb_uuid"`
	PiecesRCP         string             `json:"piecesRCP" db:"pieces_rcp"`
	GrossWeight       string             `json:"grossWeight" db:"gross_weight"`
	KgLb              string             `json:"kgLb" db:"kg_lb"`
	RateClass         string             `json:"rateClass" db:"rate_class"`
	TotalVolume       string             `json:"totalVolume" db:"total_volume"`
	ChargeableWeight  string             `json:"chargeableWeight" db:"chargeable_weight"`
	RateCharge        float64            `json:"rateCharge" db:"rate_charge"`
	Total             float64            `json:"total" db:"total"`
	NatureAndQuantity string             `json:"natureAndQuantity" db:"nature_and_quantity"`
	Dims              []DraftMAWBItemDim `json:"dims" validate:"required,dive"`
}

type DraftMAWBItemDim struct {
	ID              int    `json:"id,omitempty" db:"id"`
	DraftMAWBItemID int    `json:"draft_mawb_item_id,omitempty" db:"draft_mawb_item_id"`
	Length          string `json:"length" db:"length"`
	Width           string `json:"width" db:"width"`
	Height          string `json:"height" db:"height"`
	Count           string `json:"count" db:"count"`
}

type DraftMAWBCharge struct {
	ID            int     `json:"id,omitempty" db:"id"`
	DraftMAWBUUID string  `json:"draft_mawb_uuid,omitempty" db:"draft_mawb_uuid"`
	Key           string  `json:"key" db:"charge_key"`
	Value         float64 `json:"value" db:"charge_value"`
}

// Bind for DraftMAWB
func (dm *DraftMAWB) Bind(r *http.Request) error {
	return nil
}

// Render for DraftMAWB
func (dm *DraftMAWB) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (DraftMAWB) TableName() string {
	return "draft_mawb"
}

func (DraftMAWBItem) TableName() string {
	return "draft_mawb_items"
}

func (DraftMAWBItemDim) TableName() string {
	return "draft_mawb_item_dims"
}

func (DraftMAWBCharge) TableName() string {
	return "draft_mawb_charges"
}
