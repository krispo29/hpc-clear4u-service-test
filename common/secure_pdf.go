package common

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hpc-express-service/errors"
)

// SecurePDFConfig holds configuration for secure PDF generation
type SecurePDFConfig struct {
	MaxFileSize       int64         // Maximum PDF file size in bytes
	AllowedExtensions []string      // Allowed file extensions
	TempDir           string        // Temporary directory for PDF generation
	CleanupInterval   time.Duration // Interval for cleaning up temporary files
	EnableWatermark   bool          // Enable watermarking
	EnableEncryption  bool          // Enable PDF encryption
	MaxConcurrent     int           // Maximum concurrent PDF generations
}

// DefaultSecurePDFConfig returns default secure PDF configuration
func DefaultSecurePDFConfig() *SecurePDFConfig {
	return &SecurePDFConfig{
		MaxFileSize:       50 * 1024 * 1024, // 50MB
		AllowedExtensions: []string{".pdf"},
		TempDir:           os.TempDir(),
		CleanupInterval:   time.Hour,
		EnableWatermark:   true,
		EnableEncryption:  false, // Disabled by default for compatibility
		MaxConcurrent:     5,
	}
}

// SecurePDFGenerator provides secure PDF generation functionality
type SecurePDFGenerator struct {
	config    *SecurePDFConfig
	semaphore chan struct{} // Semaphore for limiting concurrent operations
}

// NewSecurePDFGenerator creates a new secure PDF generator
func NewSecurePDFGenerator(config *SecurePDFConfig) *SecurePDFGenerator {
	if config == nil {
		config = DefaultSecurePDFConfig()
	}

	generator := &SecurePDFGenerator{
		config:    config,
		semaphore: make(chan struct{}, config.MaxConcurrent),
	}

	// Start cleanup goroutine
	go generator.startCleanup()

	return generator
}

