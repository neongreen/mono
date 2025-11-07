package schemas

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/lib/configschema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// CompileJJSchema compiles the jj schema and returns a JSONSchemaParser
func CompileJJSchema() (*configschema.JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft4
	compiler.ExtractAnnotations = true

	if err := compiler.AddResource("jj.json", strings.NewReader(configschema.JJSchemaLatest())); err != nil {
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
	compiler.ExtractAnnotations = true

	if err := compiler.AddResource("claude.json", strings.NewReader(configschema.ClaudeSchema())); err != nil {
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
	compiler.ExtractAnnotations = true

	if err := compiler.AddResource("mise.json", strings.NewReader(configschema.MiseSchema())); err != nil {
		return nil, fmt.Errorf("failed to add mise schema resource: %w", err)
	}

	schema, err := compiler.Compile("mise.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile mise schema: %w", err)
	}

	return configschema.NewJSONSchemaParser(schema), nil
}

// CompileStarshipSchema compiles the starship schema and returns a JSONSchemaParser
func CompileStarshipSchema() (*configschema.JSONSchemaParser, error) {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	compiler.ExtractAnnotations = true

	if err := compiler.AddResource("starship.json", strings.NewReader(configschema.StarshipSchema())); err != nil {
		return nil, fmt.Errorf("failed to add starship schema resource: %w", err)
	}

	schema, err := compiler.Compile("starship.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile starship schema: %w", err)
	}

	return configschema.NewJSONSchemaParser(schema), nil
}
