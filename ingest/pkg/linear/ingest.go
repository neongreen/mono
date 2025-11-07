package linear

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/ingest/pkg/database"
)

// ResourceSession defines the subset of MCP session functionality required for Linear ingestion.
type ResourceSession interface {
	ListResources(ctx context.Context, params *sdkmcp.ListResourcesParams) (*sdkmcp.ListResourcesResult, error)
	ReadResource(ctx context.Context, params *sdkmcp.ReadResourceParams) (*sdkmcp.ReadResourceResult, error)
}

// IngestIssues pulls issue resources from the Linear MCP server and stores them in the database.
func IngestIssues(ctx context.Context, db *database.Database, runID int64, session ResourceSession) (int, error) {
	var (
		cursor string
		total  int
	)

	for {
		var params *sdkmcp.ListResourcesParams
		if cursor != "" {
			params = &sdkmcp.ListResourcesParams{Cursor: cursor}
		}

		resources, err := session.ListResources(ctx, params)
		if err != nil {
			return total, fmt.Errorf("failed to list Linear resources: %w", err)
		}

		for _, resource := range resources.Resources {
			if resource == nil {
				continue
			}
			if !strings.HasPrefix(resource.URI, "linear-issue://") {
				continue
			}

			record, raw, err := readIssue(ctx, session, resource)
			if err != nil {
				return total, fmt.Errorf("failed to fetch %s: %w", resource.URI, err)
			}

			rawCopy := raw
			dbIssue := database.LinearIssue{
				IssueID:     record.ID,
				Identifier:  record.Identifier,
				Title:       record.Title,
				Description: record.Description,
				Priority:    record.Priority,
				Status:      record.Status,
				Assignee:    record.Assignee,
				Team:        record.Team,
				URL:         record.URL,
				RawData:     &rawCopy,
			}

			if err := db.CreateLinearIssue(runID, dbIssue); err != nil {
				return total, err
			}
			total++
		}

		if resources.NextCursor == "" {
			break
		}
		cursor = resources.NextCursor
	}

	return total, nil
}

type issuePayload struct {
	ID          string  `json:"id"`
	Identifier  string  `json:"identifier"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Priority    *int    `json:"priority"`
	Status      *string `json:"status"`
	Assignee    *string `json:"assignee"`
	Team        *string `json:"team"`
	URL         *string `json:"url"`
}

type issueRecord struct {
	ID          string
	Identifier  string
	Title       string
	Description *string
	Priority    *int
	Status      *string
	Assignee    *string
	Team        *string
	URL         *string
}

func readIssue(ctx context.Context, session ResourceSession, resource *sdkmcp.Resource) (issueRecord, string, error) {
	out := issueRecord{}

	resp, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: resource.URI})
	if err != nil {
		return out, "", fmt.Errorf("read resource: %w", err)
	}
	if len(resp.Contents) == 0 {
		return out, "", fmt.Errorf("resource returned no contents")
	}

	content := resp.Contents[0]
	var raw string
	switch {
	case content.Text != "":
		raw = content.Text
	case len(content.Blob) > 0:
		raw = string(content.Blob)
	default:
		return out, "", fmt.Errorf("resource content missing text")
	}

	var payload issuePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return out, raw, fmt.Errorf("decode issue payload: %w", err)
	}

	out.ID = strings.TrimSpace(payload.ID)
	out.Identifier = strings.TrimSpace(payload.Identifier)
	out.Title = strings.TrimSpace(payload.Title)
	out.Description = trimStringPtr(payload.Description)
	out.Priority = payload.Priority
	out.Status = trimStringPtr(payload.Status)
	out.Assignee = trimStringPtr(payload.Assignee)
	out.Team = trimStringPtr(payload.Team)
	out.URL = trimStringPtr(payload.URL)

	meta := resource.Meta
	if out.ID == "" {
		out.ID = extractIssueID(resource.URI)
	}
	if out.Identifier == "" {
		out.Identifier = stringValue(meta, "identifier", "")
	}
	if out.Title == "" {
		out.Title = fallbackString(resource.Title, resource.Name)
	}
	if out.Priority == nil {
		out.Priority = intFromMeta(meta, "priority")
	}
	if out.Status == nil {
		out.Status = trimStringPtr(stringPtr(stringValue(meta, "status", "")))
	}
	if out.Assignee == nil {
		out.Assignee = trimStringPtr(stringPtr(stringValue(meta, "assignee", "")))
	}
	if out.Team == nil {
		out.Team = trimStringPtr(stringPtr(stringValue(meta, "team", "")))
	}
	if out.URL == nil {
		out.URL = trimStringPtr(stringPtr(stringValue(meta, "url", "")))
	}

	if out.ID == "" {
		return out, raw, fmt.Errorf("issue id missing in %s", resource.URI)
	}
	if out.Identifier == "" {
		out.Identifier = out.ID
	}
	if out.Title == "" {
		return out, raw, fmt.Errorf("issue title missing for %s", resource.URI)
	}

	return out, raw, nil
}

func extractIssueID(uri string) string {
	const prefix = "linear-issue://"
	if after, ok := strings.CutPrefix(uri, prefix); ok {
		value := after
		return strings.TrimLeft(value, "/")
	}
	return uri
}

func fallbackString(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func stringValue(meta map[string]any, key string, fallback string) string {
	if meta == nil {
		return fallback
	}
	if v, ok := meta[key]; ok {
		switch val := v.(type) {
		case string:
			if trimmed := strings.TrimSpace(val); trimmed != "" {
				return trimmed
			}
		case fmt.Stringer:
			if trimmed := strings.TrimSpace(val.String()); trimmed != "" {
				return trimmed
			}
		case float64:
			return strings.TrimSpace(fmt.Sprintf("%g", val))
		default:
			return strings.TrimSpace(fmt.Sprint(val))
		}
	}
	return fallback
}

func intFromMeta(meta map[string]any, key string) *int {
	if meta == nil {
		return nil
	}
	val, ok := meta[key]
	if !ok {
		return nil
	}

	switch v := val.(type) {
	case int:
		return &v
	case int64:
		i := int(v)
		return &i
	case float64:
		i := int(v)
		return &i
	case json.Number:
		if iv, err := v.Int64(); err == nil {
			i := int(iv)
			return &i
		}
	}
	return nil
}

func trimStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*in)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringPtr(value string) *string {
	return &value
}
