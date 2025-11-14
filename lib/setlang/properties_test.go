package setlang

import (
	"testing"

	"pgregory.net/rapid"
)

// Generators for rapid property testing

// genSet generates a random set of integers
func genSet() *rapid.Generator[*Set[int]] {
	return rapid.Custom(func(t *rapid.T) *Set[int] {
		items := rapid.SliceOf(rapid.IntRange(0, 100)).Draw(t, "items")
		return NewSetFrom(items...)
	})
}

// genSmallSet generates a small set for more focused testing
func genSmallSet() *rapid.Generator[*Set[int]] {
	return rapid.Custom(func(t *rapid.T) *Set[int] {
		items := rapid.SliceOfN(rapid.IntRange(0, 10), 0, 5).Draw(t, "items")
		return NewSetFrom(items...)
	})
}

// Property: Union is commutative
func TestProperty_UnionCommutative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		ab := a.Union(b)
		ba := b.Union(a)

		if !setsEqual(ab, ba) {
			t.Fatalf("union not commutative: a|b != b|a")
		}
	})
}

// Property: Union is associative
func TestProperty_UnionAssociative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")
		c := genSet().Draw(t, "c")

		ab_c := a.Union(b).Union(c)
		a_bc := a.Union(b.Union(c))

		if !setsEqual(ab_c, a_bc) {
			t.Fatalf("union not associative: (a|b)|c != a|(b|c)")
		}
	})
}

// Property: Intersection is commutative
func TestProperty_IntersectCommutative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		ab := a.Intersect(b)
		ba := b.Intersect(a)

		if !setsEqual(ab, ba) {
			t.Fatalf("intersect not commutative: a&b != b&a")
		}
	})
}

// Property: Intersection is associative
func TestProperty_IntersectAssociative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")
		c := genSet().Draw(t, "c")

		ab_c := a.Intersect(b).Intersect(c)
		a_bc := a.Intersect(b.Intersect(c))

		if !setsEqual(ab_c, a_bc) {
			t.Fatalf("intersect not associative: (a&b)&c != a&(b&c)")
		}
	})
}

// Property: Empty set is identity for union
func TestProperty_UnionIdentity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		empty := NewSet[int]()

		result := a.Union(empty)

		if !setsEqual(a, result) {
			t.Fatalf("empty not identity for union: a|∅ != a")
		}
	})
}

// Property: Intersection with self is identity
func TestProperty_IntersectIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")

		result := a.Intersect(a)

		if !setsEqual(a, result) {
			t.Fatalf("intersect not idempotent: a&a != a")
		}
	})
}

// Property: Union with self is identity
func TestProperty_UnionIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")

		result := a.Union(a)

		if !setsEqual(a, result) {
			t.Fatalf("union not idempotent: a|a != a")
		}
	})
}

// Property: Difference with self is empty
func TestProperty_DiffSelfEmpty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")

		result := a.Diff(a)

		if !result.IsEmpty() {
			t.Fatalf("diff with self not empty: a-a != ∅")
		}
	})
}

// Property: Difference with empty is identity
func TestProperty_DiffEmptyIdentity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		empty := NewSet[int]()

		result := a.Diff(empty)

		if !setsEqual(a, result) {
			t.Fatalf("diff with empty not identity: a-∅ != a")
		}
	})
}

// Property: De Morgan's Law - (A ∪ B)' = A' ∩ B'
// Since we don't have complement, we test: U - (A ∪ B) = (U - A) ∩ (U - B)
func TestProperty_DeMorganUnion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		u := genSet().Draw(t, "universe")
		a := genSmallSet().Draw(t, "a")
		b := genSmallSet().Draw(t, "b")

		// Left side: U - (A ∪ B)
		left := u.Diff(a.Union(b))

		// Right side: (U - A) ∩ (U - B)
		right := u.Diff(a).Intersect(u.Diff(b))

		if !setsEqual(left, right) {
			t.Fatalf("De Morgan's law violated for union: U-(A|B) != (U-A)&(U-B)")
		}
	})
}

// Property: De Morgan's Law - (A ∩ B)' = A' ∪ B'
// We test: U - (A ∩ B) = (U - A) ∪ (U - B)
func TestProperty_DeMorganIntersect(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		u := genSet().Draw(t, "universe")
		a := genSmallSet().Draw(t, "a")
		b := genSmallSet().Draw(t, "b")

		// Left side: U - (A ∩ B)
		left := u.Diff(a.Intersect(b))

		// Right side: (U - A) ∪ (U - B)
		right := u.Diff(a).Union(u.Diff(b))

		if !setsEqual(left, right) {
			t.Fatalf("De Morgan's law violated for intersect: U-(A&B) != (U-A)|(U-B)")
		}
	})
}

