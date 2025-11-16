package readability

// CustomEngine wraps the existing Go-based readability implementation
type CustomEngine struct{}

// NewCustomEngine creates a new custom Go engine
func NewCustomEngine() *CustomEngine {
	return &CustomEngine{}
}

// Name returns the engine name
func (e *CustomEngine) Name() string {
	return "custom"
}

// Extract extracts readable content using the custom Go implementation
func (e *CustomEngine) Extract(html []byte, sourceURL string) ([]byte, error) {
	// Use the ExtractReadableContent function from this package
	return ExtractReadableContent(html)
}

// IsAvailable always returns nil since the custom engine is built-in
func (e *CustomEngine) IsAvailable() error {
	return nil
}
