package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// BaseResponse contains common fields for all API responses
type BaseResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// DataResponse for successful responses with data
type DataResponse struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse for error responses
type ErrorResponse struct {
	BaseResponse
	Error string `json:"error,omitempty"`
	Code  string `json:"code,omitempty"`
}

// ListResponse for paginated list responses
type ListResponse struct {
	BaseResponse
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteSuccess writes a successful response
func WriteSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := DataResponse{
		BaseResponse: BaseResponse{
			Success:   true,
			Message:   message,
			Timestamp: time.Now(),
			RequestID: generateRequestID(),
		},
		Data: data,
	}
	WriteJSON(w, statusCode, response)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteErrorWithCode(w, statusCode, message, "")
}

// WriteErrorWithCode writes an error response with error code
func WriteErrorWithCode(w http.ResponseWriter, statusCode int, message, errorCode string) {
	response := ErrorResponse{
		BaseResponse: BaseResponse{
			Success:   false,
			Message:   message,
			Timestamp: time.Now(),
			RequestID: generateRequestID(),
		},
		Code: errorCode,
	}
	WriteJSON(w, statusCode, response)
}

// WriteList writes a paginated list response
func WriteList(w http.ResponseWriter, statusCode int, message string, data interface{}, total, page, limit int) {
	response := ListResponse{
		BaseResponse: BaseResponse{
			Success:   true,
			Message:   message,
			Timestamp: time.Now(),
			RequestID: generateRequestID(),
		},
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}
	WriteJSON(w, statusCode, response)
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return uuid.New().String()
}

// Helper functions for common HTTP status codes
func OK(w http.ResponseWriter, message string, data interface{}) {
	WriteSuccess(w, http.StatusOK, message, data)
}

func Created(w http.ResponseWriter, message string, data interface{}) {
	WriteSuccess(w, http.StatusCreated, message, data)
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, message)
}

func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message)
}

func InternalServerError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message)
}
