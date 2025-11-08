package remote

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

func TestPushRestoresMissingSegments(t *testing.T) {
	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "remote")
	stateDir := filepath.Join(tmpDir, "state")
	if err := os.MkdirAll(remotePath, 0o755); err != nil {
		t.Fatalf("failed to create remote dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "tk.db")
	db, err := database.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	// Insert a single event so push has something to export.
	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport: %v", err)
	}
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}
	taskUID, err := utils.GenerateTaskUUID()
	if err != nil {
		t.Fatalf("failed to generate task UUID: %v", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: time.Date(2025, 10, 24, 12, 0, 0, 0, time.UTC),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_uid":"` + taskUID + `","title":"test task","created_by":"test-user"}`),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	remoteName := "icloud"
	space := "personal"
	remoteConfig := config.RemoteConfig{
		Type:   "folder",
		Path:   remotePath,
		Spaces: []string{space},
		Push:   true,
		Pull:   true,
	}

	pushParams := PushParams{
		RemoteName:   remoteName,
		RemoteConfig: remoteConfig,
		Space:        space,
		StateDir:     stateDir,
		SyncConfig:   config.DefaultSyncConfig(),
	}

	// First push writes the segment and caches it.
	result, err := Push(db, pushParams)
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if result == nil {
		t.Fatalf("expected push result, got nil")
	}

	localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
	localIndex, err := LoadIndexFile(localIndexPath)
	if err != nil {
		t.Fatalf("failed to load local index: %v", err)
	}
	if localIndex == nil || len(localIndex.Segments) == 0 {
		t.Fatalf("expected at least one segment in local index")
	}
	segmentInfo := localIndex.Segments[0]

	cachePath := filepath.Join(stateDir, "remotes", remoteName, segmentInfo.Rel)
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cached segment at %s: %v", cachePath, err)
	}

	remoteSegmentPath := filepath.Join(remotePath, segmentInfo.Rel)
	if _, err := os.Stat(remoteSegmentPath); err != nil {
		t.Fatalf("expected remote segment at %s: %v", remoteSegmentPath, err)
	}

	// Delete the remote segment to simulate a missing file.
	if err := os.Remove(remoteSegmentPath); err != nil {
		t.Fatalf("failed to remove remote segment: %v", err)
	}

	// Second push should restore the segment from cache without error.
	if _, err := Push(db, pushParams); err != nil {
		t.Fatalf("push after deletion failed: %v", err)
	}

	if _, err := os.Stat(remoteSegmentPath); err != nil {
		t.Fatalf("expected remote segment to be restored, but stat failed: %v", err)
	}
}

