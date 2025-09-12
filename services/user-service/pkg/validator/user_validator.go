package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/ecommerce/services/user-service/internal/models"
)

type UserValidator struct {
	validator *validator.Validate
}

func NewUserValidator() *UserValidator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("strong_password", validateStrongPassword)
	
	return &UserValidator{
		validator: v,
	}
}

func (v *UserValidator) ValidateCreateUser(req *models.CreateUserRequest) error {
	// Basic validation
	if err := v.validator.Struct(req); err != nil {
		return v.formatValidationError(err)
	}
	
	// Additional custom validations
	if err := v.validateEmail(req.Email); err != nil {
		return err
	}
	
	if err := v.validatePassword(req.Password); err != nil {
		return err
	}
	
	if req.Phone != "" {
		if err := v.validatePhoneNumber(req.Phone); err != nil {
			return err
		}
	}
	
	return nil
}

func (v *UserValidator) ValidateUpdateUser(req *models.UpdateUserRequest) error {
	if err := v.validator.Struct(req); err != nil {
		return v.formatValidationError(err)
	}
	
	if req.Phone != "" {
		if err := v.validatePhoneNumber(req.Phone); err != nil {
			return err
		}
	}
	
	return nil
}

func (v *UserValidator) ValidateLogin(req *models.LoginRequest) error {
	if err := v.validator.Struct(req); err != nil {
		return v.formatValidationError(err)
	}
	
	if err := v.validateEmail(req.Email); err != nil {
		return err
	}
	
	return nil
}

func (v *UserValidator) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	
	// Basic email regex
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

func (v *UserValidator) validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	
	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	
	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	
	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	
	return nil
}

func (v *UserValidator) validatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}
	
	// Remove spaces, dashes, and parentheses
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
	
	// Check if phone number is valid (basic validation)
	phoneRegex := `^\+?[1-9]\d{1,14}$` // International format
	matched, _ := regexp.MatchString(phoneRegex, cleanPhone)
	if !matched {
		return fmt.Errorf("invalid phone number format")
	}
	
	return nil
}

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

func (v *UserValidator) getFieldErrorMessage(fieldError validator.FieldError) string {
	field := strings.ToLower(fieldError.Field())
	
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
		return