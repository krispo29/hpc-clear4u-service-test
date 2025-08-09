package mawbinfo

import (
	"net/http"
	"time"
)

type CargoManifest struct {
	UUID            string              `json:"uuid" db:"uuid"`
	MAWBInfoUUID    string              `json:"mawb_info_uuid" db:"mawb_info_uuid"`
	MAWBNumber      string              `json:"mawbNumber" db:"mawb_number" validate:"required"`
	PortOfDischarge string              `json:"portOfDischarge" db:"port_of_discharge"`
	FlightNo        string              `json:"flightNo" db:"flight_no"`
	FreightDate     string              `json:"freightDate" db:"freight_date"`
	Shipper         string              `json:"shipper" db:"shipper"`
	Consignee       string              `json:"consignee" db:"consignee"`
	TotalCtn        string              `json:"totalCtn" db:"total_ctn"`
	Transshipment   string              `json:"transshipment" db:"transshipment"`
	Status          string              `json:"status" db:"status"`
	Items           []CargoManifestItem `json:"items" validate:"required,dive"`
	CreatedAt       time.Time           `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time           `json:"updatedAt" db:"updated_at"`
}

type CargoManifestItem struct {
	ID                      int    `json:"id,omitempty" db:"id"`
	CargoManifestUUID       string `json:"cargo_manifest_uuid,omitempty" db:"cargo_manifest_uuid"`
	HAWBNo                  string `json:"hawbNo" db:"hawb_no"`
	Pkgs                    string `json:"pkgs" db:"pkgs"`
	GrossWeight             string `json:"grossWeight" db:"gross_weight"`
	Destination             string `json:"dst" db:"destination"`
	Commodity               string `json:"commodity" db:"commodity"`
	ShipperNameAndAddress   string `json:"shipperNameAndAddress" db:"shipper_name_address"`
	ConsigneeNameAndAddress string `json:"consigneeNameAndAddress" db:"consignee_name_address"`
}

// Bind for CargoManifest - for request body binding
func (cm *CargoManifest) Bind(r *http.Request) error {
	// a good place to do validation or transformation on the request body
	return nil
}

// Render for CargoManifest - for response rendering
func (cm *CargoManifest) Render(w http.ResponseWriter, r *http.Request) error {
	// a good place to do pre-processing before sending response
	return nil
}
