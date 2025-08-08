package mawbinfo

import (
	"net/http"
	"time"
)

// CargoManifest represents the cargo manifest linked to MAWB info.
type CargoManifest struct {
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
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

// CargoManifestItem represents an item within a cargo manifest.
type CargoManifestItem struct {
	ID                      int    `json:"id"`
	CargoManifestUUID       string `json:"cargo_manifest_uuid"`
	HAWBNo                  string `json:"hawbNo"`
	Pkgs                    string `json:"pkgs"`
	GrossWeight             string `json:"grossWeight"`
	Destination             string `json:"dst"`
	Commodity               string `json:"commodity"`
	ShipperNameAndAddress   string `json:"shipperNameAndAddress"`
	ConsigneeNameAndAddress string `json:"consigneeNameAndAddress"`
}

// Bind is used by chi render to bind incoming JSON.
func (c *CargoManifest) Bind(r *http.Request) error {
	return nil
}

// Render is used by chi render for responses.
func (c *CargoManifest) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
