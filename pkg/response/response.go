package response

import (
	"encoding/json"
	"net/http"
)

// Response is the standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination information
type Meta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success sends a success response
func Success(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 created response
func Created(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a success response with pagination meta
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error sends an error response
func Error(w http.ResponseWriter, status int, message, errCode string) {
	JSON(w, status, Response{
		Success: false,
		Message: message,
		Error:   errCode,
	})
}

// BadRequest sends a 400 bad request error
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message, "bad_request")
}

// Unauthorized sends a 401 unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, "unauthorized")
}

// Forbidden sends a 403 forbidden error
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, "forbidden")
}

// NotFound sends a 404 not found error
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, "not_found")
}

// Conflict sends a 409 conflict error
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, message, "conflict")
}

// ValidationError sends a 422 validation error
func ValidationError(w http.ResponseWriter, message string, details map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
		"error":   "validation_error",
		"details": details,
	})
}

// InternalError sends a 500 internal server error
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message, "internal_error")
}
