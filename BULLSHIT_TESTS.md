# Bullshit Tests to Avoid

This document catalogs types of useless tests that add no value and should not be written.

## What Are Bullshit Tests?

Bullshit tests are tests that verify basic language features or library functionality rather than your application logic. They provide no value because:

1. They test that the language or standard library works (which is not your responsibility)
2. They break only if you fundamentally misunderstand how the language works
3. They add maintenance burden without catching real bugs
4. They give false confidence (high test coverage, low actual verification)

## Categories of Bullshit Tests

### 1. Struct Field Assignment Tests

**Pattern**: Create a struct with values, then immediately verify those same values exist in the struct.

**Why it's bullshit**: You're testing that Go's struct assignment works, not your code.

**Example removed from `ingest/pkg/github/github_test.go`:**

```go
func TestIssueStructure(t *testing.T) {
	// Test that we can create and access Issue struct fields
	now := time.Now()
	issue := Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     "open",
		User:      User{Login: "testuser"},
		CreatedAt: now,
		UpdatedAt: now,
		ClosedAt:  nil,
		Labels:    []Label{{Name: "bug"}},
		Assignees: []User{{Login: "assignee1"}},
		Milestone: &Milestone{Title: "v1.0"},
	}

	if issue.Number != 1 {
		t.Errorf("Expected issue number 1, got %d", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", issue.Title)
	}
	if issue.User.Login != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", issue.User.Login)
	}
	if len(issue.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(issue.Labels))
	}
}
```

Similar tests removed:
- `TestPullRequestStructure` in `ingest/pkg/github/github_test.go`
- `TestCommentStructure` in `ingest/pkg/github/github_test.go`

### 2. Constructor Returns Non-Nil Tests

**Pattern**: Call a constructor, verify it returns non-nil.

**Why it's bullshit**: If your constructor can't fail, there's no value in testing that it returns something. If it can fail, test the failure conditions, not the success case.

**Example removed from `ingest/pkg/github/github_test.go`:**

```go
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.ctx == nil {
		t.Fatal("expected client context to be non-nil")
	}
	if client.gh == nil {
		t.Fatal("expected go-github client to be non-nil")
	}
}
```

This test provides zero value. The constructor simply calls `go-github`'s constructor. If that fails, it's their bug, not yours.

Similar tests removed:
- `TestNewClientWithContext` in `ingest/pkg/github/github_test.go`
- `TestNewClaudeSchemaParser` in `conf/pkg/schemas/claude_parser_test.go`
- `TestNewJJSchemaParser` in `conf/pkg/schemas/jj_parser_test.go`
- `TestNewMiseSchemaParser` in `conf/pkg/schemas/mise_json_parser_test.go`
- `TestNewSchemaLoader` in `conf/pkg/schemas/loader_test.go`

### 3. Constructor Parameters Copied to Fields Tests

**Pattern**: Call constructor with parameters, verify those exact parameters were copied to struct fields.

**Why it's bullshit**: You're testing that field assignment works.

**Example removed from `tk/internal/segment/writer_test.go`:**

```go
func TestNewSegmentWriter(t *testing.T) {
	sw := NewSegmentWriter("/tmp/test", "personal", "node123", 1, 1024*1024, 60)

	if sw == nil {
		t.Fatal("NewSegmentWriter() returned nil")
	}

	if sw.remotePath != "/tmp/test" {
		t.Errorf("remotePath = %v, want %v", sw.remotePath, "/tmp/test")
	}

	if sw.space != "personal" {
		t.Errorf("space = %v, want %v", sw.space, "personal")
	}

	if sw.node != "node123" {
		t.Errorf("node = %v, want %v", sw.node, "node123")
	}

	if sw.segmentSeq != 1 {
		t.Errorf("segmentSeq = %v, want %v", sw.segmentSeq, 1)
	}

	if sw.maxBytes != 1024*1024 {
		t.Errorf("maxBytes = %v, want %v", sw.maxBytes, 1024*1024)
	}

	if sw.maxAge != 60 {
		t.Errorf("maxAge = %v, want %v", sw.maxAge, 60)
	}
}
```

