package draft_mawb

import (
	"net/http"
	"strconv"
	"time"

	"hpc-express-service/common"
	customerrors "hpc-express-service/errors"
)

// DraftMAWB represents the main draft MAWB entity
type DraftMAWB struct {
	UUID                           string            `json:"uuid" db:"uuid"`
	MAWBInfoUUID                   string            `json:"mawb_info_uuid" db:"mawb_info_uuid"`
	CustomerUUID                   string            `json:"customerUUID" db:"customer_uuid"`
	AirlineLogo                    string            `json:"airlineLogo" db:"airline_logo"`
	AirlineName                    string            `json:"airlineName" db:"airline_name"`
	MAWB                           string            `json:"mawb" db:"mawb"`
	HAWB                           string            `json:"hawb" db:"hawb"`
	ShipperNameAndAddress          string            `json:"shipperNameAndAddress" db:"shipper_name_and_address"`
	ConsigneeNameAndAddress        string            `json:"consigneeNameAndAddress" db:"consignee_name_and_address"`
	IssuingCarrierAgentNameAndCity string            `json:"issuingCarrierAgentNameAndCity" db:"issuing_carrier_agent_name_and_city"`
	AccountingInformation          string            `json:"accountingInformation" db:"accounting_information"`
	AgentIATACode                  string            `json:"agentIATACode" db:"agent_iata_code"`
	AccountNo                      string            `json:"accountNo" db:"account_no"`
	AirportOfDeparture             string            `json:"airportOfDeparture" db:"airport_of_departure"`
	ReferenceNumber                string            `json:"referenceNumber" db:"reference_number"`
	To1                            string            `json:"to1" db:"to_1"`
	ByFirstCarrier                 string            `json:"byFirstCarrier" db:"by_first_carrier"`
	To2                            string            `json:"to2" db:"to_2"`
	By2                            string            `json:"by2" db:"by_2"`
	To3                            string            `json:"to3" db:"to_3"`
	By3                            string            `json:"by3" db:"by_3"`
	Currency                       string            `json:"currency" db:"currency"`
	ChgsCode                       string            `json:"chgsCode" db:"chgs_code"`
	WtValPPD                       string            `json:"wtValPPD" db:"wt_val_ppd"`
	WtValColl                      string            `json:"wtValColl" db:"wt_val_coll"`
	OtherPPD                       string            `json:"otherPPD" db:"other_ppd"`
	OtherColl                      string            `json:"otherColl" db:"other_coll"`
	DeclaredValueCarriage          string            `json:"declaredValueCarriage" db:"declared_value_carriage"`
	DeclaredValueCustoms           string            `json:"declaredValueCustoms" db:"declared_value_customs"`
	AirportOfDestination           string            `json:"airportOfDestination" db:"airport_of_destination"`
	FlightNo                       string            `json:"flightNo" db:"flight_no"`
	FlightDate                     *time.Time        `json:"flightDate" db:"flight_date"`
	InsuranceAmount                float64           `json:"insuranceAmount" db:"insurance_amount"`
	HandlingInformation            string            `json:"handlingInformation" db:"handling_information"`
	SCI                            string            `json:"sci" db:"sci"`
	TotalNoOfPieces                int               `json:"totalNoOfPieces" db:"total_no_of_pieces"`
	TotalGrossWeight               float64           `json:"totalGrossWeight" db:"total_gross_weight"`
	TotalKgLb                      string            `json:"totalKgLb" db:"total_kg_lb"`
	TotalRateClass                 string            `json:"totalRateClass" db:"total_rate_class"`
	TotalChargeableWeight          float64           `json:"totalChargeableWeight" db:"total_chargeable_weight"`
	TotalRateCharge                float64           `json:"totalRateCharge" db:"total_rate_charge"`
	TotalAmount                    float64           `json:"totalAmount" db:"total_amount"`
	ShipperCertifiesText           string            `json:"shipperCertifiesText" db:"shipper_certifies_text"`
	ExecutedOnDate                 *time.Time        `json:"executedOnDate" db:"executed_on_date"`
	ExecutedAtPlace                string            `json:"executedAtPlace" db:"executed_at_place"`
	SignatureOfShipper             string            `json:"signatureOfShipper" db:"signature_of_shipper"`
	SignatureOfIssuingCarrier      string            `json:"signatureOfIssuingCarrier" db:"signature_of_issuing_carrier"`
	Status                         string            `json:"status" db:"status"`
	Items                          []DraftMAWBItem   `json:"items"`
	Charges                        []DraftMAWBCharge `json:"charges"`
	CreatedAt                      time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt                      time.Time         `json:"updatedAt" db:"updated_at"`
}

