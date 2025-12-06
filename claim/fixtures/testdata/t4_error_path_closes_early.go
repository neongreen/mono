package t4

import "errors"

// @claim[t4]: abc can't panic due to send-on-closed-channel
// @proof[t4]:
// On error we stop all senders before closing.
// Close happens only after senders exit.

func abc(fail bool) error {
	ch := make(chan int)
	gate := make(chan struct{})

	go func() {
		<-gate
		ch <- 1 // will panic if fail==true
	}()

	if fail {
		close(ch)
		close(gate)
		return errors.New("boom")
	}

	close(gate)
	_ = <-ch
	close(ch)
	return nil
}
