package t12

import "sync"

// state models a piece of data protected by a mutex.
type state struct {
	mu    sync.Mutex
	value int
}

// @claim[t12]: foo keeps state within safe bounds
// @proof[t12]:
// helper cannot break the invariant because helper always locks the mutex
// before touching shared state, and helper normalizes negative inputs to zero
// before adding them.
// foo only calls helper for non-negative deltas, so it preserves the invariant.
//
// This claim structure mirrors the production bug: the proof depends on reasoning
// about helper that itself has supporting details. The checker should verify
// these nested claims properly.
func (s *state) helper(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if delta < 0 {
		delta = 0
	}

	s.value += delta
}

func (s *state) foo(delta int) {
	if delta < 0 {
		return
	}

	s.helper(delta)
}
