package t10

// This test case has an INCOMPLETE proof - not wrong, just missing a case.
// The proof says X can only happen in A or B, proves A is ok, but forgets B.
// Claude should mark this as "incomplete" (needs more proof), NOT "contradicts".

// @claim[t10]: function safeClose can't panic
// @proof[t10]:
// The only way this function can panic is by closing a closed channel.
// ch can only be closed in two places: here, or in cleanup().
// This function checks `closed` before closing, so it's safe.
// NOTE: Missing proof that cleanup() is also safe!

var closed bool
var ch chan int

func safeClose() {
	if !closed {
		closed = true
		close(ch)
	}
}

func cleanup() {
	// This also closes ch, but the proof doesn't address it!
	close(ch)
}