// DraftMAWBItem represents individual items in a draft MAWB
type DraftMAWBItem struct {
	ID                int                `json:"id" db:"id"`
	DraftMAWBUUID     string             `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
	PiecesRCP         string             `json:"piecesRCP" db:"pieces_rcp"`
	GrossWeight       string             `json:"grossWeight" db:"gross_weight"`
	KgLb              string             `json:"kgLb" db:"kg_lb"`
	RateClass         string             `json:"rateClass" db:"rate_class"`
	TotalVolume       float64            `json:"totalVolume" db:"total_volume"`
	ChargeableWeight  float64            `json:"chargeableWeight" db:"chargeable_weight"`
	RateCharge        float64            `json:"rateCharge" db:"rate_charge"`
	Total             float64            `json:"total" db:"total"`
	NatureAndQuantity string             `json:"natureAndQuantity" db:"nature_and_quantity"`
	Dims              []DraftMAWBItemDim `json:"dims"`
	CreatedAt         time.Time          `json:"createdAt" db:"created_at"`
}

// DraftMAWBItemDim represents dimension data for draft MAWB items
type DraftMAWBItemDim struct {
	ID              int       `json:"id" db:"id"`
	DraftMAWBItemID int       `json:"draft_mawb_item_id" db:"draft_mawb_item_id"`
	Length          string    `json:"length" db:"length"`
	Width           string    `json:"width" db:"width"`
	Height          string    `json:"height" db:"height"`
	Count           string    `json:"count" db:"count"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// DraftMAWBCharge represents charges for draft MAWB
type DraftMAWBCharge struct {
	ID            int       `json:"id" db:"id"`
	DraftMAWBUUID string    `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
	Key           string    `json:"key" db:"charge_key"`
	Value         float64   `json:"value" db:"charge_value"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// DraftMAWBRequest represents the request payload for creating/updating draft MAWB
type DraftMAWBRequest struct {
	CustomerUUID                   string                   `json:"customerUUID"`
	AirlineLogo                    string                   `json:"airlineLogo"`
	AirlineName                    string                   `json:"airlineName"`
	MAWB                           string                   `json:"mawb" validate:"required"`
	HAWB                           string                   `json:"hawb"`
	ShipperNameAndAddress          string                   `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress        string                   `json:"consigneeNameAndAddress"`
	IssuingCarrierAgentNameAndCity string                   `json:"issuingCarrierAgentNameAndCity"`
	AccountingInformation          string                   `json:"accountingInformation"`
	AgentIATACode                  string                   `json:"agentIATACode"`
	AccountNo                      string                   `json:"accountNo"`
	AirportOfDeparture             string                   `json:"airportOfDeparture"`
	ReferenceNumber                string                   `json:"referenceNumber"`
	To1                            string                   `json:"to1"`
	ByFirstCarrier                 string                   `json:"byFirstCarrier"`
	To2                            string                   `json:"to2"`
	By2                            string                   `json:"by2"`
	To3                            string                   `json:"to3"`
	By3                            string                   `json:"by3"`
	Currency                       string                   `json:"currency"`
	ChgsCode                       string                   `json:"chgsCode"`
	WtValPPD                       string                   `json:"wtValPPD"`
	WtValColl                      string                   `json:"wtValColl"`
	OtherPPD                       string                   `json:"otherPPD"`
	OtherColl                      string                   `json:"otherColl"`
	DeclaredValueCarriage          string                   `json:"declaredValueCarriage"`
	DeclaredValueCustoms           string                   `json:"declaredValueCustoms"`
	AirportOfDestination           string                   `json:"airportOfDestination"`
	FlightNo                       string                   `json:"flightNo"`
	FlightDate                     string                   `json:"flightDate"`
	InsuranceAmount                float64                  `json:"insuranceAmount"`
	HandlingInformation            string                   `json:"handlingInformation"`
	SCI                            string                   `json:"sci"`
	ShipperCertifiesText           string                   `json:"shipperCertifiesText"`
	ExecutedOnDate                 string                   `json:"executedOnDate"`
	ExecutedAtPlace                string                   `json:"executedAtPlace"`
	SignatureOfShipper             string                   `json:"signatureOfShipper"`
	SignatureOfIssuingCarrier      string                   `json:"signatureOfIssuingCarrier"`
	Items                          []DraftMAWBItemRequest   `json:"items"`
	Charges                        []DraftMAWBChargeRequest `json:"charges"`
}

// Bind implements the chi render.Binder interface for HTTP request binding
func (r *DraftMAWBRequest) Bind(req *http.Request) error {
	return nil
}

// DraftMAWBItemRequest represents the request payload for draft MAWB items
type DraftMAWBItemRequest struct {
	PiecesRCP         string                    `json:"piecesRCP"`
	GrossWeight       string                    `json:"grossWeight"`
	KgLb              string                    `json:"kgLb"`
	RateClass         string                    `json:"rateClass"`
	RateCharge        float64                   `json:"rateCharge"`
	NatureAndQuantity string                    `json:"natureAndQuantity"`
	Dims              []DraftMAWBItemDimRequest `json:"dims"`
}

// DraftMAWBItemDimRequest represents the request payload for draft MAWB item dimensions
type DraftMAWBItemDimRequest struct {
	Length string `json:"length"`
	Width  string `json:"width"`
	Height string `json:"height"`
	Count  string `json:"count"`
}

// DraftMAWBChargeRequest represents the request payload for draft MAWB charges
type DraftMAWBChargeRequest struct {
	Key   string  `json:"key" validate:"required"`
	Value float64 `json:"value"`
}

// DraftMAWBResponse represents the response payload for draft MAWB operations
type DraftMAWBResponse struct {
	UUID                           string            `json:"uuid"`
	MAWBInfoUUID                   string            `json:"mawb_info_uuid"`
	CustomerUUID                   string            `json:"customerUUID"`
	AirlineLogo                    string            `json:"airlineLogo"`
	AirlineName                    string            `json:"airlineName"`
	MAWB                           string            `json:"mawb"`
	HAWB                           string            `json:"hawb"`
	ShipperNameAndAddress          string            `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress        string            `json:"consigneeNameAndAddress"`
	IssuingCarrierAgentNameAndCity string            `json:"issuingCarrierAgentNameAndCity"`
	AccountingInformation          string            `json:"accountingInformation"`
	AgentIATACode                  string            `json:"agentIATACode"`
	AccountNo                      string            `json:"accountNo"`
	AirportOfDeparture             string            `json:"airportOfDeparture"`
	ReferenceNumber                string            `json:"referenceNumber"`
	To1                            string            `json:"to1"`
	ByFirstCarrier                 string            `json:"byFirstCarrier"`
	To2                            string            `json:"to2"`
	By2                            string            `json:"by2"`
	To3                            string            `json:"to3"`
	By3                            string            `json:"by3"`
	Currency                       string            `json:"currency"`
	ChgsCode                       string            `json:"chgsCode"`
	WtValPPD                       string            `json:"wtValPPD"`
	WtValColl                      string            `json:"wtValColl"`
	OtherPPD                       string            `json:"otherPPD"`
	OtherColl                      string            `json:"otherColl"`
	DeclaredValueCarriage          string            `json:"declaredValueCarriage"`
	DeclaredValueCustoms           string            `json:"declaredValueCustoms"`
	AirportOfDestination           string            `json:"airportOfDestination"`
	FlightNo                       string            `json:"flightNo"`
	FlightDate                     string            `json:"flightDate"`
	InsuranceAmount                float64           `json:"insuranceAmount"`
	HandlingInformation            string            `json:"handlingInformation"`
	SCI                            string            `json:"sci"`
	TotalNoOfPieces                int               `json:"totalNoOfPieces"`
	TotalGrossWeight               float64           `json:"totalGrossWeight"`
	TotalKgLb                      string            `json:"totalKgLb"`
	TotalRateClass                 string            `json:"totalRateClass"`
	TotalChargeableWeight          float64           `json:"totalChargeableWeight"`
	TotalRateCharge                float64           `json:"totalRateCharge"`
	TotalAmount                    float64           `json:"totalAmount"`
	ShipperCertifiesText           string            `json:"shipperCertifiesText"`
	ExecutedOnDate                 string            `json:"executedOnDate"`
	ExecutedAtPlace                string            `json:"executedAtPlace"`
	SignatureOfShipper             string            `json:"signatureOfShipper"`
	SignatureOfIssuingCarrier      string            `json:"signatureOfIssuingCarrier"`
	Status                         string            `json:"status"`
	Items                          []DraftMAWBItem   `json:"items"`
	Charges                        []DraftMAWBCharge `json:"charges"`
	CreatedAt                      string            `json:"createdAt"`
	UpdatedAt                      string            `json:"updatedAt"`
}

// Render implements the chi render.Renderer interface for HTTP response rendering
func (r *DraftMAWBResponse) Render(w http.ResponseWriter, req *http.Request) error {
	return nil
}

// Status constants for draft MAWB
const (
	StatusDraft     = "Draft"
	StatusPending   = "Pending"
	StatusConfirmed = "Confirmed"
	StatusRejected  = "Rejected"
)

// ValidateStatus checks if the provided status is valid
func ValidateStatus(status string) bool {
	switch status {
	case StatusDraft, StatusPending, StatusConfirmed, StatusRejected:
		return true
	default:
		return false
	}
}

// Enhanced validation using comprehensive validation utilities

// ValidateDraftMAWBRequest validates the draft MAWB request with comprehensive rules
func ValidateDraftMAWBRequest(req *DraftMAWBRequest) []customerrors.ValidationError {
	var validationErrors []customerrors.ValidationError

	// Create sanitizer and business rule validator
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())
	businessValidator := common.NewBusinessRuleValidator(sanitizer)

	// Required field validation
	if req.MAWB == "" {
		validationErrors = append(validationErrors, customerrors.NewValidationError("mawb", "MAWB number is required"))
	} else {
		// Validate MAWB number format
		if err := businessValidator.ValidateMAWBNumber(req.MAWB); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate HAWB number if provided
	if req.HAWB != "" {
		if err := businessValidator.ValidateHAWBNumber(req.HAWB); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate string lengths for key fields
	stringFields := map[string]struct {
		value  string
		maxLen int
	}{
		"airlineName":                    {req.AirlineName, 100},
		"shipperNameAndAddress":          {req.ShipperNameAndAddress, 500},
		"consigneeNameAndAddress":        {req.ConsigneeNameAndAddress, 500},
		"issuingCarrierAgentNameAndCity": {req.IssuingCarrierAgentNameAndCity, 200},
		"accountingInformation":          {req.AccountingInformation, 200},
		"agentIATACode":                  {req.AgentIATACode, 10},
		"accountNo":                      {req.AccountNo, 50},
		"referenceNumber":                {req.ReferenceNumber, 50},
		"handlingInformation":            {req.HandlingInformation, 500},
		"shipperCertifiesText":           {req.ShipperCertifiesText, 1000},
		"executedAtPlace":                {req.ExecutedAtPlace, 100},
	}

	for fieldName, fieldData := range stringFields {
		if err := sanitizer.ValidateStringLength(fieldData.value, fieldName, 0, fieldData.maxLen); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate airport codes
	if req.AirportOfDeparture != "" {
		if err := businessValidator.ValidateAirportCode(req.AirportOfDeparture, "airportOfDeparture"); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	if req.AirportOfDestination != "" {
		if err := businessValidator.ValidateAirportCode(req.AirportOfDestination, "airportOfDestination"); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate flight number
	if req.FlightNo != "" {
		if err := businessValidator.ValidateFlightNumber(req.FlightNo); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate currency
	if req.Currency != "" {
		if err := businessValidator.ValidateCurrency(req.Currency); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate numeric fields
	if req.InsuranceAmount < 0 {
		validationErrors = append(validationErrors, customerrors.NewValidationError("insuranceAmount", "Insurance amount cannot be negative"))
	}

	if req.InsuranceAmount > 10000000 { // 10 million maximum
		validationErrors = append(validationErrors, customerrors.NewValidationError("insuranceAmount", "Insurance amount cannot exceed 10,000,000"))
	}

	// Validate items count
	if err := sanitizer.ValidateArrayLength(len(req.Items), "items", 0, 50); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate charges count
	if err := sanitizer.ValidateArrayLength(len(req.Charges), "charges", 0, 20); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate items if provided
	for i, item := range req.Items {
		if itemErrors := ValidateDraftMAWBItemRequest(&item, i); len(itemErrors) > 0 {
			validationErrors = append(validationErrors, itemErrors...)
		}
	}

	// Validate charges if provided
	for i, charge := range req.Charges {
		if chargeErrors := ValidateDraftMAWBChargeRequest(&charge, i); len(chargeErrors) > 0 {
			validationErrors = append(validationErrors, chargeErrors...)
		}
	}

	return validationErrors
}

// ValidateDraftMAWBItemRequest validates individual draft MAWB item request
func ValidateDraftMAWBItemRequest(req *DraftMAWBItemRequest, index int) []customerrors.ValidationError {
	var validationErrors []customerrors.ValidationError
	indexStr := strconv.Itoa(index)

	// Create sanitizer and business rule validator
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())
	businessValidator := common.NewBusinessRuleValidator(sanitizer)

	// Gross weight validation
	if req.GrossWeight == "" {
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"items["+indexStr+"].grossWeight",
			"Gross weight is required for item",
		))
	} else {
		if err := businessValidator.ValidateWeight(req.GrossWeight, req.KgLb, "items["+indexStr+"].grossWeight"); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Rate charge validation
	if req.RateCharge < 0 {
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"items["+indexStr+"].rateCharge",
			"Rate charge cannot be negative",
		))
	}

	if req.RateCharge > 1000000 { // 1 million maximum
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"items["+indexStr+"].rateCharge",
			"Rate charge cannot exceed 1,000,000",
		))
	}

	// Validate string lengths
	if err := sanitizer.ValidateStringLength(req.PiecesRCP, "items["+indexStr+"].piecesRCP", 0, 50); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.RateClass, "items["+indexStr+"].rateClass", 0, 10); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.NatureAndQuantity, "items["+indexStr+"].natureAndQuantity", 0, 500); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate weight unit if provided
	if req.KgLb != "" {
		allowedUnits := []string{"kg", "lb"}
		if err := sanitizer.ValidateEnumValue(req.KgLb, "items["+indexStr+"].kgLb", allowedUnits); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate dimensions count
	if err := sanitizer.ValidateArrayLength(len(req.Dims), "items["+indexStr+"].dims", 0, 10); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate dimensions if provided
	for j, dim := range req.Dims {
		if dimErrors := ValidateDraftMAWBItemDimRequest(&dim, index, j); len(dimErrors) > 0 {
			validationErrors = append(validationErrors, dimErrors...)
		}
	}

	return validationErrors
}

// ValidateDraftMAWBItemDimRequest validates individual draft MAWB item dimension request
func ValidateDraftMAWBItemDimRequest(req *DraftMAWBItemDimRequest, itemIndex, dimIndex int) []customerrors.ValidationError {
	itemIndexStr := strconv.Itoa(itemIndex)
	dimIndexStr := strconv.Itoa(dimIndex)
	fieldPrefix := "items[" + itemIndexStr + "].dims[" + dimIndexStr + "]"

	// Use business rule validator for comprehensive dimension validation
	businessValidator := common.NewBusinessRuleValidator(common.NewInputSanitizer(common.DefaultValidationConfig()))

	return businessValidator.ValidateDimensions(req.Length, req.Width, req.Height, req.Count, fieldPrefix)
}

// ValidateDraftMAWBChargeRequest validates individual draft MAWB charge request
func ValidateDraftMAWBChargeRequest(req *DraftMAWBChargeRequest, index int) []customerrors.ValidationError {
	var validationErrors []customerrors.ValidationError
	indexStr := strconv.Itoa(index)

	// Create sanitizer
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	// Required field validation
	if req.Key == "" {
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"charges["+indexStr+"].key",
			"Charge key is required",
		))
	} else {
		// Validate key length and format
		if err := sanitizer.ValidateStringLength(req.Key, "charges["+indexStr+"].key", 1, 50); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}

		// Validate allowed charge keys
		allowedChargeKeys := []string{
			"weight_charge", "valuation_charge", "tax", "fuel_surcharge",
			"security_charge", "handling_charge", "documentation_fee",
			"insurance", "customs_fee", "other_charges",
		}
		if err := sanitizer.ValidateEnumValue(req.Key, "charges["+indexStr+"].key", allowedChargeKeys); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Value validation
	if req.Value < 0 {
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"charges["+indexStr+"].value",
			"Charge value cannot be negative",
		))
	}

	if req.Value > 1000000 { // 1 million maximum
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"charges["+indexStr+"].value",
			"Charge value cannot exceed 1,000,000",
		))
	}

	return validationErrors
}

