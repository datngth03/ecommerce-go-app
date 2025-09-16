package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// SuccessResponse represents a success response structure
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int         `json:"total,omitempty"`
	Page    int         `json:"page,omitempty"`
	Limit   int         `json:"limit,omitempty"`
}

// WriteErrorResponse writes an error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteErrorResponseWithDetails writes an error response with additional error details
func WriteErrorResponseWithDetails(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponse writes a success response to the HTTP response writer
func WriteSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponseWithPagination writes a success response with pagination info
func WriteSuccessResponseWithPagination(w http.ResponseWriter, statusCode int, message string, data interface{}, total, page, limit int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteJSONResponse writes a generic JSON response
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Common HTTP status messages
const (
	MessageSuccess            = "Success"
	MessageCreated            = "Resource created successfully"
	MessageUpdated            = "Resource updated successfully"
	MessageDeleted            = "Resource deleted successfully"
	MessageNotFound           = "Resource not found"
	MessageUnauthorized       = "Unauthorized access"
	MessageForbidden          = "Access forbidden"
	MessageBadRequest         = "Bad request"
	MessageInternalError      = "Internal server error"
	MessageValidationFailed   = "Validation failed"
	MessageEmailExists        = "Email already exists"
	MessageInvalidCredentials = "Invalid email or password"
	MessageTokenExpired       = "Token expired"
	MessageInvalidToken       = "Invalid token"
)

// Predefined error responses for common scenarios
func WriteUnauthorizedResponse(w http.ResponseWriter, message ...string) {
	msg := MessageUnauthorized
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	WriteErrorResponse(w, http.StatusUnauthorized, msg)
}

func WriteForbiddenResponse(w http.ResponseWriter, message ...string) {
	msg := MessageForbidden
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	WriteErrorResponse(w, http.StatusForbidden, msg)
}

func WriteNotFoundResponse(w http.ResponseWriter, message ...string) {
	msg := MessageNotFound
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	WriteErrorResponse(w, http.StatusNotFound, msg)
}

func WriteBadRequestResponse(w http.ResponseWriter, message ...string) {
	msg := MessageBadRequest
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	WriteErrorResponse(w, http.StatusBadRequest, msg)
}

func WriteInternalErrorResponse(w http.ResponseWriter, message ...string) {
	msg := MessageInternalError
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	WriteErrorResponse(w, http.StatusInternalServerError, msg)
}

func WriteValidationErrorResponse(w http.ResponseWriter, err error) {
	WriteErrorResponseWithDetails(w, http.StatusBadRequest, MessageValidationFailed, err)
}
