package mawbinfo

import (
	"fmt"
	"strings"
	"time"
)

// CustomDate is a wrapper around time.Time to handle custom date formats
type CustomDate struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (cd *CustomDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}

	// List of supported layouts
	layouts := []string{
		"2006-01-02", // YYYY-MM-DD
		"02-01-2006", // DD-MM-YYYY
		time.RFC3339,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			cd.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse date: %s", s)
}

// MarshalJSON implements the json.Marshaler interface.
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", cd.Format("2006-01-02"))), nil
}
