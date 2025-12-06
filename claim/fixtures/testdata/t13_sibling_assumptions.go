package t13

import "sync"

// This test catches proofs that assume sibling operations maintain invariants
// when they don't. The proof claims helperAdd only adds, but another goroutine
// running helperSubtract can make value negative between helperAdd and helperClamp.

type counter struct {
	mu    sync.Mutex
	value int
}

// @claim[t13]: doWork always leaves counter.value non-negative
// @proof[t13]:
// helperAdd cannot reduce counter.value because it converts negative increments
// to zero. helperClamp at the end ensures any negative value is zeroed.
// Therefore doWork always leaves counter.value non-negative.

func (c *counter) helperAdd(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if delta < 0 {
		delta = 0
	}
	c.value += delta
}

func (c *counter) helperSubtract(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value -= delta // Can make value negative!
}

func (c *counter) helperClamp() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.value < 0 {
		c.value = 0
	}
}

func (c *counter) doWork(delta int) {
	if delta < 0 {
		delta = 0
	}

	c.helperAdd(delta)
	// Another goroutine could call helperSubtract here, making value negative!
	c.helperClamp()
	// Value might be 0 from clamp, but then helperSubtract runs again...
}

func (c *counter) concurrentSubtract(delta int) {
	// This is called from another goroutine
	c.helperSubtract(delta)
}
