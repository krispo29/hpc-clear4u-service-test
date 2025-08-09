package common

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"hpc-express-service/errors"
)

// ValidationConfig holds configuration for validation rules
type ValidationConfig struct {
	MaxStringLength  int
	MaxItemsCount    int
	MaxNestedDepth   int
	AllowedFileTypes []string
	MaxFileSize      int64
	RequiredFields   map[string]bool
	NumericRanges    map[string]NumericRange
	StringPatterns   map[string]*regexp.Regexp
	BusinessRules    map[string]BusinessRule
}

// NumericRange defines min/max values for numeric fields
type NumericRange struct {
	Min float64
	Max float64
}

// BusinessRule defines custom business validation logic
type BusinessRule struct {
	Name        string
	Description string
	Validator   func(interface{}) error
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxStringLength:  1000,
		MaxItemsCount:    100,
		MaxNestedDepth:   5,
		AllowedFileTypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx"},
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		RequiredFields:   make(map[string]bool),
		NumericRanges:    make(map[string]NumericRange),
		StringPatterns:   make(map[string]*regexp.Regexp),
		BusinessRules:    make(map[string]BusinessRule),
	}
}

// InputSanitizer provides comprehensive input sanitization
type InputSanitizer struct {
	config *ValidationConfig
}

// NewInputSanitizer creates a new input sanitizer with configuration
func NewInputSanitizer(config *ValidationConfig) *InputSanitizer {
	if config == nil {
		config = DefaultValidationConfig()
	}
	return &InputSanitizer{config: config}
}

// SanitizeString performs comprehensive string sanitization
func (s *InputSanitizer) SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Remove null bytes and control characters
	input = s.removeControlCharacters(input)

	// HTML escape to prevent XSS
	input = html.EscapeString(input)

	// URL decode to handle encoded malicious content
	if decoded, err := url.QueryUnescape(input); err == nil {
		input = decoded
	}

	// Remove SQL injection patterns
	input = s.removeSQLInjectionPatterns(input)

	// Trim whitespace and limit length
	input = strings.TrimSpace(input)
	if len(input) > s.config.MaxStringLength {
		input = input[:s.config.MaxStringLength]
	}

	return input
}

// removeControlCharacters removes null bytes and control characters
func (s *InputSanitizer) removeControlCharacters(input string) string {
	return strings.Map(func(r rune) rune {
		if r == 0 || (unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t') {
			return -1 // Remove the character
		}
		return r
	}, input)
}