// Property: Distributive law - A ∩ (B ∪ C) = (A ∩ B) ∪ (A ∩ C)
func TestProperty_IntersectDistributesOverUnion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")
		c := genSet().Draw(t, "c")

		// Left side: A ∩ (B ∪ C)
		left := a.Intersect(b.Union(c))

		// Right side: (A ∩ B) ∪ (A ∩ C)
		right := a.Intersect(b).Union(a.Intersect(c))

		if !setsEqual(left, right) {
			t.Fatalf("intersect doesn't distribute over union: A&(B|C) != (A&B)|(A&C)")
		}
	})
}

// Property: Distributive law - A ∪ (B ∩ C) = (A ∪ B) ∩ (A ∪ C)
func TestProperty_UnionDistributesOverIntersect(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")
		c := genSet().Draw(t, "c")

		// Left side: A ∪ (B ∩ C)
		left := a.Union(b.Intersect(c))

		// Right side: (A ∪ B) ∩ (A ∪ C)
		right := a.Union(b).Intersect(a.Union(c))

		if !setsEqual(left, right) {
			t.Fatalf("union doesn't distribute over intersect: A|(B&C) != (A|B)&(A|C)")
		}
	})
}

// Property: Absorption law - A ∪ (A ∩ B) = A
func TestProperty_UnionAbsorption(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		result := a.Union(a.Intersect(b))

		if !setsEqual(a, result) {
			t.Fatalf("union absorption violated: A|(A&B) != A")
		}
	})
}

// Property: Absorption law - A ∩ (A ∪ B) = A
func TestProperty_IntersectAbsorption(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		result := a.Intersect(a.Union(b))

		if !setsEqual(a, result) {
			t.Fatalf("intersect absorption violated: A&(A|B) != A")
		}
	})
}

// Property: A - B and B are disjoint
func TestProperty_DiffDisjoint(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		diff := a.Diff(b)
		intersection := diff.Intersect(b)

		if !intersection.IsEmpty() {
			t.Fatalf("A-B and B not disjoint: (A-B)&B != ∅")
		}
	})
}

// Property: (A - B) ∪ B ⊇ A if A ⊆ (A ∪ B)
func TestProperty_DiffUnionSuperset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		result := a.Diff(b).Union(b)

		// Check that result contains all elements of a or b
		for _, item := range a.Items() {
			if !result.Has(item) {
				t.Fatalf("(A-B)|B doesn't contain element from A")
			}
		}
		for _, item := range b.Items() {
			if !result.Has(item) {
				t.Fatalf("(A-B)|B doesn't contain element from B")
			}
		}
	})
}

// Property: Size relationships
func TestProperty_UnionSize(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		union := a.Union(b)

		// |A ∪ B| >= max(|A|, |B|)
		if union.Size() < a.Size() || union.Size() < b.Size() {
			t.Fatalf("union size less than component sizes")
		}

		// |A ∪ B| <= |A| + |B|
		if union.Size() > a.Size()+b.Size() {
			t.Fatalf("union size greater than sum of component sizes")
		}
	})
}

// Property: Intersection size
func TestProperty_IntersectSize(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		intersection := a.Intersect(b)

		// |A ∩ B| <= min(|A|, |B|)
		minSize := a.Size()
		if b.Size() < minSize {
			minSize = b.Size()
		}

		if intersection.Size() > minSize {
			t.Fatalf("intersection size greater than min component size")
		}
	})
}

// Property: Difference size
func TestProperty_DiffSize(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := genSet().Draw(t, "b")

		diff := a.Diff(b)

		// |A - B| <= |A|
		if diff.Size() > a.Size() {
			t.Fatalf("diff size greater than first set size")
		}
	})
}

// Property: Clone creates equal but independent sets
func TestProperty_CloneIndependent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := genSet().Draw(t, "a")
		b := a.Clone()

		// Should be equal initially
		if !setsEqual(a, b) {
			t.Fatalf("clone not equal to original")
		}

		// Modify clone
		b.Add(999)

		// Original should be unchanged
		if a.Has(999) {
			t.Fatalf("modifying clone affected original")
		}
	})
}

// Helper function to check if two sets are equal
func setsEqual[T comparable](a, b *Set[T]) bool {
	if a.Size() != b.Size() {
		return false
	}

	for _, item := range a.Items() {
		if !b.Has(item) {
			return false
		}
	}

	return true
}
