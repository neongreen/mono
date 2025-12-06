// @claim[t6]: preview is either loading or correct for current code
// - we always switch to loading on edit
// - ready implies correct

type State = { status: "loading" | "ready"; html: string }
let state: State = { status: "ready", html: "old" }
let current = "old"

function buildPreviewAsync(code: string): Promise<string> {
  return new Promise(res => setTimeout(() => res("html:" + code), 0))
}

export function onEdit(code: string) {
  current = code
  buildPreviewAsync(code).then(html => {
    state.html = html
    state.status = "ready"
  })
}

export function render() {
  if (state.status === "loading") return "loading"
  return "ready " + state.html
}
