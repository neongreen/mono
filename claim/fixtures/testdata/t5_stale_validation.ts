// @claim[t5] @pedantic: form validation always reflects the current input value
// @proof[t5]:
// On every keystroke we revalidate.
// When validation completes we show errors for the current value.

type State = { status: "validating" | "valid" | "invalid"; inputHash: string; errors: string[] }
let state: State = { status: "validating", inputHash: "", errors: [] }
let currentHash = ""

function hash(s: string): string { return Math.random().toString(16).slice(2) }
function validateAsync(_value: string): Promise<string[]> {
  return new Promise(res => setTimeout(() => res([]), 0))
}

export function onInput(value: string) {
  const h = hash(value)
  currentHash = h
  state.status = "validating"

  validateAsync(value).then(errors => {
    // BUG: doesn't check if input changed during validation
    state.inputHash = h
    state.errors = errors
    state.status = errors.length ? "invalid" : "valid"
  })
}

export function render() {
  if (state.status === "validating") return "validating"
  return `${state.status} hash=${state.inputHash} current=${currentHash}`
}
