package golden

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neongreen/mono/printpdf/pkg/converter"
	"github.com/neongreen/mono/printpdf/pkg/fetcher"
)

// GoldenTestCase represents a single test case for golden testing
type GoldenTestCase struct {
	Name        string                // Test case name
	Input       string                // Input content (markdown or HTML)
	ContentType string                // fetcher.ContentTypeMarkdown or fetcher.ContentTypeHTML
	Options     converter.PageOptions // Page layout options
	Converters  []string              // List of converters to test ("typst", "prince", "weasyprint")
}

// TestConfig holds configuration for the golden test suite
type TestConfig struct {
	PDFToImageTool string  // Command to convert PDF to images (default: "pdftoppm")
	ImageDiffTool  string  // Command to compare images (default: "compare")
	DiffThreshold  float64 // Threshold for image differences (0.0 to 1.0)
	UpdateGoldens  bool    // Whether to update golden images instead of comparing
	OutputDir      string  // Directory for test outputs
	GoldenDir      string  // Directory for golden reference images
}

// DefaultTestConfig returns a reasonable default configuration
func DefaultTestConfig() TestConfig {
	return TestConfig{
		PDFToImageTool: "pdftoppm",
		ImageDiffTool:  "compare",
		DiffThreshold:  0.01, // 1% difference threshold
		UpdateGoldens:  os.Getenv("UPDATE_GOLDENS") == "1",
		OutputDir:      "testdata/output",
		GoldenDir:      "testdata/golden",
	}
}

// GoldenTestSuite manages the complete golden test suite
type GoldenTestSuite struct {
	config    TestConfig
	testCases []GoldenTestCase
}

// NewGoldenTestSuite creates a new golden test suite
func NewGoldenTestSuite(config TestConfig) *GoldenTestSuite {
	return &GoldenTestSuite{
		config:    config,
		testCases: []GoldenTestCase{},
	}
}

// AddTestCase adds a test case to the suite
func (suite *GoldenTestSuite) AddTestCase(testCase GoldenTestCase) {
	suite.testCases = append(suite.testCases, testCase)
}

// Run executes all test cases in the suite
func (suite *GoldenTestSuite) Run(t *testing.T) {
	// Check prerequisites
	if err := suite.checkPrerequisites(); err != nil {
		t.Fatalf("Prerequisites not met: %v", err)
	}

	// Ensure directories exist
	if err := os.MkdirAll(suite.config.OutputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.MkdirAll(suite.config.GoldenDir, 0755); err != nil {
		t.Fatalf("Failed to create golden directory: %v", err)
	}

	// Run each test case
	for _, testCase := range suite.testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			suite.runTestCase(t, testCase)
		})
	}
}

// checkPrerequisites verifies that required tools are available
func (suite *GoldenTestSuite) checkPrerequisites() error {
	// Check PDF to image conversion tool
	if _, err := exec.LookPath(suite.config.PDFToImageTool); err != nil {
		return fmt.Errorf("PDF to image tool not found: %s (try: apt-get install poppler-utils)", suite.config.PDFToImageTool)
	}

	// Check image diff tool
	if _, err := exec.LookPath(suite.config.ImageDiffTool); err != nil {
		return fmt.Errorf("Image diff tool not found: %s (try: apt-get install imagemagick)", suite.config.ImageDiffTool)
	}

	return nil
}

// runTestCase executes a single test case
func (suite *GoldenTestSuite) runTestCase(t *testing.T, testCase GoldenTestCase) {
	for _, converterName := range testCase.Converters {
		t.Run(converterName, func(t *testing.T) {
			// Skip if converter is not available
			conv := suite.getConverter(converterName)
			if conv == nil {
				t.Skipf("Converter %s not available", converterName)
				return
			}

			// Generate PDF
			pdfPath := filepath.Join(suite.config.OutputDir, fmt.Sprintf("%s-%s.pdf", testCase.Name, converterName))
			if err := conv.Convert([]byte(testCase.Input), testCase.ContentType, pdfPath, testCase.Options); err != nil {
				t.Fatalf("Failed to convert with %s: %v", converterName, err)
			}

			// Convert PDF to images
			imagesDir := filepath.Join(suite.config.OutputDir, fmt.Sprintf("%s-%s-images", testCase.Name, converterName))
			if err := suite.pdfToImages(pdfPath, imagesDir); err != nil {
				t.Fatalf("Failed to convert PDF to images: %v", err)
			}

			// Compare with golden images or update goldens
			goldenImagesDir := filepath.Join(suite.config.GoldenDir, fmt.Sprintf("%s-%s-images", testCase.Name, converterName))

			if suite.config.UpdateGoldens {
				suite.updateGoldenImages(t, imagesDir, goldenImagesDir)
			} else {
				suite.compareWithGoldens(t, imagesDir, goldenImagesDir)
			}
		})
	}
}

