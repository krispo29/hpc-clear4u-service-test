package cargo_manifest

import (
	"net/http"
	"strconv"
	"time"

	"hpc-express-service/common"
	customerrors "hpc-express-service/errors"
)

// CargoManifest represents the main cargo manifest entity
type CargoManifest struct {
	UUID            string              `json:"uuid" db:"uuid"`
	MAWBInfoUUID    string              `json:"mawb_info_uuid" db:"mawb_info_uuid"`
	MAWBNumber      string              `json:"mawbNumber" db:"mawb_number"`
	PortOfDischarge string              `json:"portOfDischarge" db:"port_of_discharge"`
	FlightNo        string              `json:"flightNo" db:"flight_no"`
	FreightDate     string              `json:"freightDate" db:"freight_date"`
	Shipper         string              `json:"shipper" db:"shipper"`
	Consignee       string              `json:"consignee" db:"consignee"`
	TotalCtn        string              `json:"totalCtn" db:"total_ctn"`
	Transshipment   string              `json:"transshipment" db:"transshipment"`
	Status          string              `json:"status" db:"status"`
	Items           []CargoManifestItem `json:"items"`
	CreatedAt       time.Time           `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time           `json:"updatedAt" db:"updated_at"`
}

// CargoManifestItem represents individual items in a cargo manifest
type CargoManifestItem struct {
	ID                      int       `json:"id" db:"id"`
	CargoManifestUUID       string    `json:"cargo_manifest_uuid" db:"cargo_manifest_uuid"`
	HAWBNo                  string    `json:"hawbNo" db:"hawb_no"`
	Pkgs                    string    `json:"pkgs" db:"pkgs"`
	GrossWeight             string    `json:"grossWeight" db:"gross_weight"`
	Destination             string    `json:"dst" db:"destination"`
	Commodity               string    `json:"commodity" db:"commodity"`
	ShipperNameAndAddress   string    `json:"shipperNameAndAddress" db:"shipper_name_address"`
	ConsigneeNameAndAddress string    `json:"consigneeNameAndAddress" db:"consignee_name_address"`
	CreatedAt               time.Time `json:"createdAt" db:"created_at"`
}

// CargoManifestRequest represents the request payload for creating/updating cargo manifest
type CargoManifestRequest struct {
	MAWBNumber      string                     `json:"mawbNumber" validate:"required"`
	PortOfDischarge string                     `json:"portOfDischarge"`
	FlightNo        string                     `json:"flightNo"`
	FreightDate     string                     `json:"freightDate"`
	Shipper         string                     `json:"shipper"`
	Consignee       string                     `json:"consignee"`
	TotalCtn        string                     `json:"totalCtn"`
	Transshipment   string                     `json:"transshipment"`
	Items           []CargoManifestItemRequest `json:"items"`
}

// Bind implements the chi render.Binder interface for HTTP request binding
func (r *CargoManifestRequest) Bind(req *http.Request) error {
	return nil
}

// CargoManifestItemRequest represents the request payload for cargo manifest items
type CargoManifestItemRequest struct {
	HAWBNo                  string `json:"hawbNo"`
	Pkgs                    string `json:"pkgs"`
	GrossWeight             string `json:"grossWeight"`
	Destination             string `json:"dst"`
	Commodity               string `json:"commodity"`
	ShipperNameAndAddress   string `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress string `json:"consigneeNameAndAddress"`
}

// CargoManifestResponse represents the response payload for cargo manifest operations
type CargoManifestResponse struct {
	UUID            string              `json:"uuid"`
	MAWBInfoUUID    string              `json:"mawb_info_uuid"`
	MAWBNumber      string              `json:"mawbNumber"`
	PortOfDischarge string              `json:"portOfDischarge"`
	FlightNo        string              `json:"flightNo"`
	FreightDate     string              `json:"freightDate"`
	Shipper         string              `json:"shipper"`
	Consignee       string              `json:"consignee"`
	TotalCtn        string              `json:"totalCtn"`
	Transshipment   string              `json:"transshipment"`
	Status          string              `json:"status"`
	Items           []CargoManifestItem `json:"items"`
	CreatedAt       string              `json:"createdAt"`
	UpdatedAt       string              `json:"updatedAt"`
}

// Render implements the chi render.Renderer interface for HTTP response rendering
func (r *CargoManifestResponse) Render(w http.ResponseWriter, req *http.Request) error {
	return nil
}

// Status constants for cargo manifest
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

// Enhanced validation and sanitization using common utilities

