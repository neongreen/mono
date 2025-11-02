package main

import (
	"encoding/json"
	"github.com/neongreen/mono/tk/internal/types"
	"strings"
	"testing"
)

func TestDescribeTask(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "alpha")
	taskUID := seedTask(t, db, projectUID, "Task Alpha", 4)

	info, err := describeTask(db, "alpha-4")
	if err != nil {
		t.Fatalf("describeTask failed: %v", err)
	}

	if info.TaskUID != taskUID {
		t.Fatalf("expected task uid %s, got %s", taskUID, info.TaskUID)
	}
	if info.DisplayID != "alpha-4" {
		t.Fatalf("expected display alpha-4, got %s", info.DisplayID)
	}
	if info.ProjectUID != projectUID {
		t.Fatalf("expected project %s, got %s", projectUID, info.ProjectUID)
	}
	if info.Number != 4 {
		t.Fatalf("expected number 4, got %d", info.Number)
	}
	if info.Collides {
		t.Fatalf("expected no collision")
	}
	if info.Title != "Task Alpha" {
		t.Fatalf("expected title Task Alpha, got %s", info.Title)
	}
}

func TestDescribeTaskCollisionDisplay(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "alpha")
	nodeA := "NODE_A"
	nodeB := "NODE_B"
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	taskA := seedTaskWithNode(t, db, projectUID, "Task A", 5, nodeA)
	taskB := seedTaskWithNode(t, db, projectUID, "Task B", 5, nodeB)

	info, err := describeTask(db, "alpha-5-"+types.NodeID(nodeB).Short())
	if err != nil {
		t.Fatalf("describeTask failed: %v", err)
	}
	if info.TaskUID != taskB {
		t.Fatalf("expected taskB %s, got %s", taskB, info.TaskUID)
	}
	if !info.Collides {
		t.Fatalf("expected collision detected")
	}
	if !strings.Contains(info.DisplayID, types.NodeID(nodeB).Short()) {
		t.Fatalf("expected display to include hint, got %s", info.DisplayID)
	}

	infoA, err := describeTask(db, "alpha-5-"+types.NodeID(nodeA).Short())
	if err != nil {
		t.Fatalf("describeTask failed: %v", err)
	}
	if infoA.TaskUID != taskA {
		t.Fatalf("expected taskA %s, got %s", taskA, infoA.TaskUID)
	}
	if !infoA.Collides {
		t.Fatalf("expected collision for taskA")
	}
}

func TestDescribeTaskJSONRoundTrip(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "alpha")
	seedTask(t, db, projectUID, "Task", 2)

	info, err := describeTask(db, "alpha-2")
	if err != nil {
		t.Fatalf("describeTask failed: %v", err)
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded taskIdentity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.TaskUID != info.TaskUID {
		t.Fatalf("expected roundtrip task uid %s, got %s", info.TaskUID, decoded.TaskUID)
	}
}
