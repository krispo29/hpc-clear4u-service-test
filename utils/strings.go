package utils

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

func SubStringBetween(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func SubStringBefore(value string, a string) string {
	// Get substring before a string.
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func SubStringAfter(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

func RemoveNonAlphanumeric(s string) string {

	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(s, "")

}

func BaseName(value string, a string) string {
	x := []byte(a)
	n := strings.LastIndexByte(value, x[0])
	if n == -1 {
		return value
	}
	return value[:n]
}

func ConvertStringToBoolean(s string) bool {
	var result bool
	if strings.ToLower(s) == "y" || strings.ToLower(s) == "yes" || strings.ToLower(s) == "true" {
		result = true
	} else {
		result = false
	}

	return result
}

func ConvertStringToFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 32)
	return float64(f)
}

func ConvertStringToInt(s string) int64 {
	i, _ := strconv.Atoi(s)
	return int64(i)
}

func IsEmpty(s string) string {
	if len(s) == 0 {
		return "n/a"
	}

	return s
}
