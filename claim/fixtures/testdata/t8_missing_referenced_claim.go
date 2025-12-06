package t8

// @claim[t8]: function abc can't panic
// @proof[t8]:
// foo can't panic, @see[t8-foo].
// abc only calls foo.

func abc() { foo() }

func foo() { panic("boom") }
