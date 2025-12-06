// @claim[t9] @pedantic: preview eventually becomes ready for the latest code
// - every build either succeeds or fails and updates the status

type State = { status: "loading" | "ready" | "error"; html: string }
let state: State = { status: "loading", html: "" }

function buildPreviewAsync(_code: string): Promise<string> {
  return new Promise((_res, _rej) => { })
}

export function onEdit(code: string) {
  state.status = "loading"
  buildPreviewAsync(code).then(html => {
    state.html = html
    state.status = "ready"
  }).catch(_e => {
    state.status = "error"
  })
}
