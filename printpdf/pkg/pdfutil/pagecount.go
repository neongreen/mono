package pdfutil

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// CountPages counts the number of pages in a PDF file using pdfcpu library
func CountPages(pdfPath string) (int, error) {
	// Use pdfcpu API to get page count - this is much more reliable than manual parsing
	pageCount, err := api.PageCountFile(pdfPath)
	if err != nil {
		return 0, fmt.Errorf("failed to count pages in PDF: %w", err)
	}

	if pageCount <= 0 {
		return 0, fmt.Errorf("invalid page count: %d", pageCount)
	}

	return pageCount, nil
}
