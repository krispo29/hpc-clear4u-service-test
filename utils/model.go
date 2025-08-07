package utils

type DropdownModel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetMawb struct {
	UUID              string
	FlightNo          string
	Origin            string
	Destination       string
	LotNo             string
	Mawb              string
	DepartureDatetime string
	ArrivalDatetime   string
}
