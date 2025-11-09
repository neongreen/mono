package deprecation

import (
	"fmt"
	"sync"
	"time"
)

// UsageRecord tracks usage of a deprecated field
type UsageRecord struct {
	FieldPath string    // e.g., "ProjectAliasAddPayload.Alias"
	EventKind string    // e.g., "project.alias.add"
	Count     int64     // how many times seen
	FirstSeen time.Time // when first encountered
	LastSeen  time.Time // when last encountered
}

// Tracker tracks deprecated field usage
type Tracker struct {
	mu      sync.Mutex
	records map[string]*UsageRecord // keyed by FieldPath
}

// NewTracker creates a new deprecation tracker
func NewTracker() *Tracker {
	return &Tracker{
		records: make(map[string]*UsageRecord),
	}
}

// RecordUsage records that a deprecated field was seen
func (t *Tracker) RecordUsage(fieldPath, eventKind string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	record, exists := t.records[fieldPath]
	if !exists {
		record = &UsageRecord{
			FieldPath: fieldPath,
			EventKind: eventKind,
			Count:     0,
			FirstSeen: now,
			LastSeen:  now,
		}
		t.records[fieldPath] = record
	}

	record.Count++
	record.LastSeen = now
}

// GetStats returns all usage records
func (t *Tracker) GetStats() []UsageRecord {
	t.mu.Lock()
	defer t.mu.Unlock()

	stats := make([]UsageRecord, 0, len(t.records))
	for _, record := range t.records {
		stats = append(stats, *record)
	}
	return stats
}

// GetStat returns usage record for a specific field
func (t *Tracker) GetStat(fieldPath string) (UsageRecord, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.records[fieldPath]
	if !exists {
		return UsageRecord{}, false
	}
	return *record, true
}

// Reset clears all tracking data
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.records = make(map[string]*UsageRecord)
}

// PrintSummary prints a human-readable summary
func (t *Tracker) PrintSummary() string {
	stats := t.GetStats()
	if len(stats) == 0 {
		return "No deprecated field usage found. Safe to remove all deprecated code!"
	}

	var summary string
	summary += fmt.Sprintf("Found deprecated field usage:\n\n")

	// Group by event kind
	byEventKind := make(map[string][]UsageRecord)
	for _, stat := range stats {
		byEventKind[stat.EventKind] = append(byEventKind[stat.EventKind], stat)
	}

	for eventKind, records := range byEventKind {
		summary += fmt.Sprintf("  %s:\n", eventKind)
		for _, record := range records {
			summary += fmt.Sprintf("    - %s: %d events (last: %s)\n",
				record.FieldPath,
				record.Count,
				record.LastSeen.Format("2006-01-02"))
		}
		summary += "\n"
	}

	summary += fmt.Sprintf("Total: %d deprecated field(s) in use\n", len(stats))
	return summary
}
