package dashboard

type DashboardV2Model struct {
	Summary *SummaryModel
	Data    *VolumneChartModel
}

type SummaryModel struct {
	InboundTotalHawbCount    string `json:"inboundTotalHawbCount"`
	InboundTotalNetWeight    string `json:"inboundTotalNetWeight"`
	InboundTotalGrossWeight  string `json:"inboundTotalGrossWeight"`
	OutboundTotalHawbCount   string `json:"outboundTotalHawbCount"`
	OutboundTotalNetWeight   string `json:"outboundTotalNetWeight"`
	OutboundTotalGrossWeight string `json:"outboundTotalGrossWeight"`
}

type VolumneChartModel struct {
	MmYy []string                   `json:"mmyy"`
	Data []*VolumneChartDetailModel `json:"data"`
}

type VolumneChartDetailModel struct {
	Name string  `json:"name"`
	Data []int64 `json:"data"`
}

type GetVolumneModel struct {
	MmYy                   string `json:"mmyy"`
	InboundTotalHawbCount  int64  `json:"inboundTotalHawbCount"`
	OutboundTotalHawbCount int64  `json:"outboundTotalHawbCount"`
}
