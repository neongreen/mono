// @claim[t9] @pedantic: upload eventually completes or fails
// @proof[t9]:
// Every upload either succeeds or errors and updates the status.

type State = { status: "idle" | "uploading" | "done" | "failed"; progress: number }
let state: State = { status: "idle", progress: 0 }

function uploadFileAsync(_file: Blob): Promise<void> {
  // BUG: promise never resolves or rejects!
  return new Promise((_resolve, _reject) => { })
}

export function upload(file: Blob) {
  state.status = "uploading"
  state.progress = 0
  uploadFileAsync(file).then(() => {
    state.status = "done"
  }).catch(_e => {
    state.status = "failed"
  })
}