// ValidateCargoManifestRequest validates the cargo manifest request with comprehensive rules
func ValidateCargoManifestRequest(req *CargoManifestRequest) []customerrors.ValidationError {
	var validationErrors []customerrors.ValidationError

	// Create sanitizer and business rule validator
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())
	businessValidator := common.NewBusinessRuleValidator(sanitizer)

	// Required field validation
	if req.MAWBNumber == "" {
		validationErrors = append(validationErrors, customerrors.NewValidationError("mawbNumber", "MAWB number is required"))
	} else {
		// Validate MAWB number format
		if err := businessValidator.ValidateMAWBNumber(req.MAWBNumber); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate string lengths
	if err := sanitizer.ValidateStringLength(req.PortOfDischarge, "portOfDischarge", 0, 100); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.Shipper, "shipper", 0, 500); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.Consignee, "consignee", 0, 500); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate flight number format
	if req.FlightNo != "" {
		if err := businessValidator.ValidateFlightNumber(req.FlightNo); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate airport codes
	if req.PortOfDischarge != "" {
		if err := businessValidator.ValidateAirportCode(req.PortOfDischarge, "portOfDischarge"); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate items count
	if err := sanitizer.ValidateArrayLength(len(req.Items), "items", 0, 100); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate items if provided
	for i, item := range req.Items {
		if itemErrors := ValidateCargoManifestItemRequest(&item, i); len(itemErrors) > 0 {
			validationErrors = append(validationErrors, itemErrors...)
		}
	}

	return validationErrors
}

// ValidateCargoManifestItemRequest validates individual cargo manifest item request
func ValidateCargoManifestItemRequest(req *CargoManifestItemRequest, index int) []customerrors.ValidationError {
	var validationErrors []customerrors.ValidationError
	indexStr := strconv.Itoa(index)

	// Create sanitizer and business rule validator
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())
	businessValidator := common.NewBusinessRuleValidator(sanitizer)

	// HAWB number validation
	if req.HAWBNo != "" {
		if err := businessValidator.ValidateHAWBNumber(req.HAWBNo); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				valErr.Field = "items[" + indexStr + "].hawbNo"
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Gross weight validation
	if req.GrossWeight == "" {
		validationErrors = append(validationErrors, customerrors.NewValidationError(
			"items["+indexStr+"].grossWeight",
			"Gross weight is required for item",
		))
	} else {
		if err := businessValidator.ValidateWeight(req.GrossWeight, "", "items["+indexStr+"].grossWeight"); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	// Validate string lengths
	if err := sanitizer.ValidateStringLength(req.Pkgs, "items["+indexStr+"].pkgs", 0, 50); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.Destination, "items["+indexStr+"].destination", 0, 100); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.Commodity, "items["+indexStr+"].commodity", 0, 200); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.ShipperNameAndAddress, "items["+indexStr+"].shipperNameAndAddress", 0, 500); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	if err := sanitizer.ValidateStringLength(req.ConsigneeNameAndAddress, "items["+indexStr+"].consigneeNameAndAddress", 0, 500); err != nil {
		if valErr, ok := err.(customerrors.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}

	// Validate destination as airport code if provided
	if req.Destination != "" {
		if err := businessValidator.ValidateAirportCode(req.Destination, "items["+indexStr+"].destination"); err != nil {
			if valErr, ok := err.(customerrors.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}

	return validationErrors
}

// SanitizeCargoManifestRequest sanitizes all string fields in the request
func SanitizeCargoManifestRequest(req *CargoManifestRequest) {
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	req.MAWBNumber = sanitizer.SanitizeString(req.MAWBNumber)
	req.PortOfDischarge = sanitizer.SanitizeString(req.PortOfDischarge)
	req.FlightNo = sanitizer.SanitizeString(req.FlightNo)
	req.FreightDate = sanitizer.SanitizeString(req.FreightDate)
	req.Shipper = sanitizer.SanitizeString(req.Shipper)
	req.Consignee = sanitizer.SanitizeString(req.Consignee)
	req.TotalCtn = sanitizer.SanitizeString(req.TotalCtn)
	req.Transshipment = sanitizer.SanitizeString(req.Transshipment)

	// Sanitize items
	for i := range req.Items {
		SanitizeCargoManifestItemRequest(&req.Items[i])
	}
}

// SanitizeCargoManifestItemRequest sanitizes all string fields in the item request
func SanitizeCargoManifestItemRequest(req *CargoManifestItemRequest) {
	sanitizer := common.NewInputSanitizer(common.DefaultValidationConfig())

	req.HAWBNo = sanitizer.SanitizeString(req.HAWBNo)
	req.Pkgs = sanitizer.SanitizeString(req.Pkgs)
	req.GrossWeight = sanitizer.SanitizeString(req.GrossWeight)
	req.Destination = sanitizer.SanitizeString(req.Destination)
	req.Commodity = sanitizer.SanitizeString(req.Commodity)
	req.ShipperNameAndAddress = sanitizer.SanitizeString(req.ShipperNameAndAddress)
	req.ConsigneeNameAndAddress = sanitizer.SanitizeString(req.ConsigneeNameAndAddress)
}