// GenerateSecurePDF generates a PDF with security measures
func (spg *SecurePDFGenerator) GenerateSecurePDF(ctx context.Context, data interface{}, templateName string, userID string) ([]byte, error) {
	// Acquire semaphore to limit concurrent operations
	select {
	case spg.semaphore <- struct{}{}:
		defer func() { <-spg.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Generate unique filename
	filename, err := spg.generateSecureFilename(templateName)
	if err != nil {
		return nil, errors.NewDatabaseError("pdf_generation", fmt.Errorf("failed to generate secure filename: %w", err))
	}

	// Create temporary file path
	tempPath := filepath.Join(spg.config.TempDir, filename)

	// Ensure temp directory exists
	if err := os.MkdirAll(spg.config.TempDir, 0755); err != nil {
		return nil, errors.NewDatabaseError("pdf_generation", fmt.Errorf("failed to create temp directory: %w", err))
	}

	// Generate PDF content (this would call the actual PDF generation logic)
	pdfContent, err := spg.generatePDFContent(data, templateName, userID)
	if err != nil {
		return nil, err
	}

	// Validate PDF size
	if int64(len(pdfContent)) > spg.config.MaxFileSize {
		return nil, errors.NewValidationError("pdf_size", fmt.Sprintf("PDF size exceeds maximum allowed size of %d bytes", spg.config.MaxFileSize))
	}

	// Add watermark if enabled
	if spg.config.EnableWatermark {
		pdfContent, err = spg.addWatermark(pdfContent, userID)
		if err != nil {
			log.Printf("Warning: Failed to add watermark to PDF: %v", err)
			// Continue without watermark rather than failing
		}
	}

	// Encrypt PDF if enabled
	if spg.config.EnableEncryption {
		pdfContent, err = spg.encryptPDF(pdfContent)
		if err != nil {
			log.Printf("Warning: Failed to encrypt PDF: %v", err)
			// Continue without encryption rather than failing
		}
	}

	// Write to temporary file for validation
	if err := spg.writeSecureFile(tempPath, pdfContent); err != nil {
		return nil, err
	}

	// Validate PDF file
	if err := spg.validatePDFFile(tempPath); err != nil {
		os.Remove(tempPath) // Clean up invalid file
		return nil, err
	}

	// Schedule file cleanup
	go spg.scheduleFileCleanup(tempPath, time.Minute*10)

	return pdfContent, nil
}

// generateSecureFilename generates a secure filename with random component
func (spg *SecurePDFGenerator) generateSecureFilename(templateName string) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Create filename with timestamp and random component
	timestamp := time.Now().Format("20060102_150405")
	randomHex := hex.EncodeToString(randomBytes)

	// Sanitize template name
	sanitizedTemplate := spg.sanitizeFilename(templateName)

	filename := fmt.Sprintf("%s_%s_%s.pdf", sanitizedTemplate, timestamp, randomHex[:8])
	return filename, nil
}

// sanitizeFilename removes potentially dangerous characters from filename
func (spg *SecurePDFGenerator) sanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	dangerous := []string{"/", "\\", "..", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	sanitized := filename

	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

// generatePDFContent generates the actual PDF content
func (spg *SecurePDFGenerator) generatePDFContent(data interface{}, templateName string, userID string) ([]byte, error) {
	// This is a placeholder for the actual PDF generation logic
	// In a real implementation, this would use a PDF library like gofpdf or similar

	// For now, return a simple PDF-like content
	content := fmt.Sprintf("%%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n/Contents 4 0 R\n>>\nendobj\n\n4 0 obj\n<<\n/Length 44\n>>\nstream\nBT\n/F1 12 Tf\n100 700 Td\n(Secure PDF - Template: %s) Tj\nET\nendstream\nendobj\n\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000206 00000 n \ntrailer\n<<\n/Size 5\n/Root 1 0 R\n>>\nstartxref\n299\n%%%%EOF", templateName)

	return []byte(content), nil
}

// addWatermark adds a watermark to the PDF
func (spg *SecurePDFGenerator) addWatermark(pdfContent []byte, userID string) ([]byte, error) {
	// This is a placeholder for watermark functionality
	// In a real implementation, this would modify the PDF to add a watermark

	watermarkText := fmt.Sprintf("Generated by: %s at %s", userID, time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("Adding watermark: %s", watermarkText)

	// For now, just return the original content
	return pdfContent, nil
}

// encryptPDF encrypts the PDF content
func (spg *SecurePDFGenerator) encryptPDF(pdfContent []byte) ([]byte, error) {
	// This is a placeholder for PDF encryption
	// In a real implementation, this would encrypt the PDF using a library like qpdf or similar

	log.Printf("Encrypting PDF (placeholder)")

	// For now, just return the original content
	return pdfContent, nil
}

// writeSecureFile writes content to a file securely
func (spg *SecurePDFGenerator) writeSecureFile(filepath string, content []byte) error {
	// Create file with restricted permissions
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return errors.NewDatabaseError("file_write", fmt.Errorf("failed to create file: %w", err))
	}
	defer file.Close()

	// Write content
	if _, err := file.Write(content); err != nil {
		return errors.NewDatabaseError("file_write", fmt.Errorf("failed to write file: %w", err))
	}

	return nil
}

// validatePDFFile validates that the generated file is a valid PDF
func (spg *SecurePDFGenerator) validatePDFFile(filepath string) error {
	// Open file for reading
	file, err := os.Open(filepath)
	if err != nil {
		return errors.NewDatabaseError("file_validation", fmt.Errorf("failed to open file for validation: %w", err))
	}
	defer file.Close()

	// Read first few bytes to check PDF header
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return errors.NewDatabaseError("file_validation", fmt.Errorf("failed to read file header: %w", err))
	}

	// Check PDF magic number
	if !bytes.HasPrefix(header, []byte("%PDF-")) {
		return errors.NewValidationError("pdf_format", "file is not a valid PDF")
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return errors.NewDatabaseError("file_validation", fmt.Errorf("failed to get file info: %w", err))
	}

	// Check file size
	if fileInfo.Size() > spg.config.MaxFileSize {
		return errors.NewValidationError("pdf_size", "PDF file size exceeds maximum allowed size")
	}

	if fileInfo.Size() == 0 {
		return errors.NewValidationError("pdf_size", "PDF file is empty")
	}

	return nil
}

// scheduleFileCleanup schedules a file for cleanup after a delay
func (spg *SecurePDFGenerator) scheduleFileCleanup(filepath string, delay time.Duration) {
	time.AfterFunc(delay, func() {
		if err := os.Remove(filepath); err != nil {
			log.Printf("Warning: Failed to cleanup temporary file %s: %v", filepath, err)
		}
	})
}

// startCleanup starts the periodic cleanup of temporary files
func (spg *SecurePDFGenerator) startCleanup() {
	ticker := time.NewTicker(spg.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		spg.cleanupTempFiles()
	}
}

// cleanupTempFiles removes old temporary files
func (spg *SecurePDFGenerator) cleanupTempFiles() {
	entries, err := os.ReadDir(spg.config.TempDir)
	if err != nil {
		log.Printf("Warning: Failed to read temp directory for cleanup: %v", err)
		return
	}

	cutoff := time.Now().Add(-spg.config.CleanupInterval * 2)

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pdf") {
			filepath := filepath.Join(spg.config.TempDir, entry.Name())

			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				if err := os.Remove(filepath); err != nil {
					log.Printf("Warning: Failed to cleanup old temp file %s: %v", filepath, err)
				}
			}
		}
	}
}

// SecurePDFResponse provides secure PDF response handling
func SecurePDFResponse(w http.ResponseWriter, r *http.Request, pdfData []byte, filename string) error {
	// Validate filename
	if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
		return errors.NewValidationError("filename", "invalid filename")
	}

	// Set security headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Write PDF data
	w.WriteHeader(http.StatusOK)

	// Use io.Copy for efficient streaming
	reader := bytes.NewReader(pdfData)
	if _, err := io.Copy(w, reader); err != nil {
		return errors.NewDatabaseError("pdf_response", fmt.Errorf("failed to write PDF response: %w", err))
	}

	return nil
}

// Global secure PDF generator instance
var GlobalSecurePDFGenerator = NewSecurePDFGenerator(DefaultSecurePDFConfig())
