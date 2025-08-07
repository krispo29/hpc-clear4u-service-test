package outbound

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"hpc-express-service/shopee"
	"hpc-express-service/uploadlog"
	"hpc-express-service/utils"

	"github.com/xuri/excelize/v2"
)

type OutboundExpressService interface {
	UploadManifest(ctx context.Context, userUUID, originName, templateCode string, fileBytes []byte) error
	DownloadPreExport(ctx context.Context, uploadLoggingUUID string) (string, *bytes.Buffer, error)
}

type service struct {
	selfRepo       OutboundExpressRepository
	contextTimeout time.Duration
	shopeeSvc      shopee.Service
	uploadlogSvc   uploadlog.Service
}

func NewOutboundExpressService(
	selfRepo OutboundExpressRepository,
	timeout time.Duration,
	shopeeSvc shopee.Service,
	uploadlogSvc uploadlog.Service,
) OutboundExpressService {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
		shopeeSvc:      shopeeSvc,
		uploadlogSvc:   uploadlogSvc,
	}
}

func (s *service) UploadManifest(ctx context.Context, userUUID, originName, templateCode string, fileBytes []byte) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if templateCode == "SHOPEE" {
		// Logging and Upload to GCS
		uploadLogUUID, err := s.uploadlogSvc.UploadLogFile(ctx, &uploadlog.UploadFileModel{
			// Mawb:         "", // TODO:
			UserUUID:     userUUID,
			FileName:     originName,
			TemplateCode: templateCode,
			Category:     "outbound",
			SubCategory:  "upload_manifest",
			FileBytes:    fileBytes,
			// Amount:       0,
		})
		if err != nil {
			return err
		}

		// Insert Manifest
		resultUpload, err := s.shopeeSvc.UploadPreImportManifests(ctx, uploadLogUUID, fileBytes)
		if err != nil {
			return err
		}

		log.Println(resultUpload)

		// Update Logging
		for _, v := range resultUpload {
			err := s.uploadlogSvc.Update(ctx, &uploadlog.UpdateModel{
				UUID:   uploadLogUUID,
				Mawb:   v.Mawb,
				Amount: v.Amount,
				Status: "success",
			})
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		return errors.New("not found template")
	}

	return nil

}

