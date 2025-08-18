package setting

import (
	"net/http"
	"time"
)

type MasterStatus struct {
	tableName struct{}  `pg:"master_status,alias:ms"`
	UUID      string    `json:"uuid" pg:"uuid,pk"`
	Name      string    `json:"name" pg:"name"`
	Type      string    `json:"type" pg:"type"`
	IsDefault bool      `json:"isDefault" pg:"is_default"`
	CreatedAt time.Time `json:"createdAt" pg:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" pg:"updated_at"`
}

func (ms *MasterStatus) Bind(r *http.Request) error {
	return nil
}
