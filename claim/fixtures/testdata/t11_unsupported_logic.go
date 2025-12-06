package t11

// This test case has UNSUPPORTED logic - the statements are all true,
// but they don't actually prove the parent claim.
// Claude should mark this as "unsupported" (logic doesn't follow), NOT "contradicts".

// @claim[t11]: function process never returns an error
// @proof[t11]:
// The function has proper error handling.
// All error paths are covered.
// The code follows best practices.
// NOTE: These statements are all true but don't prove "never returns an error"!

import "errors"

func process(input string) error {
	if input == "" {
		return errors.New("empty input") // Returns an error!
	}
	if len(input) > 100 {
		return errors.New("input too long") // Returns an error!
	}
	return nil
}
