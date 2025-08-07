package shopee

import (
	"context"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	InsertPreExportManifest(ctx context.Context, manifest *utils.InsertPreExportHeaderManifestModel, chunkSize int) error
}

type repository struct {
	contextTimeout time.Duration
}

func NewRepository(
	timeout time.Duration,
) Repository {
	return &repository{
		contextTimeout: timeout,
	}
}

func (r repository) InsertPreExportManifest(ctx context.Context, manifest *utils.InsertPreExportHeaderManifestModel, chunkSize int) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Insert HeaderManifest
	var headerUUID string
	{
		sqlStr :=
			`
			INSERT INTO public.tbl_pre_export_manifest_headers
				(
					upload_logging_uuid, vassel_name, departure_date, release_port, loading_port, total_package, total_package_unit_code, total_net_weight, total_net_weight_unit_code, total_gross_weight, total_gross_weight_unit_code
				)
			VALUES
				(
					?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
				)
			RETURNING uuid
		`

		// Prepare statement
		stmt, err := tx.Prepare(utils.PrepareSQL(sqlStr))
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		values := []interface{}{}
		values = append(
			values,
			manifest.UploadLoggingUUID,
			manifest.VasselName,
			manifest.DepartureDate,
			manifest.ReleasePort,
			manifest.LoadingPort,
			manifest.TotalPackage,
			manifest.TotalPackageUnitCode,
			manifest.TotalNetWeight,
			manifest.TotalNetWeightUnitCode,
			manifest.TotalGrossWeight,
			manifest.TotalGrossWeightUnitCode,
		)

		_, err = stmt.QueryOneContext(ctx, &headerUUID, values...)

		if err != nil {
			tx.Rollback()
			return err
		}
	}
	// Insert Manifest
	// Chunk slice
	chunked := utils.ChunkSlice(manifest.Details, chunkSize)
	{

		for _, chunkedRows := range chunked {
			sqlStr := `
				INSERT INTO public.tbl_pre_export_manifest_details 
					(
						header_uuid, master_air_waybill, house_air_waybill, category, consignor_company_tax_number, consignor_company_branch, consignor_name, consignor_street_and_address, consignor_district, consignor_sub_province, consignor_province, consignor_postcode, consignor_email, consignee_name, consignee_street_and_address, consignee_district, consignee_sub_province, consignee_province, consignee_postcode, consignee_country_code, consignee_email, purchase_country_code, destination_country_code, thai_description_of_goods, english_description_of_goods, quantity, quantity_unit_code, net_weight, net_weight_unit_code, gross_weight, gross_weight_unit_code, package_amount, package_unit_code, remark, fob_value_baht, fob_value_foreign, currency_code, exchange_rate, freight_amount, freight_amount_currency_code, insurance_amount, insurance_amount_currency_code, tariff_code, stat_code, tariff_sequence
					) 
					VALUES 
			`
			vals := []interface{}{}
			for _, row := range chunkedRows {
				row.HeaderUUID = headerUUID

				sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
				vals = append(vals,
					utils.NewNullString(row.HeaderUUID),
					utils.NewNullString(row.MasterAirWaybill),
					utils.NewNullString(row.HouseAirWaybill),
					utils.NewNullInt(row.Category),
					utils.NewNullString(row.ConsignorCompanyTaxNumber),
					utils.NewNullString(row.ConsignorCompanyBranch),
					utils.NewNullString(row.ConsignorName),
					utils.NewNullString(row.ConsignorStreetAndAddress),
					utils.NewNullString(row.ConsignorDistrict),
					utils.NewNullString(row.ConsignorSubProvince),
					utils.NewNullString(row.ConsignorProvince),
					utils.NewNullString(row.ConsignorPostcode),
					utils.NewNullString(row.ConsignorEmail),
					utils.NewNullString(row.ConsigneeName),
					utils.NewNullString(row.ConsigneeStreetAndAddress),
					utils.NewNullString(row.ConsigneeDistrict),
					utils.NewNullString(row.ConsigneeSubProvince),
					utils.NewNullString(row.ConsigneeProvince),
					utils.NewNullString(row.ConsigneePostcode),
					utils.NewNullString(row.ConsigneeCountryCode),
					utils.NewNullString(row.ConsigneeEmail),
					utils.NewNullString(row.PurchaseCountryCode),
					utils.NewNullString(row.DestinationCountryCode),
					utils.NewNullString(row.ThaiDescriptionOfGoods),
					utils.NewNullString(row.EnglishDescriptionOfGoods),
					utils.NewNullInt(row.Quantity),
					utils.NewNullString(row.QuantityUnitCode),
					row.NetWeight,
					utils.NewNullString(row.NetWeightUnitCode),
					row.GrossWeight,
					utils.NewNullString(row.GrossWeightUnitCode),
					utils.NewNullInt(row.PackageAmount),
					utils.NewNullString(row.PackageUnitCode),
					utils.NewNullString(row.Remark),
					row.FobValueBaht,
					row.FobValueForeign,
					utils.NewNullString(row.CurrencyCode),
					utils.NewNullInt(row.ExchangeRate),
					utils.NewNullInt(row.FreightAmount),
					utils.NewNullString(row.FreightAmountCurrencyCode),
					utils.NewNullInt(row.InsuranceAmount),
					utils.NewNullString(row.InsuranceAmountCurrencyCode),
					utils.NewNullString(row.TariffCode),
					utils.NewNullString(row.StatCode),
					utils.NewNullString(row.TariffSequence),
				)
			}

			// remove last comma,
			sqlStr = sqlStr[0 : len(sqlStr)-1]

			// Convert symbol ? to $
			sqlStr = utils.ReplaceSQL(sqlStr, "?")
			// sqlStr += " ON CONFLICT (local_no) DO NOTHING returning uuid, local_no;"

			// Prepare statement
			stmt, err := tx.Prepare(sqlStr)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer stmt.Close()

			_, err = stmt.ExecContext(ctx, vals...)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	tx.Commit()

	return nil
}