// SanitizeDraftMAWBRequest sanitizes all string fields in the request using comprehensive sanitization
func SanitizeDraftMAWBRequest(req *DraftMAWBRequest) {
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	req.CustomerUUID = sanitizer.SanitizeString(req.CustomerUUID)
	req.AirlineLogo = sanitizer.SanitizeString(req.AirlineLogo)
	req.AirlineName = sanitizer.SanitizeString(req.AirlineName)
	req.MAWB = sanitizer.SanitizeString(req.MAWB)
	req.HAWB = sanitizer.SanitizeString(req.HAWB)
	req.ShipperNameAndAddress = sanitizer.SanitizeString(req.ShipperNameAndAddress)
	req.ConsigneeNameAndAddress = sanitizer.SanitizeString(req.ConsigneeNameAndAddress)
	req.IssuingCarrierAgentNameAndCity = sanitizer.SanitizeString(req.IssuingCarrierAgentNameAndCity)
	req.AccountingInformation = sanitizer.SanitizeString(req.AccountingInformation)
	req.AgentIATACode = sanitizer.SanitizeString(req.AgentIATACode)
	req.AccountNo = sanitizer.SanitizeString(req.AccountNo)
	req.AirportOfDeparture = sanitizer.SanitizeString(req.AirportOfDeparture)
	req.ReferenceNumber = sanitizer.SanitizeString(req.ReferenceNumber)
	req.To1 = sanitizer.SanitizeString(req.To1)
	req.ByFirstCarrier = sanitizer.SanitizeString(req.ByFirstCarrier)
	req.To2 = sanitizer.SanitizeString(req.To2)
	req.By2 = sanitizer.SanitizeString(req.By2)
	req.To3 = sanitizer.SanitizeString(req.To3)
	req.By3 = sanitizer.SanitizeString(req.By3)
	req.Currency = sanitizer.SanitizeString(req.Currency)
	req.ChgsCode = sanitizer.SanitizeString(req.ChgsCode)
	req.WtValPPD = sanitizer.SanitizeString(req.WtValPPD)
	req.WtValColl = sanitizer.SanitizeString(req.WtValColl)
	req.OtherPPD = sanitizer.SanitizeString(req.OtherPPD)
	req.OtherColl = sanitizer.SanitizeString(req.OtherColl)
	req.DeclaredValueCarriage = sanitizer.SanitizeString(req.DeclaredValueCarriage)
	req.DeclaredValueCustoms = sanitizer.SanitizeString(req.DeclaredValueCustoms)
	req.AirportOfDestination = sanitizer.SanitizeString(req.AirportOfDestination)
	req.FlightNo = sanitizer.SanitizeString(req.FlightNo)
	req.FlightDate = sanitizer.SanitizeString(req.FlightDate)
	req.HandlingInformation = sanitizer.SanitizeString(req.HandlingInformation)
	req.SCI = sanitizer.SanitizeString(req.SCI)
	req.ShipperCertifiesText = sanitizer.SanitizeString(req.ShipperCertifiesText)
	req.ExecutedOnDate = sanitizer.SanitizeString(req.ExecutedOnDate)
	req.ExecutedAtPlace = sanitizer.SanitizeString(req.ExecutedAtPlace)
	req.SignatureOfShipper = sanitizer.SanitizeString(req.SignatureOfShipper)
	req.SignatureOfIssuingCarrier = sanitizer.SanitizeString(req.SignatureOfIssuingCarrier)

	// Sanitize items
	for i := range req.Items {
		SanitizeDraftMAWBItemRequest(&req.Items[i])
	}

	// Sanitize charges
	for i := range req.Charges {
		SanitizeDraftMAWBChargeRequest(&req.Charges[i])
	}
}

