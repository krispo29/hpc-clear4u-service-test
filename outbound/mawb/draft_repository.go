package outbound

import (
	"context"
	"hpc-express-service/utils"
	"time"

	"github.com/go-pg/pg/v9"
)

func (r repository) GetAllMawbDraft(ctx context.Context, start, end string) ([]*GetAllMawbDraftModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	var list []*GetAllMawbDraftModel
	_, err := db.QueryContext(ctx, &list,
		`
		SELECT
			md.uuid,
			md.mawb,
			md.hawb,
			md.shipper_name_and_address,
			md.consignee_name_and_address,
			c.name as customer_name,
			to_char(md.created_at at time zone 'utc' at time zone 'Asia/Bangkok', 'DD-MM-YYYY HH24:MI:SS') as created_at
		FROM tbl_mawb_drafts md
		left join tbl_customers c on c."uuid" = md.customer_uuid
		WHERE md.created_at::date BETWEEN SYMMETRIC ?0 AND ?1
		ORDER BY md.id DESC
	`, start, end)

	if err != nil {
		return list, err
	}

	return list, nil
}

func (r repository) GetOneMawbDraft(ctx context.Context, uuid string) (*GetMawbDraftModel, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	x := GetMawbDraftModel{}

	_, err := db.QueryOneContext(ctx, pg.Scan(
		&x.UUID,
		&x.Mawb,
		&x.Hawb,
		&x.ShipperNameAndAddress,
		&x.AwbIssuedBy,
		&x.ConsigneeNameAndAddress,
		&x.IssuingCarrierAgentName,
		&x.AccountingInfomation,
		&x.AgentsIATACode,
		&x.AccountNo,
		&x.AirportOfDeparture,
		&x.ReferenceNumber,
		&x.OptionalShippingInfo1,
		&x.OptionalShippingInfo2,
		&x.RoutingTo,
		&x.RoutingBy,
		&x.DestinationTo1,
		&x.DestinationBy1,
		&x.DestinationTo2,
		&x.DestinationBy2,
		&x.Currency,
		&x.ChgsCode,
		&x.WtValPpd,
		&x.WtValColl,
		&x.OtherPpd,
		&x.OtherColl,
		&x.DeclaredValCarriage,
		&x.DeclaredValCustoms,
		&x.AirportOfDestination,
		&x.RequestedFlightDate1,
		&x.RequestedFlightDate2,
		&x.AmountOfInsurance,
		&x.HandlingInfomation,
		&x.Sci,
		&x.TerminalChargeKey,
		&x.TerminalChargeVal,
		&x.MrKey,
		&x.MrVal,
		&x.BcKey,
		&x.BcVal,
		&x.AweFeeKey,
		&x.AweFeeVal,
		&x.Signature1,
		&x.Prepaid,
		&x.ValuationCharge,
		&x.Tax,
		&x.TotalOtherChargesDueAgent,
		&x.TotalOtherChargesDueCarrier,
		&x.TotalPrepaid,
		&x.CurrencyConversionRates,
		&x.Signature2Date,
		&x.Signature2Place,
		&x.Signature2Issuing,
		&x.CcKey,
		&x.CcVal,
		&x.CustomerUUID,
	), `
			SELECT 
				"uuid", mawb, hawb, shipper_name_and_address, awb_issued_by, consignee_name_and_address, issuing_carrier_agent_name, accounting_infomation, agents_iata_code, account_no, airport_of_departure, reference_number, optional_shipping_info1, optional_shipping_info2, routing_to, routing_by, destination_to1, destination_by1, destination_to2, destination_by2, currency, chgs_code, wt_val_ppd, wt_val_coll, other_ppd, other_coll, declared_val_carriage, declared_val_customs, airport_of_destination, requested_flight_date1, requested_flight_date2, amount_of_insurance, handling_infomation, sci, terminalcharge_key, terminalcharge_val, mr_key, mr_val, bc_key, bc_val, awe_fee_key, awe_fee_val, signature1,
				prepaid, valuation_charge, tax, total_other_charges_due_agent, total_other_charges_due_carrier, total_prepaid, currency_conversion_rates, signature2_date, signature2_place, signature2_issuing, cc_key, cc_val, customer_uuid
			FROM public.tbl_mawb_drafts
			where uuid = ?
	 `, uuid)

	if err != nil {
		return nil, err
	}

	var items []*ItemDraftDetailModel
	_, err = db.QueryContext(ctx, &items,
		`
		SELECT pieces_rcp, gross_weight, nature_and_quantity, rate_class, chargeable_weight, rate_charge, total
		FROM public.tbl_mawb_draft_details
		WHERE mawb_draft_uuid = ? and deleted_at is null
	`, uuid)

	if err != nil {
		return nil, err
	}

	x.Items = items

	return &x, nil
}

