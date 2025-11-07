package schemas

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/lib/configschema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SchemaLoader handles loading and parsing of configuration schemas
type SchemaLoader struct {
	jjSchema       *jsonschema.Schema
	claudeSchema   *jsonschema.Schema
	miseSchema     *jsonschema.Schema
	starshipSchema *jsonschema.Schema
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

	// Load Starship schema
	starshipCompiler := jsonschema.NewCompiler()
	starshipCompiler.Draft = jsonschema.Draft2020 // Starship uses JSON Schema Draft 2020-12
	starshipCompiler.ExtractAnnotations = true
	if err := starshipCompiler.AddResource("starship.json", strings.NewReader(StarshipSchema)); err != nil {
		return nil, fmt.Errorf("failed to add starship schema resource: %w", err)
	}

	starshipSchema, err := starshipCompiler.Compile("starship.json")
	if err != nil {
		fmt.Printf("Warning: starship schema compilation had issues: %v\n", err)
		loader.starshipSchema = nil
	} else {
		loader.starshipSchema = starshipSchema
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
func (s *SchemaLoader) NewJJSchemaParserFromCompiled() (*configschema.JSONSchemaParser, error) {
	if s.jjSchema == nil {
		return nil, fmt.Errorf("jj schema not compiled")
	}
	return configschema.NewJSONSchemaParser(s.jjSchema), nil
}

// NewClaudeSchemaParserFromCompiled creates a new JSONSchemaParser for claude using the compiled schema
func (s *SchemaLoader) NewClaudeSchemaParserFromCompiled() (*configschema.JSONSchemaParser, error) {
	if s.claudeSchema == nil {
		return nil, fmt.Errorf("claude schema not compiled")
	}
	return configschema.NewJSONSchemaParser(s.claudeSchema), nil
}

// NewMiseSchemaParserFromCompiled creates a new JSONSchemaParser for mise using the compiled schema
func (s *SchemaLoader) NewMiseSchemaParserFromCompiled() (*configschema.JSONSchemaParser, error) {
	if s.miseSchema == nil {
		return nil, fmt.Errorf("mise schema not compiled")
	}
	return configschema.NewJSONSchemaParser(s.miseSchema), nil
}

// CompileJJSchema compiles the jj schema and returns a JSONSchemaParser
func CompileJJSchema() (*configschema.JSONSchemaParser, error) {
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

	return configschema.NewJSONSchemaParser(schema), nil
}

// CompileClaudeSchema compiles the claude schema and returns a JSONSchemaParser
func CompileClaudeSchema() (*configschema.JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("claude.json", strings.NewReader(ClaudeSchema)); err != nil {
		return nil, fmt.Errorf("failed to add claude schema resource: %w", err)
	}

	schema, err := compiler.Compile("claude.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile claude schema: %w", err)
	}

	return configschema.NewJSONSchemaParser(schema), nil
}

// CompileMiseSchema compiles the mise schema and returns a JSONSchemaParser
func CompileMiseSchema() (*configschema.JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("mise.json", strings.NewReader(MiseJSONSchema)); err != nil {
		return nil, fmt.Errorf("failed to add mise schema resource: %w", err)
	}

	schema, err := compiler.Compile("mise.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile mise schema: %w", err)
	}

	return configschema.NewJSONSchemaParser(schema), nil
}

// GetStarshipSchema returns the compiled starship schema
func (s *SchemaLoader) GetStarshipSchema() *jsonschema.Schema {
	return s.starshipSchema
}

// NewStarshipSchemaParserFromCompiled creates a new JSONSchemaParser for starship using the compiled schema
func (s *SchemaLoader) NewStarshipSchemaParserFromCompiled() (*configschema.JSONSchemaParser, error) {
	if s.starshipSchema == nil {
		return nil, fmt.Errorf("starship schema not compiled")
	}
	return configschema.NewJSONSchemaParser(s.starshipSchema), nil
}

// CompileStarshipSchema compiles the starship schema and returns a JSONSchemaParser
func CompileStarshipSchema() (*configschema.JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020 // Starship uses JSON Schema Draft 2020-12
	compiler.ExtractAnnotations = true    // Enable extraction of title, description, default, etc.

	if err := compiler.AddResource("starship.json", strings.NewReader(StarshipSchema)); err != nil {
		return nil, fmt.Errorf("failed to add starship schema resource: %w", err)
	}

	schema, err := compiler.Compile("starship.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile starship schema: %w", err)
	}

	return configschema.NewJSONSchemaParser(schema), nil
}
