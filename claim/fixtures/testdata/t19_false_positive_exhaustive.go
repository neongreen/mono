package t19

// This is a TRUE claim with a COMPLETE enumeration.
// The proof lists all access points from grep and explains each.
// Claude should NOT demand "exhaustive verification" when grep results are addressed.
//
// If Claude returns UNPROVEN citing "incomplete enumeration", it's a FALSE POSITIVE.

// @claim[t19]: secretChan is only accessed within this file
// @context[t19]:
// grep 'secretChan' in t19_false_positive_exhaustive.go
// @proof[t19]:
// COMPLETE ENUMERATION from context grep of 'secretChan':
// 1. Declaration (var secretChan) - defines the variable
// 2. init() assignment - initializes with make()
// 3. sender() send - sends value
// 4. cleanup() close - closes channel
//
// All accesses are in this file. None of these pass secretChan to external functions.
// The variable is unexported (lowercase 's'), so external packages cannot access it directly.
// Therefore secretChan is only accessed within this file. ∎

var secretChan chan int

func init() {
	secretChan = make(chan int)
}

func sender() {
	secretChan <- 1
}

func cleanup() {
	close(secretChan)
}