func (r repository) CreateMawbDraft(ctx context.Context, data *RequestDraftModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 60*time.Second)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	/*
		Insert Mawb Draft
	*/
	var mawbDraftUUID string
	{

		sqlStr := `
		INSERT INTO public.tbl_mawb_drafts
			(
				mawb, hawb, shipper_name_and_address, awb_issued_by, consignee_name_and_address, issuing_carrier_agent_name, accounting_infomation, agents_iata_code, account_no, airport_of_departure, reference_number, optional_shipping_info1, optional_shipping_info2, routing_to, routing_by, destination_to1, destination_by1, destination_to2, destination_by2, currency, chgs_code, wt_val_ppd, wt_val_coll, other_ppd, other_coll, declared_val_carriage, declared_val_customs, airport_of_destination, requested_flight_date1, requested_flight_date2, amount_of_insurance, handling_infomation, sci, terminalcharge_key, terminalcharge_val, mr_key, mr_val, bc_key, bc_val, awe_fee_key, awe_fee_val, signature1,
				prepaid, valuation_charge, tax, total_other_charges_due_agent, total_other_charges_due_carrier, total_prepaid, currency_conversion_rates, signature2_date, signature2_place, signature2_issuing, cc_key, cc_val, customer_uuid
			)
		VALUES
			(
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			)
		RETURNING uuid
	`
		sqlStr = utils.ReplaceSQL(sqlStr, "?")
		stmt, err := tx.Prepare(sqlStr)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		_, err = stmt.QueryOneContext(ctx, &mawbDraftUUID,
			utils.NewNullString(data.Mawb),
			utils.NewNullString(data.Hawb),
			utils.NewNullString(data.ShipperNameAndAddress),
			utils.NewNullString(data.AwbIssuedBy),
			utils.NewNullString(data.ConsigneeNameAndAddress),
			utils.NewNullString(data.IssuingCarrierAgentName),
			utils.NewNullString(data.AccountingInfomation),
			utils.NewNullString(data.AgentsIATACode),
			utils.NewNullString(data.AccountNo),
			utils.NewNullString(data.AirportOfDeparture),
			utils.NewNullString(data.ReferenceNumber),
			utils.NewNullString(data.OptionalShippingInfo1),
			utils.NewNullString(data.OptionalShippingInfo2),
			utils.NewNullString(data.RoutingTo),
			utils.NewNullString(data.RoutingBy),
			utils.NewNullString(data.DestinationTo1),
			utils.NewNullString(data.DestinationBy1),
			utils.NewNullString(data.DestinationTo2),
			utils.NewNullString(data.DestinationBy2),
			utils.NewNullString(data.Currency),
			utils.NewNullString(data.ChgsCode),
			utils.NewNullString(data.WtValPpd),
			utils.NewNullString(data.WtValColl),
			utils.NewNullString(data.OtherPpd),
			utils.NewNullString(data.OtherColl),
			utils.NewNullString(data.DeclaredValCarriage),
			utils.NewNullString(data.DeclaredValCustoms),
			utils.NewNullString(data.AirportOfDestination),
			utils.NewNullString(data.RequestedFlightDate1),
			utils.NewNullString(data.RequestedFlightDate2),
			utils.NewNullString(data.AmountOfInsurance),
			utils.NewNullString(data.HandlingInfomation),
			utils.NewNullString(data.Sci),
			utils.NewNullString(data.TerminalChargeKey),
			utils.NewNullString(data.TerminalChargeVal),
			utils.NewNullString(data.MrKey),
			utils.NewNullString(data.MrVal),
			utils.NewNullString(data.BcKey),
			utils.NewNullString(data.BcVal),
			utils.NewNullString(data.AweFeeKey),
			utils.NewNullString(data.AweFeeVal),
			utils.NewNullString(data.Signature1),
			utils.NewNullString(data.Prepaid),
			utils.NewNullString(data.ValuationCharge),
			utils.NewNullString(data.Tax),
			utils.NewNullString(data.TotalOtherChargesDueAgent),
			utils.NewNullString(data.TotalOtherChargesDueCarrier),
			utils.NewNullString(data.TotalPrepaid),
			utils.NewNullString(data.CurrencyConversionRates),
			utils.NewNullString(data.Signature2Date),
			utils.NewNullString(data.Signature2Place),
			utils.NewNullString(data.Signature2Issuing),
			utils.NewNullString(data.CcKey),
			utils.NewNullString(data.CcVal),
			data.CustomerUUID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}

	}
	/*
		Insert Mawb Draft
	*/

	/*
		Insert Mawb Detail
	*/
	{

		// Insert Raw Data
		sqlStr := `INSERT INTO public.tbl_mawb_draft_details
				(
					mawb_draft_uuid, pieces_rcp, gross_weight, nature_and_quantity, rate_class, chargeable_weight, rate_charge, total
				) VALUES `
		vals := []interface{}{}
		for _, row := range data.Items {
			sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?),"
			vals = append(vals,
				utils.NewNullString(mawbDraftUUID),
				utils.NewNullString(row.PiecesRCP),
				utils.NewNullString(row.GrossWeight),
				utils.NewNullString(row.NatureAndQuantity),
				utils.NewNullString(row.RateClass),
				utils.NewNullString(row.ChargeableWeight),
				utils.NewNullString(row.RateCharge),
				utils.NewNullString(row.Total),
			)
		}

		// remove last comma,
		sqlStr = sqlStr[0 : len(sqlStr)-1]

		// Convert symbol ? to $
		sqlStr = utils.ReplaceSQL(sqlStr, "?")

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
	/*
		Insert Mawb Detail
	*/

	tx.Commit()

	return nil
}

