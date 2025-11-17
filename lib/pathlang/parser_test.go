package pathlang

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Path
		wantErr bool
	}{
		{
			name:  "root path",
			input: "/",
			want:  &Path{},
		},
		{
			name:  "single segment",
			input: "/projects",
			want: &Path{
				Segments: []Segment{
					{Name: "projects"},
				},
			},
		},
		{
			name:  "multiple segments",
			input: "/projects/tasks",
			want: &Path{
				Segments: []Segment{
					{Name: "projects"},
					{Name: "tasks"},
				},
			},
		},
		{
			name:  "segment with single predicate",
			input: "/projects[name=foo]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "projects",
						Predicates: []Predicate{
							{Field: "name", Op: OpEq, Value: "foo"},
						},
					},
				},
			},
		},
		{
			name:  "segment with multiple predicates",
			input: "/tasks[status=open,assignee=me]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "status", Op: OpEq, Value: "open"},
							{Field: "assignee", Op: OpEq, Value: "me"},
						},
					},
				},
			},
		},
		{
			name:  "complex path from spec",
			input: "/projects[name=foo]/tasks[status=open,assignee=me]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "projects",
						Predicates: []Predicate{
							{Field: "name", Op: OpEq, Value: "foo"},
						},
					},
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "status", Op: OpEq, Value: "open"},
							{Field: "assignee", Op: OpEq, Value: "me"},
						},
					},
				},
			},
		},
		{
			name:  "quoted value",
			input: `/projects[name="foo bar"]`,
			want: &Path{
				Segments: []Segment{
					{
						Name: "projects",
						Predicates: []Predicate{
							{Field: "name", Op: OpEq, Value: "foo bar"},
						},
					},
				},
			},
		},
		{
			name:  "quoted value with escapes",
			input: `/projects[name="foo\"bar\\baz\n\t"]`,
			want: &Path{
				Segments: []Segment{
					{
						Name: "projects",
						Predicates: []Predicate{
							{Field: "name", Op: OpEq, Value: "foo\"bar\\baz\n\t"},
						},
					},
				},
			},
		},
		{
			name:  "numeric value",
			input: "/tasks[id=43]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "id", Op: OpEq, Value: "43"},
						},
					},
				},
			},
		},
		{
			name:  "inequality operator",
			input: "/tasks[status!=closed]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "status", Op: OpNotEq, Value: "closed"},
						},
					},
				},
			},
		},
		{
			name:  "match operator",
			input: "/tasks[description~=bug]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "description", Op: OpMatch, Value: "bug"},
						},
					},
				},
			},
		},
		{
			name:  "identifier with dash",
			input: "/task-items",
			want: &Path{
				Segments: []Segment{
					{Name: "task-items"},
				},
			},
		},
		{
			name:  "identifier with underscore",
			input: "/task_items",
			want: &Path{
				Segments: []Segment{
					{Name: "task_items"},
				},
			},
		},
		{
			name:  "field with dash and underscore",
			input: "/tasks[my_field-name=value]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "my_field-name", Op: OpEq, Value: "value"},
						},
					},
				},
			},
		},
		{
			name:  "bare value with dot",
			input: "/tasks[url=foo.bar]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "url", Op: OpEq, Value: "foo.bar"},
						},
					},
				},
			},
		},
		{
			name:  "bare value with colon",
			input: "/tasks[url=foo:bar]",
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "url", Op: OpEq, Value: "foo:bar"},
						},
					},
				},
			},
		},
		{
			name:  "bare value with backslash",
			input: `/tasks[path=with\backslash]`,
			want: &Path{
				Segments: []Segment{
					{
						Name: "tasks",
						Predicates: []Predicate{
							{Field: "path", Op: OpEq, Value: `with\backslash`},
						},
					},
				},
			},
		},
		// Error cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "relative path",
			input:   "projects",
			wantErr: true,
		},
		{
			name:    "unclosed predicate",
			input:   "/projects[name=foo",
			wantErr: true,
		},
		{
			name:    "invalid operator",
			input:   "/projects[name>foo]",
			wantErr: true,
		},
		{
			name:    "unclosed quoted value",
			input:   `/projects[name="foo]`,
			wantErr: true,
		},
		{
			name:    "invalid escape sequence",
			input:   `/projects[name="foo\x"]`,
			wantErr: true,
		},
		{
			name:    "missing value",
			input:   "/projects[name=]",
			wantErr: true,
		},
		{
			name:    "missing field",
			input:   "/projects[=value]",
			wantErr: true,
		},
		{
			name:    "missing operator",
			input:   "/projects[name]",
			wantErr: true,
		},
		{
			name:    "trailing slash",
			input:   "/projects/",
			wantErr: true,
		},
		{
			name:    "segment name with dot",
			input:   "/foo.bar",
			wantErr: true,
		},
		{
			name:    "field name with dot",
			input:   "/tasks[foo.bar=value]",
			wantErr: true,
		},
		{
			name:    "segment name with colon",
			input:   "/foo:bar",
			wantErr: true,
		},
		// Action tests
		{
			name:  "simple action without args",
			input: "/foo-13/notes @add",
			want: &Path{
				Segments: []Segment{
					{Name: "foo-13"},
					{Name: "notes"},
				},
				Action: "add",
			},
		},
		{
			name:  "action with single arg",
			input: "/foo-13/notes @add hello",
			want: &Path{
				Segments: []Segment{
					{Name: "foo-13"},
					{Name: "notes"},
				},
				Action:     "add",
				ActionArgs: []string{"hello"},
			},
		},
		{
			name:  "action with multiple args",
			input: "/foo-13/notes @add hello world",
			want: &Path{
				Segments: []Segment{
					{Name: "foo-13"},
					{Name: "notes"},
				},
				Action:     "add",
				ActionArgs: []string{"hello", "world"},
			},
		},
		{
			name:  "action with quoted arg",
			input: `/foo-13/notes @add "hello world"`,
			want: &Path{
				Segments: []Segment{
					{Name: "foo-13"},
					{Name: "notes"},
				},
				Action:     "add",
				ActionArgs: []string{"hello world"},
			},
		},
		{
			name:  "project action",
			input: "/myproject @status",
			want: &Path{
				Segments: []Segment{
					{Name: "myproject"},
				},
				Action: "status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !pathEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // expected canonical form
	}{
		{
			name:  "root path",
			input: "/",
			want:  "/",
		},
		{
			name:  "simple path",
			input: "/projects",
			want:  "/projects",
		},
		{
			name:  "path with predicates",
			input: "/projects[name=foo]",
			want:  "/projects[name=foo]",
		},
		{
			name:  "value needing quotes",
			input: `/projects[name="foo bar"]`,
			want:  `/projects[name="foo bar"]`,
		},
		{
			name:  "value with special chars",
			input: `/projects[name="foo/bar"]`,
			want:  `/projects[name="foo/bar"]`,
		},
		{
			name:  "complex path",
			input: "/projects[name=foo]/tasks[status=open,assignee=me]",
			want:  "/projects[name=foo]/tasks[status=open,assignee=me]",
		},
		{
			name:  "escaped quotes",
			input: `/projects[name="foo\"bar"]`,
			want:  `/projects[name="foo\"bar"]`,
		},
		{
			name:  "escaped newline and tab",
			input: `/projects[description="line1\nline2\ttab"]`,
			want:  `/projects[description="line1\nline2\ttab"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got := p.String()
			if got != tt.want {
				t.Errorf("Path.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		"/",
		"/projects",
		"/projects/tasks",
		"/projects[name=foo]",
		"/tasks[status=open,assignee=me]",
		"/projects[name=foo]/tasks[status=open]",
		`/projects[name="foo bar"]`,
		`/projects[description="line1\nline2"]`,
		"/tasks[id!=42]",
		"/tasks[description~=bug]",
		"/tasks[url=foo.bar]",
		"/tasks[url=foo:bar]",
		`/tasks[path=with\backslash]`,
		// Action tests
		"/foo-13/notes @add",
		"/foo-13/notes @add hello",
		"/foo-13/notes @add hello world",
		`/foo-13/notes @add "hello world"`,
		"/myproject @status",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Parse
			p, err := Parse(input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Convert back to string
			output := p.String()

			// Parse again
			p2, err := Parse(output)
			if err != nil {
				t.Fatalf("Parse() round-trip error = %v", err)
			}

			// Should be equal
			if !pathEqual(p, p2) {
				t.Errorf("Round-trip failed:\n  original: %v\n  string:   %v\n  parsed:   %v",
					p, output, p2)
			}
		})
	}
}

// pathEqual checks if two paths are equal.
func pathEqual(a, b *Path) bool {
	if len(a.Segments) != len(b.Segments) {
		return false
	}
	for i := range a.Segments {
		if !segmentEqual(a.Segments[i], b.Segments[i]) {
			return false
		}
	}
	if a.Action != b.Action {
		return false
	}
	if len(a.ActionArgs) != len(b.ActionArgs) {
		return false
	}
	for i := range a.ActionArgs {
		if a.ActionArgs[i] != b.ActionArgs[i] {
			return false
		}
	}
	return true
}

// segmentEqual checks if two segments are equal.
func segmentEqual(a, b Segment) bool {
	if a.Name != b.Name {
		return false
	}
	if len(a.Predicates) != len(b.Predicates) {
		return false
	}
	for i := range a.Predicates {
		if !predicateEqual(a.Predicates[i], b.Predicates[i]) {
			return false
		}
	}
	return true
}

// predicateEqual checks if two predicates are equal.
func predicateEqual(a, b Predicate) bool {
	return a.Field == b.Field && a.Op == b.Op && a.Value == b.Value
}
