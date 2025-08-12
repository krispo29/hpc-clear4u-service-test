package outbound

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CustomDate ใช้สำหรับรับวันที่ในรูปแบบ dd/mm/yyyy
type CustomDate struct {
	Time  time.Time
	Valid bool
}

func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Valid = false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		cd.Time = v
		cd.Valid = true
		return nil
	case []uint8:
		// Handle byte slice (common from database)
		str := string(v)
		if str == "" {
			cd.Valid = false
			return nil
		}
		t, err := time.Parse("2006-01-02", str)
		if err != nil {
			// Try parsing with time component
			t, err = time.Parse("2006-01-02 15:04:05", str)
			if err != nil {
				return fmt.Errorf("cannot parse date string %s: %v", str, err)
			}
		}
		cd.Time = t
		cd.Valid = true
		return nil
	case string:
		if v == "" {
			cd.Valid = false
			return nil
		}
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			// Try parsing with time component
			t, err = time.Parse("2006-01-02 15:04:05", v)
			if err != nil {
				return fmt.Errorf("cannot parse date string %s: %v", v, err)
			}
		}
		cd.Time = t
		cd.Valid = true
		return nil
	default:
		return fmt.Errorf("cannot scan %T into CustomDate", value)
	}
}

func (cd CustomDate) Value() (driver.Value, error) {
	if !cd.Valid {
		return nil, nil
	}
	return cd.Time, nil
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	if str == "" || str == "null" {
		cd.Valid = false
		return nil
	}

	// แปลง dd/mm/yyyy
	parts := strings.Split(str, "/")
	if len(parts) != 3 {
		return fmt.Errorf("invalid date format: %s, expected dd/mm/yyyy", str)
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid day: %s", parts[0])
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid month: %s", parts[1])
	}
	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid year: %s", parts[2])
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	cd.Time = t
	cd.Valid = true
	return nil
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if !cd.Valid {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%02d/%02d/%04d"`, cd.Time.Day(), cd.Time.Month(), cd.Time.Year())), nil
}
