package inbound

import (
	"context"
	"errors"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
)

type InboundExpressRepository interface {
	// GetOneMawb(ctx context.Context, headerUUID string) (*GetPreImportManifestModel, error)
	GetAllMawb(ctx context.Context) ([]*GetPreImportManifestModel, error)
	InsertPreImportManifestHeader(ctx context.Context, data *InsertPreImportHeaderManifestModel) (string, error)
	UpdatePreImportManifestHeader(ctx context.Context, data *UpdatePreImportHeaderManifestModel) error
	InsertPreImportManifestDetails(ctx context.Context, headerUUID string, details []*utils.InsertPreImportDetailManifestModel, chunkSize int) error
	GetOneMawb(ctx context.Context, headerUUID string) (*GetPreImportManifestModel, error)
	UpdatePreImportManifestDetail(ctx context.Context, headerUUID string, data []*UpdatePreImportManifestDetailModel) error
	GetSummaryByHeaderUUID(ctx context.Context, headerUUID string) ([]*GetSummaryModel, error)
}

type repository struct {
	contextTimeout time.Duration
}

func NewInboundExpressRepository(
	timeout time.Duration,
) InboundExpressRepository {
	return &repository{
		contextTimeout: timeout,
	}
}

// func (r repository) GetOneMawb(ctx context.Context, headerUUID string) (*GetPreImportManifestModel, error) {
// 	db := ctx.Value("postgreSQLConn").(*pg.DB)
// 	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

// 	result := &GetPreImportManifestModel{}
// 	_, err := db.QueryOneContext(ctx, pg.Scan(
// 		&result.UUID,
// 		&result.Mawb,
// 		&result.UploadLoggingUUID,
// 		&result.DischargePort,
// 		&result.VasselName,
// 		&result.ArrivalDate,
// 		&result.CustomerName,
// 		&result.FlightNo,
// 		&result.OriginCountryCode,
// 		&result.OriginCurrencyCode,
// 		&result.IsEnableCustomsOT,
// 		&result.CreatedAt,
// 		&result.UpdatedAt,
// 	), `
// 			SELECT
// 				mh."uuid",
// 				mh.mawb,
// 				mh.upload_logging_uuid,
// 				mh.discharge_port,
// 				mh.vassel_name,
// 				mh.arrival_date,
// 				mh.customer_name,
// 				mh.flight_no,
// 				mh.origin_country_code,
// 				mh.origin_currency_code,
// 				mh.is_enable_customs_ot,
// 				mh.created_at,
// 				mh.updated_at
// 			FROM public.tbl_pre_import_manifest_headers mh
// 			WHERE mh.uuid = ?
// 	 `, headerUUID)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil

// }

func (r repository) GetAllMawb(ctx context.Context) ([]*GetPreImportManifestModel, error) {

	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	sqlStr := `
			SELECT
				mh."uuid",
				mh.mawb,
				mh.upload_logging_uuid,
				mh.discharge_port,
				mh.vassel_name,
				mh.arrival_date,
				mh.customer_name,
				mh.flight_no,
				mh.origin_country_code,
				mh.origin_currency_code,
				mh.is_enable_customs_ot,
				mh.created_at,
				mh.updated_at
			FROM public.tbl_pre_import_manifest_headers mh
			WHERE mh.deleted_at IS NULL
			ORDER BY mh.created_at DESC
	`

	var list []*GetPreImportManifestModel
	_, err := db.QueryContext(ctx, &list, sqlStr)

	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) InsertPreImportManifestHeader(ctx context.Context, data *InsertPreImportHeaderManifestModel) (string, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Insert HeaderManifest
	var headerUUID string
	{
		sqlStr :=
			`
			INSERT INTO public.tbl_pre_import_manifest_headers
				(
					mawb, discharge_port, vassel_name, arrival_date, customer_name, flight_no,  origin_country_code, origin_currency_code, is_enable_customs_ot
				)
			VALUES
				(
					?, ?, ?, ?, ?, ?, ?, ?, ?
				)
			RETURNING uuid
		`

		// Prepare statement
		stmt, err := tx.Prepare(utils.PrepareSQL(sqlStr))
		if err != nil {
			tx.Rollback()
			return "", err
		}
		defer stmt.Close()

		values := []interface{}{}
		values = append(
			values,
			data.Mawb,
			data.DischargePort,
			data.VasselName,
			data.ArrivalDate,
			data.CustomerName,
			data.FlightNo,
			data.OriginCountryCode,
			data.OriginCurrencyCode,
			data.IsEnableCustomsOT,
		)

		_, err = stmt.QueryOneContext(ctx, &headerUUID, values...)

		if err != nil {
			tx.Rollback()
			return "", err
		}
	}

	tx.Commit()

	return headerUUID, nil
}

