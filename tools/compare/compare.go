package compare

type DBDetails struct {
	GoodsEN   string  `json:"goods_en"`
	GoodsTH   string  `json:"goods_th"`
	Tariff    int     `json:"tariff"`
	Stat      int     `json:"stat"`
	UnitCode  string  `json:"unit_code"`
	DutyRate  float64 `json:"duty_rate"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	DeletedAt *string `json:"deleted_at,omitempty"`
	Remark    *string `json:"remark,omitempty"`
	HSCode    string  `json:"hs_code"`
}

type ExcelItem struct {
	Value     string     `json:"value"`
	IsMatch   bool       `json:"isMatch"`
	MatchedBy string     `json:"matchedBy,omitempty"`
	DBDetails *DBDetails `json:"dbDetails,omitempty"`
}

type CompareResponse struct {
	TotalExcelRows int         `json:"totalExcelRows"`
	TotalDBRows    int         `json:"totalDBRows"`
	MatchedRows    int         `json:"matchedRows"`
	ExcelItems     []ExcelItem `json:"excelItems"`
}
