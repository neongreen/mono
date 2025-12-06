package t20

// Tests the grep enumeration pattern for channel safety proofs.
// This is a TRUE claim that should ALWAYS return PROVEN.
//
// Pattern: grep finds N access sites, proof explains each one is safe via axioms.

// @claim[localChan-unexported]: localChan is unexported (lowercase)
// @proof[localChan-unexported]:
// The field is declared as 'localChan' with lowercase 'l'.

// @claim[helperA-safe]: helperA cannot cause send-on-closed-channel panic
// @proof[helperA-safe]:
// helperA sends but never closes. Sends before close are safe.

// @claim[helperB-safe]: helperB's close is safe
// @proof[helperB-safe]:
// helperB closes only after waiting for done signal, which comes after helperA exits.

// @claim[t20]: No send-on-closed-channel panic can occur in process() or its spawned goroutines
// @context[t20]:
// grep 'localChan' in t20_grep_enumeration_pattern.go
// @proof[t20]:
// COMPLETE ENUMERATION from context grep of 'localChan':
// 1. Struct field declaration - safe (not a send)
// 2. process() assignment - safe (not a send)
// 3. helperA() send - safe by @see[helperA-safe]
// 4. helperB() close - safe by @see[helperB-safe]
// 5. process() receive - safe (receives don't panic on closed channels)
//
// All accesses enumerated. By @see[localChan-unexported], external code cannot access. ∎

type Processor struct {
	localChan chan int
	done      chan struct{}
}

func process() {
	p := &Processor{
		localChan: make(chan int),
		done:      make(chan struct{}),
	}
	go helperA(p)
	go helperB(p)
	<-p.localChan
}

func helperA(p *Processor) {
	p.localChan <- 1
	close(p.done)
}

func helperB(p *Processor) {
	<-p.done
	close(p.localChan)
}
