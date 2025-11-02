# tk Timestamp and Event Ordering Model

## Document Metadata

**Written**: November 2, 2025  
**Author**: AI assistant (Claude), working with Emily  
**Context**: Created during metadata implementation work to document timestamp handling

**Sources**:
- `~/code/tk1.md` - Original ChatGPT design dialogue (5,229 lines)
- `tk/specs/spec-v1.md` - v1 sync specification
- `tk/internal/database/db_events.go` - Event storage and retrieval implementation
- `tk/internal/reducer/reducer.go` - Event application and conflict resolution
- `tk/cmd/test_helpers.go` - Testing patterns

**How to use this document**:

1. **For implementing new commands**: See section 4 for the correct pattern to create events
2. **For writing tests**: See section 5 for proper test patterns
3. **For understanding event ordering**: See sections 2-3 for why we have two timestamps
4. **For multi-machine sync**: See section 3 for how Lamport timestamps enable sync

---

## 1. Overview

tk uses an **event-sourced architecture** where all changes are recorded as immutable events. To support multi-machine sync while maintaining deterministic ordering, events have **two timestamps**:

1. **Lamport timestamp (`TS`)** - Logical clock for distributed ordering
2. **Wall clock (`CreatedAt`)** - Physical timestamp for humans

This document explains when and how to use each timestamp.

---

## 2. The Two Timestamps

### 2.1 Lamport Timestamp (`TS`)

```go
type Event struct {
    TS int64  // Lamport timestamp
    ...
}
```

**Purpose**: Provide a partial order across events from different machines.

**Guarantees**: If event A causally precedes event B, then `A.TS < B.TS`.

**Use cases**:
- Ordering events for replay (primary sort key)
- Conflict resolution in reducer
- Detecting causal relationships
- Multi-machine sync

**How it works**:
- Each machine maintains a Lamport counter in the database
- Counter increments on every local write: 1, 2, 3, 4...
- When ingesting events from another machine, bump counter to `max(local, remote.TS) + 1`
- Persisted in `metadata` table with key `lamport_counter`

### 2.2 Wall Clock (`CreatedAt`)

```go
type Event struct {
    CreatedAt time.Time  // Wall clock timestamp
    ...
}
```

**Purpose**: Record actual time the event occurred.

**Use cases**:
- Displaying timestamps to users
- Audit trails
- Time-based queries (e.g., "events from last week")

**How it works**:
- Set to `time.Now()` when event is created
- Stored as nanoseconds since epoch in database
- May differ across machines (timezone, clock skew)

### 2.3 Why Both?

**Lamport solves**: "What happened first?" (logical causality)  
**Wall clock solves**: "When did it happen?" (human time)

Example:
```
Machine A creates event at 10:00 AM local time → TS=1
Machine B creates event at 9:00 AM local time → TS=2

Logical order: TS=1, then TS=2 (correct causal order)
Wall clock order: 9:00 AM, then 10:00 AM (different!)
```

---

## 3. Event Ordering

### 3.1 Database Ordering

All event queries use **Lamport timestamp ordering**:

```sql
SELECT * FROM events ORDER BY ts, id
```

**Sort key priority**:
1. **Primary**: Lamport timestamp (`ts`)
2. **Tie-breaker**: Event ID (`id`)

This ensures deterministic replay across all machines regardless of wall clock differences.

### 3.2 Why Event ID as Tie-Breaker?

When multiple events have the same Lamport timestamp (concurrent events from different machines), we need a deterministic tie-breaker.

Event IDs are globally unique and sortable:
- Format: `ev-<sequence>-<node>`
- Example: `ev-42-abc123`
- Lexicographic ordering provides determinism

### 3.3 Multi-Machine Sync Example

```
Machine A (timezone GMT):
  Event 1: TS=1, created_at=2025-11-02 12:00:00 GMT

Machine B (timezone PST, 8 hours behind):
  Event 2: TS=2, created_at=2025-11-02 04:00:00 PST (= 12:00:00 GMT)
  
When Machine A ingests Event 2:
  - Bump local Lamport: max(1, 2) + 1 = 3
  - Sort events by TS: Event 1 (TS=1), Event 2 (TS=2)
  - Wall clocks don't affect ordering
```

---

## 4. Creating Events in Commands

### 4.1 Correct Pattern

```go
// Get next Lamport timestamp
ts, err := db.GetNextLamportTS()
if err != nil {
    return err
}

// Create event
event := types.Event{
    ID:        eventID,
    TS:        ts,                // Lamport timestamp
    CreatedAt: time.Now(),        // Wall clock
    Actor:     actor,
    Role:      role,
    Kind:      "task.status.set",
    Payload:   payloadJSON,
}

// Insert event
if err := db.InsertEvent(event); err != nil {
    return err
}
```

### 4.2 Critical Rules

1. **Always call `GetNextLamportTS()`**: Never hardcode or reuse Lamport timestamps
2. **Always set `CreatedAt`**: Use `time.Now()`, never leave it zero
3. **Set both timestamps**: Both are required for proper operation

### 4.3 What NOT to Do

```go
// ❌ WRONG: TS=0
event := types.Event{TS: 0, CreatedAt: time.Now(), ...}

// ❌ WRONG: Missing CreatedAt
event := types.Event{TS: ts, ...}  // CreatedAt defaults to zero!

// ❌ WRONG: Reusing timestamp
ts := 1
event1 := types.Event{TS: ts, ...}
event2 := types.Event{TS: ts, ...}  // Should increment!
```

---

## 5. Writing Tests

### 5.1 Integration Test Pattern

