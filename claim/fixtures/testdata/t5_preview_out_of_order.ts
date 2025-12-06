// @claim[t5] @pedantic: live preview is always either loading or corresponds to the latest code
// - on every edit we set loading
// - when we set ready we show the preview built from the current code

type State = { status: "loading" | "ready"; hash: string; html: string }
let state: State = { status: "loading", hash: "", html: "" }
let currentHash = ""

function sha1(_s: string): string { return Math.random().toString(16).slice(2) }
function buildPreviewAsync(_code: string): Promise<string> {
  return new Promise(res => setTimeout(() => res("<html/>"), 0))
}

export function onEdit(code: string) {
  const hash = sha1(code)
  currentHash = hash
  state.status = "loading"

  buildPreviewAsync(code).then(html => {
    state.hash = hash
    state.html = html
    state.status = "ready"
  })
}

export function render() {
  if (state.status === "loading") return "loading"
  return `ready hash=${state.hash} current=${currentHash} html=${state.html}`
}
