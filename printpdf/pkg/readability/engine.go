package readability

import "fmt"

// Engine defines the interface for readability extraction engines
type Engine interface {
	// Name returns the engine name
	Name() string

	// Extract extracts readable content from HTML
	Extract(html []byte, sourceURL string) ([]byte, error)

	// IsAvailable checks if the engine is available for use
	IsAvailable() error
}

// EngineType represents different types of engines
type EngineType string

const (
	EngineTypeMozilla   EngineType = "mozilla"
	EngineTypeDefuddle  EngineType = "defuddle"
	EngineTypePostlight EngineType = "postlight"
	EngineTypePureMD    EngineType = "pure-md"
	EngineTypeJina      EngineType = "jina"
	EngineTypeCustom    EngineType = "custom"
)

// Registry manages available readability engines
type Registry struct {
	engines map[EngineType]Engine
}

// NewRegistry creates a new engine registry
func NewRegistry() *Registry {
	return &Registry{
		engines: make(map[EngineType]Engine),
	}
}

// Register adds an engine to the registry
func (r *Registry) Register(engineType EngineType, engine Engine) {
	r.engines[engineType] = engine
}

// Get retrieves an engine by type
func (r *Registry) Get(engineType EngineType) (Engine, error) {
	engine, ok := r.engines[engineType]
	if !ok {
		return nil, fmt.Errorf("unknown readability engine: %s", engineType)
	}
	return engine, nil
}

// Available returns list of available engine types
func (r *Registry) Available() []EngineType {
	var available []EngineType
	for engineType, engine := range r.engines {
		if engine.IsAvailable() == nil {
			available = append(available, engineType)
		}
	}
	return available
}

// All returns all registered engine types
func (r *Registry) All() []EngineType {
	var all []EngineType
	for engineType := range r.engines {
		all = append(all, engineType)
	}
	return all
}
