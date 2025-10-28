package main

import "testing"

func TestMigrationHandlersCoverAllEventKinds(t *testing.T) {
	for _, kind := range AllEventKinds {
		handler, ok := migrationHandlerForKind(kind)
		if !ok {
			t.Fatalf("no migration handler registered for %s", kind)
		}
		if handler == nil {
			t.Fatalf("nil migration handler registered for %s", kind)
		}
	}
}
