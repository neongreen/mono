package cli

import (
	"strings"
	"testing"
)

func TestColorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		contains string
	}{
		{"Success", Success, "test", "test"},
		{"Warning", Warning, "test", "test"},
		{"Error", Error, "test", "test"},
		{"Key", Key, "test", "test"},
		{"Path", Path, "test", "test"},
		{"Type", Type, "test", "test"},
		{"Secondary", Secondary, "test", "test"},
		{"Muted", Muted, "test", "test"},
		{"Header", Header, "test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("%s(%q) should contain %q, got %q", tt.name, tt.input, tt.contains, result)
			}
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"string", "test"},
		{"int", 42},
		{"bool", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Value(tt.input)
			if result == "" {
				t.Errorf("Value(%v) should not be empty", tt.input)
			}
		})
	}
}

func TestFormatFunctions(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(string, ...interface{}) string
		format string
		args   []interface{}
		want   string
	}{
		{"Successf", Successf, "test %d", []interface{}{42}, "42"},
		{"Warningf", Warningf, "test %s", []interface{}{"foo"}, "foo"},
		{"Errorf", Errorf, "test %s", []interface{}{"bar"}, "bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.format, tt.args...)
			if !strings.Contains(result, tt.want) {
				t.Errorf("%s should contain %q, got %q", tt.name, tt.want, result)
			}
		})
	}
}
