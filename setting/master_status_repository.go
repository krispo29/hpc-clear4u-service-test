package setting

import (
	"context"
	"hpc-express-service/utils"
)

type MasterStatusRepository interface {
	CreateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error)
	GetAllMasterStatuses(ctx context.Context) ([]MasterStatus, error)
	GetMasterStatusesByType(ctx context.Context, statusType string) ([]MasterStatus, error)
	GetMasterStatusByUUID(ctx context.Context, uuid string) (*MasterStatus, error)
	UpdateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error)
	DeleteMasterStatus(ctx context.Context, uuid string) error
	GetDefaultStatusByType(ctx context.Context, statusType string) (*MasterStatus, error)
	GetStatusByNameAndType(ctx context.Context, name, statusType string) (*MasterStatus, error)
}

type masterStatusRepository struct{}

func NewMasterStatusRepository() MasterStatusRepository {
	return &masterStatusRepository{}
}

func (r *masterStatusRepository) CreateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	_, err = db.ModelContext(ctx, status).Insert()
	return status, err
}

func (r *masterStatusRepository) GetAllMasterStatuses(ctx context.Context) ([]MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var statuses []MasterStatus
	err = db.ModelContext(ctx, &statuses).Select()
	return statuses, err
}

func (r *masterStatusRepository) GetMasterStatusesByType(ctx context.Context, statusType string) ([]MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var statuses []MasterStatus
	err = db.ModelContext(ctx, &statuses).Where("type = ?", statusType).Select()
	return statuses, err
}

func (r *masterStatusRepository) GetMasterStatusByUUID(ctx context.Context, uuid string) (*MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	status := new(MasterStatus)
	err = db.ModelContext(ctx, status).Where("uuid = ?", uuid).Select()
	return status, err
}

func (r *masterStatusRepository) UpdateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	_, err = db.ModelContext(ctx, status).WherePK().Update()
	return status, err
}

func (r *masterStatusRepository) DeleteMasterStatus(ctx context.Context, uuid string) error {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return err
	}
	_, err = db.ModelContext(ctx, &MasterStatus{}).Where("uuid = ?", uuid).Delete()
	return err
}

func (r *masterStatusRepository) GetDefaultStatusByType(ctx context.Context, statusType string) (*MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	status := new(MasterStatus)
	_, err = db.QueryOne(status, `SELECT * FROM master_status WHERE type = ? AND is_default = true LIMIT 1`, statusType)
	return status, err
}

func (r *masterStatusRepository) GetStatusByNameAndType(ctx context.Context, name, statusType string) (*MasterStatus, error) {
	db, err := utils.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	status := new(MasterStatus)
	err = db.ModelContext(ctx, status).Where("name = ?", name).Where("type = ?", statusType).First()
	return status, err
}
