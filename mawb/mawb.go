package mawb

import "net/http"

type RequestDraftModel struct {
	CustomerUUID                string             `json:"customerUuid" validate:"required"`
	Mawb                        string             `json:"mawb"`
	Hawb                        string             `json:"hawb"`
	ShipperNameAndAddress       string             `json:"shipperNameAndAddress"`
	AwbIssuedBy                 string             `json:"awbIssuedBy"`
	ConsigneeNameAndAddress     string             `json:"consigneeNameAndAddress"`
	IssuingCarrierAgentName     string             `json:"issuingCarrierAgentName"`
	AccountingInfomation        string             `json:"accountingInfomation"`
	AgentsIATACode              string             `json:"agentsIATACode"`
	AccountNo                   string             `json:"accountNo"`
	AirportOfDeparture          string             `json:"airportOfDeparture"`
	ReferenceNumber             string             `json:"referenceNumber"`
	OptionalShippingInfo1       string             `json:"optionalShippingInfo1"`
	OptionalShippingInfo2       string             `json:"optionalShippingInfo2"`
	RoutingTo                   string             `json:"routingTo"`
	RoutingBy                   string             `json:"routingBy"`
	DestinationTo1              string             `json:"destinationTo1"`
	DestinationBy1              string             `json:"destinationBy1"`
	DestinationTo2              string             `json:"destinationTo2"`
	DestinationBy2              string             `json:"destinationBy2"`
	Currency                    string             `json:"currency"`
	ChgsCode                    string             `json:"chgsCode"`
	WtValPpd                    string             `json:"wtValPpd"`
	WtValColl                   string             `json:"wtValColl"`
	OtherPpd                    string             `json:"otherPpd"`
	OtherColl                   string             `json:"otherColl"`
	DeclaredValCarriage         string             `json:"declaredValCarriage"`
	DeclaredValCustoms          string             `json:"declaredValCustoms"`
	AirportOfDestination        string             `json:"airportOfDestination"`
	RequestedFlightDate1        string             `json:"requestedFlightDate1"`
	RequestedFlightDate2        string             `json:"requestedFlightDate2"`
	AmountOfInsurance           string             `json:"amountOfInsurance"`
	HandlingInfomation          string             `json:"handlingInfomation"`
	Sci                         string             `json:"sci"`
	Items                       []*ItemDetailModel `json:"items"`
	TerminalChargeKey           string             `json:"terminalChargeKey"`
	TerminalChargeVal           string             `json:"terminalChargeVal"`
	MrKey                       string             `json:"mrKey"`
	MrVal                       string             `json:"mrVal"`
	BcKey                       string             `json:"bcKey"`
	BcVal                       string             `json:"bcVal"`
	CcKey                       string             `json:"ccKey"`
	CcVal                       string             `json:"ccVal"`
	AweFeeKey                   string             `json:"aweFeeKey"`
	AweFeeVal                   string             `json:"aweFeeVal"`
	Signature1                  string             `json:"signature1"`
	Prepaid                     string             `json:"prepaid"`
	ValuationCharge             string             `json:"valuationCharge"`
	Tax                         string             `json:"tax"`
	TotalOtherChargesDueAgent   string             `json:"totalOtherChargesDueAgent"`
	TotalOtherChargesDueCarrier string             `json:"totalOtherChargesDueCarrier"`
	TotalPrepaid                string             `json:"totalPrepaid"`
	CurrencyConversionRates     string             `json:"currencyConversionRates"`
	Signature2Date              string             `json:"signature2Date"`
	Signature2Place             string             `json:"signature2Place"`
	Signature2Issuing           string             `json:"signature2Issuing"`
}

func (o *RequestDraftModel) Bind(r *http.Request) error {
	return nil
}

func (o *RequestDraftModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ItemDetailModel struct {
	PiecesRCP         string `json:"piecesRCP"`
	GrossWeight       string `json:"grossWeight"`
	NatureAndQuantity string `json:"natureAndQuantity"`
	RateClass         string `json:"rateClass"`
	ChargeableWeight  string `json:"chargeableWeight"`
	RateCharge        string `json:"rateCharge"`
	Total             string `json:"total"`
}

type OtherServiceModel struct {
	Key   string
	Value string
}

type GetAllMawbDraftModel struct {
	UUID                    string `json:"uuid"`
	Mawb                    string `json:"mawb"`
	Hawb                    string `json:"hawb"`
	ShipperNameAndAddress   string `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress string `json:"consigneeNameAndAddress"`
	CustomerName            string `json:"customerName"`
	CreatedAt               string `json:"createdAt"`
}

type GetMawbDraftModel struct {
	UUID string `json:"uuid"`
	RequestDraftModel
}

type RequestUpdateMawbDraftModel struct {
	UUID string `json:"uuid"`
	RequestDraftModel
}

func (o *RequestUpdateMawbDraftModel) Bind(r *http.Request) error {
	return nil
}

func (o *RequestUpdateMawbDraftModel) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
