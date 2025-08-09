package pdf

import (
	"hpc-express-service/cargo_manifest"
	"hpc-express-service/draft_mawb"
)

// InitializePDFServices initializes PDF generation services for all modules
func InitializePDFServices() error {
	// Create a factory function for PDF generator
	createGenerator := func() (interface{}, error) {
		return NewPDFGenerator()
	}

	// Set the PDF generator factory for cargo manifest service
	cargo_manifest.NewPDFGenerator = func() (cargo_manifest.PDFGenerator, error) {
		generator, err := createGenerator()
		if err != nil {
			return nil, err
		}
		return generator.(*PDFGenerator), nil
	}

	// Set the PDF generator factory for draft MAWB service
	draft_mawb.NewPDFGenerator = func() (draft_mawb.PDFGenerator, error) {
		generator, err := createGenerator()
		if err != nil {
			return nil, err
		}
		return generator.(*PDFGenerator), nil
	}

	return nil
}
