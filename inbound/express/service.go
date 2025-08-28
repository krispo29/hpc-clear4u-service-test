package inbound

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"hpc-express-service/ship2cu"
	"hpc-express-service/uploadlog"
	"hpc-express-service/utils"
)

type InboundExpressService interface {
	GetAllMawb(ctx context.Context) ([]*GetPreImportManifestModel, error)
	// GetMawb(ctx context.Context, uuid string)
	InsertPreImportManifestHeader(ctx context.Context, data *InsertPreImportHeaderManifestModel) (string, error)
	UpdatePreImportManifestHeader(ctx context.Context, data *UpdatePreImportHeaderManifestModel) error
	UploadManifestDetails(ctx context.Context, userUUID, headerUUID, originName, templateCode string, fileBytes []byte) error
	DownloadPreImport(ctx context.Context, headerUUID string) (string, *bytes.Buffer, error)
	DownloadRawPreImport(ctx context.Context, headerUUID string) (string, *bytes.Buffer, error)
	UploadUpdateRawPreImport(ctx context.Context, userUUID, headerUUID, originName string, fileBytes []byte) error
	GetOneByHeaderUUID(ctx context.Context, headerUUID string) (*GetPreImportManifestModel, error)
	GetSummaryByHeaderUUID(ctx context.Context, headerUUID string) (*UploadSummaryModel, error)
}

type service struct {
	selfRepo       InboundExpressRepository
	contextTimeout time.Duration
	ship2cuSvc     ship2cu.Service
	uploadlogSvc   uploadlog.Service
	ship2cuRepo    ship2cu.Repository
}

func NewInboundExpressService(
	selfRepo InboundExpressRepository,
	timeout time.Duration,
	ship2cuSvc ship2cu.Service,
	uploadlogSvc uploadlog.Service,
	ship2cuRepo ship2cu.Repository,
) InboundExpressService {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
		ship2cuSvc:     ship2cuSvc,
		uploadlogSvc:   uploadlogSvc,
		ship2cuRepo:    ship2cuRepo,
	}
}