func (s *service) DownloadPreExport(ctx context.Context, uploadLoggingUUID string) (string, *bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	var err error
	uploadLogData, err := s.uploadlogSvc.Get(ctx, uploadLoggingUUID)
	if err != nil {
		return "", nil, err
	}

	/*
		Prepare Data
	*/
	manifest := &utils.GetHeaderManifestPreExport{}
	if uploadLogData.TemplateCode == "SHOPEE" {
		manifest, err = s.selfRepo.GetAllManifestToPreExport(ctx, uploadLogData.UUID)
		if err != nil {
			return "", nil, err
		}
	} else {
		return "", nil, errors.New("invalid template")
	}

	// Create in-memory ZIP buffer
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	// Split into chunks of 50
	chunks := utils.ChunkSlice(manifest.Details, 10000)
	// Print the result
	for i, chunk := range chunks {

		f := excelize.NewFile()
		sheetName := "Output"
		f.SetSheetName("Sheet1", sheetName)

		for k, h := range []string{"//1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "40", "41", "42", "43", "44", "45"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 1), h)
		}
		for k, h := range []string{"//Row header (=1)", "Vassel Name", "Departure Date", "Release Port", "Loading Port", "Total Package", "Total Package Unit Code", "Total Net Weight", "Total Net Weight Unit Code", "Total Gross Weight", "Total Gross Weight Uit Code"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 2), h)
		}

		for k, h := range []string{"//Row header (=2)", "Master Air Waybill", "House Air Waybill", "Category", "Consignor Company Tax Number", "Consignor Company Branch", "Consignor Name", "Consignor Street and Address", "Consignor District", "Consignor Sub Province", "Consignor Province", "Consignor Postcode", "Consignor E-mail", "Consignee Name", "Consignee Street and Address", "Consignee District", "Consignee Sub Province", "Consignee Province", "Consignee Postcode", "Consignee Country Code", "Consignee E-mail", "Purchase Country Code", "Destination Country Code", "Thai Description of Goods", "English Description of Goods", "Quantity", "Quantity Unit Code", "Net Weight", "Net Weight Unit Code", "Gross Weight ", "Gross Weight Unit Code", "Package Amount", "Package Unit Code", "Remark", "FOB Value Baht", "FOB Value Foreign", "Currency Code", "Exchange Rate", "Freight Amount", "Freight Amount Currency Code", "Insurance Amount", "Insurance Amount Currency Code", "Tariff Code", "Stat. Code", "Tariff Sequence"} {
			colName, _ := excelize.ColumnNumberToName(k + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 3), h)
		}

		startRow := 5

		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", startRow-1), 1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", startRow-1), manifest.VasselName)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", startRow-1), manifest.DepartureDate)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "D", startRow-1), manifest.ReleasePort)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "E", startRow-1), manifest.LoadingPort)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "F", startRow-1), manifest.TotalPackage)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "G", startRow-1), manifest.TotalPackageUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "H", startRow-1), manifest.TotalNetWeight)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "I", startRow-1), manifest.TotalNetWeightUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "J", startRow-1), manifest.TotalGrossWeight)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "K", startRow-1), manifest.TotalGrossWeightUnitCode)

		for k, v := range chunk {
			rowNum := (k + startRow)

			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", rowNum), 2)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", rowNum), v.MasterAirWaybill)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", rowNum), v.HouseAirWaybill)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "D", rowNum), v.Category)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "E", rowNum), v.ConsignorCompanyTaxNumber)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "F", rowNum), v.ConsignorCompanyBranch)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "G", rowNum), v.ConsignorName)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "H", rowNum), v.ConsignorStreetAndAddress)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "I", rowNum), v.ConsignorDistrict)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "J", rowNum), v.ConsignorSubProvince)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "K", rowNum), v.ConsignorProvince)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "L", rowNum), v.ConsignorPostcode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "M", rowNum), v.ConsignorEmail)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "N", rowNum), v.ConsigneeName)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "O", rowNum), v.ConsigneeStreetAndAddress)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "P", rowNum), v.ConsigneeDistrict)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Q", rowNum), v.ConsigneeSubProvince)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "R", rowNum), v.ConsigneeProvince)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "S", rowNum), v.ConsigneePostcode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "T", rowNum), v.ConsigneeCountryCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "U", rowNum), v.ConsigneeEmail)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "V", rowNum), v.PurchaseCountryCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "W", rowNum), v.DestinationCountryCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "X", rowNum), v.ThaiDescriptionOfGoods)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Y", rowNum), v.EnglishDescriptionOfGoods)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Z", rowNum), v.Quantity)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AA", rowNum), v.QuantityUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AB", rowNum), v.NetWeight)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AC", rowNum), fmt.Sprintf("%.2f", v.NetWeight))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AD", rowNum), v.GrossWeight)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AE", rowNum), v.GrossWeightUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AF", rowNum), v.PackageAmount)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AG", rowNum), v.PackageUnitCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AH", rowNum), v.Remark)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AI", rowNum), v.FobValueBaht)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AJ", rowNum), v.FobValueForeign)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AK", rowNum), v.CurrencyCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AL", rowNum), v.ExchangeRate)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AM", rowNum), v.FreightAmount)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AN", rowNum), v.FreightAmountCurrencyCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AO", rowNum), v.InsuranceAmount)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AP", rowNum), v.InsuranceAmountCurrencyCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AQ", rowNum), v.TariffCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AR", rowNum), v.StatCode)
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "AS", rowNum), v.TariffSequence)
		}
		// Save Excel to a buffer
		fileName := fmt.Sprintf("pre_export_%v_%v_%v.xlsx", uploadLogData.TemplateCode, uploadLogData.Mawb, strconv.Itoa(i+1))

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

	return fmt.Sprintf("pre_export_%v_%v", uploadLogData.TemplateCode, uploadLogData.Mawb), &zipBuf, nil
}
