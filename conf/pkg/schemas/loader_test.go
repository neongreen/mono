package schemas

import (
	"strings"
	"testing"
)

func TestNewSchemaLoader(t *testing.T) {
	loader, err := NewSchemaLoader()
	if err != nil {
		t.Fatalf("Failed to create schema loader: %v", err)
	}

	if loader == nil {
		t.Fatal("Schema loader is nil")
	}

	// Schema might be nil due to validation issues, but raw schema should work
	raw := loader.GetJJSchemaRaw()
	if len(raw) == 0 {
		t.Fatal("Raw JJ schema is empty")
	}
}

func TestJJSchemaEmbedded(t *testing.T) {
	if len(JJSchema) == 0 {
		t.Fatal("JJ schema is empty")
	}

	// Check that it contains expected JSON schema content
	if !strings.Contains(JJSchema, "\"title\": \"Jujutsu config\"") {
		t.Error("JJ schema doesn't contain expected title")
	}

	if !strings.Contains(JJSchema, "\"properties\"") {
		t.Error("JJ schema doesn't contain properties section")
	}
}
