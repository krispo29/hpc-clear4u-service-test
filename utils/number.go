package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func RoundUpInt(x float64) int64 {
	return int64(math.Ceil(x))
}

// ConvertStringToFloat64WithDecimals converts a string to float64 with exactly 2 decimal places
// Returns error if the string cannot be parsed as a valid number
func ConvertStringToFloat64WithDecimals(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string cannot be converted to float64")
	}

	// Parse the string to float64
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", s)
	}

	// Round to 2 decimal places
	rounded := math.Round(f*100) / 100
	return rounded, nil
}

// ValidateDateFormat validates if the date string is in YYYY-MM-DD format
// Returns error if the date format is invalid
func ValidateDateFormat(dateStr string) error {
	if dateStr == "" {
		return fmt.Errorf("date string cannot be empty")
	}

	// Parse the date using the expected format
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", dateStr)
	}

	return nil
}
