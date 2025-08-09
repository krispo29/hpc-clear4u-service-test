package outbound

import "time"

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
	Items           []CargoManifestItem `json:"items" db:"-"` // db:"-" to ignore this field in main struct scans
	CreatedAt       time.Time           `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time           `json:"updatedAt" db:"updated_at"`
}

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
	CreatedAt               time.Time `json:"-" db:"created_at"`
}
