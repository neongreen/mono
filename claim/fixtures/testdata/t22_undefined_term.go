package t22

// This test catches when a claim uses an UNDEFINED TERM and the checker
// just accepts whatever definition the proof provides.
//
// The claim says "Handlers never panic" - but what is a Handler?
// The proof says "LogHandler has recover() in defer" and the checker
// accepts this as proof about ALL handlers.
//
// The correct behavior: REJECT because "handlers" is undefined.
// The checker has NO CONTEXT about what handlers are in this codebase.
// It should say: "I cannot verify this claim - 'handlers' is not defined."
//
// Instead, the checker accepts the proof's implicit definition that
// handlers = LogHandler, which is a false narrowing.
//
// BUG: The checker should demand a definition of "handlers" before
// attempting to verify ANY proof about them.

// @claim[t22]: Handlers never panic
// @proof[t22]:
// Looking at the LogHandler implementation:
//
// func (h *LogHandler) Handle(msg Message) {
//     defer func() {
//         if r := recover(); r != nil {
//             log.Printf("recovered: %v", r)
//         }
//     }()
//     h.process(msg)
// }
//
// The defer/recover pattern catches any panic from process().
// The panic is logged and the function returns normally.
//
// Therefore handlers never panic.

type Message struct {
	Data string
}

type LogHandler struct{}

func (h *LogHandler) Handle(msg Message) {
	defer func() {
		if r := recover(); r != nil {
			// recovered
		}
	}()
	h.process(msg)
}

func (h *LogHandler) process(msg Message) {
	// safe processing
}

// The actual code - proof completely ignores CrashHandler!
type CrashHandler struct{}

func (h *CrashHandler) Handle(msg Message) {
	// No recover! This WILL panic!
	if msg.Data == "" {
		panic("empty message") // <- Handlers CAN panic!
	}
}