// removeSQLInjectionPatterns removes common SQL injection patterns
func (s *InputSanitizer) removeSQLInjectionPatterns(input string) string {
	// Common SQL injection patterns
	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(drop\s+table)`,
		`(?i)(delete\s+from)`,
		`(?i)(insert\s+into)`,
		`(?i)(update\s+set)`,
		`(?i)(exec\s*\()`,
		`(?i)(script\s*>)`,
		`(?i)(<\s*script)`,
		`(?i)(javascript\s*:)`,
		`(?i)(vbscript\s*:)`,
		`(?i)(onload\s*=)`,
		`(?i)(onerror\s*=)`,
		`(?i)(onclick\s*=)`,
	}

	for _, pattern := range sqlPatterns {
		re := regexp.MustCompile(pattern)
		input = re.ReplaceAllString(input, "")
	}

	return input
}

// ValidateNumericString validates and sanitizes numeric string inputs
func (s *InputSanitizer) ValidateNumericString(input, fieldName string, allowNegative bool) (string, error) {
	if input == "" {
		return input, nil
	}

	// Remove non-numeric characters except decimal point and minus sign
	cleaned := regexp.MustCompile(`[^\d.-]`).ReplaceAllString(input, "")

	// Validate format
	if !allowNegative {
		cleaned = strings.ReplaceAll(cleaned, "-", "")
	}

	// Check if it's a valid number
	if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
		return "", errors.NewValidationError(fieldName, "must be a valid number")
	}

	return cleaned, nil
}

// ValidateEmail validates email format
func (s *InputSanitizer) ValidateEmail(email string) error {
	if email == "" {
		return nil
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.NewValidationError("email", "invalid email format")
	}

	return nil
}

// ValidateUUID validates UUID format
func (s *InputSanitizer) ValidateUUID(uuid, fieldName string) error {
	if uuid == "" {
		return nil
	}

	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(uuid) {
		return errors.NewValidationError(fieldName, "invalid UUID format")
	}

	return nil
}

// ValidateStringLength validates string length constraints
func (s *InputSanitizer) ValidateStringLength(input, fieldName string, minLength, maxLength int) error {
	length := len(input)

	if minLength > 0 && length < minLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must be at least %d characters long", minLength))
	}

	if maxLength > 0 && length > maxLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must not exceed %d characters", maxLength))
	}

	return nil
}

// ValidateEnumValue validates that a value is in allowed enum values
func (s *InputSanitizer) ValidateEnumValue(value, fieldName string, allowedValues []string) error {
	if value == "" {
		return nil
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return errors.NewValidationError(fieldName, fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", ")))
}

// ValidateArrayLength validates array/slice length constraints
func (s *InputSanitizer) ValidateArrayLength(length int, fieldName string, minLength, maxLength int) error {
	if minLength > 0 && length < minLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must contain at least %d items", minLength))
	}

	if maxLength > 0 && length > maxLength {
		return errors.NewValidationError(fieldName, fmt.Sprintf("must not contain more than %d items", maxLength))
	}

	return nil
}

// BusinessRuleValidator provides business rule validation
type BusinessRuleValidator struct {
	sanitizer *InputSanitizer
}

// NewBusinessRuleValidator creates a new business rule validator
func NewBusinessRuleValidator(sanitizer *InputSanitizer) *BusinessRuleValidator {
	return &BusinessRuleValidator{sanitizer: sanitizer}
}

// ValidateMAWBNumber validates MAWB number format and business rules
func (v *BusinessRuleValidator) ValidateMAWBNumber(mawbNumber string) error {
	if mawbNumber == "" {
		return errors.NewValidationError("mawbNumber", "MAWB number is required")
	}

	// MAWB number should be 11 digits (3-digit airline code + 8-digit serial number)
	mawbRegex := regexp.MustCompile(`^\d{3}-?\d{8}$`)
	if !mawbRegex.MatchString(mawbNumber) {
		return errors.NewValidationError("mawbNumber", "MAWB number must be in format XXX-XXXXXXXX or XXXXXXXXXXX")
	}

	return nil
}

// ValidateHAWBNumber validates HAWB number format
func (v *BusinessRuleValidator) ValidateHAWBNumber(hawbNumber string) error {
	if hawbNumber == "" {
		return nil // HAWB is optional in most cases
	}

	// HAWB number validation - alphanumeric, 6-20 characters
	hawbRegex := regexp.MustCompile(`^[A-Za-z0-9]{6,20}$`)
	if !hawbRegex.MatchString(hawbNumber) {
		return errors.NewValidationError("hawbNumber", "HAWB number must be 6-20 alphanumeric characters")
	}

	return nil
}

// ValidateWeight validates weight values and units
func (v *BusinessRuleValidator) ValidateWeight(weight, unit, fieldName string) error {
	if weight == "" {
		return nil
	}

	// Validate numeric value
	weightValue, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		return errors.NewValidationError(fieldName, "weight must be a valid number")
	}

	if weightValue < 0 {
		return errors.NewValidationError(fieldName, "weight cannot be negative")
	}

	if weightValue > 100000 { // 100 tons maximum
		return errors.NewValidationError(fieldName, "weight cannot exceed 100,000 kg")
	}

	// Validate unit if provided
	if unit != "" {
		allowedUnits := []string{"kg", "lb", "g", "oz"}
		if err := v.sanitizer.ValidateEnumValue(unit, fieldName+"Unit", allowedUnits); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDimensions validates dimension values
func (v *BusinessRuleValidator) ValidateDimensions(length, width, height, count, fieldPrefix string) []errors.ValidationError {
	var validationErrors []errors.ValidationError

	// Validate each dimension
	dimensions := map[string]string{
		"length": length,
		"width":  width,
		"height": height,
	}

	for dimName, dimValue := range dimensions {
		if dimValue != "" {
			dimFloat, err := strconv.ParseFloat(dimValue, 64)
			if err != nil {
				validationErrors = append(validationErrors, errors.NewValidationError(
					fieldPrefix+"."+dimName,
					"must be a valid number",
				))
				continue
			}

			if dimFloat <= 0 {
				validationErrors = append(validationErrors, errors.NewValidationError(
					fieldPrefix+"."+dimName,
					"must be greater than zero",
				))
			}

			if dimFloat > 1000 { // 10 meters maximum
				validationErrors = append(validationErrors, errors.NewValidationError(
					fieldPrefix+"."+dimName,
					"cannot exceed 1000 cm",
				))
			}
		}
	}

	// Validate count
	if count != "" {
		countInt, err := strconv.Atoi(count)
		if err != nil {
			validationErrors = append(validationErrors, errors.NewValidationError(
				fieldPrefix+".count",
				"count must be a valid integer",
			))
		} else if countInt <= 0 {
			validationErrors = append(validationErrors, errors.NewValidationError(
				fieldPrefix+".count",
				"count must be greater than zero",
			))
		} else if countInt > 10000 {
			validationErrors = append(validationErrors, errors.NewValidationError(
				fieldPrefix+".count",
				"count cannot exceed 10,000",
			))
		}
	}

	return validationErrors
}

// ValidateCurrency validates currency codes
func (v *BusinessRuleValidator) ValidateCurrency(currency string) error {
	if currency == "" {
		return nil
	}

	// Common currency codes
	allowedCurrencies := []string{"USD", "EUR", "GBP", "JPY", "THB", "CNY", "SGD", "HKD", "AUD", "CAD"}
	return v.sanitizer.ValidateEnumValue(currency, "currency", allowedCurrencies)
}

// ValidateAirportCode validates IATA airport codes
func (v *BusinessRuleValidator) ValidateAirportCode(code, fieldName string) error {
	if code == "" {
		return nil
	}

	// IATA airport codes are 3 uppercase letters
	airportRegex := regexp.MustCompile(`^[A-Z]{3}$`)
	if !airportRegex.MatchString(code) {
		return errors.NewValidationError(fieldName, "must be a valid 3-letter IATA airport code")
	}

	return nil
}

// ValidateFlightNumber validates flight number format
func (v *BusinessRuleValidator) ValidateFlightNumber(flightNo string) error {
	if flightNo == "" {
		return nil
	}

	// Flight number format: 2-3 letter airline code + 1-4 digit flight number
	flightRegex := regexp.MustCompile(`^[A-Z]{2,3}\d{1,4}$`)
	if !flightRegex.MatchString(flightNo) {
		return errors.NewValidationError("flightNo", "must be in format like TG123 or BA1234")
	}

	return nil
}
