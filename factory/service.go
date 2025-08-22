package factory

import (
	"time"

	"hpc-express-service/auth"
	"hpc-express-service/common"
	"hpc-express-service/config"
	"hpc-express-service/customer"
	"hpc-express-service/dashboard"
	"hpc-express-service/dropdown"
	"hpc-express-service/gcs"
	inbound "hpc-express-service/inbound/express"
	"hpc-express-service/mawb"
	cargoManifest "hpc-express-service/outbound/cargoManifest"
	draftMawb "hpc-express-service/outbound/draftMawb"
	outboundExpress "hpc-express-service/outbound/express"
	outboundMawb "hpc-express-service/outbound/mawb"
	"hpc-express-service/outbound/mawbinfo"
	weightSlip "hpc-express-service/outbound/weightSlip"
	"hpc-express-service/setting"
	"hpc-express-service/ship2cu"
	"hpc-express-service/shopee"
	"hpc-express-service/tools/compare"
	"hpc-express-service/uploadlog"
	"hpc-express-service/user"
)

type ServiceFactory struct {
	AuthSvc                   auth.Service
	CommonSvc                 common.Service
	CompareSvc                compare.ExcelServiceInterface
	DropdownSvc               dropdown.Service
	InboundExpressServiceSvc  inbound.InboundExpressService
	Ship2cuSvc                ship2cu.Service
	UploadlogSvc              uploadlog.Service
	OutboundExpressServiceSvc outboundExpress.OutboundExpressService
	OutboundMawbServiceSvc    outboundMawb.OutboundMawbService
	ShopeeSvc                 shopee.Service
	MawbSvc                   mawb.Service
	MawbInfoSvc               mawbinfo.Service
	CustomerSvc               customer.Service
	DashboardSvc              dashboard.Service
	UserSvc                   user.Service
	SettingSvc                setting.Service
	CargoManifestSvc          cargoManifest.CargoManifestService
	DraftMAWBSvc              draftMawb.DraftMAWBService
	WeightSlipSvc             weightSlip.WeightSlipService
	MasterStatusSvc           setting.MasterStatusService
}

func NewServiceFactory(repo *RepositoryFactory, gcsClient *gcs.Client, conf *config.Config) *ServiceFactory {
	timeoutContext := time.Duration(60) * time.Second

	/*
	* Sharing Services
	 */

	// setting
	settingSvc := setting.NewService(
		repo.SettingRepo,
		timeoutContext,
	)

	// MasterStatus
	masterStatusSvc := setting.NewMasterStatusService(
		repo.MasterStatusRepo,
		timeoutContext,
	)

	// Ship2cu
	ship2cuSvc := ship2cu.NewService(
		repo.Ship2cuRepo,
		timeoutContext,
	)

	// Shopee
	shopeeSvc := shopee.NewService(
		repo.ShopeeRepo,
		timeoutContext,
	)

	// Upload Logging
	uploadlogSvc := uploadlog.NewService(
		repo.UploadlogRepo,
		timeoutContext,
		gcsClient,
	)

	// MAWB
	mawbSvc := mawb.NewService(
		repo.MawbRepo,
		timeoutContext,
	)

	// Customer
	customerSvc := customer.NewService(
		repo.CustomerRepo,
		timeoutContext,
	)

	// User
	userSvc := user.NewService(
		repo.UserRepo,
		timeoutContext,
	)

	// MawbInfo
	mawbInfoSvc := mawbinfo.NewService(
		repo.MawbInfoRepo,
		timeoutContext,
		gcsClient,
		conf,
	)
	/*
	* Sharing Services
	 */
	// Compare Service
	compareSvc := compare.NewExcelService(repo.CompareRepo)
	// Auth
	authSvc := auth.NewService(
		repo.AuthRepo,
		timeoutContext,
	)

	// Common
	dashboardSvc := dashboard.NewService(
		repo.DashboardRepo,
		timeoutContext,
	)

	// Common
	commonSvc := common.NewService(
		repo.CommonRepo,
		timeoutContext,
	)

	// Dropdown
	dropdownSvc := dropdown.NewService(
		repo.DropdownRepo,
		timeoutContext,
	)

	// Inbound Express
	inboundExpressServiceSvc := inbound.NewInboundExpressService(
		repo.InboundExpressRepositoryRepo,
		timeoutContext,
		ship2cuSvc,
		uploadlogSvc,
		repo.Ship2cuRepo,
	)

	// Outbound Express
	outboundExpressServiceSvc := outboundExpress.NewOutboundExpressService(
		repo.OutboundExpressRepositoryRepo,
		timeoutContext,
		shopeeSvc,
		uploadlogSvc,
	)

	// Outbound Mawb
	outboundMawbServiceSvc := outboundMawb.NewOutboundMawbService(
		repo.OutboundMawbRepositoryRepo,
		timeoutContext,
		gcsClient,
		conf,
	)

	// Cargo Manifest
	cargoManifestSvc := cargoManifest.NewCargoManifestService(repo.CargoManifestRepo, masterStatusSvc)

	// Draft MAWB
	draftMAWBSvc := draftMawb.NewDraftMAWBService(repo.DraftMAWBRepo, masterStatusSvc)

	// Weight Slip
	weightSlipSvc := weightSlip.NewWeightSlipService(repo.WeightSlipRepo, masterStatusSvc)

	return &ServiceFactory{
		AuthSvc:                   authSvc,
		CommonSvc:                 commonSvc,
		DropdownSvc:               dropdownSvc,
		InboundExpressServiceSvc:  inboundExpressServiceSvc,
		Ship2cuSvc:                ship2cuSvc,
		UploadlogSvc:              uploadlogSvc,
		OutboundExpressServiceSvc: outboundExpressServiceSvc,
		OutboundMawbServiceSvc:    outboundMawbServiceSvc,
		ShopeeSvc:                 shopeeSvc,
		MawbSvc:                   mawbSvc,
		MawbInfoSvc:               mawbInfoSvc,
		CustomerSvc:               customerSvc,
		DashboardSvc:              dashboardSvc,
		UserSvc:                   userSvc,
		CompareSvc:                compareSvc,
		SettingSvc:                settingSvc,
		CargoManifestSvc:          cargoManifestSvc,
		DraftMAWBSvc:              draftMAWBSvc,
		WeightSlipSvc:             weightSlipSvc,
		MasterStatusSvc:           masterStatusSvc,
	}
}
