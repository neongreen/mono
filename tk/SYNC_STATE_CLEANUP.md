# tk Sync State Cleanup

## Changes in This Update

The export state tracking has been moved from separate JSON files to the database itself. This ensures that export state is automatically invalidated when the database is deleted.

## Required Cleanup

After updating to this version, delete the following on **all machines**:

### 1. Old Export State Files

```bash
rm -rf ~/.local/state/tk/export/
```

### 2. All State Files (moved to ~/.tk)

```bash
rm -rf ~/.local/state/tk/
```

The state directory has been moved from `~/.local/state/tk/` to `~/.tk/`. The new location:
- Remote index mirrors: `~/.tk/remotes/{remote}/{space}/index.json`
- Segment cache: `~/.tk/cache/segments/...`

### 3. Remote Segments (will be regenerated)

```bash
# For iCloud remote:
rm -rf ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events/

# Or wherever your remotes are configured
```

### 4. Run Sync

After cleanup, run `tk sync` on each machine to regenerate everything from the databases.

## What Changed

1. **Export state now in database**: The `export_state` table in the database tracks the last exported event ID per remote/space.

2. **State directory moved**: From `~/.local/state/tk/` to `~/.tk/` for consistency (the database is also in `~/.tk/`).

3. **Warning on stale state**: If export state references an event ID that doesn't exist in the database (e.g., after deleting the database), a warning is logged and all events are exported.

## Example

```bash
# On machine 1 (emily)
rm -rf ~/.local/state/tk/
rm -rf ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events/
tk sync

# On machine 2 (imac)
rm -rf ~/.local/state/tk/
tk sync
```

