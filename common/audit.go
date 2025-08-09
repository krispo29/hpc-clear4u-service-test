package common

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id"`
	UserID     string                 `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	StatusCode int                    `json:"status_code,omitempty"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Sensitive  bool                   `json:"sensitive"`
}

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	enabled bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(enabled bool) *AuditLogger {
	return &AuditLogger{enabled: enabled}
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event *AuditEvent) {
	if !al.enabled {
		return
	}

	// Mask sensitive data if needed
	if event.Sensitive {
		event = al.maskSensitiveData(event)
	}

	// Convert to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("AUDIT_ERROR: Failed to marshal audit event: %v", err)
		return
	}

	// Log the audit event
	log.Printf("AUDIT: %s", string(eventJSON))
}

// LogHTTPRequest logs an HTTP request audit event
func (al *AuditLogger) LogHTTPRequest(r *http.Request, userID, action, resource, resourceID string, statusCode int, success bool, err error, duration time.Duration, metadata map[string]interface{}) {
	event := &AuditEvent{
		Timestamp:  time.Now(),
		RequestID:  getRequestIDFromContext(r.Context()),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Method:     r.Method,
		Path:       r.URL.Path,
		IPAddress:  getClientIPFromRequest(r),
		UserAgent:  r.UserAgent(),
		StatusCode: statusCode,
		Success:    success,
		Duration:   duration,
		Metadata:   metadata,
		Sensitive:  al.isSensitiveOperation(action, resource),
	}

	if err != nil {
		event.Error = err.Error()
	}

	al.LogEvent(event)
}

// LogStatusChange logs status change operations (sensitive)
func (al *AuditLogger) LogStatusChange(r *http.Request, userID, resource, resourceID, oldStatus, newStatus string, success bool, err error) {
	metadata := map[string]interface{}{
		"old_status": oldStatus,
		"new_status": newStatus,
	}

	al.LogHTTPRequest(r, userID, "status_change", resource, resourceID, 0, success, err, 0, metadata)
}

// LogDataAccess logs data access operations
func (al *AuditLogger) LogDataAccess(r *http.Request, userID, resource, resourceID string, success bool, err error) {
	al.LogHTTPRequest(r, userID, "data_access", resource, resourceID, 0, success, err, 0, nil)
}

// LogDataModification logs data modification operations
func (al *AuditLogger) LogDataModification(r *http.Request, userID, resource, resourceID string, operation string, success bool, err error, changes map[string]interface{}) {
	metadata := map[string]interface{}{
		"operation": operation,
		"changes":   changes,
	}

	al.LogHTTPRequest(r, userID, "data_modification", resource, resourceID, 0, success, err, 0, metadata)
}

// LogPDFGeneration logs PDF generation operations
func (al *AuditLogger) LogPDFGeneration(r *http.Request, userID, resource, resourceID string, success bool, err error, fileSize int64) {
	metadata := map[string]interface{}{
		"file_size": fileSize,
	}

	al.LogHTTPRequest(r, userID, "pdf_generation", resource, resourceID, 0, success, err, 0, metadata)
}

// maskSensitiveData masks sensitive information in audit logs
func (al *AuditLogger) maskSensitiveData(event *AuditEvent) *AuditEvent {
	// Create a copy to avoid modifying the original
	maskedEvent := *event

	// Mask sensitive metadata
	if maskedEvent.Metadata != nil {
		maskedMetadata := make(map[string]interface{})
		for key, value := range maskedEvent.Metadata {
			if al.isSensitiveField(key) {
				maskedMetadata[key] = al.maskValue(value)
			} else {
				maskedMetadata[key] = value
			}
		}
		maskedEvent.Metadata = maskedMetadata
	}

	// Mask user agent if it contains sensitive information
	maskedEvent.UserAgent = al.maskUserAgent(maskedEvent.UserAgent)

	return &maskedEvent
}