func (s *service) GetAllMawb(ctx context.Context) ([]*GetPreImportManifestModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetAllMawb(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) InsertPreImportManifestHeader(ctx context.Context, data *InsertPreImportHeaderManifestModel) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	uuid, err := s.selfRepo.InsertPreImportManifestHeader(ctx, data)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (s *service) UpdatePreImportManifestHeader(ctx context.Context, data *UpdatePreImportHeaderManifestModel) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	err := s.selfRepo.UpdatePreImportManifestHeader(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UploadManifestDetails(ctx context.Context, userUUID, headerUUID, originName, templateCode string, fileBytes []byte) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if templateCode == "SHIP2CU" {
		// Logging and Upload to GCS
		uploadLogUUID, err := s.uploadlogSvc.UploadLogFile(ctx, &uploadlog.UploadFileModel{
			// Mawb:         "", // TODO:
			UserUUID:     userUUID,
			FileName:     originName,
			TemplateCode: templateCode,
			Category:     "inbound",
			SubCategory:  "upload_manifest",
			FileBytes:    fileBytes,
			// Amount:       0,
		})
		if err != nil {
			return err
		}

		// Insert Manifest
		details, err := s.ship2cuSvc.ConvertToPreImportDetails(ctx, uploadLogUUID, fileBytes)
		if err != nil {
			s.uploadlogSvc.Update(ctx, &uploadlog.UpdateModel{
				UUID:   uploadLogUUID,
				Mawb:   "",
				Amount: 0,
				Status: "failed",
				Remark: err.Error(),
			})
			return err
		} else {

			err := s.selfRepo.InsertPreImportManifestDetails(ctx, headerUUID, details, 200)
			if err != nil {
				return err
			}

			s.uploadlogSvc.Update(ctx, &uploadlog.UpdateModel{
				UUID: uploadLogUUID,
				// Mawb:   resultUpload.Mawb,
				Amount: int64(len(details)),
				Status: "success",
			})
		}
		// }
	} else {
		return errors.New("not found template")
	}

	return nil
}

func (s *service) DownloadPreImport(ctx context.Context, headerUUID string) (string, *bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	/*
		Prepare Data
	*/
	// allData := []*GetDataManifestImport{}
	// if uploadLogData.TemplateCode == "TOPGLS" {
	preImportData, err := s.selfRepo.GetOneMawb(ctx, headerUUID)
	if err != nil {
		return "", nil, err
	}
	// } else {
	// 	return "", nil, errors.New("invalid template")
	// }

	// mawbInfo, err := s.selfRepo.GetMawbByTimstamp(ctx, uploadLogData.Mawb)
	// if err != nil {
	// 	log.Println("#2 ", err)
	// 	// return nil, err
	// 	mawbInfo = &utils.GetMawb{}
	// }

	// Create in-memory ZIP buffer
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	// Split into chunks of 50
	chunks := utils.ChunkSlice(preImportData.Details, 250)
	// Print the result
	for i, chunk := range chunks {

		f := excelize.NewFile()
		sheetName := "Output"
		f.SetSheetName("Sheet1", sheetName)

		for k, h := range []string{"//1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "50", "51"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 1), h)
		}

		for k, h := range []string{"//Row header (=1)", "Discharge port", "Vessel name", "Arrival date", "Customer Name"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 2), h)
		}

		for k, h := range []string{"//Row header (=2)", "Master Air Waybill", "House Air Waybill", "Category", "Consignee Tax", "Consignee Branch", "Consignee Name", "Consignee Address", "Consignee District", "Consignee Subprovince", "Consignee Province", "Consignee Postcode", "Consignee Country Code", "Consignee E-mail", "Consignee Phone Number", "Shipper Name", "Shipper Address", "Shipper District", "Shipper Subprovince", "Shipper Province", "Shipper Postcode", "Shipper Country Code", "Shipper E-mail", "Shipper Phone Number", "Tariff Code", "Tariff Sequence", "Statistical Code", "English Description of Good", "Thai Description of Good", "Quantity", "Quantity Unit Code", "Net Weight", "Net Weight Unit Code", "Gross Weight ", "Gross Weight Unit Code", "Package", "Package Unit Code", "CIF Value Foreign", "FOB Value Foreign", "Exchange Rate", "Currency Code", "Shipping Mark", "Consignment Country", "Freight Value Foreign", "Freight Currency Code", "Insurance Value Foreign", "Insurance Currency Code", "Other Charge Value Foreign", "Other Charge Currency Code", "Invoice No", "Invoice Date"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 3), h)
		}

		for k, h := range []string{"//Row header (=3)", "Invoice No", "Invoice Date"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 4), h)
		}

		startRow := 6

		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", startRow-1), 1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", startRow-1), preImportData.DischargePort)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", startRow-1), preImportData.VasselName)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "D", startRow-1), preImportData.ArrivalDate)
		for i := 0; i < len(chunk)*2; i = i + 2 {

			// for kk, v := range data {
			v := chunk[i/2]

			rowNum := (i + startRow)

			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", rowNum), 2)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", rowNum), v.MasterAirWaybill)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", rowNum), v.HouseAirWaybill)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "D", rowNum), v.Category)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "E", rowNum), v.ConsigneeTax)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "F", rowNum), v.ConsigneeBranch)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "G", rowNum), v.ConsigneeName)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "H", rowNum), v.ConsigneeAddress)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "I", rowNum), v.ConsigneeDistrict)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "J", rowNum), v.ConsigneeSubprovince)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "K", rowNum), v.ConsigneeProvince)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "L", rowNum), v.ConsigneePostcode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "M", rowNum), v.ConsigneeCountryCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "N", rowNum), v.ConsigneeEmail)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "O", rowNum), v.ConsigneePhoneNumber)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "P", rowNum), v.ShipperName)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Q", rowNum), utils.IsEmpty(v.ShipperAddress))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "R", rowNum), utils.IsEmpty(v.ShipperDistrict))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "S", rowNum), utils.IsEmpty(v.ShipperSubprovince))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "T", rowNum), utils.IsEmpty(v.ShipperProvince))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "U", rowNum), utils.IsEmpty(v.ShipperPostcode))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "V", rowNum), v.ShipperCountryCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "W", rowNum), v.ShipperEmail)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "X", rowNum), v.ShipperPhoneNumber)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Y", rowNum), utils.IsEmpty(v.TariffCode))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Z", rowNum), utils.IsEmpty(v.TariffSequence))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AA", rowNum), utils.IsEmpty(v.StatisticalCode))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AB", rowNum), v.EnglishDescriptionOfGood)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AC", rowNum), "") // v.THDescriptionGoods
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AD", rowNum), v.Quantity)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AE", rowNum), v.QuantityUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AF", rowNum), fmt.Sprintf("%.2f", v.NetWeight))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AG", rowNum), v.NetWeightUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AH", rowNum), fmt.Sprintf("%.2f", v.GrossWeight))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AI", rowNum), v.GrossWeightUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AJ", rowNum), v.Package)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AK", rowNum), v.PackageUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AL", rowNum), fmt.Sprintf("%.2f", v.CifValueForeign))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AM", rowNum), v.FobValueForeign)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AN", rowNum), v.ExchangeRate)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AO", rowNum), v.CurrencyCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AP", rowNum), v.HouseAirWaybill)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AQ", rowNum), v.ConsignmentCountry)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AR", rowNum), fmt.Sprintf("%.2f", v.FreightValueForeign))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AS", rowNum), v.FreightCurrencyCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AT", rowNum), fmt.Sprintf("%.2f", v.InsuranceValueForeign))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AU", rowNum), v.InsuranceCurrencyCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AV", rowNum), v.OtherChargeValueForeign)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AW", rowNum), v.OtherChargeCurrencyCode)

			layout := "2006-01-02 15:04:05"
			t, _ := time.Parse(layout, preImportData.ArrivalDate)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", rowNum+1), 3)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", rowNum+1), v.HouseAirWaybill)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", rowNum+1), t.Format("20060102"))

		}
		// Save Excel to a buffer
		fileName := fmt.Sprintf("pre_import_%v_%v.xlsx", preImportData.Mawb, strconv.Itoa(i+1))

		var excelBuf bytes.Buffer
		if err := f.Write(&excelBuf); err != nil {
			return "", nil, err
		}

		// Set header with compression
		header := &zip.FileHeader{
			Name:   fileName,
			Method: zip.Deflate, // 🔥 Use compression
		}
		zipFileWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create ZIP entry: %w", err)
		}

		if _, err := io.Copy(zipFileWriter, &excelBuf); err != nil {
			return "", nil, err
		}

	}

	/*
		Create Excel
	*/

	// Close the zip writer to finalize the zip structure
	if err := zipWriter.Close(); err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("pre_import_%v", preImportData.Mawb), &zipBuf, nil
}