func (r repository) UpdatePreImportManifestHeader(ctx context.Context, data *UpdatePreImportHeaderManifestModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err := db.ExecOneContext(ctx,
		`
			UPDATE public.tbl_pre_import_manifest_headers
				SET
					mawb=?1,
					discharge_port=?2,
					vassel_name=?3,
					arrival_date=?4,
					customer_name=?5,
					flight_no=?6,
					origin_country_code=?7,
					origin_currency_code=?8,
					is_enable_customs_ot=?9,
					updated_at = NOW()
			WHERE "uuid" = ?0;
		`,
		data.UUID,
		utils.NewNullString(data.Mawb),
		utils.NewNullString(data.DischargePort),
		utils.NewNullString(data.VasselName),
		utils.NewNullString(data.ArrivalDate),
		utils.NewNullString(data.CustomerName),
		utils.NewNullString(data.FlightNo),
		utils.NewNullString(data.OriginCountryCode),
		utils.NewNullString(data.OriginCurrencyCode),
		data.IsEnableCustomsOT,
	)

	if err != nil {
		if err == pg.ErrNoRows {
			return errors.New("not found")
		}
		return err
	}

	return nil
}

func (r repository) InsertPreImportManifestDetails(ctx context.Context, headerUUID string, details []*utils.InsertPreImportDetailManifestModel, chunkSize int) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert Manifest
	// Chunk slice
	chunked := utils.ChunkSlice(details, chunkSize)
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

