package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"hpc-express-service/constant"
	"hpc-express-service/errors"

	"github.com/go-chi/render"
)

// ErrorRecoveryMiddleware provides panic recovery and error logging
func ErrorRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Printf("PANIC: %v\nStack trace:\n%s", err, debug.Stack())

				// Create a standardized error response
				errResponse := &errors.ErrResponse{
					Err:            fmt.Errorf("internal server error: %v", err),
					HTTPStatusCode: http.StatusInternalServerError,
					AppCode:        constant.CodeError,
					StatusText:     "Internal Server Error",
					Message:        "An unexpected error occurred",
				}

				// Set content type and render error response
				w.Header().Set("Content-Type", "application/json")
				render.Render(w, r, errResponse)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RequestLoggingMiddleware logs requests with error context for debugging
func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Add request ID to context for tracing
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		r = r.WithContext(ctx)

		// Log request start
		log.Printf("REQUEST_START [%s] %s %s %s - User-Agent: %s",
			requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Process request
		next.ServeHTTP(wrappedWriter, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion with status
		if wrappedWriter.statusCode >= 400 {
			log.Printf("REQUEST_ERROR [%s] %s %s - Status: %d, Duration: %v",
				requestID, r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
		} else {
			log.Printf("REQUEST_SUCCESS [%s] %s %s - Status: %d, Duration: %v",
				requestID, r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
		}
	})
}

// ErrorResponseStandardizationMiddleware ensures consistent error response format
func ErrorResponseStandardizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to intercept error responses
		wrappedWriter := &errorStandardizationWriter{
			ResponseWriter: w,
			request:        r,
		}

		next.ServeHTTP(wrappedWriter, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// errorStandardizationWriter wraps http.ResponseWriter to standardize error responses
type errorStandardizationWriter struct {
	http.ResponseWriter
	request *http.Request
}

func (esw *errorStandardizationWriter) WriteHeader(code int) {
	// If it's an error status code and no content-type is set, ensure JSON
	if code >= 400 && esw.Header().Get("Content-Type") == "" {
		esw.Header().Set("Content-Type", "application/json")
	}
	esw.ResponseWriter.WriteHeader(code)
}

// generateRequestID generates a unique request ID for tracing
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// LogError logs errors with context information
func LogError(r *http.Request, err error, message string) {
	requestID := r.Context().Value("requestID")
	if requestID == nil {
		requestID = "unknown"
	}

	log.Printf("ERROR [%s] %s %s - %s: %v",
		requestID, r.Method, r.URL.Path, message, err)
}

// LogErrorWithContext logs errors with additional context
func LogErrorWithContext(r *http.Request, err error, message string, context map[string]interface{}) {
	requestID := r.Context().Value("requestID")
	if requestID == nil {
		requestID = "unknown"
	}

	contextStr := ""
	for key, value := range context {
		contextStr += fmt.Sprintf(" %s=%v", key, value)
	}

	log.Printf("ERROR [%s] %s %s - %s: %v%s",
		requestID, r.Method, r.URL.Path, message, err, contextStr)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(r *http.Request) string {
	if requestID := r.Context().Value("requestID"); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}
