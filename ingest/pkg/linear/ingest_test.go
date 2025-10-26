package linear

import (
	"context"
	"fmt"
	"github.com/neongreen/mono/ingest/pkg/database"
	"os"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeSession struct {
	resources []*sdkmcp.Resource
	payloads  map[string]string
}

func (f *fakeSession) ListResources(_ context.Context, params *sdkmcp.ListResourcesParams) (*sdkmcp.ListResourcesResult, error) {
	if params != nil && params.Cursor != "" {
		return &sdkmcp.ListResourcesResult{Resources: []*sdkmcp.Resource{}}, nil
	}
	return &sdkmcp.ListResourcesResult{Resources: f.resources}, nil
}

func (f *fakeSession) ReadResource(_ context.Context, params *sdkmcp.ReadResourceParams) (*sdkmcp.ReadResourceResult, error) {
	raw, ok := f.payloads[params.URI]
	if !ok {
		return nil, fmt.Errorf("unexpected resource %s", params.URI)
	}
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{
			{URI: params.URI, Text: raw},
		},
	}, nil
}

func TestIngestIssues(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	db, err := database.Open()
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	runID, err := db.CreateRun("linear", "linear")
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	session := &fakeSession{
		resources: []*sdkmcp.Resource{
			{
				URI:         "linear-issue:///ISSUE-1",
				Name:        "Fix login bug",
				Description: "Linear issue LIN-1",
				Meta: sdkmcp.Meta{
					"identifier": "LIN-1",
					"status":     "Todo",
					"priority":   1,
					"team":       "Platform",
				},
			},
			{
				URI:  "linear-issue:///ISSUE-2",
				Name: "Implement feature toggle",
				Meta: sdkmcp.Meta{
					"identifier": "LIN-2",
					"status":     "In Progress",
					"priority":   float64(3),
					"team":       "Core",
				},
			},
			{
				URI: "linear-organization:",
			},
		},
		payloads: map[string]string{
			"linear-issue:///ISSUE-1": `{"id":"ISSUE-1","identifier":"LIN-1","title":"Fix login bug","description":"Users cannot log in","priority":1,"status":"Todo","assignee":"Alice","team":"Platform","url":"https://linear.app/example/issue/LIN-1"}`,
			"linear-issue:///ISSUE-2": `{"id":"ISSUE-2","title":"Implement feature toggle"}`,
		},
	}

	count, err := IngestIssues(context.Background(), db, runID, session)
	if err != nil {
		t.Fatalf("ingest issues: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 issues ingested, got %d", count)
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		t.Fatalf("update run item count: %v", err)
	}

	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].ItemCount != 2 {
		t.Errorf("expected run item count 2, got %d", runs[0].ItemCount)
	}

	results, err := db.Query("SELECT identifier, priority, status, assignee, team, url FROM linear_issues ORDER BY identifier")
	if err != nil {
		t.Fatalf("query linear issues: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 stored issues, got %d", len(results))
	}

	first := results[0]
	if first["identifier"] != "LIN-1" {
		t.Errorf("first identifier mismatch, got %v", first["identifier"])
	}
	if first["priority"] != int64(1) {
		t.Errorf("first priority mismatch, got %v", first["priority"])
	}
	if first["assignee"] != "Alice" {
		t.Errorf("expected assignee Alice, got %v", first["assignee"])
	}
	if first["url"] != "https://linear.app/example/issue/LIN-1" {
		t.Errorf("unexpected url %v", first["url"])
	}

	second := results[1]
	if second["identifier"] != "LIN-2" {
		t.Errorf("second identifier mismatch, got %v", second["identifier"])
	}
	if second["priority"] != int64(3) {
		t.Errorf("expected fallback priority 3, got %v", second["priority"])
	}
	if second["status"] != "In Progress" {
		t.Errorf("expected fallback status, got %v", second["status"])
	}
	if second["team"] != "Core" {
		t.Errorf("expected fallback team, got %v", second["team"])
	}
	if second["assignee"] != nil {
		t.Errorf("expected nil assignee, got %v", second["assignee"])
	}
}
