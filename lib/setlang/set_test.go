package setlang

import (
	"sort"
	"testing"
)

func TestNewSet(t *testing.T) {
	s := NewSet[int]()
	if s == nil {
		t.Fatal("NewSet returned nil")
	}
	if s.Size() != 0 {
		t.Errorf("expected size 0, got %d", s.Size())
	}
	if !s.IsEmpty() {
		t.Error("expected set to be empty")
	}
}

func TestNewSetFrom(t *testing.T) {
	s := NewSetFrom(1, 2, 3, 2, 1) // duplicates should be ignored
	if s.Size() != 3 {
		t.Errorf("expected size 3, got %d", s.Size())
	}
	if !s.Has(1) || !s.Has(2) || !s.Has(3) {
		t.Error("set missing expected items")
	}
}

func TestSetAdd(t *testing.T) {
	s := NewSet[string]()
	s.Add("foo")
	s.Add("bar")
	s.Add("foo") // duplicate

	if s.Size() != 2 {
		t.Errorf("expected size 2, got %d", s.Size())
	}
	if !s.Has("foo") || !s.Has("bar") {
		t.Error("set missing expected items")
	}
}

func TestSetHas(t *testing.T) {
	s := NewSetFrom("a", "b", "c")

	if !s.Has("a") {
		t.Error("expected set to have 'a'")
	}
	if s.Has("d") {
		t.Error("expected set not to have 'd'")
	}
}

func TestSetRemove(t *testing.T) {
	s := NewSetFrom(1, 2, 3)
	s.Remove(2)

	if s.Size() != 2 {
		t.Errorf("expected size 2, got %d", s.Size())
	}
	if s.Has(2) {
		t.Error("expected 2 to be removed")
	}
	if !s.Has(1) || !s.Has(3) {
		t.Error("other items should remain")
	}
}

func TestSetItems(t *testing.T) {
	s := NewSetFrom(3, 1, 2)
	items := s.Items()

	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// Sort for consistent comparison
	sort.Ints(items)
	expected := []int{1, 2, 3}
	for i, v := range expected {
		if items[i] != v {
			t.Errorf("expected item %d to be %d, got %d", i, v, items[i])
		}
	}
}

func TestSetUnion(t *testing.T) {
	s1 := NewSetFrom(1, 2, 3)
	s2 := NewSetFrom(3, 4, 5)

	result := s1.Union(s2)

	if result.Size() != 5 {
		t.Errorf("expected size 5, got %d", result.Size())
	}
	for i := 1; i <= 5; i++ {
		if !result.Has(i) {
			t.Errorf("expected result to have %d", i)
		}
	}
}

func TestSetIntersect(t *testing.T) {
	s1 := NewSetFrom(1, 2, 3, 4)
	s2 := NewSetFrom(3, 4, 5, 6)

	result := s1.Intersect(s2)

	if result.Size() != 2 {
		t.Errorf("expected size 2, got %d", result.Size())
	}
	if !result.Has(3) || !result.Has(4) {
		t.Error("expected result to have 3 and 4")
	}
	if result.Has(1) || result.Has(5) {
		t.Error("result should not have 1 or 5")
	}
}

func TestSetDiff(t *testing.T) {
	s1 := NewSetFrom(1, 2, 3, 4)
	s2 := NewSetFrom(3, 4, 5, 6)

	result := s1.Diff(s2)

	if result.Size() != 2 {
		t.Errorf("expected size 2, got %d", result.Size())
	}
	if !result.Has(1) || !result.Has(2) {
		t.Error("expected result to have 1 and 2")
	}
	if result.Has(3) || result.Has(4) {
		t.Error("result should not have 3 or 4")
	}
}

func TestSetClone(t *testing.T) {
	s1 := NewSetFrom("a", "b", "c")
	s2 := s1.Clone()

	if s2.Size() != s1.Size() {
		t.Error("clone should have same size")
	}

	// Modify clone
	s2.Add("d")
	if s1.Has("d") {
		t.Error("original set should not be modified")
	}
	if !s2.Has("d") {
		t.Error("clone should have new item")
	}
}

func TestSetOperationsWithEmpty(t *testing.T) {
	s1 := NewSetFrom(1, 2, 3)
	s2 := NewSet[int]()

	union := s1.Union(s2)
	if union.Size() != 3 {
		t.Errorf("union with empty should equal s1, got size %d", union.Size())
	}

	intersect := s1.Intersect(s2)
	if !intersect.IsEmpty() {
		t.Error("intersect with empty should be empty")
	}

	diff := s1.Diff(s2)
	if diff.Size() != 3 {
		t.Errorf("diff with empty should equal s1, got size %d", diff.Size())
	}
}

func TestSetOperationsWithNil(t *testing.T) {
	s1 := NewSetFrom(1, 2, 3)

	union := s1.Union(nil)
	if union.Size() != 3 {
		t.Errorf("union with nil should equal s1, got size %d", union.Size())
	}

	intersect := s1.Intersect(nil)
	if !intersect.IsEmpty() {
		t.Error("intersect with nil should be empty")
	}

	diff := s1.Diff(nil)
	if diff.Size() != 3 {
		t.Errorf("diff with nil should equal s1, got size %d", diff.Size())
	}
}
