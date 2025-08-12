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
	ID                int                `json:"id" db:"id"`
	DraftMAWBUUID     string             `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
	PiecesRCP         string             `json:"pieces_rcp" db:"pieces_rcp"`
	GrossWeight       string             `json:"gross_weight" db:"gross_weight"`
	KgLb              string             `json:"kg_lb" db:"kg_lb"`
	RateClass         string             `json:"rate_class" db:"rate_class"`
	TotalVolume       string             `json:"total_volume" db:"total_volume"`
	ChargeableWeight  string             `json:"chargeable_weight" db:"chargeable_weight"`
	RateCharge        float64            `json:"rate_charge" db:"rate_charge"`
	Total             float64            `json:"total" db:"total"`
	NatureAndQuantity string             `json:"nature_and_quantity" db:"nature_and_quantity"`
	Dims              []DraftMAWBItemDim `json:"dims,omitempty"`
}

type DraftMAWBItemDim struct {
	tableName       struct{} `pg:"public.draft_mawb_dims"`
	ID              int      `json:"id" db:"id"`
	DraftMAWBItemID int      `json:"draft_mawb_item_id" db:"draft_mawb_item_id"`
	Length          string   `json:"length" db:"length"`
	Width           string   `json:"width" db:"width"`
	Height          string   `json:"height" db:"height"`
	Count           string   `json:"count" db:"count"`
}

type DraftMAWBCharge struct {
	tableName     struct{} `pg:"public.draft_mawb_charges"`
	ID            int      `json:"id" db:"id"`
	DraftMAWBUUID string   `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
	Key           string   `json:"key" db:"charge_key"`
	Value         float64  `json:"value" db:"charge_value"`
}

// DraftMAWBListItem represents a draft MAWB item in the list view
type DraftMAWBListItem struct {
	UUID                    string `json:"uuid"`
	MAWB                    string `json:"mawb"`
	HAWB                    string `json:"hawb"`
	Airline                 string `json:"airline"`
	ShipperNameAndAddress   string `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress string `json:"consigneeNameAndAddress"`
	CustomerName            string `json:"customerName"`
	CreatedAt               string `json:"createdAt"`
}

// DraftMAWBInput is used for API input that includes items and charges
type DraftMAWBInput struct {
	DraftMAWB
	Items   []DraftMAWBItemInput   `json:"items,omitempty"`
	Charges []DraftMAWBChargeInput `json:"charges,omitempty"`
}

type DraftMAWBItemInput struct {
	ID                int                     `json:"id,omitempty"`
	PiecesRCP         string                  `json:"piecesRCP"`
	GrossWeight       string                  `json:"grossWeight"`
	KgLb              string                  `json:"kgLb"`
	RateClass         string                  `json:"rateClass"`
	TotalVolume       string                  `json:"totalVolume"`
	ChargeableWeight  string                  `json:"chargeableWeight"`
	RateCharge        float64                 `json:"rateCharge"`
	Total             float64                 `json:"total"`
	NatureAndQuantity string                  `json:"natureAndQuantity"`
	Dims              []DraftMAWBItemDimInput `json:"dims,omitempty"`
}

type DraftMAWBItemDimInput struct {
	ID     int    `json:"id,omitempty"`
	Length string `json:"length"`
	Width  string `json:"width"`
	Height string `json:"height"`
	Count  string `json:"count"`
}

type DraftMAWBChargeInput struct {
	ID    int     `json:"id,omitempty"`
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

func (d *DraftMAWBInput) Bind(r *http.Request) error {
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	return nil
}
