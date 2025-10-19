package schemas

import (
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"strings"
)

// SchemaLoader handles loading and parsing of configuration schemas
type SchemaLoader struct {
	jjSchema *jsonschema.Schema
}

// NewSchemaLoader creates a new schema loader instance
func NewSchemaLoader() (*SchemaLoader, error) {
	loader := &SchemaLoader{}

	// Load jj schema - disable validation since the upstream schema has some issues
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft4 // Use draft-04 as specified in schema

	if err := compiler.AddResource("jj.json", strings.NewReader(JJSchema)); err != nil {
		return nil, fmt.Errorf("failed to add jj schema resource: %w", err)
	}

	schema, err := compiler.Compile("jj.json")
	if err != nil {
		// Log the error but continue - schema might still be usable for our purposes
		fmt.Printf("Warning: jj schema compilation had issues: %v\n", err)
		// For now, we'll work without the compiled schema and parse manually
		loader.jjSchema = nil
	} else {
		loader.jjSchema = schema
	}

	return loader, nil
}

// GetJJSchema returns the compiled jj schema
func (s *SchemaLoader) GetJJSchema() *jsonschema.Schema {
	return s.jjSchema
}

// GetJJSchemaRaw returns the raw jj schema JSON
func (s *SchemaLoader) GetJJSchemaRaw() string {
	return JJSchema
}
