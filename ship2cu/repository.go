package ship2cu

import (
	"context"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	InsertPreImportManifest(ctx context.Context, manifest *utils.InsertPreImportHeaderManifestModel, chunkSize int) error
	GetMawb(ctx context.Context, timestamp string) (*utils.GetMawb, error)
	GetShipperBrands(ctx context.Context) ([]*GetShipperBrandModel, error)
	GetMasterHsCode(ctx context.Context) ([]*GetMasterHsCodeModel, error)
	GetFreightData(ctx context.Context, uploadLogUUID, countryCode, currencyCode string) (*GetFreightDataModel, error)
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

func (r repository) InsertPreImportManifest(ctx context.Context, manifest *utils.InsertPreImportHeaderManifestModel, chunkSize int) error {
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
			INSERT INTO public.tbl_pre_import_manifest_headers
				(
					upload_logging_uuid, discharge_port, vassel_name, arrival_date, customer_name, origin_country_code, origin_currency_code
				)
			VALUES
				(
					?, ?, ?, ?, ?, ?, ?
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
			manifest.DischargePort,
			manifest.VasselName,
			manifest.ArrivalDate,
			manifest.CustomerName,
			manifest.OriginCountryCode,
			manifest.OriginCurrencyCode,
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
			sqlStr :=
				`
				INSERT INTO public.tbl_pre_import_manifest_details 
					(
						header_uuid, master_air_waybill, house_air_waybill, category, consignee_tax, consignee_branch, consignee_name, consignee_address, consignee_district, consignee_subprovince, consignee_province, consignee_postcode, consignee_country_code, consignee_email, consignee_phone_number, shipper_name, shipper_address, shipper_district, shipper_subprovince, shipper_province, shipper_postcode, shipper_country_code, shipper_email, shipper_phone_number, tariff_code, tariff_sequence, statistical_code, english_description_of_good, thai_description_of_good, quantity, quantity_unit_code, net_weight, net_weight_unit_code, gross_weight, gross_weight_unit_code, package, package_unit_code, cif_value_foreign, fob_value_foreign, exchange_rate, currency_code, shipping_mark, consignment_country, freight_value_foreign, freight_currency_code, insurance_value_foreign, insurance_currency_code, other_charge_value_foreign, other_charge_currency_code, invoice_no, invoice_date
					) 
					VALUES 
			`
			vals := []interface{}{}
			for _, row := range chunkedRows {
				row.HeaderUUID = headerUUID

				sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
				vals = append(vals,
					utils.NewNullString(row.HeaderUUID),
					utils.NewNullString(row.MasterAirWaybill),
					utils.NewNullString(row.HouseAirWaybill),
					utils.NewNullString(row.Category),
					utils.NewNullString(row.ConsigneeTax),
					utils.NewNullString(row.ConsigneeBranch),
					utils.NewNullString(row.ConsigneeName),
					utils.NewNullString(row.ConsigneeAddress),
					utils.NewNullString(row.ConsigneeDistrict),
					utils.NewNullString(row.ConsigneeSubprovince),
					utils.NewNullString(row.ConsigneeProvince),
					utils.NewNullString(row.ConsigneePostcode),
					utils.NewNullString(row.ConsigneeCountryCode),
					utils.NewNullString(row.ConsigneeEmail),
					utils.NewNullString(row.ConsigneePhoneNumber),
					utils.NewNullString(row.ShipperName),
					utils.NewNullString(row.ShipperAddress),
					utils.NewNullString(row.ShipperDistrict),
					utils.NewNullString(row.ShipperSubprovince),
					utils.NewNullString(row.ShipperProvince),
					utils.NewNullString(row.ShipperPostcode),
					utils.NewNullString(row.ShipperCountryCode),
					utils.NewNullString(row.ShipperEmail),
					utils.NewNullString(row.ShipperPhoneNumber),
					utils.NewNullString(row.TariffCode),
					utils.NewNullString(row.TariffSequence),
					utils.NewNullString(row.StatisticalCode),
					utils.NewNullString(row.EnglishDescriptionOfGood),
					utils.NewNullString(row.ThaiDescriptionOfGood),
					row.Quantity,
					utils.NewNullString(row.QuantityUnitCode),
					row.NetWeight,
					utils.NewNullString(row.NetWeightUnitCode),
					row.GrossWeight,
					utils.NewNullString(row.GrossWeightUnitCode),
					utils.NewNullString(row.Package),
					utils.NewNullString(row.PackageUnitCode),
					row.CifValueForeign,
					row.FobValueForeign,
					row.ExchangeRate,
					utils.NewNullString(row.CurrencyCode),
					utils.NewNullString(row.ShippingMark),
					utils.NewNullString(row.ConsignmentCountry),
					row.FreightValueForeign,
					utils.NewNullString(row.FreightCurrencyCode),
					row.InsuranceValueForeign,
					utils.NewNullString(row.InsuranceCurrencyCode),
					utils.NewNullString(row.OtherChargeValueForeign),
					utils.NewNullString(row.OtherChargeCurrencyCode),
					utils.NewNullString(row.InvoiceNo),
					utils.NewNullString(row.InvoiceDate),
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

func (r repository) GetShipperBrands(ctx context.Context) ([]*GetShipperBrandModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	sqlStr := `
		SELECT
			sb.name AS shipper_name,
			sb.address AS shipper_address,
			sb.district AS shipper_district,
			sb.sub_district AS shipper_subprovince,
			sb.province AS shipper_province,
			sb.postal_code AS shipper_postcode,
			sb.country_code AS shipper_country_code
		FROM ship2cu.master_shipper_brands sb 
	`

	var list []*GetShipperBrandModel
	_, err := db.QueryContext(ctx, &list, sqlStr)

	if err != nil {
		return list, err
	}

	return list, nil

}

func (r repository) GetMasterHsCode(ctx context.Context) ([]*GetMasterHsCodeModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	sqlStr := `
		SELECT * FROM get_hs_code_data();
	`

	var list []*GetMasterHsCodeModel
	_, err := db.QueryContext(ctx, &list, sqlStr)

	if err != nil {
		return list, err
	}

	return list, nil

}

func (r repository) GetMawb(ctx context.Context, mawb string) (*utils.GetMawb, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := utils.GetMawb{}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.UUID,
		&x.FlightNo,
		&x.Origin,
		&x.Destination,
		&x.Mawb,
		&x.LotNo,
		&x.DepartureDatetime,
		&x.ArrivalDatetime,
		&x.Origin,
	), `
			SELECT 
				maw.uuid,
				maw.flight_no ,
				maw.origin_code ,
				maw.destination_code ,
				maw.lot_no_code ,
				maw.mawb,
				maw.lot_no_code,
				TO_CHAR(maw.departure_date_time at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS') as departure_datetime,
				TO_CHAR(maw.arrival_date_time at time zone 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS') as arrival_datetime
			FROM ship2cu.company_master_airway_bill maw 
			WHERE maw.mawb = ?
			AND maw.deleted_at IS NULL
			LIMIT 1
	 `, mawb)

	if err != nil {
		return nil, err
	}
	return &x, nil
}

func (r repository) GetFreightData(ctx context.Context, uploadLogUUID, countryCode, currencyCode string) (*GetFreightDataModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := GetFreightDataModel{}
	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.FreightZone,
		&x.FreightRate,
	), `
			SELECT (
				SELECT 
					rate
				FROM public.master_inbound_express_freight_zones 
				WHERE country_code = ?1
				LIMIT 1
			) AS freight_zone,
			(
				SELECT
						CASE
								WHEN mct."type" = 'inbound' THEN import_exchange_rate / cer.ratio::numeric
								WHEN mct."type" = 'outbound' THEN export_exchange_rate / cer.ratio::numeric
								ELSE 0
						end as exchange_rate
				FROM
						ship2cu.customs_exchange_rate cer
						join public.tbl_upload_loggings tul on tul.uuid = ?0
						join public.master_convert_templates mct on mct.code  = tul.template_code
				WHERE
						cer.currency_code = ?2
				LIMIT 1
			) AS freight_rate
	`, uploadLogUUID, countryCode, currencyCode)

	if err != nil {
		return nil, err
	}

	return &x, nil
}
