package t1

// @claim[t1]: function abc can't panic (no send-on-closed-channel originates here)
// @proof[t1]:
// The channel is closed only after all senders are done.
// Only abc closes the channel.
// abc sends only after checking the channel is open.

func abc() {
	ch := make(chan int)
	gate := make(chan struct{})

	go func() {
		<-gate
		ch <- 1 // will panic
	}()

	close(ch)
	close(gate)
}
