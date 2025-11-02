# tk Timestamp and Event Ordering Model

## Document Metadata

**Written**: November 2, 2025  
**Author**: AI assistant (Claude), working with Emily  
**Context**: Created during metadata implementation work when timestamp handling bugs were encountered

**Sources studied**:
- `~/code/tk1.md` - Original ChatGPT design dialogue (5,229 lines)
- `~/code/tk2.md` - Multi-prefix implementation review dialogue
- `tk/specs/spec-v1.md` - v1 sync specification
- `tk/specs/v4.md` and `v4-migration.md` - v4 project specs
- `tk/internal/database/db_events.go` - Event storage and retrieval implementation
- `tk/internal/reducer/reducer.go` - Event application and conflict resolution
- `tk/cmd/test_helpers.go` - Testing patterns
- Git history (commits ce811b0, 67d8678) - Bug fixes and workarounds

**Why this document exists**:

During implementation of metadata claims (`task.meta.set` events), the AI agent encountered confusion about timestamp handling:
- Integration tests were failing with "task not found" errors
- Root cause: `CreatedAt` field was not being set, defaulting to zero
- This caused events to sort incorrectly (corrupt negative timestamps)
- Investigation revealed a discrepancy between the spec and implementation

The user requested a comprehensive study to understand:
1. How event ordering actually works in tk
2. What the difference is between Lamport timestamps (TS) and wall clock (CreatedAt)
3. When and why to use each timestamp
4. Whether the current implementation is correct

**How to use this document**:

1. **For implementing new commands**: See section 6.3 for the correct pattern to create events
2. **For writing tests**: See section 6.2 for proper test patterns (avoid TS=0)
3. **For understanding bugs**: See section 3 for bug history and why things are the way they are
4. **For multi-machine sync**: See sections 4.1 and 7.2 for when current implementation could break
5. **For architectural decisions**: See section 5 for spec vs implementation inconsistencies

**Key finding**: The implementation orders events by wall clock (`created_at`) but the spec says to use Lamport timestamps (`ts`). This works for single-machine usage but could break with multi-machine sync across different timezones.

---

## Introduction

This document explains how tk handles timestamps and event ordering, based on studying the design documents, specifications, and implementation.

## 1. Design Intent

### 1.1 Two Clocks

tk events have **two timestamps**:

1. **`TS` (Lamport timestamp)** - Logical clock for distributed ordering
   - Type: `int64`
   - Purpose: Provide a partial order across events from different machines
   - Guarantees: If event A causally precedes event B, then A.TS < B.TS

2. **`CreatedAt` (Wall clock)** - Physical timestamp
   - Type: `time.Time` (stored as nanoseconds since epoch)
   - Purpose: Record actual time event occurred
   - Use: Human-readable timestamps, audit trails

### 1.2 Original Specification

From `spec-v1.md` line 113:

> **6.2 projection order = (lamport, event.id) deterministic across machines**

The spec clearly states events should be ordered by:
1. Lamport timestamp (`TS`) first
2. Event ID second (for tie-breaking concurrent events)

This ordering ensures deterministic event replay across different machines in a distributed system.

### 1.3 Why Lamport Timestamps?

Lamport timestamps solve the distributed ordering problem:

- Each machine maintains its own Lamport counter
- Counter increments on every local write
- When receiving events from other machines, bump counter to max(local, remote) + 1
- Result: Causally related events maintain their order across machines

**Example**:
```
Machine A: Creates task (TS=1)
Machine A: Syncs, pushes event
Machine B: Pulls event, sees TS=1, bumps local counter to 2
Machine B: Updates task (TS=2)
Machine B: Syncs, pushes event
Machine A: Pulls event, sees TS=2, bumps local counter to 3
```

Both machines can now reconstruct events in the same order: TS=1, then TS=2.

---

## 2. Current Implementation

### 2.1 Event Storage

Events are stored in the `events` table with both timestamps:

```sql
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    ts INTEGER,           -- Lamport timestamp
    created_at INTEGER,   -- Wall clock (nanoseconds)
    actor TEXT,
    role TEXT,
    kind TEXT,
    payload TEXT,
    ...
)
```

