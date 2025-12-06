package t2

// @claim[t2]: function abc can't panic
// - abc closes the channel at most once
// - all closes are guarded by control flow

func abc() {
	ch := make(chan int)
	close(ch)
	close(ch) // panic
}
