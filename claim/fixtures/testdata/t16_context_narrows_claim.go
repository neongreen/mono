package t16

// This test checks that claims which would be interpreted differently
// with vs without context are flagged as ambiguous.
//
// The claim says "the function never panics" - but WHICH function?
// The context shows a specific function, but the claim doesn't name it.
// Without context, the claim is meaningless.
// WITH context, the claim is narrowed to that specific function.
// This should be caught as "Ambiguous claim".

// @claim[t16]: The function never panics
// @context[t16]:
// function processItems in this file
// @proof[t16]:
// Looking at processItems, it has no panic calls and handles all errors.

func processItems(items []string) error {
	for _, item := range items {
		if item == "" {
			return nil // safely returns on empty
		}
	}
	return nil
}

func otherFunction() {
	panic("this one does panic!")
}
