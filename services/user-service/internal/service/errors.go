// File: internal/service/errors.go
package service

import "errors"

// Custom error types
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
)

type ServiceError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewValidationError(message string) error {
	return &ServiceError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

func NewNotFoundError(message string) error {
	return &ServiceError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

func NewConflictError(message string) error {
	return &ServiceError{
		Type:    ErrorTypeConflict,
		Message: message,
	}
}

func NewUnauthorizedError(message string) error {
	return &ServiceError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

// Error type checkers
func IsValidationError(err error) bool {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Type == ErrorTypeValidation
	}
	return false
}

func IsNotFoundError(err error) bool {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Type == ErrorTypeNotFound
	}
	return false
}

func IsConflictError(err error) bool {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Type == ErrorTypeConflict
	}
	return false
}

func IsUnauthorizedError(err error) bool {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Type == ErrorTypeUnauthorized
	}
	return false
}
