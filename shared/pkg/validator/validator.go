package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Email regex pattern (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone regex pattern (international format)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// Alphanumeric with spaces, hyphens, and underscores
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)

	// UUID v4 pattern
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// Common validation errors
var (
	ErrRequired         = errors.New("field is required")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidPhone     = errors.New("invalid phone number format")
	ErrInvalidUUID      = errors.New("invalid UUID format")
	ErrTooShort         = errors.New("value is too short")
	ErrTooLong          = errors.New("value is too long")
	ErrInvalidCharacter = errors.New("contains invalid characters")
	ErrOutOfRange       = errors.New("value is out of valid range")
)

// ValidateRequired checks if a string field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s %w", fieldName, ErrRequired)
	}
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return ErrRequired
	}

	email = strings.TrimSpace(strings.ToLower(email))

	if len(email) > 254 {
		return fmt.Errorf("email %w: max 254 characters", ErrTooLong)
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) error {
	if phone == "" {
		return ErrRequired
	}

	// Remove spaces and hyphens for validation
	cleaned := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	if !phoneRegex.MatchString(cleaned) {
		return ErrInvalidPhone
	}

	return nil
}

// ValidatePassword validates password strength
// Requirements: min 8 chars, at least 1 uppercase, 1 lowercase, 1 digit
func ValidatePassword(password string) error {
	if password == "" {
		return ErrRequired
	}

	if len(password) < 8 {
		return fmt.Errorf("password %w: minimum 8 characters", ErrTooShort)
	}

	if len(password) > 128 {
		return fmt.Errorf("password %w: maximum 128 characters", ErrTooLong)
	}

	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain at least 1 uppercase, 1 lowercase, and 1 digit")
	}

	return nil
}

// ValidateLength validates string length
func ValidateLength(value string, min, max int, fieldName string) error {
	length := len(strings.TrimSpace(value))

	if length < min {
		return fmt.Errorf("%s %w: minimum %d characters", fieldName, ErrTooShort, min)
	}

	if length > max {
		return fmt.Errorf("%s %w: maximum %d characters", fieldName, ErrTooLong, max)
	}

	return nil
}

// ValidateAlphanumeric validates that string contains only alphanumeric characters, spaces, hyphens, underscores
func ValidateAlphanumeric(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s %w", fieldName, ErrRequired)
	}

	if !alphanumericRegex.MatchString(value) {
		return fmt.Errorf("%s %w: only letters, numbers, spaces, hyphens, and underscores allowed", fieldName, ErrInvalidCharacter)
	}

	return nil
}

// ValidateUUID validates UUID v4 format
func ValidateUUID(uuid, fieldName string) error {
	if uuid == "" {
		return fmt.Errorf("%s %w", fieldName, ErrRequired)
	}

	uuid = strings.ToLower(uuid)

	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("%s %w", fieldName, ErrInvalidUUID)
	}

	return nil
}

// ValidatePositiveInt validates that integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidatePositiveFloat validates that float is positive
func ValidatePositiveFloat(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateRange validates that integer is within range
func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s %w: must be between %d and %d", fieldName, ErrOutOfRange, min, max)
	}
	return nil
}

// ValidateEnum validates that value is in allowed list
func ValidateEnum(value string, allowed []string, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s %w", fieldName, ErrRequired)
	}

	for _, a := range allowed {
		if value == a {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
}

// SanitizeHTML removes HTML tags to prevent XSS
func SanitizeHTML(input string) string {
	// Simple HTML tag removal - for production, use library like bluemonday
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	return htmlTagRegex.ReplaceAllString(input, "")
}

// SanitizeString trims whitespace and removes control characters
func SanitizeString(input string) string {
	// Remove control characters
	input = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)

	// Trim whitespace
	return strings.TrimSpace(input)
}

// ValidateURL validates basic URL format
func ValidateURL(url, fieldName string) error {
	if url == "" {
		return fmt.Errorf("%s %w", fieldName, ErrRequired)
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("%s must start with http:// or https://", fieldName)
	}

	if len(url) > 2048 {
		return fmt.Errorf("%s %w: maximum 2048 characters", fieldName, ErrTooLong)
	}

	return nil
}

// ValidatePaginationParams validates page and page size parameters
func ValidatePaginationParams(page, pageSize int) error {
	if page < 1 {
		return errors.New("page must be at least 1")
	}

	if pageSize < 1 {
		return errors.New("page_size must be at least 1")
	}

	if pageSize > 100 {
		return errors.New("page_size cannot exceed 100")
	}

	return nil
}
