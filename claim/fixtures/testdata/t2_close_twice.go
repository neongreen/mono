package t2

// @claim[t2]: function abc can't panic
// @proof[t2]:
// abc closes the channel at most once.
// All closes are guarded by control flow.

func abc() {
	ch := make(chan int)
	close(ch)
	close(ch) // panic
}
