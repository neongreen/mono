package t3

import "sync"

// @claim[t3]: no panics from waitgroup usage
// @proof[t3]:
// Every Done has a matching Add.
// Add happens before any goroutine can call Done.

func abc() {
	var wg sync.WaitGroup
	go func() { wg.Done() }()
	wg.Wait()
}
