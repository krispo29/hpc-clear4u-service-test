package cargomanifest

import (
	"net/http"
	"time"
)

type CargoManifest struct {
	tableName       struct{}            `pg:"public.cargo_manifest"`
	UUID            string              `json:"uuid" pg:"uuid"`
	MAWBInfoUUID    string              `json:"mawbInfoUuid" pg:"mawb_info_uuid"`
	MAWBNumber      string              `json:"mawbNumber" pg:"mawb_number"`
	PortOfDischarge string              `json:"portOfDischarge" pg:"port_of_discharge"`
	FlightNo        string              `json:"flightNo" pg:"flight_no"`
	FreightDate     string              `json:"freightDate" pg:"freight_date"`
	Shipper         string              `json:"shipper" pg:"shipper"`
	Consignee       string              `json:"consignee" pg:"consignee"`
	TotalCtn        string              `json:"totalCtn" pg:"total_ctn"`
	Transshipment   string              `json:"transshipment" pg:"transshipment"`
	StatusUUID      string              `json:"statusUuid" pg:"status_uuid"`
	Status          string              `json:"status" pg:"-"`
	Items           []CargoManifestItem `json:"items"`
	CreatedAt       time.Time           `json:"createdAt" pg:"created_at"`
	UpdatedAt       time.Time           `json:"updatedAt" pg:"updated_at"`
}

func (c *CargoManifest) Bind(r *http.Request) error {
	return nil
}

type CargoManifestItem struct {
	tableName               struct{} `pg:"public.cargo_manifest_items"`
	ID                      int      `json:"id" db:"id"`
	CargoManifestUUID       string   `json:"cargoManifestUuid" db:"cargo_manifest_uuid"`
	HAWBNo                  string   `json:"hawbNo" db:"hawb_no"`
	Pkgs                    string   `json:"pkgs" db:"pkgs"`
	GrossWeight             string   `json:"grossWeight" db:"gross_weight"`
	Destination             string   `json:"destination" db:"destination"`
	Commodity               string   `json:"commodity" db:"commodity"`
	ShipperNameAndAddress   string   `json:"shipperNameAndAddress" db:"shipper_name_address"`
	ConsigneeNameAndAddress string   `json:"consigneeNameAndAddress" db:"consignee_name_address"`
}

// CargoManifestListItem represents a cargo manifest item in the list view
// with limited fields for listing endpoints.
type CargoManifestListItem struct {
	UUID         string `json:"uuid"`
	MAWBInfoUUID string `json:"mawbInfoUuid"`
	MAWBNumber   string `json:"mawbNumber"`
	CustomerName string `json:"customerName"`
	CreatedAt    string `json:"createdAt"`
	Status       string `json:"status"`
	IsDeleted    bool   `json:"isDeleted"`
}