### 2.2 Event Creation (Commands)

When a command creates an event (e.g., `tk meta set`):

```go
ts, err := db.GetNextLamportTS()  // Get incremented Lamport: 1, 2, 3...
event := types.Event{
    ID:        eventID,
    TS:        ts,
    CreatedAt: time.Now(),        // Current wall clock
    ...
}
```

**Lamport Counter Persistence**:
- Stored in `metadata` table with key `lamport_counter`
- Increments atomically: 1 → 2 → 3 → 4...
- Persists across command invocations
- Never resets (unless database is rebuilt)

### 2.3 Event Creation (Tests)

Test seed helpers use **different values**:

```go
// From test_helpers.go
event := types.Event{
    TS:        0,           // All seed events use TS=0
    CreatedAt: now,         // Same `now` variable for batch
    ...
}
```

**Why TS=0 works**:
- All seed events in a test have the same `CreatedAt`
- When `ORDER BY created_at, id` runs, events sort by ID
- Event IDs are generated sequentially, so order is preserved

### 2.4 Event Retrieval - **THE DISCREPANCY**

```go
// From db_events.go line 33
func (d *DB) GetEvents() ([]types.Event, error) {
    query := `SELECT id, ts, created_at, ...
              FROM events ORDER BY created_at, id`
    ...
}
```

**Current behavior**: Events are ordered by **wall clock (`created_at`)**, not Lamport (`ts`).

This contradicts the spec which says `ORDER BY lamport, event.id`.

### 2.5 Event Sync (Multi-Machine)

During `tk ingest` (importing events from another machine):

```go
// From ingest.go line 95
db.BumpLamport(event.TS)
```

When ingesting an event:
1. Check if event.TS > local Lamport counter
2. If yes, update local counter to event.TS
3. This ensures local clock is always ahead of any seen event

**This is standard Lamport clock behavior.**

### 2.6 Conflict Resolution (Reducer)

The reducer uses **Lamport timestamps** for conflict resolution:

```go
// From reducer.go line 224-229
latestTS := int64(0)
for _, claim := range axis.Claims {
    if claim.TS > latestTS {
        latestTS = claim.TS
    }
}
```

**Key insight**: The reducer finds claims with the highest `TS` (Lamport timestamp), then resolves conflicts among concurrent claims by authority level.

---

## 3. The Bug History

### 3.1 Original Bug (Fixed in commit 67d8678)

**Problem**: Lamport counter was resetting to 1 on each command run.

