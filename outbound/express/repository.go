package outbound

import (
	"context"
	"hpc-express-service/utils"
	"log"
	"time"

	"github.com/go-pg/pg/v9"
)

type OutboundExpressRepository interface {
	// GetMawbByTimstamp(ctx context.Context, timestamp string) (*GetMawb, error)
	GetAllManifestToPreExport(ctx context.Context, uploadLoggingUUID string) (*utils.GetHeaderManifestPreExport, error)
}

type repository struct {
	contextTimeout time.Duration
}

func NewOutboundExpressRepository(
	timeout time.Duration,
) OutboundExpressRepository {
	return &repository{
		contextTimeout: timeout,
	}
}

func (r repository) GetAllManifestToPreExport(ctx context.Context, uploadLoggingUUID string) (*utils.GetHeaderManifestPreExport, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	result := &utils.GetHeaderManifestPreExport{}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&result.UUID,
		&result.UploadLoggingUUID,
		&result.VasselName,
		&result.DepartureDate,
		&result.ReleasePort,
		&result.LoadingPort,
		&result.TotalPackage,
		&result.TotalPackageUnitCode,
		&result.TotalNetWeight,
		&result.TotalNetWeightUnitCode,
		&result.TotalGrossWeight,
		&result.TotalGrossWeightUnitCode,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	), `
			SELECT 
				"uuid", 
				upload_logging_uuid, 
				vassel_name, 
				departure_date, 
				release_port, 
				loading_port, 
				total_package, 
				total_package_unit_code, 
				total_net_weight, 
				total_net_weight_unit_code, 
				total_gross_weight, 
				total_gross_weight_unit_code, 
				created_at, 
				updated_at, 
				deleted_at
			FROM public.tbl_pre_export_manifest_headers
			WHERE upload_logging_uuid = ?
	 `, uploadLoggingUUID)

	if err != nil {
		log.Println("xxx1")
		return nil, err
	}

	stmt, err := db.Prepare(`
	SELECT 
			"uuid", 
			header_uuid, 
			master_air_waybill, 
			house_air_waybill, 
			category, 
			consignor_company_tax_number, 
			consignor_company_branch, 
			consignor_name, 
			consignor_street_and_address, 
			consignor_district, 
			consignor_sub_province, 
			consignor_province, 
			consignor_postcode, 
			consignor_email, 
			consignee_name, 
			consignee_street_and_address, 
			consignee_district, 
			consignee_sub_province, 
			consignee_province, 
			consignee_postcode, 
			consignee_country_code, 
			consignee_email, 
			purchase_country_code, 
			destination_country_code, 
			thai_description_of_goods, 
			english_description_of_goods, 
			quantity, 
			quantity_unit_code, 
			net_weight, 
			net_weight_unit_code, 
			gross_weight, 
			gross_weight_unit_code, 
			package_amount, 
			package_unit_code, 
			remark, 
			fob_value_baht, 
			fob_value_foreign, 
			currency_code, 
			exchange_rate, 
			freight_amount, 
			freight_amount_currency_code, 
			insurance_amount, 
			insurance_amount_currency_code, 
			tariff_code, 
			stat_code, 
			tariff_sequence, 
			created_at, 
			updated_at, 
			deleted_at
		FROM public.tbl_pre_export_manifest_details
		WHERE header_uuid = $1
	`)
	if err != nil {
		log.Println("xxx")
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.QueryContext(ctx, &result.Details, result.UUID)
	if err != nil {
		return nil, err
	}

	return result, nil
}
