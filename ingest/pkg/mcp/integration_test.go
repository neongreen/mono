package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSessionCallToolWithInMemoryServer(t *testing.T) {
	session := newTestSession(t)
	defer session.Close()

	result, err := session.CallTool(context.Background(), "hello", map[string]any{"name": "tester"})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	text := collectText(t, result)
	if text != `{"message":"hello tester"}` {
		t.Fatalf("unexpected response %s", text)
	}
}

func TestSessionListTools(t *testing.T) {
	session := newTestSession(t)
	defer session.Close()

	tools, err := session.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "hello" {
		t.Fatalf("unexpected tools %+v", tools)
	}
}

func newTestSession(t *testing.T) *Session {
	t.Helper()

	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test-server", Version: "1.0"}, nil)
	server.AddTool(&sdkmcp.Tool{Name: "hello", InputSchema: map[string]any{
		"type":       "object",
		"properties": map[string]any{"name": map[string]any{"type": "string"}},
		"required":   []string{"name"},
	}}, func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
		var args struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, err
		}
		payload, _ := json.Marshal(map[string]string{"message": "hello " + args.Name})
		return &sdkmcp.CallToolResult{Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(payload)}}}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect failed: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}

	return &Session{session: cs}
}

func collectText(t *testing.T, res *sdkmcp.CallToolResult) string {
	t.Helper()
	if res == nil {
		t.Fatal("nil result")
	}
	if res.IsError {
		t.Fatal("result indicates error")
	}
	if len(res.Content) == 0 {
		t.Fatal("empty content")
	}
	txt, ok := res.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("unexpected content type %T", res.Content[0])
	}
	return txt.Text
}