// isSensitiveOperation determines if an operation is sensitive
func (al *AuditLogger) isSensitiveOperation(action, resource string) bool {
	sensitiveActions := []string{"status_change", "data_modification", "pdf_generation"}
	for _, sensitiveAction := range sensitiveActions {
		if action == sensitiveAction {
			return true
		}
	}

	sensitiveResources := []string{"cargo_manifest", "draft_mawb", "mawb_info"}
	for _, sensitiveResource := range sensitiveResources {
		if resource == sensitiveResource {
			return true
		}
	}

	return false
}

// isSensitiveField determines if a field contains sensitive data
func (al *AuditLogger) isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"shipper", "consignee", "shipperNameAndAddress", "consigneeNameAndAddress",
		"accountNo", "signatureOfShipper", "signatureOfIssuingCarrier",
		"password", "token", "key", "secret", "credit_card", "ssn",
	}

	for _, sensitiveField := range sensitiveFields {
		if fieldName == sensitiveField {
			return true
		}
	}

	return false
}

// maskValue masks a sensitive value
func (al *AuditLogger) maskValue(value interface{}) string {
	if str, ok := value.(string); ok {
		if len(str) <= 4 {
			return "****"
		}
		// Show first 2 and last 2 characters, mask the rest
		return str[:2] + strings.Repeat("*", len(str)-4) + str[len(str)-2:]
	}
	return "****"
}

// maskUserAgent masks potentially sensitive information in user agent
func (al *AuditLogger) maskUserAgent(userAgent string) string {
	// Keep only the browser/application name, remove version details
	if len(userAgent) > 50 {
		return userAgent[:50] + "..."
	}
	return userAgent
}

// getRequestIDFromContext extracts request ID from context
func getRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value("requestID"); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}

// getClientIPFromRequest extracts client IP from request
func getClientIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if commaIdx := strings.Index(xff, ","); commaIdx > 0 {
			return strings.TrimSpace(xff[:commaIdx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Global audit logger instance
var GlobalAuditLogger = NewAuditLogger(true)

// Helper functions for common audit operations

// AuditCargoManifestAccess logs cargo manifest access
func AuditCargoManifestAccess(r *http.Request, userID, mawbUUID string, success bool, err error) {
	GlobalAuditLogger.LogDataAccess(r, userID, "cargo_manifest", mawbUUID, success, err)
}

// AuditCargoManifestModification logs cargo manifest modification
func AuditCargoManifestModification(r *http.Request, userID, mawbUUID, operation string, success bool, err error, changes map[string]interface{}) {
	GlobalAuditLogger.LogDataModification(r, userID, "cargo_manifest", mawbUUID, operation, success, err, changes)
}

// AuditCargoManifestStatusChange logs cargo manifest status changes
func AuditCargoManifestStatusChange(r *http.Request, userID, mawbUUID, oldStatus, newStatus string, success bool, err error) {
	GlobalAuditLogger.LogStatusChange(r, userID, "cargo_manifest", mawbUUID, oldStatus, newStatus, success, err)
}

// AuditDraftMAWBAccess logs draft MAWB access
func AuditDraftMAWBAccess(r *http.Request, userID, mawbUUID string, success bool, err error) {
	GlobalAuditLogger.LogDataAccess(r, userID, "draft_mawb", mawbUUID, success, err)
}

// AuditDraftMAWBModification logs draft MAWB modification
func AuditDraftMAWBModification(r *http.Request, userID, mawbUUID, operation string, success bool, err error, changes map[string]interface{}) {
	GlobalAuditLogger.LogDataModification(r, userID, "draft_mawb", mawbUUID, operation, success, err, changes)
}

// AuditDraftMAWBStatusChange logs draft MAWB status changes
func AuditDraftMAWBStatusChange(r *http.Request, userID, mawbUUID, oldStatus, newStatus string, success bool, err error) {
	GlobalAuditLogger.LogStatusChange(r, userID, "draft_mawb", mawbUUID, oldStatus, newStatus, success, err)
}

// AuditPDFGeneration logs PDF generation
func AuditPDFGeneration(r *http.Request, userID, resource, resourceID string, success bool, err error, fileSize int64) {
	GlobalAuditLogger.LogPDFGeneration(r, userID, resource, resourceID, success, err, fileSize)
}
