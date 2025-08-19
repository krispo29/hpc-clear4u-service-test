package setting

import (
	"context"
	"hpc-express-service/common"
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
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	_, err = db.Model(status).Insert()
	return status, err
}

func (r *masterStatusRepository) GetAllMasterStatuses(ctx context.Context) ([]MasterStatus, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	var statuses []MasterStatus
	err = db.Model(&statuses).Select()
	return statuses, err
}

func (r *masterStatusRepository) GetMasterStatusesByType(ctx context.Context, statusType string) ([]MasterStatus, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	var statuses []MasterStatus
	err = db.Model(&statuses).Where("type = ?", statusType).Select()
	return statuses, err
}

func (r *masterStatusRepository) GetMasterStatusByUUID(ctx context.Context, uuid string) (*MasterStatus, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	status := new(MasterStatus)
	err = db.Model(status).Where("uuid = ?", uuid).Select()
	return status, err
}

func (r *masterStatusRepository) UpdateMasterStatus(ctx context.Context, status *MasterStatus) (*MasterStatus, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	_, err = db.Model(status).WherePK().Update()
	return status, err
}

func (r *masterStatusRepository) DeleteMasterStatus(ctx context.Context, uuid string) error {
	db, err := common.GetQer(ctx)
	if err != nil {
		return err
	}
	_, err = db.Model(&MasterStatus{}).Where("uuid = ?", uuid).Delete()
	return err
}

func (r *masterStatusRepository) GetDefaultStatusByType(ctx context.Context, statusType string) (*MasterStatus, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	status := new(MasterStatus)
	err = db.Model(status).Where("type = ?", statusType).Where("is_default = ?", true).First()
	return status, err
}

func (r *masterStatusRepository) GetStatusByNameAndType(ctx context.Context, name, statusType string) (*MasterStatus, error) {
	db, err := common.GetQer(ctx)
	if err != nil {
		return nil, err
	}
	status := new(MasterStatus)
	err = db.Model(status).Where("name = ?", name).Where("type = ?", statusType).First()
	return status, err
}
