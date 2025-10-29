package schemas

import (
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"strings"
)

// SchemaLoader handles loading and parsing of configuration schemas
type SchemaLoader struct {
	jjSchema     *jsonschema.Schema
	claudeSchema *jsonschema.Schema
	miseSchema   *jsonschema.Schema
}

// NewSchemaLoader creates a new schema loader instance
func NewSchemaLoader() (*SchemaLoader, error) {
	loader := &SchemaLoader{}

	// Load jj schema
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft4 // Use draft-04 as specified in schema
	compiler.ExtractAnnotations = true // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("jj.json", strings.NewReader(JJSchema)); err != nil {
		return nil, fmt.Errorf("failed to add jj schema resource: %w", err)
	}

	schema, err := compiler.Compile("jj.json")
	if err != nil {
		// Log the error but continue - schema might still be usable for our purposes
		fmt.Printf("Warning: jj schema compilation had issues: %v\n", err)
		loader.jjSchema = nil
	} else {
		loader.jjSchema = schema
	}

	// Load Claude schema
	claudeCompiler := jsonschema.NewCompiler()
	claudeCompiler.ExtractAnnotations = true
	if err := claudeCompiler.AddResource("claude.json", strings.NewReader(ClaudeSchema)); err != nil {
		return nil, fmt.Errorf("failed to add claude schema resource: %w", err)
	}

	claudeSchema, err := claudeCompiler.Compile("claude.json")
	if err != nil {
		fmt.Printf("Warning: claude schema compilation had issues: %v\n", err)
		loader.claudeSchema = nil
	} else {
		loader.claudeSchema = claudeSchema
	}

	// Load Mise schema
	miseCompiler := jsonschema.NewCompiler()
	miseCompiler.ExtractAnnotations = true
	if err := miseCompiler.AddResource("mise.json", strings.NewReader(MiseJSONSchema)); err != nil {
		return nil, fmt.Errorf("failed to add mise schema resource: %w", err)
	}

	miseSchema, err := miseCompiler.Compile("mise.json")
	if err != nil {
		fmt.Printf("Warning: mise schema compilation had issues: %v\n", err)
		loader.miseSchema = nil
	} else {
		loader.miseSchema = miseSchema
	}

	return loader, nil
}

// GetJJSchema returns the compiled jj schema
func (s *SchemaLoader) GetJJSchema() *jsonschema.Schema {
	return s.jjSchema
}

// GetClaudeSchema returns the compiled claude schema
func (s *SchemaLoader) GetClaudeSchema() *jsonschema.Schema {
	return s.claudeSchema
}

// GetMiseSchema returns the compiled mise schema
func (s *SchemaLoader) GetMiseSchema() *jsonschema.Schema {
	return s.miseSchema
}

// GetJJSchemaRaw returns the raw jj schema JSON
func (s *SchemaLoader) GetJJSchemaRaw() string {
	return JJSchema
}

// NewJJSchemaParserFromCompiled creates a new JSONSchemaParser for jj using the compiled schema
func (s *SchemaLoader) NewJJSchemaParserFromCompiled() (*JSONSchemaParser, error) {
	if s.jjSchema == nil {
		return nil, fmt.Errorf("jj schema not compiled")
	}
	return NewJSONSchemaParser(s.jjSchema), nil
}

// NewClaudeSchemaParserFromCompiled creates a new JSONSchemaParser for claude using the compiled schema
func (s *SchemaLoader) NewClaudeSchemaParserFromCompiled() (*JSONSchemaParser, error) {
	if s.claudeSchema == nil {
		return nil, fmt.Errorf("claude schema not compiled")
	}
	return NewJSONSchemaParser(s.claudeSchema), nil
}

// NewMiseSchemaParserFromCompiled creates a new JSONSchemaParser for mise using the compiled schema
func (s *SchemaLoader) NewMiseSchemaParserFromCompiled() (*JSONSchemaParser, error) {
	if s.miseSchema == nil {
		return nil, fmt.Errorf("mise schema not compiled")
	}
	return NewJSONSchemaParser(s.miseSchema), nil
}

// CompileJJSchema compiles the jj schema and returns a JSONSchemaParser
func CompileJJSchema() (*JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft4
	compiler.ExtractAnnotations = true // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("jj.json", strings.NewReader(JJSchema)); err != nil {
		return nil, fmt.Errorf("failed to add jj schema resource: %w", err)
	}

	schema, err := compiler.Compile("jj.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile jj schema: %w", err)
	}

	return NewJSONSchemaParser(schema), nil
}

// CompileClaudeSchema compiles the claude schema and returns a JSONSchemaParser
func CompileClaudeSchema() (*JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("claude.json", strings.NewReader(ClaudeSchema)); err != nil {
		return nil, fmt.Errorf("failed to add claude schema resource: %w", err)
	}

	schema, err := compiler.Compile("claude.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile claude schema: %w", err)
	}

	return NewJSONSchemaParser(schema), nil
}

// CompileMiseSchema compiles the mise schema and returns a JSONSchemaParser
func CompileMiseSchema() (*JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("mise.json", strings.NewReader(MiseJSONSchema)); err != nil {
		return nil, fmt.Errorf("failed to add mise schema resource: %w", err)
	}

	schema, err := compiler.Compile("mise.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile mise schema: %w", err)
	}

	return NewJSONSchemaParser(schema), nil
}