func (s *service) DownloadRawPreImport(ctx context.Context, headerUUID string) (string, *bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()
	/*
		Prepare Data
	*/
	preImportData, err := s.selfRepo.GetOneMawb(ctx, headerUUID)
	if err != nil {
		return "", nil, err
	}

	f := excelize.NewFile()
	sheetName := "Output"
	f.SetSheetName("Sheet1", sheetName)

	for k, h := range []string{"//1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "50", "51", "52"} {
		colName, _ := excelize.ColumnNumberToName(k + 1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 1), h)
	}

	for k, h := range []string{"//Row header (=2)", "Master Air Waybill", "House Air Waybill", "Category", "Consignee Tax", "Consignee Branch", "Consignee Name", "Consignee Address", "Consignee District", "Consignee Subprovince", "Consignee Province", "Consignee Postcode", "Consignee Country Code", "Consignee E-mail", "Consignee Phone Number", "Shipper Name", "Shipper Address", "Shipper District", "Shipper Subprovince", "Shipper Province", "Shipper Postcode", "Shipper Country Code", "Shipper E-mail", "Shipper Phone Number", "Tariff Code", "Tariff Sequence", "Statistical Code", "English Description of Good", "Thai Description of Good", "Quantity", "Quantity Unit Code", "Net Weight", "Net Weight Unit Code", "Gross Weight ", "Gross Weight Unit Code", "Package", "Package Unit Code", "CIF Value Foreign", "FOB Value Foreign", "Exchange Rate", "Currency Code", "Shipping Mark", "Consignment Country", "Freight Value Foreign", "Freight Currency Code", "Insurance Value Foreign", "Insurance Currency Code", "Other Charge Value Foreign", "Other Charge Currency Code", "Invoice No", "Invoice Date", "Unique Key"} {
		colName, _ := excelize.ColumnNumberToName(k + 1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 2), h)
	}

	for i, row := range preImportData.Details {
		startRow := 3

		rowNum := (i + startRow)

		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", rowNum), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", rowNum), row.MasterAirWaybill)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", rowNum), row.HouseAirWaybill)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "D", rowNum), row.Category)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "E", rowNum), row.ConsigneeTax)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "F", rowNum), row.ConsigneeBranch)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "G", rowNum), row.ConsigneeName)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "H", rowNum), row.ConsigneeAddress)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "I", rowNum), row.ConsigneeDistrict)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "J", rowNum), row.ConsigneeSubprovince)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "K", rowNum), row.ConsigneeProvince)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "L", rowNum), row.ConsigneePostcode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "M", rowNum), row.ConsigneeCountryCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "N", rowNum), row.ConsigneeEmail)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "O", rowNum), row.ConsigneePhoneNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "P", rowNum), row.ShipperName)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Q", rowNum), utils.IsEmpty(row.ShipperAddress))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "R", rowNum), utils.IsEmpty(row.ShipperDistrict))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "S", rowNum), utils.IsEmpty(row.ShipperSubprovince))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "T", rowNum), utils.IsEmpty(row.ShipperProvince))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "U", rowNum), utils.IsEmpty(row.ShipperPostcode))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "V", rowNum), row.ShipperCountryCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "W", rowNum), row.ShipperEmail)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "X", rowNum), row.ShipperPhoneNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Y", rowNum), utils.IsEmpty(row.TariffCode))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Z", rowNum), utils.IsEmpty(row.TariffSequence))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AA", rowNum), utils.IsEmpty(row.StatisticalCode))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AB", rowNum), row.EnglishDescriptionOfGood)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AC", rowNum), "") // row.THDescriptionGoods
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AD", rowNum), row.Quantity)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AE", rowNum), row.QuantityUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AF", rowNum), fmt.Sprintf("%.2f", row.NetWeight))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AG", rowNum), row.NetWeightUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AH", rowNum), fmt.Sprintf("%.2f", row.GrossWeight))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AI", rowNum), row.GrossWeightUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AJ", rowNum), row.Package)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AK", rowNum), row.PackageUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AL", rowNum), fmt.Sprintf("%.2f", row.CifValueForeign))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AM", rowNum), row.FobValueForeign)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AN", rowNum), row.ExchangeRate)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AO", rowNum), row.CurrencyCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AP", rowNum), row.HouseAirWaybill)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AQ", rowNum), row.ConsignmentCountry)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AR", rowNum), fmt.Sprintf("%.2f", row.FreightValueForeign))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AS", rowNum), row.FreightCurrencyCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AT", rowNum), fmt.Sprintf("%.2f", row.InsuranceValueForeign))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AU", rowNum), row.InsuranceCurrencyCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AV", rowNum), row.OtherChargeValueForeign)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AW", rowNum), row.OtherChargeCurrencyCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AX", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AY", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AZ", rowNum), row.UUID)

	}

	var excelBuf bytes.Buffer
	if err := f.Write(&excelBuf); err != nil {
		return "", nil, err
	}

	fileName := fmt.Sprintf("raw_pre_import_%v.xlsx", preImportData.Mawb)

	return fileName, &excelBuf, nil
}

