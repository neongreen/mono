package pathlang

import (
	"context"
	"fmt"
)

// Node represents a domain object in the resolver's graph.
// It is opaque to pathlang; the resolver interprets it.
type Node any

// Resolver defines how to navigate a domain-specific node graph.
// Implementations map path segments to actual domain objects.
type Resolver interface {
	// Root returns the logical root node for path resolution.
	Root(ctx context.Context) (Node, error)

	// Children resolves a segment under a parent node.
	// It should:
	//   - Filter children by segment.Name
	//   - Apply all predicates (implicit AND)
	//   - Return all matching child nodes
	Children(ctx context.Context, parent Node, seg Segment) ([]Node, error)
}

// Eval evaluates a path from the resolver's root.
// Returns all nodes matching the path, or an error.
func Eval(ctx context.Context, r Resolver, p *Path) ([]Node, error) {
	root, err := r.Root(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get root: %w", err)
	}
	return EvalFrom(ctx, r, root, p)
}

// EvalFrom evaluates a path from a specific starting node.
// This is useful for evaluating relative paths or sub-paths.
func EvalFrom(ctx context.Context, r Resolver, start Node, p *Path) ([]Node, error) {
	// Start with the initial node set
	nodes := []Node{start}

	// Apply each segment in sequence
	for _, seg := range p.Segments {
		var nextNodes []Node

		// For each node in the current set, get matching children
		for _, node := range nodes {
			children, err := r.Children(ctx, node, seg)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve segment %q: %w", seg.Name, err)
			}
			nextNodes = append(nextNodes, children...)
		}

		nodes = nextNodes

		// If we have no nodes at this point, short-circuit
		if len(nodes) == 0 {
			return nil, nil
		}
	}

	return nodes, nil
}
