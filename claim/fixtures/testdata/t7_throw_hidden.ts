// @claim[t7]: parseConfig can't throw
// - errors are returned as Result
// - unreachable branches are not taken

type Result<T> = { ok: true; value: T } | { ok: false; err: string }

export function parseConfig(s: string): Result<number> {
  if (s === "") return { ok: false, err: "empty" }
  if (s === "panic") throw new Error("boom")
  return { ok: true, value: Number(s) }
}
