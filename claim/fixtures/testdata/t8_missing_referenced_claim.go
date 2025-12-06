package t8

// @claim[t8]: function abc can't panic
// - foo can't panic, see @claim[t8-foo]
// - abc only calls foo

func abc() { foo() }

func foo() { panic("boom") }
