package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	// You need to update the import path to match your project's module name.
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
)

// UserValidator holds the validator instance.
type UserValidator struct {
	validator *validator.Validate
}

// validatePhone is a custom validator function to check phone number format.
// It returns true if the phone number is valid, or if it's an empty string.
// The 'required' tag should be used separately if the field is mandatory.
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Allow empty strings to pass, in case the field is optional.
	// The 'required' tag will handle the empty case if needed.
	if phone == "" {
		return true
	}

	// Remove spaces, dashes, and parentheses for consistent validation
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	// Basic regex for international phone number format (+ followed by 1-14 digits)
	phoneRegex := `^\+?[1-9]\d{1,14}$`
	matched, _ := regexp.MatchString(phoneRegex, cleanPhone)
	return matched
}

// validateStrongPassword is a custom validator function to check password strength.
// It checks for uppercase, lowercase, and a number.
// The minimum length check is handled by the `min` tag in the struct.
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)

	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasUpper && hasLower && hasNumber
}

// NewUserValidator creates and returns a new UserValidator instance with
// custom validators registered.
func NewUserValidator() *UserValidator {
	v := validator.New()

	// Register the custom validator functions
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("strong_password", validateStrongPassword)
	// The `email` tag is a built-in validator, so no need to register it.

	return &UserValidator{
		validator: v,
	}
}

// ValidateCreateUser validates a CreateUserRequest struct using tags.
func (v *UserValidator) ValidateCreateUser(req *models.CreateUserRequest) error {
	// The validator library handles all validation rules defined by struct tags.
	if err := v.validator.Struct(req); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// ValidateUpdateUser validates an UpdateUserRequest struct.
// func (v *UserValidator) ValidateUpdateUser(req *models.UpdateUserRequest) error {
// 	if err := v.validator.Struct(req); err != nil {
// 		return v.formatValidationError(err)
// 	}
// 	return nil
// }

// ValidateLogin validates a LoginRequest struct.
func (v *UserValidator) ValidateLogin(req *models.LoginRequest) error {
	if err := v.validator.Struct(req); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// formatValidationError converts a validator.ValidationErrors to a single, readable error string.
func (v *UserValidator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			messages = append(messages, v.getFieldErrorMessage(fieldError))
		}
		return fmt.Errorf(strings.Join(messages, "; "))
	}
	return err
}

// getFieldErrorMessage returns a user-friendly error message for a given validation error.
func (v *UserValidator) getFieldErrorMessage(fieldError validator.FieldError) string {
	// fieldError.Field() returns the struct field name, e.g., "Password"
	field := strings.ToLower(fieldError.Field())

	// Customize messages based on the validation tag
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldError.Param())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least one uppercase letter, one lowercase letter, and one number", field)
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, fieldError.Tag())
	}
}
