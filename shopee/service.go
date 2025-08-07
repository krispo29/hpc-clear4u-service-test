package shopee

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hpc-express-service/utils"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type Service interface {
	UploadPreImportManifests(ctx context.Context, uploadLogUUID string, fileBytes []byte) ([]*ResponseUploadManifestModel, error)
}

type service struct {
	selfRepo       Repository
	contextTimeout time.Duration
}

func NewService(
	selfRepo Repository,
	timeout time.Duration,
) Service {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
	}
}

func (s *service) UploadPreImportManifests(ctx context.Context, uploadLogUUID string, fileBytes []byte) ([]*ResponseUploadManifestModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	file := bytes.NewReader(fileBytes)

	if file.Len() == 0 {
		return nil, errors.New("empty")
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}

	// Get all sheet names
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("No sheets found in Excel file")
	}
	sheet := sheets[2] // or specify the sheet name directly

	// Read all rows from the first sheet
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("Failed to read rows: %v", err)
	}

	if len(rows) == 0 {
		return nil, errors.New("Excel file is empty!")
	}

	headerRow := rows[0]
	if err := validateHeadersUploadPreImportManifests(headerRow); err != nil {
		return nil, fmt.Errorf("Header validation failed: %vs", err)
	}

	var data []*UploadManifestModel
	for i, row := range rows {
		// Skip header row
		if i == 0 {
			continue
		}

		// Safely access columns
		temp := &UploadManifestModel{}

		if len(row) > 0 {
			temp.OutboundTime = row[0]
		}
		if len(row) > 1 {
			temp.LMTracking = row[1]
		}
		if len(row) > 2 {
			temp.ShopeeTracking = row[2]
		}

		if len(row) > 3 {
			temp.OrderSN = row[3]
		}

		if len(row) > 4 {
			temp.InvoiceDate = row[4]
		}

		if len(row) > 5 {
			temp.CartonNo = row[5]
		}

		if len(row) > 6 {
			temp.UnitCode = row[6]
		}

		if len(row) > 7 {
			temp.DispatchNumber = row[7]
		}

		if len(row) > 8 {
			temp.CartonSize = row[8]
		}

		if len(row) > 9 {
			if s, err := strconv.ParseFloat(row[9], 64); err == nil {
				temp.ParcelWeight = s
			}
		}

		if len(row) > 10 {
			temp.ParcelSize = row[10]
		}

		if len(row) > 11 {
			if s, err := strconv.ParseFloat(row[11], 64); err == nil {
				temp.CartonWeight = s
			}
		}

		if len(row) > 12 {
			temp.CartonVolume = row[12]
		}

		if len(row) > 13 {
			if n, err := strconv.Atoi(row[13]); err == nil {
				temp.ParcelVolume = int64(n)
			}
		}

		if len(row) > 14 {
			temp.Transportation = row[14]
		}

		if len(row) > 15 {
			temp.Channel = row[15]
		}

		if len(row) > 16 {
			temp.ServiceCode = row[16]
		}

		if len(row) > 17 {
			temp.Country = row[17]
		}

		if len(row) > 18 {
			temp.DestinationCode = row[18]
		}

		if len(row) > 19 {
			temp.ReceiverName = row[19]
		}

		if len(row) > 20 {
			temp.ReceiverProvinceOrState = row[20]
		}

		if len(row) > 21 {
			temp.ReceiverCity = row[21]
		}

		if len(row) > 22 {
			temp.PostalCode = row[22]
		}

		if len(row) > 23 {
			temp.ReceiverTelephone = row[23]
		}

		if len(row) > 24 {
			temp.ReceiverAddress = row[24]
		}

		if len(row) > 25 {
			temp.SenderName = row[25]
		}

		if len(row) > 26 {
			temp.SenderCountry = row[26]
		}

		if len(row) > 27 {
			temp.SenderProvince = row[27]
		}

		if len(row) > 28 {
			temp.SenderCity = row[28]
		}

		if len(row) > 29 {
			temp.SenderAddress = row[29]
		}

		if len(row) > 30 {
			temp.SenderTelephone = row[30]
		}

		if len(row) > 31 {
			temp.SellerTaxNumber = row[31]
		}

		if len(row) > 32 {
			temp.BrInvoiceNumber = row[32]
		}

		if len(row) > 33 {
			temp.Declareusername = row[33]
		}

		if len(row) > 34 {
			temp.Declareusertelephone = row[34]
		}

		if len(row) > 35 {
			temp.DeclareuserID = row[35]
		}

		if len(row) > 36 {
			temp.KYCPopulationID = row[36]
		}

		if len(row) > 37 {
			temp.IncludingShoes = row[37]
		}

		if len(row) > 38 {
			temp.FootwearQuantity = row[38]
		}

		if len(row) > 39 {
			temp.FootwearDeclareValue = row[39]
		}

		if len(row) > 40 {
			temp.PackageQTY = row[40]
		}

		// Product Name 1 starts at column index 4
		for j := 41; j < 321; j += 7 {

			if j < len(row) {
				tempDetail := &DeclaredDetailModel{}
				tempDetail.DeclaredName = row[j]
				if j+6 < len(row) {
					tempDetail.HSCode = row[j+1]
				}
				if j+2 < len(row) {
					tempDetail.ProductName = row[j+2]
				}
				if j+3 < len(row) {
					if s, err := strconv.ParseFloat(row[j+3], 64); err == nil {
						tempDetail.DeclaredValue = s
					}
				}
				if j+4 < len(row) {
					if n, err := strconv.Atoi(row[j+4]); err == nil {
						tempDetail.DeclaredQTY = int64(n)
					}
				}
				if j+5 < len(row) {
					tempDetail.DeclaredCategoryid = row[j+5]
				}
				if j+6 < len(row) {
					tempDetail.DeclaredNameLocal = row[j+6]
				}
				// Add only if there is at least a name or price
				if tempDetail.DeclaredName != "" {
					temp.DeclaredDetails = append(temp.DeclaredDetails, tempDetail)
				}
			}
		}

		data = append(data, temp)
	}

	var totalNetWeight int64
	var totalGrossWeight float64
	details := []*utils.InsertPreExportDetailManifestModel{}
	for _, v := range data {
		details = append(details, v.ConvertToManifest())
		totalNetWeight += int64(v.ParcelWeight)
		totalGrossWeight += v.ParcelWeight
	}

	manifest := &utils.InsertPreExportHeaderManifestModel{
		UploadLoggingUUID:        uploadLogUUID,
		VasselName:               "TG620",
		DepartureDate:            "10/4/2017", // Outbound Time
		ReleasePort:              1193,
		LoadingPort:              1190,
		TotalPackage:             int64(len(details)),
		TotalPackageUnitCode:     "PK",
		TotalNetWeight:           totalNetWeight,
		TotalNetWeightUnitCode:   "KGM",
		TotalGrossWeight:         totalGrossWeight,
		TotalGrossWeightUnitCode: "KGM",
		Details:                  details,
	}

	err = s.selfRepo.InsertPreExportManifest(ctx, manifest, 200)
	if err != nil {
		return nil, err
	}

	result := []*ResponseUploadManifestModel{}

	result = append(result, &ResponseUploadManifestModel{
		// Mawb:   ,
		Amount: manifest.TotalPackage,
	})

	return result, nil
}
