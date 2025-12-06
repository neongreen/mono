package t18

// This is a TRUE claim with VALID axiom usage (NOT circular).
// Claim A references axiom B. B's proof does NOT reference A.
// This is a directed acyclic dependency, not circular.
//
// If Claude returns UNPROVEN citing "circular reasoning", it's a FALSE POSITIVE.

// @claim[channel-closed-once]: ch is only closed in cleanup()
// @proof[channel-closed-once]:
// The code shows close(ch) only in cleanup(). No other close() calls exist.

// @claim[t18]: process() cannot panic from closing a closed channel
// @proof[t18]:
// By @see[channel-closed-once], ch is only closed in cleanup().
// cleanup() is only called once at the end of process().
// Therefore ch cannot be closed twice, so no panic. ∎

var ch chan int

func process() {
	ch = make(chan int)
	// ... do work ...
	cleanup()
}

func cleanup() {
	close(ch)
}
