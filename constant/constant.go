package constant

import "errors"

const (
	CodeSuccess      = 200
	CodeError        = 400
	CodeUnauthorized = 401
)

const (
	API_LOG_STATUS_SUCCESS = "success"
	API_LOG_STATUS_FAILED  = "failed"
	API_LOG_STATUS_NEW     = "new"

	API_LOG_TYPE_REQUEST = "request"
	API_LOG_TYPE_WEBHOOK = "webhook"
)

var (
	ErrInvalidArgument             = errors.New("invalid argument")
	ErrUsernameOrPasswordIncorrect = errors.New("username or password incorrect")
	ErrNotFound                    = errors.New("not found")
	ErrEmpty                       = errors.New("data empty")
	ErrRequireUUID                 = errors.New("required uuid")
	ErrTemplateVersion             = errors.New("template version is incorrect")
	ErrDatabaseConnectionNil       = errors.New("db connection null")
	ErrForbiddenQrCode             = errors.New("Forbidden to use QR code")
)

type DropdownModel struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type InsertHistory struct {
	ParentUUID string `json:"parentUUID"`
	UserUUID   string `json:"userUUID"`
	Status     string `json:"status"`
	Remark     string `json:"remark"`
	ZoneUUID   string `json:"zoneUUID"`
}

type ZoneDropdownModel struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	AreaType string `json:"areaType"`
}
