package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"hpc-express-service/constant"
	"hpc-express-service/errors"

	"github.com/go-chi/render"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	MaxRequestSize    int64         // Maximum request body size in bytes
	RateLimitRequests int           // Number of requests allowed per window
	RateLimitWindow   time.Duration // Time window for rate limiting
	EnableRateLimit   bool          // Enable/disable rate limiting
	EnableSizeLimit   bool          // Enable/disable request size limiting
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		RateLimitRequests: 100,              // 100 requests
		RateLimitWindow:   time.Minute,      // per minute
		EnableRateLimit:   true,
		EnableSizeLimit:   true,
	}
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this IP
	requests := rl.requests[ip]

	// Remove old requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests

	return true
}

// cleanup removes old entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2) // Keep some buffer

		for ip, requests := range rl.requests {
			validRequests := make([]time.Time, 0)
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// RateLimitMiddleware provides rate limiting functionality
func RateLimitMiddleware(rateLimiter *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			clientIP := getClientIP(r)

			// Check rate limit
			if !rateLimiter.Allow(clientIP) {
				// Rate limit exceeded
				errResponse := &errors.ErrResponse{
					Err:            fmt.Errorf("rate limit exceeded"),
					HTTPStatusCode: http.StatusTooManyRequests,
					AppCode:        constant.CodeError,
					StatusText:     "Too Many Requests",
					Message:        "Rate limit exceeded. Please try again later.",
				}

				render.Render(w, r, errResponse)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from request headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, char := range xff {
					if char == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return xff[:commaIdx]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Remove server information
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")

		next.ServeHTTP(w, r)
	})
}

// InputValidationMiddleware provides request validation
func InputValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add validation context
		ctx := context.WithValue(r.Context(), "validation_enabled", true)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequestTimeoutMiddleware adds request timeout handling
func RequestTimeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			// Channel to signal completion
			done := make(chan struct{})

			go func() {
				defer close(done)
				next.ServeHTTP(w, r)
			}()

			select {
			case <-done:
				// Request completed normally
				return
			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					errResponse := &errors.ErrResponse{
						Err:            fmt.Errorf("request timeout"),
						HTTPStatusCode: http.StatusRequestTimeout,
						AppCode:        constant.CodeError,
						StatusText:     "Request Timeout",
						Message:        "Request took too long to process",
					}

					render.Render(w, r, errResponse)
				}
				return
			}
		})
	}
}

// ContentTypeValidationMiddleware validates content type for POST/PUT requests
func ContentTypeValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate content type for requests with body
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			contentType := r.Header.Get("Content-Type")

			// Allow JSON and multipart form data
			validContentTypes := []string{
				"application/json",
				"multipart/form-data",
				"application/x-www-form-urlencoded",
			}

			isValid := false
			for _, validType := range validContentTypes {
				if contentType == validType || (len(contentType) > len(validType) && contentType[:len(validType)] == validType) {
					isValid = true
					break
				}
			}

			if !isValid && contentType != "" {
				errResponse := &errors.ErrResponse{
					Err:            fmt.Errorf("unsupported content type: %s", contentType),
					HTTPStatusCode: http.StatusUnsupportedMediaType,
					AppCode:        constant.CodeError,
					StatusText:     "Unsupported Media Type",
					Message:        "Content-Type must be application/json or multipart/form-data",
				}

				render.Render(w, r, errResponse)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// CreateSecurityMiddlewareChain creates a chain of security middleware
func CreateSecurityMiddlewareChain(config *SecurityConfig) []func(http.Handler) http.Handler {
	var middlewares []func(http.Handler) http.Handler

	// Always add security headers
	middlewares = append(middlewares, SecurityHeadersMiddleware)

	// Add request size limiting if enabled
	if config.EnableSizeLimit {
		middlewares = append(middlewares, RequestSizeLimitMiddleware(config.MaxRequestSize))
	}

	// Add rate limiting if enabled
	if config.EnableRateLimit {
		rateLimiter := NewRateLimiter(config.RateLimitRequests, config.RateLimitWindow)
		middlewares = append(middlewares, RateLimitMiddleware(rateLimiter))
	}

	// Add content type validation
	middlewares = append(middlewares, ContentTypeValidationMiddleware)

	// Add input validation context
	middlewares = append(middlewares, InputValidationMiddleware)

	// Add request timeout (30 seconds for API requests)
	middlewares = append(middlewares, RequestTimeoutMiddleware(30*time.Second))

	return middlewares
}
