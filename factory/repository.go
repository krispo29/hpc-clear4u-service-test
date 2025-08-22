package factory

import (
	"hpc-express-service/auth"
	"hpc-express-service/common"
	"hpc-express-service/customer"
	"hpc-express-service/dashboard"
	"hpc-express-service/dropdown"
	inbound "hpc-express-service/inbound/express"
	"hpc-express-service/mawb"
	cargoManifest "hpc-express-service/outbound/cargomanifest"
	draftMawb "hpc-express-service/outbound/draftmawb"
	outboundExpress "hpc-express-service/outbound/express"
	outboundMawb "hpc-express-service/outbound/mawb"
	"hpc-express-service/outbound/mawbinfo"
	"hpc-express-service/setting"
	"hpc-express-service/ship2cu"
	"hpc-express-service/shopee"
	"hpc-express-service/tools/compare"
	"hpc-express-service/uploadlog"
	"hpc-express-service/user"
	"time"
)

type RepositoryFactory struct {
	AuthRepo                      auth.Repository
	CommonRepo                    common.Repository
	CompareRepo                   compare.ExcelRepositoryInterface
	DropdownRepo                  dropdown.Repository
	InboundExpressRepositoryRepo  inbound.InboundExpressRepository
	Ship2cuRepo                   ship2cu.Repository
	UploadlogRepo                 uploadlog.Repository
	OutboundExpressRepositoryRepo outboundExpress.OutboundExpressRepository
	OutboundMawbRepositoryRepo    outboundMawb.OutboundMawbRepository
	ShopeeRepo                    shopee.Repository
	MawbRepo                      mawb.Repository
	MawbInfoRepo                  mawbinfo.Repository
	CustomerRepo                  customer.Repository
	DashboardRepo                 dashboard.Repository
	UserRepo                      user.Repository
	SettingRepo                   setting.Repository
	CargoManifestRepo             cargoManifest.CargoManifestRepository
	DraftMAWBRepo                 draftMawb.DraftMAWBRepository
	MasterStatusRepo              setting.MasterStatusRepository
}

func NewRepositoryFactory() *RepositoryFactory {
	timeoutContext := time.Duration(60) * time.Second

	return &RepositoryFactory{
		AuthRepo:                      auth.NewRepository(timeoutContext),
		CommonRepo:                    common.NewRepository(timeoutContext),
		DropdownRepo:                  dropdown.NewRepository(),
		InboundExpressRepositoryRepo:  inbound.NewInboundExpressRepository(timeoutContext),
		Ship2cuRepo:                   ship2cu.NewRepository(timeoutContext),
		UploadlogRepo:                 uploadlog.NewRepository(timeoutContext),
		OutboundExpressRepositoryRepo: outboundExpress.NewOutboundExpressRepository(timeoutContext),
		OutboundMawbRepositoryRepo:    outboundMawb.NewOutboundMawbRepository(timeoutContext),
		ShopeeRepo:                    shopee.NewRepository(timeoutContext),
		MawbRepo:                      mawb.NewRepository(timeoutContext),
		MawbInfoRepo:                  mawbinfo.NewRepository(timeoutContext),
		CustomerRepo:                  customer.NewRepository(timeoutContext),
		DashboardRepo:                 dashboard.NewRepository(timeoutContext),
		UserRepo:                      user.NewRepository(timeoutContext),
		CompareRepo:                   compare.NewExcelRepository(timeoutContext),
		SettingRepo:                   setting.NewRepository(timeoutContext),
		CargoManifestRepo:             cargoManifest.NewCargoManifestRepository(),
		DraftMAWBRepo:                 draftMawb.NewDraftMAWBRepository(),
		MasterStatusRepo:              setting.NewMasterStatusRepository(),
	}
}
