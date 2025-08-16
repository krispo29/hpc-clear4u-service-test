package outbound

import (
	"net/http"
	"time"
)

type CargoManifest struct {
	tableName       struct{}            `pg:"public.cargo_manifest"`
	UUID            string              `json:"uuid" db:"uuid"`
	MAWBInfoUUID    string              `json:"mawbInfoUuid" db:"mawb_info_uuid"`
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