```go
func TestMyFeature(t *testing.T) {
    db := openTempDB(t)
    projectUID := seedProject(t, db, "test")
    taskUID := seedTask(t, db, projectUID, "Test task", 1)
    
    // Create additional events with proper timestamps
    ts, _ := db.GetNextLamportTS()
    event := types.Event{
        ID:        string(types.NewEventID()),
        TS:        ts,                // Real Lamport timestamp
        CreatedAt: time.Now(),        // Real wall clock
        Actor:     "tester",
        Role:      "human",
        Kind:      string(types.EventKindTaskStatusSet),
        Payload:   mustJSON(t, payload),
    }
    db.InsertEvent(event)
    
    // Test the feature...
}
```

### 5.2 Seed Helpers Pattern

The `seedProject` and `seedTask` helpers use a special pattern:

```go
now := time.Now()
// All seed events use same wall clock time
event1 := types.Event{TS: 0, CreatedAt: now, ...}
event2 := types.Event{TS: 0, CreatedAt: now, ...}
```

**Why TS=0 works for seeds**:
- All seed events have identical `CreatedAt` timestamp
- When sorted by `ORDER BY ts, id`, events with same TS sort by ID
- Event IDs are sequential, so order is preserved
- This is acceptable for test setup since events don't need causal ordering

**When to use each pattern**:
- **Seed helpers** (TS=0): Setting up initial test state (projects, tasks)
- **Real timestamps**: Testing actual features that depend on ordering

### 5.3 Testing Multi-Machine Scenarios

To test that Lamport ordering works correctly:

```go
// Machine A creates event at TS=1, wall clock = now
eventA := types.Event{
    TS:        1,
    CreatedAt: time.Now(),
    ...
}

// Machine B creates event at TS=2, wall clock = 1 hour ago
// (simulates machine with clock skew)
eventB := types.Event{
    TS:        2,  // Logically happens AFTER eventA
    CreatedAt: time.Now().Add(-1 * time.Hour),  // Wall clock shows earlier
    ...
}

// Events should be ordered: eventA, eventB (by TS)
// Not: eventB, eventA (by created_at)
```

---

## 6. Conflict Resolution

The reducer uses Lamport timestamps to determine which claims are latest:

```go
// Find the latest timestamp
latestTS := int64(0)
for _, claim := range axis.Claims {
    if claim.TS > latestTS {
        latestTS = claim.TS
    }
}

// Get all claims at the latest timestamp (concurrent claims)
var concurrentClaims []types.Claim
for _, claim := range axis.Claims {
    if claim.TS == latestTS {
        concurrentClaims = append(concurrentClaims, claim)
    }
}

// Sort concurrent claims by authority
sort.Slice(concurrentClaims, func(i, j int) bool {
    return GetRoleAuthority(i.Role) > GetRoleAuthority(j.Role)
})

// Highest authority wins
effectiveClaim := concurrentClaims[0]
```

**Key insight**: Among concurrent claims (same TS), authority decides. Among sequential claims (different TS), latest TS wins.

---

## 7. Sync and Lamport Clocks

### 7.1 Ingest Process

When ingesting events from another machine:

```go
// From ingest.go
for event := range remoteEvents {
    // Insert event
    db.InsertEvent(event)
    
    // Bump local Lamport clock if remote TS is higher
    db.BumpLamport(event.TS)
}
```

### 7.2 BumpLamport Logic

```go
func (d *DB) BumpLamport(newValue int64) error {
    currentCounter := getCounterFromDB()
    if newValue > currentCounter {
        updateCounterInDB(newValue)
    }
    return nil
}
```

This ensures the local Lamport clock always reflects the highest timestamp seen from any machine.

### 7.3 Multi-Machine Example

```
Initial state:
  Machine A: Lamport = 0
  Machine B: Lamport = 0

Machine A creates task:
  Event: TS=1
  Machine A: Lamport = 1

Machine A syncs to Machine B:
  Machine B ingests event with TS=1
  Machine B bumps Lamport: max(0, 1) = 1
  Machine B: Lamport = 1

Machine B updates task:
  GetNextLamportTS() returns 2
  Event: TS=2
  Machine B: Lamport = 2

Machine B syncs to Machine A:
  Machine A ingests event with TS=2
  Machine A bumps Lamport: max(1, 2) = 2
  Machine A: Lamport = 2

Both machines now have same Lamport clock and same event order.
```

---

## 8. Summary

### 8.1 Quick Reference

| Aspect | Use This |
|--------|----------|
| Ordering events | `TS` (Lamport) |
| Displaying to users | `CreatedAt` (wall clock) |
| Conflict resolution | `TS` (Lamport) |
| Causality detection | `TS` (Lamport) |
| Audit trails | `CreatedAt` (wall clock) |
| Sync/bump counter | `TS` (Lamport) |

### 8.2 Implementation Checklist

When creating events:
- ✅ Call `db.GetNextLamportTS()`
- ✅ Set `TS` to the returned value
- ✅ Set `CreatedAt` to `time.Now()`
- ✅ Never hardcode timestamps
- ✅ Never leave timestamps unset

When querying events:
- ✅ Use `ORDER BY ts, id` for deterministic replay
- ✅ Trust the database to return events in correct order
- ✅ Don't rely on wall clock ordering

---

## Appendix: Code Locations

- **Event type**: `tk/internal/types/event.go`
- **Event queries**: `tk/internal/database/db_events.go`
- **Lamport counter**: `tk/internal/database/db_counters.go`
- **Conflict resolution**: `tk/internal/reducer/reducer.go:211-257`
- **Test helpers**: `tk/cmd/test_helpers.go`
- **Sync/ingest**: `tk/cmd/ingest.go`
- **Ordering tests**: `tk/internal/database/db_events_test.go`
- **v1 Spec**: `tk/specs/spec-v1.md:113`
