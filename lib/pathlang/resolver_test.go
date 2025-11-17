package pathlang

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// testNode represents a node in our test domain.
type testNode struct {
	kind     string
	fields   map[string]string
	children []*testNode
}

// testResolver implements Resolver for testing.
type testResolver struct {
	root *testNode
}

func (r *testResolver) Root(ctx context.Context) (Node, error) {
	return r.root, nil
}

func (r *testResolver) Children(ctx context.Context, parent Node, seg Segment) ([]Node, error) {
	node, ok := parent.(*testNode)
	if !ok {
		return nil, fmt.Errorf("invalid node type")
	}

	var matches []Node

	for _, child := range node.children {
		// Check if the segment name matches the child's kind
		if child.kind != seg.Name {
			continue
		}

		// Check all predicates
		if !matchesPredicates(child, seg.Predicates) {
			continue
		}

		matches = append(matches, child)
	}

	return matches, nil
}

// matchesPredicates checks if a node satisfies all predicates (implicit AND).
func matchesPredicates(node *testNode, preds []Predicate) bool {
	for _, pred := range preds {
		value, exists := node.fields[pred.Field]
		if !exists {
			return false
		}

		switch pred.Op {
		case OpEq:
			if value != pred.Value {
				return false
			}
		case OpNotEq:
			if value == pred.Value {
				return false
			}
		case OpMatch:
			// Simple substring match
			if !strings.Contains(strings.ToLower(value), strings.ToLower(pred.Value)) {
				return false
			}
		}
	}
	return true
}

func TestEval(t *testing.T) {
	// Build a test node hierarchy
	//   root
	//     projects (name=foo)
	//       tasks (status=open, assignee=alice)
	//       tasks (status=closed, assignee=bob)
	//     projects (name=bar)
	//       tasks (status=open, assignee=bob)

	root := &testNode{
		kind: "root",
		children: []*testNode{
			{
				kind: "projects",
				fields: map[string]string{
					"name": "foo",
				},
				children: []*testNode{
					{
						kind: "tasks",
						fields: map[string]string{
							"status":   "open",
							"assignee": "alice",
						},
					},
					{
						kind: "tasks",
						fields: map[string]string{
							"status":   "closed",
							"assignee": "bob",
						},
					},
				},
			},
			{
				kind: "projects",
				fields: map[string]string{
					"name": "bar",
				},
				children: []*testNode{
					{
						kind: "tasks",
						fields: map[string]string{
							"status":   "open",
							"assignee": "bob",
						},
					},
				},
			},
		},
	}

	resolver := &testResolver{root: root}
	ctx := context.Background()

	tests := []struct {
		name      string
		path      string
		wantCount int
	}{
		{
			name:      "root path",
			path:      "/",
			wantCount: 1, // just the root
		},
		{
			name:      "all projects",
			path:      "/projects",
			wantCount: 2,
		},
		{
			name:      "project by name",
			path:      "/projects[name=foo]",
			wantCount: 1,
		},
		{
			name:      "all tasks under all projects",
			path:      "/projects/tasks",
			wantCount: 3,
		},
		{
			name:      "tasks under specific project",
			path:      "/projects[name=foo]/tasks",
			wantCount: 2,
		},
		{
			name:      "open tasks under specific project",
			path:      "/projects[name=foo]/tasks[status=open]",
			wantCount: 1,
		},
		{
			name:      "tasks by assignee",
			path:      "/projects/tasks[assignee=bob]",
			wantCount: 2,
		},
		{
			name:      "tasks with multiple predicates",
			path:      "/projects/tasks[status=open,assignee=bob]",
			wantCount: 1,
		},
		{
			name:      "inequality operator",
			path:      "/projects/tasks[status!=closed]",
			wantCount: 2,
		},
		{
			name:      "match operator",
			path:      "/projects[name~=fo]", // matches "foo"
			wantCount: 1,
		},
		{
			name:      "no matches",
			path:      "/projects[name=nonexistent]",
			wantCount: 0,
		},
		{
			name:      "no matches in child",
			path:      "/projects[name=foo]/tasks[status=invalid]",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.path)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			results, err := Eval(ctx, resolver, p)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("Eval() returned %d nodes, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestEvalFrom(t *testing.T) {
	// Create a simple hierarchy for testing
	root := &testNode{
		kind: "root",
		children: []*testNode{
			{
				kind: "projects",
				fields: map[string]string{
					"name": "foo",
				},
				children: []*testNode{
					{
						kind: "tasks",
						fields: map[string]string{
							"id": "1",
						},
					},
				},
			},
		},
	}

	resolver := &testResolver{root: root}
	ctx := context.Background()

	// Start from a project node instead of root
	startNode := root.children[0]

	// Path relative to the project
	p, err := Parse("/tasks[id=1]")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	results, err := EvalFrom(ctx, resolver, startNode, p)
	if err != nil {
		t.Fatalf("EvalFrom() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("EvalFrom() returned %d nodes, want 1", len(results))
	}
}

func TestEvalErrors(t *testing.T) {
	// Test error handling in evaluation

	// Resolver that always returns an error
	errorResolver := &errorTestResolver{
		rootErr:     nil,
		childrenErr: fmt.Errorf("test error"),
	}

	ctx := context.Background()
	p, _ := Parse("/projects")

	_, err := Eval(ctx, errorResolver, p)
	if err == nil {
		t.Error("Eval() expected error, got nil")
	}
}

// errorTestResolver is a resolver that returns errors for testing.
type errorTestResolver struct {
	rootErr     error
	childrenErr error
}

func (r *errorTestResolver) Root(ctx context.Context) (Node, error) {
	if r.rootErr != nil {
		return nil, r.rootErr
	}
	return &testNode{kind: "root"}, nil
}

func (r *errorTestResolver) Children(ctx context.Context, parent Node, seg Segment) ([]Node, error) {
	return nil, r.childrenErr
}
