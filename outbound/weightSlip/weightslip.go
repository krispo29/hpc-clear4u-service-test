package outbound

import (
	"net/http"
	"time"
)

type WeightSlip struct {
	tableName     struct{}              `pg:"public.weight_slip"`
	UUID          string                `json:"uuid" pg:"uuid"`
	MAWBInfoUUID  string                `json:"mawbInfoUuid" pg:"mawb_info_uuid"`
	SlipNo        string                `json:"slipNo" pg:"slip_no"`
	WSID          string                `json:"wsid" pg:"wsid"`
	DateTime      time.Time             `json:"dateTime" pg:"date_time"`
	PSEQ          string                `json:"pseq" pg:"pseq"`
	Staff         string                `json:"staff" pg:"staff"`
	MAWB          string                `json:"mawb" pg:"mawb"`
	HAWB          string                `json:"hawb" pg:"hawb"`
	Dest          string                `json:"dest" pg:"dest"`
	AgentCode     string                `json:"-" pg:"agent_code"`
	AgentName     string                `json:"-" pg:"agent_name"`
	Agent         Agent                 `json:"agent" pg:"-"`
	Flight        string                `json:"flight" pg:"flight"`
	NatureOfGoods string                `json:"natureOfGoods" pg:"nature_of_goods"`
	EWS           bool                  `json:"ews" pg:"ews"`
	PCS           int                   `json:"pcs" pg:"pcs"`
	GW            float64               `json:"-" pg:"gw"`
	TW            float64               `json:"-" pg:"tw"`
	NW            float64               `json:"-" pg:"nw"`
	DimWeight     float64               `json:"-" pg:"dim_weight"`
	VolumeM3      float64               `json:"-" pg:"volume_m3"`
	Weights       Weights               `json:"weights" pg:"-"`
	StatusUUID    string                `json:"statusUuid" pg:"status_uuid"`
	Status        string                `json:"status" pg:"-"`
	Dimensions    []WeightSlipDimension `json:"dimensions"`
	CreatedAt     time.Time             `json:"createdAt" pg:"created_at"`
	UpdatedAt     time.Time             `json:"updatedAt" pg:"updated_at"`
}

type Agent struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Weights struct {
	GW        float64 `json:"gw"`
	TW        float64 `json:"tw"`
	NW        float64 `json:"nw"`
	DimWeight float64 `json:"dimWeight"`
	VolumeM3  float64 `json:"volumeM3"`
}

type WeightSlipDimension struct {
	tableName      struct{} `pg:"public.weight_slip_dimensions"`
	ID             int      `json:"id" pg:"id"`
	WeightSlipUUID string   `json:"weightslipUuid" pg:"weightslip_uuid"`
	No             int      `json:"no" pg:"no"`
	Lcm            float64  `json:"L_cm" pg:"l_cm"`
	Wcm            float64  `json:"W_cm" pg:"w_cm"`
	Hcm            float64  `json:"H_cm" pg:"h_cm"`
	PCS            int      `json:"pcs" pg:"pcs"`
}

func (w *WeightSlip) Bind(r *http.Request) error {
	w.AgentCode = w.Agent.Code
	w.AgentName = w.Agent.Name
	w.GW = w.Weights.GW
	w.TW = w.Weights.TW
	w.NW = w.Weights.NW
	w.DimWeight = w.Weights.DimWeight
	w.VolumeM3 = w.Weights.VolumeM3
	return nil
}

func (w *WeightSlip) AfterSelect(ctx interface{}) error {
	w.Agent = Agent{Code: w.AgentCode, Name: w.AgentName}
	w.Weights = Weights{GW: w.GW, TW: w.TW, NW: w.NW, DimWeight: w.DimWeight, VolumeM3: w.VolumeM3}
	return nil
}
