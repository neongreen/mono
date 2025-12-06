package t14

// This test catches when a proof only covers a subset of what the claim states.
// The claim says "no send on ANY closed channel" but the proof only covers ch1.
// ch2 is also closed and has a sender, but the proof doesn't mention it.

// @claim[t14]: process() cannot panic from sending on any closed channel
// @proof[t14]:
// ch1 is closed only after the sender goroutine exits.
// The sender goroutine exits when it reads from done.
// done is closed before ch1 is closed.
// Therefore no send on ch1 can happen after ch1 is closed.

func process() {
	ch1 := make(chan int)
	ch2 := make(chan string) // NOT covered by proof!
	done := make(chan struct{})

	// Sender for ch1 - properly synchronized
	go func() {
		<-done
		// exits before ch1 closes
	}()

	// Sender for ch2 - will panic!
	go func() {
		ch2 <- "oops" // sends after close
	}()

	close(ch2) // closes before sender exits
	close(done)
	close(ch1)
}
