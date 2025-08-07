package compare

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

type ExcelServiceInterface interface {
	CompareExcelWithDB(ctx context.Context, excelFileBytes []byte, columnName string) (*CompareResponse, error)
}

type ExcelValue struct {
	Value  string
	HSCode string
}

func readExcelColumnFromBytes(fileBytes []byte, columnName string) (map[string]ExcelValue, error) {
	values := make(map[string]ExcelValue)
	file, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found in sheet")
	}

	columnIndex := -1
	hsCodeIndex := -1
	for i, header := range rows[0] {
		if header == columnName {
			columnIndex = i
		}
		if header == "hs_code" {
			hsCodeIndex = i
		}
	}
	if columnIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found in Excel file", columnName)
	}
	if (columnName == "goods_en" || columnName == "goods_th") && hsCodeIndex == -1 {
		return nil, fmt.Errorf("column 'hs_code' is required when comparing '%s'", columnName)
	}

	for _, row := range rows[1:] { // ข้าม header
		if columnIndex < len(row) && row[columnIndex] != "" {
			hsCodeValue := ""
			if hsCodeIndex != -1 && hsCodeIndex < len(row) {
				hsCodeValue = row[hsCodeIndex]
			}
			values[row[columnIndex]] = ExcelValue{
				Value:  row[columnIndex],
				HSCode: hsCodeValue,
			}
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no valid values found in column '%s'", columnName)
	}

	return values, nil
}

type excelService struct {
	repo ExcelRepositoryInterface
}

func NewExcelService(repo ExcelRepositoryInterface) ExcelServiceInterface {
	return &excelService{repo: repo}
}

func (s *excelService) CompareExcelWithDB(ctx context.Context, excelFileBytes []byte, columnName string) (*CompareResponse, error) {
	// Validate columnName first
	if columnName == "" {
		return nil, fmt.Errorf("columnName cannot be empty") // Or a more specific error type
	}
	allowedColumns := map[string]bool{
		"goods_en":  true,
		"goods_th":  true,
		"hs_code":   true,
		"tariff":    true,
		"unit_code": true,
		"duty_rate": true,
	}
	if !allowedColumns[columnName] {
		return nil, fmt.Errorf("column '%s' is not allowed for comparison", columnName) // Or a more specific error type
	}

	excelValues, err := readExcelColumnFromBytes(excelFileBytes, columnName)
	if err != nil {
		return nil, fmt.Errorf("Error processing Excel file for column '%s': %w", columnName, err)
	}
	if len(excelValues) == 0 {
		// This case should ideally be handled by readExcelColumnFromBytes returning an error,
		// but we can double-check or rely on the error from there.
		// For now, let's assume readExcelColumnFromBytes handles it.
		// If not, an explicit check and error might be needed here.
	}

	// ดึงข้อมูลจากฐานข้อมูล
	dbValuesSlice, err := s.repo.GetValuesFromDB(ctx, columnName)
	if err != nil {
		return nil, fmt.Errorf("failed to get values from DB: %w", err)
	}

	// สร้าง map สำหรับเปรียบเทียบตาม columnName และ hs_code
	// dbValuesMap stores DBDetails structs (not pointers) for direct value lookup.
	dbValuesMap := make(map[string]DBDetails)
	// hsCodeMap stores slices of DBDetails structs (not pointers) for hs_code fallback.
	hsCodeMap := make(map[string][]DBDetails)

	for _, dbRowPtr := range dbValuesSlice { // dbRowPtr is *DBDetails
		if dbRowPtr == nil { // Safety check
			continue
		}
		row := *dbRowPtr // Dereference to get DBDetails struct

		var val string
		switch columnName {
		case "goods_en":
			val = row.GoodsEN
		case "goods_th":
			val = row.GoodsTH
		case "hs_code":
			val = row.HSCode
		case "tariff":
			val = fmt.Sprintf("%d", row.Tariff)
		case "unit_code":
			val = row.UnitCode
		case "duty_rate":
			val = fmt.Sprintf("%f", row.DutyRate)
		}
		if val != "" {
			dbValuesMap[val] = row // Store the actual struct
		}
		if row.HSCode != "" {
			// Check for inconsistent data before appending
			for _, existingRow := range hsCodeMap[row.HSCode] {
				if existingRow.GoodsEN != row.GoodsEN || existingRow.GoodsTH != row.GoodsTH {
					log.Printf("Warning: Potential inconsistent data for HSCode %s. Record 1: GoodsEN='%s', GoodsTH='%s'. Record 2: GoodsEN='%s', GoodsTH='%s'", row.HSCode, existingRow.GoodsEN, existingRow.GoodsTH, row.GoodsEN, row.GoodsTH)
				}
			}
			hsCodeMap[row.HSCode] = append(hsCodeMap[row.HSCode], row) // Store the actual struct
		}
	}

	matchedRows := 0
	excelItems := make([]ExcelItem, 0, len(excelValues))

	for _, excelVal := range excelValues {
		item := ExcelItem{Value: excelVal.Value}
		// Attempt to match by direct value
		if dbDetailFromMap, exists := dbValuesMap[excelVal.Value]; exists {
			item.IsMatch = true
			item.MatchedBy = "column"
			item.DBDetails = &dbDetailFromMap // Assign address of the struct from map
			matchedRows++
		} else if (columnName == "goods_en" || columnName == "goods_th") && excelVal.HSCode != "" {
			// Fallback to HSCode match if applicable
			if dbDetailsSliceFromMap, ok := hsCodeMap[excelVal.HSCode]; ok && len(dbDetailsSliceFromMap) > 0 {
				matchType := "hs_code_fallback" // Default match type for HS code fallback
				var bestMatch *DBDetails

				for i := range dbDetailsSliceFromMap { // Iterate by index to get addressable elements
					record := dbDetailsSliceFromMap[i]
					if (columnName == "goods_en" && record.GoodsEN == excelVal.Value) ||
						(columnName == "goods_th" && record.GoodsTH == excelVal.Value) {
						bestMatch = &record // Found a specific match
						if columnName == "goods_en" {
							matchType = "hs_code_specific_en"
						} else {
							matchType = "hs_code_specific_th"
						}
						break // Specific match found, no need to check further in this HSCode group
					}
				}

				if bestMatch != nil {
					item.DBDetails = bestMatch
					item.IsMatch = true
					matchedRows++
				} else {
					// If no specific match, use the first record from the HSCode group as a fallback
					item.DBDetails = &dbDetailsSliceFromMap[0]
					item.IsMatch = true // Still considered a match due to HSCode
					matchedRows++
					// matchType remains "hs_code_fallback"
				}
				item.MatchedBy = matchType
			}
		}
		excelItems = append(excelItems, item)
	}

	return &CompareResponse{
		TotalExcelRows: len(excelValues),
		TotalDBRows:    len(dbValuesMap),
		MatchedRows:    matchedRows,
		ExcelItems:     excelItems,
	}, nil
}