func (r repository) UpdateMawbDraft(ctx context.Context, data *RequestUpdateMawbDraftModel) error {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	/*
		1. Start Delete Details
	*/
	{
		// DELETE FROM public.mawb_draft_details WHERE mawb_draft_uuid = $1
		sqlStr := (`UPDATE public.tbl_mawb_draft_details SET deleted_at = now() WHERE mawb_draft_uuid = $1`)
		// Prepare statement
		stmt, err := tx.Prepare(sqlStr)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx, data.UUID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	/*
		1. End Delete Details
	*/

	/*
		2. Start Update Mawb Draft
	*/
	{
		sqlStr := `
			UPDATE public.tbl_mawb_drafts
			SET 
				mawb=?,
				hawb=?,
				shipper_name_and_address=?,
				awb_issued_by=?,
				consignee_name_and_address=?,
				issuing_carrier_agent_name=?,
				accounting_infomation=?,
				agents_iata_code=?,
				account_no=?,
				airport_of_departure=?,
				reference_number=?,
				optional_shipping_info1=?,
				optional_shipping_info2=?,
				routing_to=?,
				routing_by=?,
				destination_to1=?,
				destination_by1=?,
				destination_to2=?,
				destination_by2=?,
				currency=?,
				chgs_code=?,
				wt_val_ppd=?,
				wt_val_coll=?,
				other_ppd=?,
				other_coll=?,
				declared_val_carriage=?,
				declared_val_customs=?,
				airport_of_destination=?,
				requested_flight_date1=?,
				requested_flight_date2=?,
				amount_of_insurance=?,
				handling_infomation=?,
				sci=?,
				terminalcharge_key=?,
				terminalcharge_val=?,
				mr_key=?,
				mr_val=?,
				bc_key=?,
				bc_val=?,
				awe_fee_key=?,
				awe_fee_val=?,
				signature1=?,
				prepaid=?,
				valuation_charge=?,
				tax=?,
				total_other_charges_due_agent=?,
				total_other_charges_due_carrier=?,
				total_prepaid=?,
				currency_conversion_rates=?,
				signature2_date=?,
				signature2_place=?,
				signature2_issuing=?,
				cc_key=?,
				cc_val=?,
				customer_uuid=?
			WHERE uuid=?
		`

		sqlStr = utils.ReplaceSQL(sqlStr, "?")
		stmt, err := tx.Prepare(sqlStr)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		_, err = stmt.ExecOneContext(ctx,
			utils.NewNullString(data.Mawb),
			utils.NewNullString(data.Hawb),
			utils.NewNullString(data.ShipperNameAndAddress),
			utils.NewNullString(data.AwbIssuedBy),
			utils.NewNullString(data.ConsigneeNameAndAddress),
			utils.NewNullString(data.IssuingCarrierAgentName),
			utils.NewNullString(data.AccountingInfomation),
			utils.NewNullString(data.AgentsIATACode),
			utils.NewNullString(data.AccountNo),
			utils.NewNullString(data.AirportOfDeparture),
			utils.NewNullString(data.ReferenceNumber),
			utils.NewNullString(data.OptionalShippingInfo1),
			utils.NewNullString(data.OptionalShippingInfo2),
			utils.NewNullString(data.RoutingTo),
			utils.NewNullString(data.RoutingBy),
			utils.NewNullString(data.DestinationTo1),
			utils.NewNullString(data.DestinationBy1),
			utils.NewNullString(data.DestinationTo2),
			utils.NewNullString(data.DestinationBy2),
			utils.NewNullString(data.Currency),
			utils.NewNullString(data.ChgsCode),
			utils.NewNullString(data.WtValPpd),
			utils.NewNullString(data.WtValColl),
			utils.NewNullString(data.OtherPpd),
			utils.NewNullString(data.OtherColl),
			utils.NewNullString(data.DeclaredValCarriage),
			utils.NewNullString(data.DeclaredValCustoms),
			utils.NewNullString(data.AirportOfDestination),
			utils.NewNullString(data.RequestedFlightDate1),
			utils.NewNullString(data.RequestedFlightDate2),
			utils.NewNullString(data.AmountOfInsurance),
			utils.NewNullString(data.HandlingInfomation),
			utils.NewNullString(data.Sci),
			utils.NewNullString(data.TerminalChargeKey),
			utils.NewNullString(data.TerminalChargeVal),
			utils.NewNullString(data.MrKey),
			utils.NewNullString(data.MrVal),
			utils.NewNullString(data.BcKey),
			utils.NewNullString(data.BcVal),
			utils.NewNullString(data.AweFeeKey),
			utils.NewNullString(data.AweFeeVal),
			utils.NewNullString(data.Signature1),
			utils.NewNullString(data.Prepaid),
			utils.NewNullString(data.ValuationCharge),
			utils.NewNullString(data.Tax),
			utils.NewNullString(data.TotalOtherChargesDueAgent),
			utils.NewNullString(data.TotalOtherChargesDueCarrier),
			utils.NewNullString(data.TotalPrepaid),
			utils.NewNullString(data.CurrencyConversionRates),
			utils.NewNullString(data.Signature2Date),
			utils.NewNullString(data.Signature2Place),
			utils.NewNullString(data.Signature2Issuing),
			utils.NewNullString(data.CcKey),
			utils.NewNullString(data.CcVal),
			data.CustomerUUID,
			utils.NewNullString(data.UUID),
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}
	/*
		2. Start Update Mawb Draft
	*/

	/*
		3. Start Insert Details
	*/
	{
		// Insert Raw Data
		sqlStr := `INSERT INTO public.tbl_mawb_draft_details
				(
					mawb_draft_uuid, pieces_rcp, gross_weight, nature_and_quantity, rate_class, chargeable_weight, rate_charge, total
				) VALUES `
		vals := []interface{}{}
		for _, row := range data.Items {
			sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?),"
			vals = append(vals,
				utils.NewNullString(data.UUID),
				utils.NewNullString(row.PiecesRCP),
				utils.NewNullString(row.GrossWeight),
				utils.NewNullString(row.NatureAndQuantity),
				utils.NewNullString(row.RateClass),
				utils.NewNullString(row.ChargeableWeight),
				utils.NewNullString(row.RateCharge),
				utils.NewNullString(row.Total),
			)
		}

		// remove last comma,
		sqlStr = sqlStr[0 : len(sqlStr)-1]

		// Convert symbol ? to $
		sqlStr = utils.ReplaceSQL(sqlStr, "?")

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
	/*
		3. End Insert Details
	*/

	tx.Commit()

	return nil
}