// SanitizeDraftMAWBItemRequest sanitizes all string fields in the item request
func SanitizeDraftMAWBItemRequest(req *DraftMAWBItemRequest) {
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	req.PiecesRCP = sanitizer.SanitizeString(req.PiecesRCP)
	req.GrossWeight = sanitizer.SanitizeString(req.GrossWeight)
	req.KgLb = sanitizer.SanitizeString(req.KgLb)
	req.RateClass = sanitizer.SanitizeString(req.RateClass)
	req.NatureAndQuantity = sanitizer.SanitizeString(req.NatureAndQuantity)

	// Sanitize dimensions
	for i := range req.Dims {
		SanitizeDraftMAWBItemDimRequest(&req.Dims[i])
	}
}

// SanitizeDraftMAWBItemDimRequest sanitizes all string fields in the item dimension request
func SanitizeDraftMAWBItemDimRequest(req *DraftMAWBItemDimRequest) {
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	req.Length = sanitizer.SanitizeString(req.Length)
	req.Width = sanitizer.SanitizeString(req.Width)
	req.Height = sanitizer.SanitizeString(req.Height)
	req.Count = sanitizer.SanitizeString(req.Count)
}

// SanitizeDraftMAWBChargeRequest sanitizes all string fields in the charge request
func SanitizeDraftMAWBChargeRequest(req *DraftMAWBChargeRequest) {
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	req.Key = sanitizer.SanitizeString(req.Key)
}