### 4. Framework Registration Tests

**Pattern**: Verify that commands/routes/handlers were registered with a framework.

**Why it's bullshit**: If registration doesn't happen, your code literally won't compile or will panic on startup. The test adds no value.

**Example removed from `conf/cmd/main_test.go`:**

```go
func TestCobraCommandsExist(t *testing.T) {
	// Test that root command has the expected subcommands
	if rootCmd == nil {
		t.Fatal("Root command should not be nil")
	}

	// Check subcommands exist
	jjSubCmd, _, err := rootCmd.Find([]string{"jj"})
	if err != nil {
		t.Fatalf("jj subcommand should exist: %v", err)
	}
	if jjSubCmd.Use != "jj [config.path] [value]" {
		t.Errorf("jj command has wrong usage: %s", jjSubCmd.Use)
	}

	miseSubCmd, _, err := rootCmd.Find([]string{"mise"})
	if err != nil {
		t.Fatalf("mise subcommand should exist: %v", err)
	}
	if miseSubCmd.Use != "mise [config.path] [value]" {
		t.Errorf("mise command has wrong usage: %s", miseSubCmd.Use)
	}

	completionSubCmd, _, err := rootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("completion subcommand should exist: %v", err)
	}
	if completionSubCmd.Use != "completion [bash|zsh|fish]" {
		t.Errorf("completion command has wrong usage: %s", completionSubCmd.Use)
	}
}

func TestCommandStructure(t *testing.T) {
	if rootCmd.Use != "conf" {
		t.Errorf("Root command should be 'conf', got: %s", rootCmd.Use)
	}

	if len(rootCmd.Commands()) < 3 {
		t.Errorf("Root command should have at least 3 subcommands, got: %d", len(rootCmd.Commands()))
	}
}
```

### 5. Dry-Run Flag Tests

**Pattern**: Create object with dry-run flag, verify flag was set.

**Why it's bullshit**: You're testing field assignment again.

**Example removed from `conf/pkg/tools/shims/shims_test.go`:**

```go
func TestNewShimsTool(t *testing.T) {
	tool, err := NewShimsTool()
	if err != nil {
		t.Fatalf("Failed to create shims tool: %v", err)
	}

	if tool.shimsDir == "" {
		t.Error("Expected shims directory to be set")
	}

	if !strings.Contains(tool.shimsDir, "conf-shims") {
		t.Errorf("Expected shims directory to contain 'conf-shims', got %s", tool.shimsDir)
	}

	if tool.dryRun {
		t.Error("Expected dry-run to be false by default")
	}
}

func TestNewShimsToolWithDryRun(t *testing.T) {
	tool, err := NewShimsToolWithDryRun(true)
	if err != nil {
		t.Fatalf("Failed to create shims tool: %v", err)
	}

	if !tool.IsDryRun() {
		t.Error("Expected dry-run to be true")
	}
}
```

## What to Test Instead

Test actual behavior and logic:

**Good tests:**
- Business logic and algorithms
- Error conditions and edge cases  
- Integration between components
- Data transformations
- Side effects (file I/O, API calls, database operations)
- Failure modes and recovery

**Examples of good tests** (not removed):
- `TestClientFetchIssuesFiltersPRsAndAggregatesPages` - Tests actual GitHub API interaction with pagination
- `TestClientFetchPullRequestsAggregatesPages` - Tests data fetching and aggregation logic
- `TestClientFetchPRCommentsCombinesIssueAndReview` - Tests combining data from multiple sources
- `TestClientFetchIssuesError` - Tests error handling

## Summary

**Don't test:**
- That constructors return non-nil
- That parameters are copied to fields
- That struct field assignment works
- That framework registration works (it will fail at compile/runtime if broken)
- Basic language features

**Do test:**
- Business logic
- Error cases
- Data transformation
- Integration points
- Edge cases

