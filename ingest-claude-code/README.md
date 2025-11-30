# ingest-claude-code

A simple tool to extract sessions and messages from Claude Code conversation traces and output them as JSONL.

## Usage

### List all sessions

```bash
ingest-claude-code sessions
```

Outputs one JSONL line per conversation file:

```json
{"session_id":"549bdd72-9c4b-49bb-a8af-29598824d93f","summary":"Backstage auth fix","file_path":"~/.claude/projects/-Users-artyom-code-neongreen-mono/549bdd72-9c4b-49bb-a8af-29598824d93f.jsonl","project_path":"/Users/artyom/code/neongreen/mono","message_count":2830,"mod_time":"2025-11-23T09:23:52Z"}
```

### Extract all messages

```bash
ingest-claude-code messages
```

Outputs one JSONL line per message (user, assistant, tool_result):

```json
{"session_id":"549bdd72-9c4b-49bb-a8af-29598824d93f","message_id":"bc5a32f9-a6c6-4ec5-bb79-6dc54c3864b5","parent_id":"70fa72d7-30c4-4529-a5ab-332fd32ccb59","timestamp":"2025-11-21T19:13:03.636Z","type":"assistant","content":[{"type":"text","text":"Let me create those tasks..."}],"model":"claude-sonnet-4-5-20250929","cwd":"/Users/artyom/code/neongreen/mono","version":"2.0.49"}
```

## Data Sources

The tool searches for traces in:
- `~/.claude/projects/` - Project-specific conversation traces
- `~/.claude/debug/` - Debug traces

## Output Format

### Sessions

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Unique session identifier |
| `summary` | string | Conversation summary (if available) |
| `file_path` | string | Path to the trace file |
| `project_path` | string | Project directory path |
| `message_count` | int | Number of messages in the session |
| `mod_time` | timestamp | File modification time |

### Messages

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Session this message belongs to |
| `message_id` | string | Unique message identifier (UUID) |
| `parent_id` | string | Parent message ID (forms conversation tree) |
| `timestamp` | string | Message timestamp |
| `type` | string | Message type: `user`, `assistant`, or `tool_result` |
| `content` | any | Full message content (format varies by type) |
| `model` | string | Model name (for assistant messages) |
| `cwd` | string | Working directory |
| `tool_use_id` | string | Tool use ID (for tool results) |
| `is_error` | bool | Error flag (for tool results) |
| `git_branch` | string | Git branch |
| `version` | string | Claude Code version |
| `request_id` | string | API request ID |

## Building

```bash
cd ingest-claude-code
go build
```

## Development

For quick iteration without building, use the dev script from the repo root:

```bash
dev/ingest-claude-code sessions
dev/ingest-claude-code messages
```

This runs the tool using `go run`.

## Example Usage

Count total sessions:
```bash
ingest-claude-code sessions | wc -l
```

Extract messages for a specific session:
```bash
ingest-claude-code messages | jq 'select(.session_id=="549bdd72-9c4b-49bb-a8af-29598824d93f")'
```

Count messages by type:
```bash
ingest-claude-code messages | jq -r '.type' | sort | uniq -c
```