// getConverter returns a converter instance by name, or nil if not available
func (suite *GoldenTestSuite) getConverter(name string) converter.Converter {
	switch strings.ToLower(name) {
	case "typst":
		// Check if typst is available
		conv := converter.NewTypstConverter()
		testErr := conv.Convert([]byte("# Test"), fetcher.ContentTypeMarkdown, "/tmp/test-typst.pdf", converter.PageOptions{})
		os.Remove("/tmp/test-typst.pdf")
		if testErr != nil && strings.Contains(testErr.Error(), "failed to get typst") {
			return nil // Typst not available
		}
		return conv
	case "prince":
		// Check if prince is available
		conv := converter.NewPrinceConverter()
		testErr := conv.Convert([]byte("# Test"), fetcher.ContentTypeMarkdown, "/tmp/test-prince.pdf", converter.PageOptions{})
		os.Remove("/tmp/test-prince.pdf")
		if testErr != nil && strings.Contains(testErr.Error(), "prince not found") {
			return nil // Prince not available
		}
		return conv
	case "weasyprint":
		conv := converter.NewWeasyPrintConverter()
		testErr := conv.Convert([]byte("# Test"), fetcher.ContentTypeMarkdown, "/tmp/test-weasy.pdf", converter.PageOptions{})
		os.Remove("/tmp/test-weasy.pdf")
		if testErr != nil && strings.Contains(testErr.Error(), "weasyprint not found") {
			return nil // WeasyPrint not available
		}
		return conv
	default:
		return nil
	}
}

// pdfToImages converts a PDF file to a series of PNG images
func (suite *GoldenTestSuite) pdfToImages(pdfPath, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	// Use pdftoppm to convert PDF to images
	// pdftoppm -png -r 150 input.pdf output_prefix
	outputPrefix := filepath.Join(outputDir, "page")
	cmd := exec.Command(suite.config.PDFToImageTool, "-png", "-r", "150", pdfPath, outputPrefix)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pdftoppm failed: %w\nStderr: %s", err, stderr.String())
	}

	return nil
}

// updateGoldenImages replaces golden reference images with current output
func (suite *GoldenTestSuite) updateGoldenImages(t *testing.T, currentDir, goldenDir string) {
	// Remove existing golden images
	os.RemoveAll(goldenDir)

	// Copy current images to golden directory
	if err := suite.copyDir(currentDir, goldenDir); err != nil {
		t.Fatalf("Failed to update golden images: %v", err)
	}

	t.Logf("Updated golden images in %s", goldenDir)
}

// compareWithGoldens compares current images with golden reference images
func (suite *GoldenTestSuite) compareWithGoldens(t *testing.T, currentDir, goldenDir string) {
	// Check if golden directory exists
	if _, err := os.Stat(goldenDir); os.IsNotExist(err) {
		t.Fatalf("Golden directory does not exist: %s\nRun tests with UPDATE_GOLDENS=1 to create initial goldens", goldenDir)
	}

	// Get list of current images
	currentImages, err := filepath.Glob(filepath.Join(currentDir, "*.png"))
	if err != nil {
		t.Fatalf("Failed to list current images: %v", err)
	}

	if len(currentImages) == 0 {
		t.Fatalf("No images generated in %s", currentDir)
	}

	// Compare each image
	for _, currentImage := range currentImages {
		imageName := filepath.Base(currentImage)
		goldenImage := filepath.Join(goldenDir, imageName)

		if _, err := os.Stat(goldenImage); os.IsNotExist(err) {
			t.Errorf("Golden image missing: %s", goldenImage)
			continue
		}

		if err := suite.compareImages(currentImage, goldenImage); err != nil {
			t.Errorf("Image comparison failed for %s: %v", imageName, err)
		}
	}
}

// compareImages compares two images and returns an error if they differ significantly
func (suite *GoldenTestSuite) compareImages(image1, image2 string) error {
	// Use ImageMagick's compare tool
	// compare -metric AE image1.png image2.png null: 2>&1
	cmd := exec.Command(suite.config.ImageDiffTool, "-metric", "AE", image1, image2, "null:")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	_ = cmd.Run() // compare returns exit code 1 if images differ, but that's expected
	output := stderr.String()

	// We only care about the metric output
	if output == "0" {
		return nil // Images are identical
	}

	// For now, just report the difference - we could parse the metric and apply threshold
	return fmt.Errorf("images differ (metric: %s)", strings.TrimSpace(output))
}

// copyDir recursively copies a directory
func (suite *GoldenTestSuite) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the target path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = srcFile.WriteTo(dstFile)
		return err
	})
}
