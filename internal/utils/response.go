package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"Targeting-Engine/internal/models"
)

// WriteErrorResponse writes an error response in JSON format
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := models.ErrorResponse{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// LogError logs an error with context
func LogError(message string, err error) {
	log.Printf("ERROR: %s: %v", message, err)
}

// LogRequest logs HTTP request details for monitoring
func LogRequest(r *http.Request, statusCode int, duration time.Duration) {
	log.Printf(
		"REQUEST: %s %s - Status: %d - Duration: %v - IP: %s - UserAgent: %s",
		r.Method,
		r.URL.String(),
		statusCode,
		duration,
		getClientIP(r),
		r.UserAgent(),
	)
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for load balancers/proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// LogInfo logs informational messages
func LogInfo(message string) {
	log.Printf("INFO: %s", message)
}

// LogDebug logs debug messages (only in development)
func LogDebug(message string) {
	// You can add environment check here
	log.Printf("DEBUG: %s", message)
}
