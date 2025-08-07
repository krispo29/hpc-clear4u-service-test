package ship2cu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hpc-express-service/utils"
	"log"
	"time"

	"github.com/xuri/excelize/v2"
)

type Service interface {
	UploadPreImportManifests(ctx context.Context, uploadLogUUID string, fileBytes []byte) (*ResponseUploadManifestModel, error)
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

func (s *service) UploadPreImportManifests(ctx context.Context, uploadLogUUID string, fileBytes []byte) (*ResponseUploadManifestModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	file := bytes.NewReader(fileBytes)

	if file.Len() == 0 {
		return nil, errors.New("empty")
	}

	x, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}

	// Validate Sheet name
	sheetName := "Output" //TODO:

	rows, err := x.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, errors.New("Excel file is empty!")
	}

	headerRow := rows[0]
	if err := validateHeadersUploadPreImportManifests(headerRow); err != nil {
		return nil, fmt.Errorf("Header validation failed: %vs", err)
	}

	list := []*UploadManifestModel{}
	resultMap := make(map[string]int64)
	var mawb, countryCode, currencyCode string
	var hawbTotal int64
	for j := 2; j <= len(rows); j++ {
		data := UploadManifestModel{}

		data.Mawb, _ = x.GetCellValue(sheetName, fmt.Sprintf("B%d", j))
		data.BagNo, _ = x.GetCellValue(sheetName, fmt.Sprintf("C%d", j))
		data.Hawb, _ = x.GetCellValue(sheetName, fmt.Sprintf("D%d", j))
		data.HsCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("E%d", j))
		data.Origin, _ = x.GetCellValue(sheetName, fmt.Sprintf("F%d", j))
		data.ShipperName, _ = x.GetCellValue(sheetName, fmt.Sprintf("G%d", j))
		data.ConsigneeName, _ = x.GetCellValue(sheetName, fmt.Sprintf("H%d", j))

		wgt, _ := x.GetCellValue(sheetName, fmt.Sprintf("I%d", j))
		data.WgtValue = utils.ConvertStringToFloat(wgt)

		data.WgtUnit, _ = x.GetCellValue(sheetName, fmt.Sprintf("J%d", j))
		data.Packaging, _ = x.GetCellValue(sheetName, fmt.Sprintf("K%d", j))
		data.ShipperAddress, _ = x.GetCellValue(sheetName, fmt.Sprintf("L%d", j))
		data.ConsigneeAddress, _ = x.GetCellValue(sheetName, fmt.Sprintf("M%d", j))
		data.Province, _ = x.GetCellValue(sheetName, fmt.Sprintf("N%d", j))
		data.District, _ = x.GetCellValue(sheetName, fmt.Sprintf("O%d", j))
		data.Postcode, _ = x.GetCellValue(sheetName, fmt.Sprintf("P%d", j))
		data.Pcs, _ = x.GetCellValue(sheetName, fmt.Sprintf("Q%d", j))

		qty, _ := x.GetCellValue(sheetName, fmt.Sprintf("R%d", j))
		data.Qty = utils.ConvertStringToInt(qty)

		data.Goods, _ = x.GetCellValue(sheetName, fmt.Sprintf("S%d", j))
		data.GoodsEN, _ = x.GetCellValue(sheetName, fmt.Sprintf("T%d", j))
		data.Currency, _ = x.GetCellValue(sheetName, fmt.Sprintf("U%d", j))
		totalPrice, _ := x.GetCellValue(sheetName, fmt.Sprintf("V%d", j))
		data.TotalPrice = utils.ConvertStringToFloat(totalPrice)

		fob, _ := x.GetCellValue(sheetName, fmt.Sprintf("W%d", j))
		data.FOB = utils.ConvertStringToFloat(fob)

		freight, _ := x.GetCellValue(sheetName, fmt.Sprintf("X%d", j))
		data.Freight = utils.ConvertStringToFloat(freight)

		insurance, _ := x.GetCellValue(sheetName, fmt.Sprintf("Y%d", j))
		data.Insurance = utils.ConvertStringToFloat(insurance)

		cif, _ := x.GetCellValue(sheetName, fmt.Sprintf("Z%d", j))
		data.CIF = utils.ConvertStringToFloat(cif)

		data.Cat, _ = x.GetCellValue(sheetName, fmt.Sprintf("AA%d", j))

		duty, _ := x.GetCellValue(sheetName, fmt.Sprintf("AB%d", j))
		data.Duty = utils.ConvertStringToFloat(duty)

		vat, _ := x.GetCellValue(sheetName, fmt.Sprintf("AC%d", j))
		data.Vat = utils.ConvertStringToFloat(vat)

		cost, _ := x.GetCellValue(sheetName, fmt.Sprintf("AD%d", j))
		data.Cost = utils.ConvertStringToFloat(cost)

		data.LocalTrackingNo, _ = x.GetCellValue(sheetName, fmt.Sprintf("AE%d", j))
		data.Reference1, _ = x.GetCellValue(sheetName, fmt.Sprintf("AF%d", j))
		data.Reference2, _ = x.GetCellValue(sheetName, fmt.Sprintf("AG%d", j))
		data.CustomerCode, _ = x.GetCellValue(sheetName, fmt.Sprintf("AH%d", j))
		data.ShipperTel, _ = x.GetCellValue(sheetName, fmt.Sprintf("AI%d", j))
		data.ConsigneeTel, _ = x.GetCellValue(sheetName, fmt.Sprintf("AJ%d", j))
		data.Dimension, _ = x.GetCellValue(sheetName, fmt.Sprintf("AK%d", j))
		data.DimensionRepacking, _ = x.GetCellValue(sheetName, fmt.Sprintf("AL%d", j))

		width, _ := x.GetCellValue(sheetName, fmt.Sprintf("AM%d", j))
		data.Width = utils.ConvertStringToFloat(width)

		length, _ := x.GetCellValue(sheetName, fmt.Sprintf("AN%d", j))
		data.Length = utils.ConvertStringToFloat(length)

		height, _ := x.GetCellValue(sheetName, fmt.Sprintf("AO%d", j))
		data.Height = utils.ConvertStringToFloat(height)

		volumeWeight, _ := x.GetCellValue(sheetName, fmt.Sprintf("AP%d", j))
		data.VolumeWeight = utils.ConvertStringToFloat(volumeWeight)

		if _, exists := resultMap[data.Mawb]; !exists {
			resultMap[data.Mawb] = 1
		} else {
			resultMap[data.Mawb]++
		}

		countryCode = data.Origin
		currencyCode = data.Currency

		mawb = data.Mawb
		hawbTotal++
		list = append(list, &data)
	}

	if len(resultMap) > 1 {
		return nil, errors.New("MAWB are more than 1")
	}

	mawbData, err := s.selfRepo.GetMawb(ctx, mawb)
	if err != nil {
		log.Println("GetMawb: ", err.Error())
		// return nil, err
		mawbData = &utils.GetMawb{}
	}

	shipperBrandsData, err := s.selfRepo.GetShipperBrands(ctx)
	if err != nil {
		log.Println("GetShipperBrands: ", err.Error())
		return nil, err
	}

	masterHsCodeData, err := s.selfRepo.GetMasterHsCode(ctx)
	if err != nil {
		log.Println("GetMasterHsCode: ", err.Error())
		return nil, err
	}

	freightConfig, err := s.selfRepo.GetFreightData(ctx, uploadLogUUID, countryCode, currencyCode)
	if err != nil {
		log.Println("GetFreightData: ", err.Error())
		return nil, err
	}

	details := []*utils.InsertPreImportDetailManifestModel{}
	for k, v := range list {

		if err != nil {
			log.Println("dataOther: ", k, "=> ", err)
			continue
		}
		details = append(details, v.ConvertToManifest(shipperBrandsData, masterHsCodeData, freightConfig))
	}

	manifest := &utils.InsertPreImportHeaderManifestModel{
		UploadLoggingUUID:  uploadLogUUID,
		DischargePort:      "1190",
		VasselName:         mawbData.FlightNo,        // SHIP2CU
		ArrivalDate:        mawbData.ArrivalDatetime, // SHIP2CU
		CustomerName:       "",
		OriginCountryCode:  countryCode,
		OriginCurrencyCode: currencyCode,
		Details:            details,
	}

	err = s.selfRepo.InsertPreImportManifest(ctx, manifest, 200)
	if err != nil {
		return nil, err
	}

	return &ResponseUploadManifestModel{
		Mawb:   mawb,
		Amount: hawbTotal,
	}, nil
}