func (s *service) UploadUpdateRawPreImport(ctx context.Context, userUUID, headerUUID, originName string, fileBytes []byte) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	// Logging and Upload to GCS
	uploadLogUUID, err := s.uploadlogSvc.UploadLogFile(ctx, &uploadlog.UploadFileModel{
		// Mawb:         "", // TODO:
		UserUUID: userUUID,
		FileName: originName,
		// TemplateCode: templateCode,
		Category:    "inbound",
		SubCategory: "upload_raw_pre_import",
		FileBytes:   fileBytes,
		// Amount:       0,
	})
	if err != nil {
		return err
	}

	file := bytes.NewReader(fileBytes)

	if file.Len() == 0 {
		return errors.New("empty")
	}

	x, err := excelize.OpenReader(file)
	if err != nil {
		return err
	}

	// Validate Sheet name
	sheetName := "Output" //TODO:

	rows, err := x.GetRows(sheetName)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return errors.New("Excel file is empty!")
	}

	headerRow := rows[0]
	if err := validateHeadersUploadUpdateRawPreImport(headerRow); err != nil {
		return fmt.Errorf("Header validation failed: %vs", err)
	}

	listUpdateData := []*UpdatePreImportManifestDetailModel{}
	for j := 2; j <= len(rows); j++ {
		data := UpdatePreImportManifestDetailModel{}
		data.MasterAirWaybill, _ = x.GetCellValue(sheetName, fmt.Sprintf("B%d", j))
		data.HouseAirWaybill, _ = x.GetCellValue(sheetName, fmt.Sprintf("C%d", j))
		data.Category, _ = x.GetCellValue(sheetName, fmt.Sprintf("D%d", j))
		data.ConsigneeTax, _ = x.GetCellValue(sheetName, fmt.Sprintf("E%d", j))
		data.ConsigneeBranch, _ = x.GetCellValue(sheetName, fmt.Sprintf("F%d", j))
		data.ConsigneeName, _ = x.GetCellValue(sheetName, fmt.Sprintf("G%d", j))
		data.ConsigneeAddress, _ = x.GetCellValue(sheetName, fmt.Sprintf("H%d", j))
		data.ConsigneeDistrict, _ = x.GetCellValue(sheetName, fmt.Sprintf("I%d", j))
		data.ConsigneeSubprovince, _ = x.GetCellValue(sheetName, fmt.Sprintf("J%d", j))
		data.ConsigneeProvince, _ = x.GetCellValue(sheetName, fmt.Sprintf("K%d", j))
		data.ConsigneePostcode, _ = x.GetCellValue(sheetName, fmt.Sprintf("L%d", j))
		data.ConsigneeCountryCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("M%d", j))
		data.ConsigneeEmail, _ = x.GetCellValue(sheetName, fmt.Sprintf("N%d", j))
		data.ConsigneePhoneNumber, _ = x.GetCellValue(sheetName, fmt.Sprintf("O%d", j))
		data.ShipperName, _ = x.GetCellValue(sheetName, fmt.Sprintf("P%d", j))
		data.ShipperAddress, _ = x.GetCellValue(sheetName, fmt.Sprintf("Q%d", j))
		data.ShipperDistrict, _ = x.GetCellValue(sheetName, fmt.Sprintf("R%d", j))
		data.ShipperSubprovince, _ = x.GetCellValue(sheetName, fmt.Sprintf("S%d", j))
		data.ShipperProvince, _ = x.GetCellValue(sheetName, fmt.Sprintf("T%d", j))
		data.ShipperPostcode, _ = x.GetCellValue(sheetName, fmt.Sprintf("U%d", j))
		data.ShipperCountryCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("V%d", j))
		data.ShipperEmail, _ = x.GetCellValue(sheetName, fmt.Sprintf("W%d", j))
		data.ShipperPhoneNumber, _ = x.GetCellValue(sheetName, fmt.Sprintf("X%d", j))
		data.TariffCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("Y%d", j))
		data.TariffSequence, _ = x.GetCellValue(sheetName, fmt.Sprintf("Z%d", j))
		data.StatisticalCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AA%d", j))
		data.EnglishDescriptionOfGood, _ = x.GetCellValue(sheetName, fmt.Sprintf("AB%d", j))
		data.ThaiDescriptionOfGood, _ = x.GetCellValue(sheetName, fmt.Sprintf("AC%d", j))

		// Quantity
		qty, _ := x.GetCellValue(sheetName, fmt.Sprintf("AD%d", j))
		data.Quantity = utils.ConvertStringToInt(qty)
		// Quantity

		data.QuantityUnitCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AE%d", j))

		// NetWeight
		netWeight, _ := x.GetCellValue(sheetName, fmt.Sprintf("AF%d", j))
		data.NetWeight = utils.ConvertStringToFloat(netWeight)
		// NetWeight

		data.NetWeightUnitCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AG%d", j))

		// GrossWeight
		grossWeight, _ := x.GetCellValue(sheetName, fmt.Sprintf("AH%d", j))
		data.GrossWeight = utils.ConvertStringToFloat(grossWeight)
		// GrossWeight

		data.GrossWeightUnitCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AI%d", j))
		data.Package, _ = x.GetCellValue(sheetName, fmt.Sprintf("AJ%d", j))
		data.PackageUnitCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AK%d", j))

		// CifValueForeign
		cifValueForeign, _ := x.GetCellValue(sheetName, fmt.Sprintf("AL%d", j))
		data.CifValueForeign = utils.ConvertStringToFloat(cifValueForeign)
		// CifValueForeign

		// FobValueForeign
		fobValueForeign, _ := x.GetCellValue(sheetName, fmt.Sprintf("AM%d", j))
		data.FobValueForeign = utils.ConvertStringToFloat(fobValueForeign)
		// FobValueForeign

		// ExchangeRate
		exchangeRate, _ := x.GetCellValue(sheetName, fmt.Sprintf("AN%d", j))
		data.ExchangeRate = utils.ConvertStringToFloat(exchangeRate)
		// ExchangeRate

		data.CurrencyCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AO%d", j))
		data.ShippingMark, _ = x.GetCellValue(sheetName, fmt.Sprintf("AP%d", j))
		data.ConsignmentCountry, _ = x.GetCellValue(sheetName, fmt.Sprintf("AQ%d", j))

		// FreightValueForeign
		freightValueForeign, _ := x.GetCellValue(sheetName, fmt.Sprintf("AR%d", j))
		data.FreightValueForeign = utils.ConvertStringToFloat(freightValueForeign)
		// FreightValueForeign

		data.FreightCurrencyCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AS%d", j))

		// InsuranceValueForeign
		insuranceValueForeign, _ := x.GetCellValue(sheetName, fmt.Sprintf("AT%d", j))
		data.InsuranceValueForeign = utils.ConvertStringToFloat(insuranceValueForeign)
		// InsuranceValueForeign

		data.InsuranceCurrencyCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AU%d", j))
		data.OtherChargeValueForeign, _ = x.GetCellValue(sheetName, fmt.Sprintf("AV%d", j))
		data.OtherChargeCurrencyCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AW%d", j))
		data.InvoiceNo, _ = x.GetCellValue(sheetName, fmt.Sprintf("AX%d", j))
		data.InvoiceDate, _ = x.GetCellValue(sheetName, fmt.Sprintf("AY%d", j))
		data.UUID, _ = x.GetCellValue(sheetName, fmt.Sprintf("AZ%d", j))

		listUpdateData = append(listUpdateData, &data)
	}

	if err := s.selfRepo.UpdatePreImportManifestDetail(ctx, headerUUID, listUpdateData); err != nil {
		return err
	}

	err = s.uploadlogSvc.Update(ctx, &uploadlog.UpdateModel{
		UUID: uploadLogUUID,
		// Mawb:   v.Mawb,
		Amount: int64(len(listUpdateData)),
		Status: "success",
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetOneByHeaderUUID(ctx context.Context, headerUUID string) (*GetPreImportManifestModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	result, err := s.selfRepo.GetOneMawb(ctx, headerUUID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetSummaryByHeaderUUID(ctx context.Context, headerUUID string) (*UploadSummaryModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	mawbInfo, err := s.selfRepo.GetOneMawb(ctx, headerUUID)
	if err != nil {
		return nil, err
	}

	list, err := s.selfRepo.GetSummaryByHeaderUUID(ctx, headerUUID)
	if err != nil {
		return nil, err
	}

	totalHawb := len(list)
	customFee := calcCustomsFee(int(totalHawb))
	otCustomsFee := &OTCustomFeeModel{}
	if mawbInfo.IsEnableCustomsOT {
		otCustomsFee = calcOTCustomsFee(int(totalHawb))
	}
	bankFee := calcBankFee(int(totalHawb))
	cargoPermitFee := calcCargoPermitFee(int(totalHawb))
	expressDelvieryFee := calcExpressDeliveryFee(int(totalHawb))

	result := &UploadSummaryModel{}
	cat2 := &CatogorySummaryModel{Category: "2"}
	cat3 := &CatogorySummaryModel{Category: "3"}
	otherCatogory := &CatogorySummaryModel{Category: "other"}
	for i, v := range list {
		switch v.Category {
		case "2":
			cat2.Total++
			cat2.Vat += v.Vat
			cat2.CustomFee = cat2.CustomFee.Add(customFee.PerHawbFees[i])
			if mawbInfo.IsEnableCustomsOT {
				cat2.OTCustomsFee = cat2.OTCustomsFee.Add(otCustomsFee.PerHawbFees[i])
			}
			cat2.BankFee = cat2.BankFee.Add(bankFee.PerHawbFees[i])
			cat2.CargoPermitFee = cat2.CargoPermitFee.Add(cargoPermitFee.PerHawbFees[i])
			cat2.ExpressDeliveryFee = cat2.ExpressDeliveryFee.Add(expressDelvieryFee.PerHawbFees[i])
		case "3":
			cat3.Total++
			cat3.Vat += v.Vat
			cat3.Duty += v.Duty
			cat3.DutyAndVat += (v.Duty + v.Vat)
			cat3.CustomFee = cat3.CustomFee.Add(customFee.PerHawbFees[i])
			if mawbInfo.IsEnableCustomsOT {
				cat3.OTCustomsFee = cat3.OTCustomsFee.Add(otCustomsFee.PerHawbFees[i])
			}
			cat3.BankFee = cat3.BankFee.Add(bankFee.PerHawbFees[i])
			cat3.CargoPermitFee = cat3.CargoPermitFee.Add(cargoPermitFee.PerHawbFees[i])
			cat3.ExpressDeliveryFee = cat3.ExpressDeliveryFee.Add(expressDelvieryFee.PerHawbFees[i])
		default:
			otherCatogory.Total++
			otherCatogory.Vat += v.Vat
			otherCatogory.Duty += v.Duty
			otherCatogory.DutyAndVat += (v.Duty + v.Vat)
			otherCatogory.CustomFee = otherCatogory.CustomFee.Add(customFee.PerHawbFees[i])
			if mawbInfo.IsEnableCustomsOT {
				otherCatogory.OTCustomsFee = otherCatogory.OTCustomsFee.Add(otCustomsFee.PerHawbFees[i])
			}
			otherCatogory.BankFee = otherCatogory.BankFee.Add(bankFee.PerHawbFees[i])
			otherCatogory.CargoPermitFee = otherCatogory.CargoPermitFee.Add(cargoPermitFee.PerHawbFees[i])
			otherCatogory.ExpressDeliveryFee = otherCatogory.ExpressDeliveryFee.Add(expressDelvieryFee.PerHawbFees[i])
		}
	}

	result.CustomFee = customFee
	result.OTCustomFee = otCustomsFee
	result.BankFeeFee = bankFee
	result.CargoPermitFee = cargoPermitFee
	result.ExpressDeliveryFee = expressDelvieryFee
	result.Catogory2 = cat2
	result.Catogory3 = cat3
	result.OtherCatogory = otherCatogory
	result.TotalTax = cat2.Vat + cat3.DutyAndVat
	result.TotalHawb = cat2.Total + cat3.Total + otherCatogory.Total

	return result, nil
}
