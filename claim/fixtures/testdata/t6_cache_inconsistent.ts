// @claim[t6]: cache is either stale or consistent with source
// @proof[t6]:
// We always mark stale on source change.
// Fresh implies consistent.

type State = { status: "stale" | "fresh"; data: string }
let state: State = { status: "fresh", data: "old" }
let source = "old"

function fetchFromSource(key: string): Promise<string> {
  return new Promise(res => setTimeout(() => res("data:" + key), 0))
}

export function onSourceChange(key: string) {
  // BUG: doesn't set status to "stale" first!
  source = key
  fetchFromSource(key).then(data => {
    state.data = data
    state.status = "fresh"
  })
}

export function read() {
  if (state.status === "stale") return "[stale]"
  return "fresh: " + state.data
}
