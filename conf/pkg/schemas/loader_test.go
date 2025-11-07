package schemas

import (
	"strings"
	"testing"
)

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
