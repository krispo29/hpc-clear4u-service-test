package pdf

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"hpc-express-service/cargo_manifest"
	"hpc-express-service/common"
	"hpc-express-service/draft_mawb"

	"github.com/jung-kurt/gofpdf"
)

// PDFGenerator handles PDF generation for various document types with caching
type PDFGenerator struct {
	fonts         map[string][]byte
	fontMutex     sync.RWMutex
	templateCache *common.PDFTemplateCache
	fontCache     *common.PDFTemplateCache
}

// FontConfig holds font configuration
type FontConfig struct {
	Name string
	Path string
}

// NewPDFGenerator creates a new PDF generator instance with font loading and caching
func NewPDFGenerator() (*PDFGenerator, error) {
	// Create cache instances
	memoryCache := common.NewMemoryCache(1000)       // Cache up to 1000 items
	memoryCache.StartCleanupRoutine(5 * time.Minute) // Cleanup every 5 minutes

	templateCache := common.NewPDFTemplateCache(memoryCache, 30*time.Minute) // 30 min TTL for templates
	fontCache := common.NewPDFTemplateCache(memoryCache, 60*time.Minute)     // 60 min TTL for fonts

	generator := &PDFGenerator{
		fonts:         make(map[string][]byte),
		templateCache: templateCache,
		fontCache:     fontCache,
	}

	// Load fonts using existing patterns with caching
	if err := generator.loadFonts(); err != nil {
		return nil, fmt.Errorf("failed to load fonts: %w", err)
	}

	return generator, nil
}

// loadFonts loads all required fonts following existing patterns with caching
func (g *PDFGenerator) loadFonts() error {
	fontConfigs := []FontConfig{
		{"THSarabunNew", "assets/THSarabunNew.ttf"},
		{"THSarabunNew Bold", "assets/THSarabunNew Bold.ttf"},
		{"THSarabunNew BoldItalic", "assets/THSarabunNew BoldItalic.ttf"},
		{"THSarabunNew Italic", "assets/THSarabunNew Italic.ttf"},
		{"Pridi-Regular", "assets/Pridi-Regular.ttf"},
		{"Pridi-Bold", "assets/Pridi-Bold.ttf"},
		{"Pridi-Light", "assets/Pridi-Light.ttf"},
	}

	g.fontMutex.Lock()
	defer g.fontMutex.Unlock()

	ctx := context.Background()

	for _, config := range fontConfigs {
		// Try to get font from cache first
		if cachedFont, exists := g.fontCache.GetFont(ctx, config.Name); exists {
			g.fonts[config.Name] = cachedFont
			continue
		}

		// Load font from file if not in cache
		fontData, err := os.ReadFile(config.Path)
		if err != nil {
			log.Printf("Warning: Failed to load font %s from %s: %v", config.Name, config.Path, err)
			continue
		}

		// Store in both memory and cache
		g.fonts[config.Name] = fontData
		if err := g.fontCache.SetFont(ctx, config.Name, fontData); err != nil {
			log.Printf("Warning: Failed to cache font %s: %v", config.Name, err)
		}
	}

	// Ensure we have at least the basic Thai font
	if _, exists := g.fonts["THSarabunNew"]; !exists {
		return fmt.Errorf("required font THSarabunNew not found")
	}

	return nil
}

// addFontsToPDF adds all loaded fonts to the PDF document
func (g *PDFGenerator) addFontsToPDF(pdf *gofpdf.Fpdf) {
	g.fontMutex.RLock()
	defer g.fontMutex.RUnlock()

	for fontName, fontData := range g.fonts {
		pdf.AddUTF8FontFromBytes(fontName, "", fontData)
	}
}

// GenerateCargoManifestPDF generates a PDF document for cargo manifest
func (g *PDFGenerator) GenerateCargoManifestPDF(manifest *cargo_manifest.CargoManifest) ([]byte, error) {
	if manifest == nil {
		return nil, fmt.Errorf("manifest cannot be nil")
	}

	// Create new PDF document
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        gofpdf.PageSizeA4,
	})

	// Add fonts to PDF
	g.addFontsToPDF(pdf)

	// Add page
	pdf.AddPage()
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)

	// Generate cargo manifest content
	if err := g.generateCargoManifestContent(pdf, manifest); err != nil {
		return nil, fmt.Errorf("failed to generate cargo manifest content: %w", err)
	}

	// Output PDF to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateDraftMAWBPDF generates a PDF document for draft MAWB with complex layout
