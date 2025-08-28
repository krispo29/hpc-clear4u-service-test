package setting

import (
	"bytes"
	"context"
	"fmt"
	"hpc-express-service/utils"
	"time"

	"github.com/xuri/excelize/v2"
)

type Service interface {
	// HS Code
	CreateHsCode(ctx context.Context, data *CreateHsCodeModel) (string, error)
	GetAllHsCode(ctx context.Context) ([]*GetHsCodeModel, error)
	GetHsCodeByUUID(ctx context.Context, uuid string) (*GetHsCodeModel, error)
	UpdateHsCode(ctx context.Context, data *UpdateHsCodeModel) error
	UpdateStatusHsCode(ctx context.Context, uuid string) error
	ExportHsCode(ctx context.Context) (*bytes.Buffer, error)
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

func (s *service) CreateHsCode(ctx context.Context, data *CreateHsCodeModel) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if uuid, err := s.selfRepo.CreateHsCode(ctx, data); err != nil {
		return "", err
	} else {
		return uuid, nil
	}
}

func (s *service) GetAllHsCode(ctx context.Context) ([]*GetHsCodeModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if result, err := s.selfRepo.GetAllHsCode(ctx, ""); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (s *service) GetHsCodeByUUID(ctx context.Context, uuid string) (*GetHsCodeModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if result, err := s.selfRepo.GetHsCodeByUUID(ctx, uuid); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (s *service) UpdateHsCode(ctx context.Context, data *UpdateHsCodeModel) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.UpdateHsCode(ctx, data); err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateStatusHsCode(ctx context.Context, uuid string) error {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.UpdateStatusHsCode(ctx, uuid); err != nil {
		return err
	}

	return nil
}

func (s *service) ExportHsCode(ctx context.Context) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	hscodeData, err := s.selfRepo.GetAllHsCode(ctx, "ORDER BY goods_en asc")
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Output"
	f.SetSheetName("Sheet1", sheetName)

	for k, h := range []string{"No.", "Goods EN", "Goods TH", "HsCode", "Tariff", "Stat", "Unit Code", "Duty Rate", "Remark", "Air Service Charge", "Sea Service Charge", "Fob Price Control", "Fob Price Control Origin Currency Code", "Fob Price Control Origin Country Code", "Weight Control", "Weight Control Unit Code", "Cif Control", "Cif Control Destination Currency Code", "Cif Control Destination Country Code", "Created At", "Updated At", "Deleted At", "Is Deleted"} {
		colName, _ := excelize.ColumnNumberToName(k + 1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, 1), h)
	}

	for i, row := range hscodeData {
		startRow := 2

		rowNum := (i + startRow)

		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "A", rowNum), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "B", rowNum), row.GoodsEN)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "C", rowNum), row.GoodsTH)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "D", rowNum), row.HsCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "E", rowNum), row.Tariff)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "F", rowNum), row.Stat)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "G", rowNum), row.UnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "H", rowNum), row.DutyRate)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "I", rowNum), row.Remark)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "J", rowNum), row.AirServiceCharge)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "K", rowNum), row.SeaServiceCharge)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "L", rowNum), row.FobPriceControl)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "M", rowNum), row.FobPriceControlOriginCurrencyCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "N", rowNum), row.FobPriceControlOriginCountryCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "O", rowNum), row.WeightControl)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "P", rowNum), row.WeightControlUnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "Q", rowNum), fmt.Sprintf("%d", row.CifControl))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "R", rowNum), utils.IsEmpty(row.CifControlDestinationCurrencyCode))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "S", rowNum), utils.IsEmpty(row.CifControlDestinationCountryCode))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "T", rowNum), utils.IsEmpty(row.CreatedAt))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "U", rowNum), utils.IsEmpty(row.UpdatedAt))
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "V", rowNum), row.DeletedAt)
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", "W", rowNum), row.IsDeleted)

	}

	var excelBuf bytes.Buffer
	if err := f.Write(&excelBuf); err != nil {
		return nil, err
	}

	return &excelBuf, nil
}
