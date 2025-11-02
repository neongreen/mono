package types

import (
	"encoding/json"
	"testing"
)

func TestMetadataClaim_JSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		claim MetadataClaim
	}{
		{
			name: "number value",
			claim: MetadataClaim{
				Value:     json.RawMessage(`1`),
				Role:      "human",
				Tentative: false,
				TS:        100,
			},
		},
		{
			name: "string value",
			claim: MetadataClaim{
				Value:     json.RawMessage(`"test"`),
				Role:      "agent",
				Tentative: true,
				TS:        200,
			},
		},
		{
			name: "array value",
			claim: MetadataClaim{
				Value:     json.RawMessage(`["bug","urgent"]`),
				Role:      "qa",
				Tentative: false,
				TS:        300,
			},
		},
		{
			name: "object value",
			claim: MetadataClaim{
				Value:     json.RawMessage(`{"severity":"high","impact":"critical"}`),
				Role:      "human",
				Tentative: false,
				TS:        400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.claim)
			if err != nil {
				t.Fatalf("Failed to marshal claim: %v", err)
			}

			// Unmarshal back
			var decoded MetadataClaim
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal claim: %v", err)
			}

			// Compare
			if decoded.Role != tt.claim.Role {
				t.Errorf("Role mismatch: got %s, want %s", decoded.Role, tt.claim.Role)
			}
			if decoded.Tentative != tt.claim.Tentative {
				t.Errorf("Tentative mismatch: got %v, want %v", decoded.Tentative, tt.claim.Tentative)
			}
			if decoded.TS != tt.claim.TS {
				t.Errorf("TS mismatch: got %d, want %d", decoded.TS, tt.claim.TS)
			}
			if string(decoded.Value) != string(tt.claim.Value) {
				t.Errorf("Value mismatch: got %s, want %s", decoded.Value, tt.claim.Value)
			}
		})
	}
}

func TestMetadataStatus_JSONSerialization(t *testing.T) {
	status := MetadataStatus{
		Effective: json.RawMessage(`1`),
		Claims: []MetadataClaim{
			{
				Value:     json.RawMessage(`1`),
				Role:      "human",
				Tentative: false,
				TS:        100,
			},
			{
				Value:     json.RawMessage(`3`),
				Role:      "agent",
				Tentative: true,
				TS:        90,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal status: %v", err)
	}

	// Unmarshal back
	var decoded MetadataStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal status: %v", err)
	}

	// Compare
	if string(decoded.Effective) != string(status.Effective) {
		t.Errorf("Effective mismatch: got %s, want %s", decoded.Effective, status.Effective)
	}
	if len(decoded.Claims) != len(status.Claims) {
		t.Errorf("Claims count mismatch: got %d, want %d", len(decoded.Claims), len(status.Claims))
	}
}
