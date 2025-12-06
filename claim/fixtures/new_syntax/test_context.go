package test

// @claim[ctx-test]: processData returns nil when input is empty
// @context[ctx-test]:
// function processData in this file
// @proof[ctx-test]:
// The function checks if len(data) == 0 and returns nil immediately.
// This is the first check in the function body.

func processData(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	// ... processing logic ...
	return nil
}
