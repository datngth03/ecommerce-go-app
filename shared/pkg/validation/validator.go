package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Email validation regex (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone validation regex (international format)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// Alphanumeric with common special characters
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,']+$`)

	// URL validation regex
	urlRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9\-._~:/?#[\]@!$&'()*+,;=%]+$`)

	// UUID validation regex
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Dangerous HTML/Script patterns
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // onclick, onerror, etc.
		regexp.MustCompile(`(?i)<embed[^>]*>`),
		regexp.MustCompile(`(?i)<object[^>]*>`),
	}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateEmail checks if the email is valid
func ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}

	if len(email) > 254 {
		return &ValidationError{Field: "email", Message: "email too long (max 254 characters)"}
	}

	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}

	return nil
}

// ValidatePhone checks if the phone number is valid
func ValidatePhone(phone string) error {
	if phone == "" {
		return &ValidationError{Field: "phone", Message: "phone is required"}
	}

	// Remove common formatting characters
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	if !phoneRegex.MatchString(cleanPhone) {
		return &ValidationError{Field: "phone", Message: "invalid phone format (use international format, e.g., +1234567890)"}
	}

	return nil
}

// ValidatePassword checks password strength
func ValidatePassword(password string) error {
	if password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}

	if len(password) > 128 {
		return &ValidationError{Field: "password", Message: "password too long (max 128 characters)"}
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &ValidationError{Field: "password", Message: "password must contain at least one uppercase letter"}
	}

	if !hasLower {
		return &ValidationError{Field: "password", Message: "password must contain at least one lowercase letter"}
	}

	if !hasNumber {
		return &ValidationError{Field: "password", Message: "password must contain at least one number"}
	}

	if !hasSpecial {
		return &ValidationError{Field: "password", Message: "password must contain at least one special character"}
	}

	return nil
}

// ValidateAlphanumeric checks if string contains only alphanumeric characters and common punctuation
func ValidateAlphanumeric(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if !alphanumericRegex.MatchString(value) {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s contains invalid characters", field)}
	}

	return nil
}

// ValidateStringLength checks if string length is within bounds
func ValidateStringLength(field, value string, minLen, maxLen int) error {
	if value == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	length := len(value)

	if length < minLen {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be at least %d characters", field, minLen)}
	}

	if maxLen > 0 && length > maxLen {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must not exceed %d characters", field, maxLen)}
	}

	return nil
}

// ValidateURL checks if the URL is valid
func ValidateURL(url string) error {
	if url == "" {
		return &ValidationError{Field: "url", Message: "url is required"}
	}

	if len(url) > 2048 {
		return &ValidationError{Field: "url", Message: "url too long (max 2048 characters)"}
	}

	if !urlRegex.MatchString(url) {
		return &ValidationError{Field: "url", Message: "invalid URL format"}
	}

	return nil
}

// ValidateUUID checks if the UUID is valid
func ValidateUUID(field, uuid string) error {
	if uuid == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if !uuidRegex.MatchString(strings.ToLower(uuid)) {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is not a valid UUID", field)}
	}

	return nil
}

// SanitizeHTML removes potentially dangerous HTML/script content
func SanitizeHTML(input string) string {
	sanitized := input

	for _, pattern := range dangerousPatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "")
	}

	return sanitized
}

// ValidateRequired checks if value is not empty
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}
	return nil
}

// ValidateNumericRange checks if numeric value is within range
func ValidateNumericRange(field string, value, min, max float64) error {
	if value < min {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be at least %.2f", field, min)}
	}

	if value > max {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must not exceed %.2f", field, max)}
	}

	return nil
}

// ValidateEnum checks if value is in allowed list
func ValidateEnum(field, value string, allowedValues []string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowedValues, ", ")),
	}
}

// ValidateSliceLength checks if slice length is within bounds
func ValidateSliceLength(field string, length, minLen, maxLen int) error {
	if length < minLen {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must contain at least %d items", field, minLen)}
	}

	if maxLen > 0 && length > maxLen {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must not exceed %d items", field, maxLen)}
	}

	return nil
}
