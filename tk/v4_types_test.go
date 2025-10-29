package main

import (
	"strings"
	"testing"
)

// Test ProjectUID validation and string conversion
func TestProjectUID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		uid     ProjectUID
		wantErr bool
	}{
		{
			name:    "valid project UID",
			uid:     NewProjectUID(),
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			uid:     ProjectUID("task_01HQWX3YGPKQVWDXQH7J8FMYZE"),
			wantErr: true,
		},
		{
			name:    "invalid ULID",
			uid:     ProjectUID("prj_INVALID"),
			wantErr: true,
		},
		{
			name:    "empty",
			uid:     ProjectUID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.uid.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectUID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectUID_String(t *testing.T) {
	uid := NewProjectUID()
	s := uid.String()
	if !strings.HasPrefix(s, "prj_") {
		t.Errorf("ProjectUID.String() = %v, want prefix 'prj_'", s)
	}
}

// Test Alias validation and string conversion
func TestAlias_Validate(t *testing.T) {
	tests := []struct {
		name    string
		alias   Alias
		wantErr bool
	}{
		{
			name:    "valid alias",
			alias:   Alias("myproject"),
			wantErr: false,
		},
		{
			name:    "valid with dash",
			alias:   Alias("my-project"),
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			alias:   Alias("my_project"),
			wantErr: false,
		},
		{
			name:    "too short",
			alias:   Alias("x"),
			wantErr: true,
		},
		{
			name:    "too long",
			alias:   Alias("verylongaliasnameover"),
			wantErr: true,
		},
		{
			name:    "invalid characters",
			alias:   Alias("my project"),
			wantErr: true,
		},
		{
			name:    "invalid special chars",
			alias:   Alias("my@project"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.alias.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Alias.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlias_String(t *testing.T) {
	alias := Alias("test")
	if alias.String() != "test" {
		t.Errorf("Alias.String() = %v, want %v", alias.String(), "test")
	}
}

// Test ProjectType validation
func TestProjectType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ptype   ProjectType
		wantErr bool
	}{
		{
			name:    "local",
			ptype:   ProjectTypeLocal,
			wantErr: false,
		},
		{
			name:    "github",
			ptype:   ProjectTypeGithub,
			wantErr: false,
		},
		{
			name:    "linear",
			ptype:   ProjectTypeLinear,
			wantErr: false,
		},
		{
			name:    "jira",
			ptype:   ProjectTypeJira,
			wantErr: false,
		},
		{
			name:    "invalid",
			ptype:   ProjectType("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ptype.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test TaskUID validation and string conversion
func TestTaskUID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		uid     TaskUID
		wantErr bool
	}{
		{
			name:    "valid task UID",
			uid:     NewTaskUID(),
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			uid:     TaskUID("prj_01HQWX3YGPKQVWDXQH7J8FMYZE"),
			wantErr: true,
		},
		{
			name:    "invalid ULID",
			uid:     TaskUID("tsk_INVALID"),
			wantErr: true,
		},
		{
			name:    "empty",
			uid:     TaskUID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.uid.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TaskUID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskUID_String(t *testing.T) {
	uid := NewTaskUID()
	s := uid.String()
	if !strings.HasPrefix(s, "tsk_") {
		t.Errorf("TaskUID.String() = %v, want prefix 'tsk_'", s)
	}
}

// Test TaskNumber validation and string conversion
func TestTaskNumber_Validate(t *testing.T) {
	tests := []struct {
		name    string
		number  TaskNumber
		wantErr bool
	}{
		{
			name:    "valid positive",
			number:  TaskNumber(1),
			wantErr: false,
		},
		{
			name:    "valid large",
			number:  TaskNumber(1000),
			wantErr: false,
		},
		{
			name:    "zero",
			number:  TaskNumber(0),
			wantErr: true,
		},
		{
			name:    "negative",
			number:  TaskNumber(-1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.number.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TaskNumber.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskNumber_String(t *testing.T) {
	num := TaskNumber(42)
	if num.String() != "42" {
		t.Errorf("TaskNumber.String() = %v, want %v", num.String(), "42")
	}
}

// Test NodeID validation and string conversion
func TestNodeID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  NodeID
		wantErr bool
	}{
		{
			name:    "valid node ID",
			nodeID:  NewNodeID(),
			wantErr: false,
		},
		{
			name:    "invalid ULID",
			nodeID:  NodeID("INVALID"),
			wantErr: true,
		},
		{
			name:    "empty",
			nodeID:  NodeID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.nodeID.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeID_String(t *testing.T) {
	nodeID := NewNodeID()
	s := nodeID.String()
	if len(s) != 26 { // ULID length
		t.Errorf("NodeID.String() length = %v, want %v", len(s), 26)
	}
}

func TestNodeID_Short(t *testing.T) {
	nodeID := NodeID("01HQWX3YGPKQVWDXQH7J8FMYZE")
	short := nodeID.Short()
	if short != "8FMYZE" {
		t.Errorf("NodeID.Short() = %v, want %v", short, "8FMYZE")
	}

	// Test short ID (less than 6 chars)
	shortID := NodeID("ABC")
	if shortID.Short() != "ABC" {
		t.Errorf("NodeID.Short() = %v, want %v", shortID.Short(), "ABC")
	}
}

// Test DisplayID parsing and string conversion
func TestDisplayID_Parse(t *testing.T) {
	tests := []struct {
		name         string
		displayID    DisplayID
		wantAlias    string
		wantNumber   int64
		wantNodeHint string
		wantErr      bool
	}{
		{
			name:         "simple format",
			displayID:    DisplayID("proj-42"),
			wantAlias:    "proj",
			wantNumber:   42,
			wantNodeHint: "",
			wantErr:      false,
		},
		{
			name:         "with node hint",
			displayID:    DisplayID("proj-42-abc123"),
			wantAlias:    "proj",
			wantNumber:   42,
			wantNodeHint: "abc123",
			wantErr:      false,
		},
		{
			name:      "invalid - too few parts",
			displayID: DisplayID("proj"),
			wantErr:   true,
		},
		{
			name:      "invalid - too many parts",
			displayID: DisplayID("proj-42-hint-extra"),
			wantErr:   true,
		},
		{
			name:      "invalid number",
			displayID: DisplayID("proj-abc"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias, number, nodeHint, err := tt.displayID.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("DisplayID.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if alias != tt.wantAlias {
					t.Errorf("DisplayID.Parse() alias = %v, want %v", alias, tt.wantAlias)
				}
				if number != tt.wantNumber {
					t.Errorf("DisplayID.Parse() number = %v, want %v", number, tt.wantNumber)
				}
				if nodeHint != tt.wantNodeHint {
					t.Errorf("DisplayID.Parse() nodeHint = %v, want %v", nodeHint, tt.wantNodeHint)
				}
			}
		})
	}
}

func TestDisplayID_String(t *testing.T) {
	displayID := DisplayID("proj-42")
	if displayID.String() != "proj-42" {
		t.Errorf("DisplayID.String() = %v, want %v", displayID.String(), "proj-42")
	}
}

// Test EventID validation and string conversion
func TestEventID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		eventID EventID
		wantErr bool
	}{
		{
			name:    "valid event ID",
			eventID: NewEventID(),
			wantErr: false,
		},
		{
			name:    "invalid ULID",
			eventID: EventID("INVALID"),
			wantErr: true,
		},
		{
			name:    "empty",
			eventID: EventID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.eventID.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EventID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventID_String(t *testing.T) {
	eventID := NewEventID()
	s := eventID.String()
	if len(s) != 26 { // ULID length
		t.Errorf("EventID.String() length = %v, want %v", len(s), 26)
	}
}

// Test TaskLabel validation
func TestTaskLabel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		label   TaskLabel
		wantErr bool
	}{
		{
			name: "valid label",
			label: TaskLabel{
				ProjectUID: NewProjectUID(),
				Number:     TaskNumber(42),
			},
			wantErr: false,
		},
		{
			name: "invalid project UID",
			label: TaskLabel{
				ProjectUID: ProjectUID("invalid"),
				Number:     TaskNumber(42),
			},
			wantErr: true,
		},
		{
			name: "invalid number",
			label: TaskLabel{
				ProjectUID: NewProjectUID(),
				Number:     TaskNumber(0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.label.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TaskLabel.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test NumberPolicy validation
func TestNumberPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  NumberPolicy
		wantErr bool
	}{
		{
			name: "keep mode",
			policy: NumberPolicy{
				Mode: "keep",
			},
			wantErr: false,
		},
		{
			name: "auto mode",
			policy: NumberPolicy{
				Mode: "auto",
			},
			wantErr: false,
		},
		{
			name: "force mode with valid number",
			policy: NumberPolicy{
				Mode:   "force",
				Number: TaskNumber(42),
			},
			wantErr: false,
		},
		{
			name: "force mode with invalid number",
			policy: NumberPolicy{
				Mode:   "force",
				Number: TaskNumber(0),
			},
			wantErr: true,
		},
		{
			name: "fail mode",
			policy: NumberPolicy{
				Mode: "fail",
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			policy: NumberPolicy{
				Mode: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("NumberPolicy.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