func (g *PDFGenerator) GenerateDraftMAWBPDF(draftMAWB *draft_mawb.DraftMAWB) ([]byte, error) {
	if draftMAWB == nil {
		return nil, fmt.Errorf("draft MAWB cannot be nil")
	}

	// Create new PDF document
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        gofpdf.PageSizeA4,
	})

	// Add fonts to PDF
	g.addFontsToPDF(pdf)

	// Add page
	pdf.AddPage()
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(true, 0)

	// Generate draft MAWB content using existing patterns
	if err := g.generateDraftMAWBContent(pdf, draftMAWB); err != nil {
		return nil, fmt.Errorf("failed to generate draft MAWB content: %w", err)
	}

	// Output PDF to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// generateCargoManifestContent creates the cargo manifest PDF content
func (g *PDFGenerator) generateCargoManifestContent(pdf *gofpdf.Fpdf, manifest *cargo_manifest.CargoManifest) error {
	// Set default font
	pdf.SetFont("THSarabunNew Bold", "", 16)

	// Title
	pdf.SetXY(10, 20)
	pdf.CellFormat(190, 10, "CARGO MANIFEST", "0", 1, "C", false, 0, "")

	// MAWB Information Section
	pdf.SetFont("THSarabunNew Bold", "", 12)
	pdf.SetXY(10, 35)
	pdf.CellFormat(190, 8, "MAWB INFORMATION", "0", 1, "L", false, 0, "")

	pdf.SetFont("THSarabunNew", "", 10)
	y := 45.0

	// MAWB details
	fields := []struct {
		label string
		value string
	}{
		{"MAWB Number:", manifest.MAWBNumber},
		{"Port of Discharge:", manifest.PortOfDischarge},
		{"Flight No:", manifest.FlightNo},
		{"Freight Date:", manifest.FreightDate},
		{"Shipper:", manifest.Shipper},
		{"Consignee:", manifest.Consignee},
		{"Total CTN:", manifest.TotalCtn},
		{"Transshipment:", manifest.Transshipment},
		{"Status:", manifest.Status},
	}

	for _, field := range fields {
		pdf.SetXY(10, y)
		pdf.SetFont("THSarabunNew Bold", "", 10)
		pdf.CellFormat(50, 6, field.label, "0", 0, "L", false, 0, "")
		pdf.SetFont("THSarabunNew", "", 10)
		pdf.CellFormat(140, 6, field.value, "0", 1, "L", false, 0, "")
		y += 7
	}

	// Items Section
	y += 10
	pdf.SetFont("THSarabunNew Bold", "", 12)
	pdf.SetXY(10, y)
	pdf.CellFormat(190, 8, "CARGO MANIFEST ITEMS", "0", 1, "L", false, 0, "")

	y += 10

	// Items table header
	pdf.SetFont("THSarabunNew Bold", "", 9)
	pdf.SetXY(10, y)

	headers := []struct {
		text  string
		width float64
	}{
		{"HAWB No", 25},
		{"Pkgs", 15},
		{"Gross Weight", 25},
		{"Destination", 20},
		{"Commodity", 30},
		{"Shipper", 35},
		{"Consignee", 35},
	}

	x := 10.0
	for _, header := range headers {
		pdf.SetXY(x, y)
		pdf.CellFormat(header.width, 8, header.text, "1", 0, "C", false, 0, "")
		x += header.width
	}

	y += 8

	// Items data
	pdf.SetFont("THSarabunNew", "", 8)
	for _, item := range manifest.Items {
		if y > 250 { // Check if we need a new page
			pdf.AddPage()
			y = 20
		}

		x = 10.0
		itemData := []struct {
			text  string
			width float64
		}{
			{item.HAWBNo, 25},
			{item.Pkgs, 15},
			{item.GrossWeight, 25},
			{item.Destination, 20},
			{item.Commodity, 30},
			{g.truncateText(item.ShipperNameAndAddress, 25), 35},
			{g.truncateText(item.ConsigneeNameAndAddress, 25), 35},
		}

		for _, data := range itemData {
			pdf.SetXY(x, y)
			pdf.CellFormat(data.width, 6, data.text, "1", 0, "L", false, 0, "")
			x += data.width
		}

		y += 6
	}

	// Footer
	y += 20
	pdf.SetFont("THSarabunNew", "", 8)
	pdf.SetXY(10, y)
	pdf.CellFormat(190, 6, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")), "0", 1, "R", false, 0, "")

	return nil
}

// generateDraftMAWBContent creates the draft MAWB PDF content using existing patterns
func (g *PDFGenerator) generateDraftMAWBContent(pdf *gofpdf.Fpdf, draftMAWB *draft_mawb.DraftMAWB) error {
	// Set default font
	pdf.SetFont("THSarabunNew Bold", "", 10)

	// Add background image if available
	width, height := pdf.GetPageSize()

	// Try to add background image (non-critical if it fails)
	pdf.ImageOptions(
		"assets/bg-mawb.png",
		0, 0,
		width, height,
		false,
		gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
		0,
		"",
	)

	// MAWB and HAWB numbers (following existing pattern)
	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(7, 4)
	pdf.MultiCell(51, 5, draftMAWB.MAWB, "0", "L", false)

	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(143, 4)
	pdf.MultiCell(51, 5, draftMAWB.HAWB, "0", "C", false)

	// Shipper information
	pdf.SetFont("THSarabunNew Bold", "", 10)
	pdf.SetXY(9, 19)
	pdf.MultiCell(89, 3, draftMAWB.ShipperNameAndAddress, "0", "LT", false)

	// Airline information
	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(118, 13)
	pdf.MultiCell(77, 20, draftMAWB.AirlineName, "0", "C", false)

	// Consignee information
	pdf.SetFont("THSarabunNew Bold", "", 10)
	pdf.SetXY(9, 45)
	pdf.MultiCell(89, 3, draftMAWB.ConsigneeNameAndAddress, "0", "LT", false)

	// Issuing carrier agent
	pdf.SetFont("THSarabunNew Bold", "", 14)
	pdf.SetXY(8, 65)
	pdf.CellFormat(89, 15, draftMAWB.IssuingCarrierAgentNameAndCity, "0", 0, "C", false, 0, "")

	// Accounting information
	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(98, 67)
	pdf.MultiCell(97, 6, draftMAWB.AccountingInformation, "0", "C", false)

	// Agent IATA code and account
	pdf.SetFont("THSarabunNew Bold", "", 10)
	pdf.SetXY(8, 84)
	pdf.CellFormat(44, 6, draftMAWB.AgentIATACode, "0", 0, "L", false, 0, "")

	pdf.SetXY(53, 84)
	pdf.MultiCell(44, 6, draftMAWB.AccountNo, "0", "L", false)

	// Airport information
	pdf.SetXY(8, 92)
	pdf.MultiCell(89, 6, draftMAWB.AirportOfDeparture, "0", "C", false)

	pdf.SetXY(97, 92)
	pdf.MultiCell(35, 6, draftMAWB.ReferenceNumber, "0", "C", false)

	// Flight information
	pdf.SetXY(132, 92)
	pdf.MultiCell(30, 6, draftMAWB.FlightNo, "0", "C", false)

	// Routing information
	pdf.SetXY(8, 100)
	pdf.MultiCell(30, 6, draftMAWB.To1, "0", "C", false)

	pdf.SetXY(38, 100)
	pdf.MultiCell(30, 6, draftMAWB.ByFirstCarrier, "0", "C", false)

	pdf.SetXY(68, 100)
	pdf.MultiCell(30, 6, draftMAWB.To2, "0", "C", false)

	pdf.SetXY(98, 100)
	pdf.MultiCell(30, 6, draftMAWB.By2, "0", "C", false)

	pdf.SetXY(128, 100)
	pdf.MultiCell(30, 6, draftMAWB.To3, "0", "C", false)

	pdf.SetXY(158, 100)
	pdf.MultiCell(30, 6, draftMAWB.By3, "0", "C", false)

	// Currency and charges
	pdf.SetXY(8, 108)
	pdf.MultiCell(20, 6, draftMAWB.Currency, "0", "C", false)

	pdf.SetXY(28, 108)
	pdf.MultiCell(20, 6, draftMAWB.ChgsCode, "0", "C", false)

	// Weight/Valuation charges
	pdf.SetXY(48, 108)
	pdf.MultiCell(25, 6, draftMAWB.WtValPPD, "0", "C", false)

	pdf.SetXY(73, 108)
	pdf.MultiCell(25, 6, draftMAWB.WtValColl, "0", "C", false)

	pdf.SetXY(98, 108)
	pdf.MultiCell(25, 6, draftMAWB.OtherPPD, "0", "C", false)

	pdf.SetXY(123, 108)
	pdf.MultiCell(25, 6, draftMAWB.OtherColl, "0", "C", false)

	// Declared values
	pdf.SetXY(148, 108)
	pdf.MultiCell(25, 6, draftMAWB.DeclaredValueCarriage, "0", "C", false)

	pdf.SetXY(173, 108)
	pdf.MultiCell(25, 6, draftMAWB.DeclaredValueCustoms, "0", "C", false)

	// Airport of destination
	pdf.SetXY(8, 116)
	pdf.MultiCell(89, 6, draftMAWB.AirportOfDestination, "0", "C", false)

	// Flight date
	if draftMAWB.FlightDate != nil {
		pdf.SetXY(97, 116)
		pdf.MultiCell(35, 6, draftMAWB.FlightDate.Format("2006-01-02"), "0", "C", false)
	}

	// Insurance amount
	pdf.SetXY(132, 116)
	pdf.MultiCell(28, 7, fmt.Sprintf("%.2f", draftMAWB.InsuranceAmount), "0", "C", false)

	// Handling information
	pdf.SetFont("THSarabunNew Bold", "", 14)
	pdf.SetXY(9, 119)
	pdf.MultiCell(156, 12, draftMAWB.HandlingInformation, "0", "L", false)

	// SCI
	pdf.SetXY(165, 119)
	pdf.MultiCell(30, 7, draftMAWB.SCI, "0", "L", false)

	// Items section
	pdf.SetFont("THSarabunNew Bold", "", 10)
	startX := 8.0
	startY := 143.0
	currentY := startY

	// Items table headers
	pdf.SetXY(startX, currentY)
	pdf.CellFormat(15, 6, "Pieces", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 6, "Gross Wt", "1", 0, "C", false, 0, "")
	pdf.CellFormat(15, 6, "Kg/Lb", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 6, "Rate Class", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 6, "Chg Wt", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 6, "Rate", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 6, "Total", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 6, "Nature and Quantity", "1", 0, "C", false, 0, "")

	currentY += 6

	// Items data
	pdf.SetFont("THSarabunNew", "", 9)
	for _, item := range draftMAWB.Items {
		if currentY > 220 { // Check if we need a new page
			pdf.AddPage()
			currentY = 20
		}

		pdf.SetXY(startX, currentY)
		pdf.CellFormat(15, 6, item.PiecesRCP, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, item.GrossWeight, "1", 0, "C", false, 0, "")
		pdf.CellFormat(15, 6, item.KgLb, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, item.RateClass, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", item.ChargeableWeight), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", item.RateCharge), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", item.Total), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 6, g.truncateText(item.NatureAndQuantity, 35), "1", 0, "L", false, 0, "")

		currentY += 6
	}

	// Totals section
	currentY += 10
	pdf.SetFont("THSarabunNew Bold", "", 12)
	pdf.SetXY(startX, currentY)
	pdf.CellFormat(100, 8, "TOTALS", "0", 1, "L", false, 0, "")

	currentY += 10
	pdf.SetFont("THSarabunNew", "", 10)

	totals := []struct {
		label string
		value string
	}{
		{"Total Pieces:", fmt.Sprintf("%d", draftMAWB.TotalNoOfPieces)},
		{"Total Gross Weight:", fmt.Sprintf("%.2f kg", draftMAWB.TotalGrossWeight)},
		{"Total Chargeable Weight:", fmt.Sprintf("%.2f", draftMAWB.TotalChargeableWeight)},
		{"Total Rate Charge:", fmt.Sprintf("%.2f", draftMAWB.TotalRateCharge)},
		{"Total Amount:", fmt.Sprintf("%.2f", draftMAWB.TotalAmount)},
	}

	for _, total := range totals {
		pdf.SetXY(startX, currentY)
		pdf.SetFont("THSarabunNew Bold", "", 10)
		pdf.CellFormat(50, 6, total.label, "0", 0, "L", false, 0, "")
		pdf.SetFont("THSarabunNew", "", 10)
		pdf.CellFormat(50, 6, total.value, "0", 1, "L", false, 0, "")
		currentY += 7
	}

	// Signatures section
	currentY += 10
	pdf.SetFont("THSarabunNew Bold", "", 14)
	pdf.SetXY(82, currentY)
	pdf.MultiCell(120, 5, draftMAWB.SignatureOfShipper, "0", "C", false)

	currentY += 15
	pdf.SetXY(82, currentY)
	pdf.MultiCell(120, 5, draftMAWB.SignatureOfIssuingCarrier, "0", "C", false)

	// Final MAWB number at bottom
	pdf.SetFont("THSarabunNew Bold", "", 18)
	pdf.SetXY(136, 280)
	pdf.MultiCell(51, 3, draftMAWB.MAWB, "0", "L", false)

	return nil
}

// truncateText truncates text to fit within specified character limit
func (g *PDFGenerator) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}