func (r repository) GetOneMawb(ctx context.Context, headerUUID string) (*GetPreImportManifestModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	result := &GetPreImportManifestModel{}
	_, err := db.QueryOneContext(ctx, pg.Scan(
		&result.UUID,
		&result.Mawb,
		&result.UploadLoggingUUID,
		&result.DischargePort,
		&result.VasselName,
		&result.ArrivalDate,
		&result.CustomerName,
		&result.FlightNo,
		&result.OriginCountryCode,
		&result.OriginCurrencyCode,
		&result.CreatedAt,
		&result.UpdatedAt,
	), `
			SELECT
				mh."uuid",
				mh.mawb,
				mh.upload_logging_uuid,
				mh.discharge_port,
				mh.vassel_name,
				mh.arrival_date,
				mh.customer_name,
				mh.flight_no,
				mh.origin_country_code,
				mh.origin_currency_code,
				mh.created_at,
				mh.updated_at
			FROM public.tbl_pre_import_manifest_headers mh
			WHERE mh.uuid = ?
	 `, headerUUID)

	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(`
		SELECT DISTINCT
			tpimd.uuid,
			tpimd.master_air_waybill,
			tpimd.house_air_waybill,
			tpimd.category,
			tpimd.consignee_tax,
			tpimd.consignee_branch,
			tpimd.consignee_name,
			tpimd.consignee_address,
			tpimd.consignee_district,
			tpimd.consignee_subprovince,
			tpimd.consignee_province,
			tpimd.consignee_postcode,
			tpimd.consignee_country_code,
			tpimd.consignee_email,
			tpimd.consignee_phone_number,
			tpimd.shipper_name,
			tpimd.shipper_address,
			tpimd.shipper_district,
			tpimd.shipper_subprovince,
			tpimd.shipper_province,
			tpimd.shipper_postcode,
			tpimd.shipper_country_code,
			tpimd.shipper_email,
			tpimd.shipper_phone_number,
			tpimd.tariff_code,
			tpimd.tariff_sequence,
			tpimd.statistical_code,
			tpimd.english_description_of_good,
			tpimd.thai_description_of_good,
			tpimd.quantity,
			tpimd.quantity_unit_code,
			tpimd.net_weight,
			tpimd.net_weight_unit_code,
			tpimd.gross_weight,
			tpimd.gross_weight_unit_code,
			tpimd.package,
			tpimd.package_unit_code,
			tpimd.cif_value_foreign,
			tpimd.fob_value_foreign,
			tpimd.exchange_rate,
			tpimd.currency_code,
			tpimd.shipping_mark,
			tpimd.consignment_country,
			tpimd.freight_value_foreign,
			tpimd.freight_currency_code,
			tpimd.insurance_value_foreign,
			tpimd.insurance_currency_code,
			tpimd.other_charge_value_foreign,
			tpimd.other_charge_currency_code,
			tpimd.invoice_no,
			tpimd.invoice_date,
			tpimd.created_at,
			tpimd.updated_at,
				tpimd.cif_value_foreign*0.07 as vat,
				tpimd.cif_value_foreign*(mhcv.duty_rate/100) as duty,
				CASE
						WHEN mhcv.duty_rate is NULL THEN FALSE
						ELSE TRUE
				end as is_goods_matched
		FROM public.tbl_pre_import_manifest_details tpimd
		LEFT JOIN LATERAL (
		    SELECT duty_rate
		    FROM master_hs_code_v2
		    WHERE TRIM(UPPER(goods_en)) = TRIM(UPPER(tpimd.english_description_of_good))
		    LIMIT 1
		) mhcv ON true
		WHERE tpimd.header_uuid = $1
		ORDER BY tpimd.category ASC
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.QueryContext(ctx, &result.Details, result.UUID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r repository) UpdatePreImportManifestDetail(ctx context.Context, headerUUID string, data []*UpdatePreImportManifestDetailModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	sqlStr := `
	UPDATE public.tbl_pre_import_manifest_details as t 
	SET
		master_air_waybill = c.master_air_waybill::text,
		house_air_waybill = c.house_air_waybill::text,
		category = c.category::text,
		consignee_tax = c.consignee_tax::text,
		consignee_branch = c.consignee_branch::text,
		consignee_name = c.consignee_name::text,
		consignee_address = c.consignee_address::text,
		consignee_district = c.consignee_district::text,
		consignee_subprovince = c.consignee_subprovince::text,
		consignee_province = c.consignee_province::text,
		consignee_postcode = c.consignee_postcode::text,
		consignee_country_code = c.consignee_country_code::text,
		consignee_email = c.consignee_email::text,
		consignee_phone_number = c.consignee_phone_number::text,
		shipper_name = c.shipper_name::text,
		shipper_address = c.shipper_address::text,
		shipper_district = c.shipper_district::text,
		shipper_subprovince = c.shipper_subprovince::text,
		shipper_province = c.shipper_province::text,
		shipper_postcode = c.shipper_postcode::text,
		shipper_country_code = c.shipper_country_code::text,
		shipper_email = c.shipper_email::text,
		shipper_phone_number = c.shipper_phone_number::text,
		tariff_code = c.tariff_code::text,
		tariff_sequence = c.tariff_sequence::text,
		statistical_code = c.statistical_code::text,
		english_description_of_good = c.english_description_of_good::text,
		thai_description_of_good = c.thai_description_of_good::text,
		quantity = c.quantity::integer,
		quantity_unit_code = c.quantity_unit_code::text,
		net_weight = c.net_weight::float,
		net_weight_unit_code = c.net_weight_unit_code::text,
		gross_weight = c.gross_weight::float,
		gross_weight_unit_code = c.gross_weight_unit_code::text,
		package = c.package::text,
		package_unit_code = c.package_unit_code::text,
		cif_value_foreign = c.cif_value_foreign::float,
		fob_value_foreign = c.fob_value_foreign::float,
		exchange_rate = c.exchange_rate::float,
		currency_code = c.currency_code::text,
		shipping_mark = c.shipping_mark::text,
		consignment_country = c.consignment_country::text,
		freight_value_foreign = c.freight_value_foreign::float,
		freight_currency_code = c.freight_currency_code::text,
		insurance_value_foreign = c.insurance_value_foreign::float,
		insurance_currency_code = c.insurance_currency_code::text,
		other_charge_value_foreign = c.other_charge_value_foreign::text,
		other_charge_currency_code = c.other_charge_currency_code::text,
		invoice_no = c.invoice_no::text,
		invoice_date = c.invoice_date::text,
		updated_at=now()
	from (values
	`
	vals := []interface{}{}
	for _, row := range data {
		sqlStr += "( ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text, ?::integer, ?::text, ?::float, ?::text, ?::float, ?::text, ?::text, ?::text, ?::float, ?::float, ?::float, ?::text, ?::text, ?::text, ?::float, ?::text, ?::float, ?::text, ?::text, ?::text, ?::text, ?::text, ?::text),"
		vals = append(vals, row.MasterAirWaybill,
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
			utils.NewNullString(row.UUID),
			headerUUID,
		)
	}

	// lot_no mawb bag_no hawb menifest actual_weight

	// remove last comma,
	sqlStr = sqlStr[0 : len(sqlStr)-1]

	// Convert symbol ? to $
	// sqlStr = utils.ReplaceSQL(sqlStr, "?")

	// Concat sql
	sqlStr += `) as c(
		master_air_waybill, house_air_waybill, category, consignee_tax, consignee_branch, consignee_name, consignee_address, consignee_district, consignee_subprovince, consignee_province, consignee_postcode, consignee_country_code, consignee_email, consignee_phone_number, shipper_name, shipper_address, shipper_district, shipper_subprovince, shipper_province, shipper_postcode, shipper_country_code, shipper_email, shipper_phone_number, tariff_code, tariff_sequence, statistical_code, english_description_of_good, thai_description_of_good, quantity, quantity_unit_code, net_weight, net_weight_unit_code, gross_weight, gross_weight_unit_code, package, package_unit_code, cif_value_foreign, fob_value_foreign, exchange_rate, currency_code, shipping_mark, consignment_country, freight_value_foreign, freight_currency_code, insurance_value_foreign, insurance_currency_code, other_charge_value_foreign, other_charge_currency_code, invoice_no, invoice_date, uuid, header_uuid
	)
	WHERE c.uuid::text = t.uuid::text
	AND c.header_uuid = t.header_uuid
	`

	// Prepare statement
	// stmt, err := db.Prepare(sqlStr)
	// if err != nil {
	// 	log.Println("#3.1")
	// 	return err
	// }
	// defer stmt.Close()

	// log.Println(sqlStr)

	_, err := db.ExecContext(ctx, sqlStr, vals...)
	if err != nil {
		return err
	}

	return nil
}

func (r repository) GetSummaryByHeaderUUID(ctx context.Context, headerUUID string) ([]*GetSummaryModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	var list []*GetSummaryModel
	_, err := db.QueryContext(ctx, &list,
		`
		SELECT distinct
			tpimd.house_air_waybill as hawb,
			tpimd.category,
			tpimd.cif_value_foreign*0.07 as vat,
			tpimd.cif_value_foreign*(mhcv.duty_rate/100) as duty
		FROM public.tbl_pre_import_manifest_headers mh
		left join tbl_pre_import_manifest_details tpimd on tpimd.header_uuid = mh."uuid"
		LEFT JOIN LATERAL (
		    SELECT duty_rate
		    FROM master_hs_code_v2
		    WHERE TRIM(UPPER(goods_en)) = TRIM(UPPER(tpimd.english_description_of_good))
		    LIMIT 1
		) mhcv ON true
		where mh.uuid = ?0
	`, headerUUID)

	if err != nil {
		return list, err
	}

	return list, nil
}
