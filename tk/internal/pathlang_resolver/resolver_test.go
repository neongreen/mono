package pathlang_resolver_test

import (
	"context"
	"testing"

	"github.com/neongreen/mono/lib/pathlang"
	"github.com/neongreen/mono/tk/internal/pathlang_resolver"
)

// TestResolverInterfaceImplementation tests that TkResolver implements pathlang.Resolver
func TestResolverInterfaceImplementation(t *testing.T) {
	// This test ensures that TkResolver implements the Resolver interface
	// It will fail to compile if the interface is not properly implemented
	var _ pathlang.Resolver = (*pathlang_resolver.TkResolver)(nil)
}

// TestNodeTypes tests the node type constants
func TestNodeTypes(t *testing.T) {
	tests := []struct {
		nodeType pathlang_resolver.NodeType
		expected string
	}{
		{pathlang_resolver.NodeTypeRoot, "root"},
		{pathlang_resolver.NodeTypeProject, "project"},
		{pathlang_resolver.NodeTypeTask, "task"},
		{pathlang_resolver.NodeTypeTasks, "tasks"},
		{pathlang_resolver.NodeTypeSubtasks, "subtasks"},
		{pathlang_resolver.NodeTypeBlockers, "blockers"},
		{pathlang_resolver.NodeTypeNotes, "notes"},
		{pathlang_resolver.NodeTypeRelations, "relations"},
		{pathlang_resolver.NodeTypeJSON, "json"},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			if string(tt.nodeType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.nodeType)
			}
		})
	}
}

// TestRootNode tests that Root returns a valid root node
func TestRootNode(t *testing.T) {
	// Create a nil resolver just to test the structure
	// In a real scenario, this would be created with proper db and reducer
	resolver := &pathlang_resolver.TkResolver{}

	ctx := context.Background()
	root, err := resolver.Root(ctx)
	if err != nil {
		t.Fatalf("Root() failed: %v", err)
	}

	node, ok := root.(*pathlang_resolver.Node)
	if !ok {
		t.Fatal("Root() did not return *Node")
	}

	if node.Type != pathlang_resolver.NodeTypeRoot {
		t.Errorf("expected root node type, got %v", node.Type)
	}
}
