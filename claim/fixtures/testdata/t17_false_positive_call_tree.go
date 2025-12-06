package t17

// This is a TRUE claim with a VALID proof.
// The proof correctly uses grep + axioms to establish safety.
// Claude should return PROVEN, not demand additional verification.
//
// If Claude returns UNPROVEN, it's a FALSE POSITIVE.

// @claim[ch-not-closed]: main() never closes ch
// @proof[ch-not-closed]:
// main() only creates ch and receives from it. No close() call exists.

// @claim[t17]: main() cannot panic from send-on-closed-channel
// @context[t17]:
// grep 'close' in this file
// @proof[t17]:
// From context grep, there are no close() calls in this file.
// By @see[ch-not-closed], ch is never closed.
// Since ch is never closed, no send-on-closed-channel panic can occur. ∎

func main() {
	ch := make(chan int)
	go helper(ch)
	<-ch
}

func helper(ch chan int) {
	ch <- 1
}