**Symptoms**:
- All events from one session had TS=1
- All events from next session had TS=1 again
- Conflict resolution broke (couldn't determine latest claim)

**Fix**: Store Lamport counter in database `metadata` table instead of global variable.

### 3.2 Workaround (Commit ce811b0)

**Problem**: Event ordering was using Lamport timestamps, but they were all 1.

**Workaround**: Change `GetEvents()` from `ORDER BY ts, id` to `ORDER BY created_at, id`.

**Justification from commit message**:
> "Fix event ordering to use created_at instead of Lamport timestamp
> - Lamport counter resets on each run, causing incorrect ordering"

### 3.3 Current State

1. ✅ Lamport counter bug is **fixed** (counter increments properly: 1, 2, 3...)
2. ⚠️ **Workaround still in place**: Code still uses `ORDER BY created_at, id`
3. ⚠️ **Spec violation**: Ordering doesn't match spec's `ORDER BY lamport, event.id`

---

## 4. Why This Works (For Now)

The current implementation (ordering by `created_at`) works in single-machine scenarios because:

1. **Wall clocks are monotonic**: `time.Now()` increases with each event
2. **Events are created sequentially**: No concurrent event creation in CLI
3. **Test helpers use same `now`**: Events sort by ID when `created_at` is equal

### 4.1 When It Could Break

The current implementation could produce incorrect ordering in these scenarios:

1. **Clock adjustments**: NTP or manual clock changes could make newer events appear older
2. **Fast sequences**: Events created in same microsecond might sort incorrectly
3. **Sync from past**: Ingesting events with old `created_at` but high `TS` could sort wrong
4. **Multi-machine**: Two machines with different wall clocks would order events differently

**Example failure scenario**:
```
Machine A (wall clock ahead):
  Event 1: TS=1, created_at=2025-11-02 12:00:00

Machine B (wall clock behind):  
  Event 2: TS=2, created_at=2025-11-02 11:59:00

With ORDER BY created_at: Event 2, Event 1 (WRONG)
With ORDER BY ts:         Event 1, Event 2 (CORRECT)
```

---

## 5. Inconsistencies Found

### 5.1 Spec vs Implementation

| Aspect | Spec (spec-v1.md) | Implementation |
|--------|-------------------|----------------|
| Event ordering | `(lamport, event.id)` | `(created_at, id)` |
| Purpose of Lamport | Deterministic cross-machine order | Conflict resolution only |
| Primary sort key | Logical time (TS) | Physical time (CreatedAt) |

### 5.2 Internal Consistency

Within the implementation:

✅ **Consistent**:
- Lamport counter increments properly
- BumpLamport works correctly during sync
- Reducer uses Lamport for conflict resolution

❌ **Inconsistent**:
- Reducer expects events in some order (uses TS for conflicts)
- Database returns events in different order (by created_at)
- Test helpers assume ordering by ID (use TS=0)

### 5.3 Questions Raised

1. **Was the workaround meant to be temporary?**
   - The Lamport bug (commit 67d8678) was fixed after the workaround (commit ce811b0)
   - But the workaround was never reverted

2. **Does sync actually work correctly?**
   - Synced events have both TS and created_at from remote machine
   - If machines have different wall clocks, ordering could be wrong
   - Has this been tested with machines in different timezones?

3. **Do test patterns hide the issue?**
   - Tests use TS=0 and same `now` value
   - This makes the issue invisible in tests
   - Real-world multi-machine scenarios might behave differently

---

## 6. Recommendations

### 6.1 For Immediate Use

**Current code works for**:
- Single machine usage
- Clock-synchronized machines (same timezone)
- Test scenarios

**Use caution with**:
- Multi-machine sync across timezones
- Clock adjustments during operation
- High-frequency event creation

### 6.2 For Writing Tests

**Pattern for integration tests**:

```go
// Create events with proper timestamps
ts, _ := db.GetNextLamportTS()      // Get incrementing TS
event := types.Event{
    TS:        ts,                   // Use real Lamport timestamp
    CreatedAt: time.Now(),           // Each event gets current time
    ...
}
```

**Do NOT use this pattern**:
```go
// This pattern hides ordering issues
now := time.Now()
event1 := types.Event{TS: 0, CreatedAt: now, ...}
event2 := types.Event{TS: 0, CreatedAt: now, ...}
```

### 6.3 For Writing Commands

**Always set both timestamps**:

```go
ts, err := db.GetNextLamportTS()
if err != nil {
    return err
}

event := types.Event{
    ID:        eventID,
    TS:        ts,                // Lamport timestamp
    CreatedAt: time.Now(),        // Wall clock
    ...
}
```

**Never leave CreatedAt unset** - it will default to zero and break ordering.

### 6.4 Potential Fixes (Not Implemented)

**Option A**: Revert to spec's `ORDER BY ts, id`
- Pros: Matches spec, correct for distributed systems
- Cons: Need to test all edge cases, could break something subtle
- Risk: Medium

**Option B**: Keep current behavior, document it
- Pros: No risk, works for current usage
- Cons: Spec violation, potential future issues
- Risk: Low now, higher if multi-machine sync grows

**Option C**: Use both: `ORDER BY ts, created_at, id`
- Pros: Preserves Lamport ordering, created_at as secondary
- Cons: Mixing logical and physical time might be confusing
- Risk: Low

---

## 7. Critical Insights

### 7.1 The Real Ordering Guarantee

**What tk currently guarantees**:
> Events are ordered by the wall-clock time they were created on each machine, with event ID as tie-breaker.

**What the spec says tk should guarantee**:
> Events are ordered by their Lamport timestamp (logical time), with event ID as tie-breaker, ensuring deterministic replay across machines.

### 7.2 When It Matters

The difference matters when:
- Syncing events between machines
- One machine's clock is ahead/behind
- Events created close together in time
- Correctness depends on causal order

### 7.3 Current Safety

The current implementation is **safe** because:
- CLI is single-threaded (one event at a time)
- Wall clocks are usually monotonic within a session
- Event IDs provide deterministic tie-breaking
- Sync is rare (most users use single machine)

But it's **not correct** according to distributed systems principles.

---

## 8. Testing Implications

### 8.1 Why Tests Pass

Seed helpers use this pattern:
```go
now := time.Now()
// All events get same CreatedAt
seedProject(..., CreatedAt: now, TS: 0)
seedTask(..., CreatedAt: now, TS: 0)
```

With `ORDER BY created_at, id`:
- Events with same `created_at` sort by `id`
- Event IDs are generated sequentially
- So events appear in creation order

**This pattern accidentally matches the spec's intended behavior.**

### 8.2 What Tests Don't Catch

Tests don't catch:
- Ordering issues when `created_at` differs
- Sync scenarios with events from different machines
- Clock adjustment edge cases
- Lamport timestamp conflicts

### 8.3 Better Test Pattern

To properly test distributed scenarios:

```go
// Simulate two machines
nodeA := "nodeA"
nodeB := "nodeB"

// Machine A creates event at TS=1
eventA := types.Event{
    TS:        1,
    CreatedAt: time.Now().Add(-1 * time.Hour), // Hour ago
    Node:      nodeA,
}

// Machine B creates event at TS=2
eventB := types.Event{
    TS:        2,
    CreatedAt: time.Now(),                      // Now
    Node:      nodeB,
}

// When ordered by TS: A, B (correct causal order)
// When ordered by created_at: B, A (wrong!)
```

---

## 9. For Future Contributors

### 9.1 Key Principles

1. **Always increment Lamport timestamps**: Use `db.GetNextLamportTS()`
2. **Always set wall clock**: Use `time.Now()`
3. **Understand both clocks**: Lamport for order, wall clock for humans
4. **Test with both**: Create tests with varying timestamps

### 9.2 When to Use Which Timestamp

| Use Case | Use This |
|----------|----------|
| Sorting events for replay | `TS` (Lamport) per spec, `created_at` currently |
| Displaying to user | `CreatedAt` (wall clock) |
| Conflict resolution | `TS` (Lamport) |
| Detecting causality | `TS` (Lamport) |
| Audit trails | `CreatedAt` (wall clock) |
| Sync/bump counter | `TS` (Lamport) |

### 9.3 Red Flags

Watch out for:
- Setting `TS: 0` in non-test code
- Leaving `CreatedAt` unset (defaults to zero)
- Comparing wall clocks across machines
- Assuming events sort by creation order without checking timestamps

---

## 10. Conclusion

**TL;DR**:

1. The **spec says**: Order events by Lamport timestamp
2. The **code does**: Order events by wall clock
3. **Why it works**: Single-machine usage with monotonic clocks
4. **When it breaks**: Multi-machine sync with clock skew
5. **What to do**: Use proper Lamport + wall clock timestamps in all code

The current implementation is a **working workaround** that diverges from the spec but functions correctly in practice for the current usage patterns. Future work on multi-machine sync may require revisiting this decision.

---

## Appendix: Relevant Code Locations

- Event type definition: `tk/internal/types/event.go`
- Event retrieval: `tk/internal/database/db_events.go:33` (ORDER BY)
- Lamport counter: `tk/internal/database/db_counters.go`
- Conflict resolution: `tk/internal/reducer/reducer.go:211-257`
- Test helpers: `tk/cmd/test_helpers.go`
- Sync/ingest: `tk/cmd/ingest.go:95` (BumpLamport)
- v1 Spec: `tk/specs/spec-v1.md:113`

