package uploadlog

type GetUploadloggingModel struct {
	UUID         string `json:"uuid"`
	Mawb         string `json:"mawb"`
	FileName     string `json:"fileName"`
	FileURL      string `json:"fileURL"`
	TemplateCode string `json:"templateCode"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	Remark       string `json:"remark"`
	Creator      string `json:"creator"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type UploadFileModel struct {
	Mawb         string
	UserUUID     string
	FileName     string
	TemplateCode string
	Category     string
	SubCategory  string
	FileBytes    []byte
	Amount       int64
}

type InsertModel struct {
	Mawb         string
	FileName     string
	FileUrl      string
	TemplateCode string
	Category     string
	SubCategory  string
	CreatorUUID  string
	Status       string
	Amount       int64
}

type UpdateModel struct {
	UUID   string
	Mawb   string
	Amount int64
	Status string
	Remark string
}
