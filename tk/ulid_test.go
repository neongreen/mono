package main

import (
	"strings"
	"testing"
)

func TestGenerateTaskUUID(t *testing.T) {
	uuid, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("GenerateTaskUUID() error = %v", err)
	}

	if !strings.HasPrefix(uuid, "task-") {
		t.Errorf("GenerateTaskUUID() = %v, want prefix 'task-'", uuid)
	}

	// ULID should be lowercase
	ulidPart := strings.TrimPrefix(uuid, "task-")
	if ulidPart != strings.ToLower(ulidPart) {
		t.Errorf("GenerateTaskUUID() ULID part = %v, want lowercase", ulidPart)
	}

	// Generate another and ensure they're different
	uuid2, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("GenerateTaskUUID() second call error = %v", err)
	}

	if uuid == uuid2 {
		t.Errorf("GenerateTaskUUID() generated duplicate UUIDs: %v", uuid)
	}
}

// TestGenerateTaskID - removed - GenerateTaskID is deprecated, prefix functionality has been removed

func TestGenerateEventID(t *testing.T) {
	db := openTempDB(t)

	eventID, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("GenerateEventID() error = %v", err)
	}

	// Should have format: ev-number-node
	if !strings.HasPrefix(eventID, "ev-") {
		t.Errorf("GenerateEventID() = %v, want prefix 'ev-'", eventID)
	}

	parts := strings.Split(eventID, "-")
	if len(parts) < 3 {
		t.Errorf("GenerateEventID() = %v, want format 'ev-number-node'", eventID)
	}

	// Second call should increment the number
	eventID2, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("GenerateEventID() second call error = %v", err)
	}

	if eventID == eventID2 {
		t.Errorf("GenerateEventID() generated duplicate event IDs: %v", eventID)
	}
}

func TestSplitEventID(t *testing.T) {
	tests := []struct {
		name     string
		eventID  string
		wantLen  int
		wantPart string
	}{
		{
			name:     "standard format",
			eventID:  "ev-123-node",
			wantLen:  3,
			wantPart: "ev",
		},
		{
			name:     "with more dashes",
			eventID:  "ev-123-node-extra",
			wantLen:  4,
			wantPart: "ev",
		},
		{
			name:     "simple",
			eventID:  "ev-123",
			wantLen:  2,
			wantPart: "ev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitEventID(tt.eventID)
			if len(parts) != tt.wantLen {
				t.Errorf("splitEventID() len = %v, want %v", len(parts), tt.wantLen)
			}
			if parts[0] != tt.wantPart {
				t.Errorf("splitEventID() first part = %v, want %v", parts[0], tt.wantPart)
			}
		})
	}
}
