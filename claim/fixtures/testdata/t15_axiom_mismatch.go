package t15

// This test catches when axioms prove something about a specific channel
// but the claim is stated generically about "any channel".
// The mismatch between specific axioms and generic claim should be flagged.

// @claim[ch1-only-closed-in-one-place]: ch1 is only closed in cleanup()
// @proof[ch1-only-closed-in-one-place]:
// The code shows close(ch1) only in cleanup().

// @claim[t15]: process() cannot panic from sending on a closed channel
// @proof[t15]:
// By @see[ch1-only-closed-in-one-place], ch1 is only closed in cleanup().
// cleanup() is called after all senders exit.
// Therefore no send-on-closed-channel panic can occur.

func process() {
	ch1 := make(chan int)
	ch2 := make(chan string) // Axiom doesn't mention this!
	done := make(chan struct{})

	go func() {
		select {
		case ch1 <- 1:
		case <-done:
		}
	}()

	go func() {
		ch2 <- "oops" // Will panic - not covered by axioms
	}()

	close(done)
	close(ch2) // Closes ch2 while sender might still be running
	cleanup(ch1)
}

func cleanup(ch chan int) {
	close(ch)
}
