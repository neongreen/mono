# tk — v1 Sync Spec (Event Remotes, Two-Laptop iCloud Flow)

> v0 is live using a 6-char alphanumeric **node id** stored in sqlite, and **task ids** like `tk-1-<node>`, `tk-2-<node>`
> this spec **embraces that format** and makes it safe for multi-machine sync

---

## 1. scope & goals
1.1 offline-first sync between two mac laptops using an iCloud Drive folder as the first remote
1.2 keep sqlite as a **local cache**, sync **immutable event segments**
1.3 **no renaming** of existing `tk-<n>-<node>` task ids
1.4 deterministic, idempotent ingest, no merge prompts
1.5 minimal cli to `remote add | sync | push | pull | export | ingest | status sync`

---

## 2. identifiers & clocks
2.1 **node id**: 6-char `[A-Za-z0-9]`, persisted in sqlite (already in v0)
2.2 **task id**: `tk-<task_seq>-<node>` where `<task_seq>` is a per-node monotonic integer (already in v0)
2.3 **event id**: keep/introduce `ev-<event_seq>-<node>` (per-node monotonic) **or** ULID if you already have it
 2.3.1 if v0 lacked stable event ids, synthesize `ev-<rowid>-<node>` during export and persist a backfill table for dedupe
2.4 **lamport**: per-node lamport clock stored in sqlite
 2.4.1 bump on local write and when ingesting an event with a higher lamport

---

## 3. segment format (immutable files)
3.1 **path layout** (iCloud example):

```
/Library/Mobile Documents/com~apple~CloudDocs/tk-events/
personal/
segments/
2025/10/24/
2025-10-24T15-07-33Z_ab12CD_v1_s000001.jsonl.zst
2025-10-24T15-12-10Z_ab12CD_v1_s000002.jsonl.zst
index.json
```

3.2 **file name**: `YYYY-MM-DDThh-mm-ssZ_<node>_v1_s<segment_seq>.jsonl.zst`
3.3 **one line = one event** (canonicalized JSON):
```json
{
  "schema":"tk.event.v1",
  "id":"ev-42-ab12CD",            // or ULID if you prefer, but must be globally unique
  "lamport":12345,
  "ts":"2025-10-24T15:07:33.512Z",
  "node":"ab12CD",
  "space":"personal",
  "actor":"emily",
  "role":"human",
  "kind":"task.status.set",
  "payload":{"task_id":"tk-7-ab12CD","axis":"generic","state":"in_progress"},
  "ctx":{"repo_uuid":null,"branch":null,"commit":null,"jj_op_id":null}
}
```

3.4 index.json (per space):
```json
{
  "schema":"tk.index.v1",
  "space":"personal",
  "segments":[
    {"rel":"segments/2025/10/24/2025-10-24T15-07-33Z_ab12CD_v1_s000001.jsonl.zst",
     "sha256":"…","size":1234,"mtime":"2025-10-24T15:07:35Z"}
  ]
}
```

3.5 segments are append-only, never edited or deleted; update index.json last

⸻

## 4 event remotes

4.1 folder remote (v1): iCloud Drive path
4.2 config ~/.config/tk/config.toml:

```toml
[remotes.icloud]
type = "folder"
path = "~/Library/Mobile Documents/com~apple~CloudDocs/tk-events"
spaces = ["personal"]
push = true
pull = true
```

4.3 later remote types (git/http/s3) plug into the same interface

⸻

5 export / ingest / sync flow

5.1 export: roll unsent local events into a new segment
 5.1.1 rotate by size (e.g. 2 MB) or age (e.g. 120 s)
 5.1.2 write *.partial → fsync → rename to .jsonl.zst atomically
 5.1.3 update local mirror of index.json at `~/.local/state/tk/remotes/<remote_name>/<space>/index.json`
5.2 pull: read remote index, download any missing segment files, verify sha256
 5.2.1 if index.json missing on remote, reconstruct by scanning segments/** and write it (iCloud resilience)
 5.2.2 always scan directory as fallback to catch segments not yet in index (iCloud eventual consistency)
5.3 ingest: stream each segment; for each event
 5.3.1 dedupe by event.id; ignore duplicates
 5.3.2 bump lamport if needed
 5.3.3 record high-watermark in `~/.local/state/tk/ingest_watermarks/<remote_name>/<space>.json`
5.4 push: upload new local segments first, then upload updated index.json
5.5 sync: pull → ingest → export → push → status

⸻

6 dedupe & ordering

6.1 dedupe key = event.id
6.2 projection order = (lamport, event.id) deterministic across machines
6.3 claim conflicts resolved by existing MV/authority rules — sync doesn’t resolve anything, it just ships events

⸻

7 node id collision policy

7.1 on first sync, compare your node id against index.json producer nodes
7.2 if collision detected (same 6-char node seen from different machine fingerprints)
 7.2.1 block sync and print remediation steps
 7.2.2 user runs tk node regen to mint a new 6-char node id locally
 7.2.3 keep old task ids (tk-*-oldNode) as-is; new tasks will use the new node id
 7.2.4 record node.alias event so UIs can label both node ids as “this device”
7.3 rationale: we keep append-only and avoid rewriting existing ids

⸻

8 cli

8.1 tk remote add icloud folder ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events
8.2 tk remote ls → list remotes and discovered nodes per remote
8.3 tk push --all [--space personal]
8.4 tk pull icloud [--space personal]
8.5 tk ingest [path]
8.6 tk push icloud [--space personal]
8.7 tk sync icloud [--space personal]
8.8 tk status sync → divergence summary
8.9 tk node show → current 6-char node id
8.10 tk node regen → mint a new node id (guarded; warns about id change semantics)

⸻

9 migration from your actual v0

9.1 keep task ids unchanged (tk-<n>-<node>), no renaming
9.2 event ids
 9.2.1 if v0 events already have unique ids: reuse them
 9.2.2 else: on first push --all, synthesize ev-<rowid>-<node> and persist a mapping table event_id_map(rowid,event_id) for stable dedupe
9.3 seed segments
 9.3.1 tk push --all creates initial segments from sqlite
 9.3.2 add the iCloud folder remote, then tk sync icloud on both laptops
9.4 lock out db file sync
 9.4.1 stop syncing the sqlite tk.db itself; only sync segments

⸻

10 tests

10.1 round-trip: 1k events → push --all → ingest into fresh db → projection matches
10.2 dup ingest: same segment twice → event count stable
10.3 concurrent writers: two nodes produce segments, merge via sync → deterministic state
10.4 checksum fail: corrupt file detected and doesn’t poison sqlite
10.5 partial write recovery: leftover *.partial cleaned on startup
10.6 node collision: simulate collision → sync blocked with actionable error
10.7 perf: ingest 10k events < 2s on M-series

⸻

11 ergonomics & mac specifics

11.1 iCloud is eventually consistent; after tk push on laptop A, wait for Finder’s “uploaded” indicator before pulling on laptop B
11.2 prefer a flat tk-events/personal/segments/... tree — avoid long paths to help iCloud
11.3 tk status sync prints something like:

icloud/personal: local +1 seg, remote +0 seg, diverged: no, last_sync: 1m


⸻

12 configuration defaults

12.1 sync.segment_max_bytes = 2_000_000 (2 MB)
12.2 sync.segment_max_age = 120 (seconds)
12.3 sync.compress = zstd level 6
12.4 sync.safe_mode = true (verify sha256 on every ingest; recommended for iCloud)
12.5 spaces = ["personal"] (v1 fixed to single space)

⸻

13 implementation tasks (bead-sized)

13.1 persist node id in sqlite if not present; expose tk node show
13.2 event id backfill: push --all synthesizes ev-<rowid>-<node> when needed
13.3 segment writer: rotate by size/age, zstd, atomic rename
13.4 folder remote adapter: read/write files, generate/consume index.json
13.5 ingest: stream zstd, dedupe by event.id, bump lamport
13.6 sync orchestrator: pull → ingest → push → status
13.7 collision check: scan producer nodes in remote; block on clash
13.8 cli: remote add|ls, ingest|push|pull|sync, status sync, node show|regen
13.9 tests: round-trip, dup, concurrent, checksum, partial, collision, perf
13.10 icloud advisory: detect iCloud path and warn if files "waiting to upload"; suggest "Keep Downloaded"
13.11 docs: two-Mac iCloud quickstart

⸻

14 acceptance criteria

14.1 both laptops create tasks offline and see each other's changes after tk sync icloud
14.2 no data loss; re-running sync is a no-op
14.3 existing tk-<n>-<node> ids remain valid; new tasks continue the per-node counter
14.4 node id collision is detected and blocked with clear remedy
14.5 initial push --all migrates v0 cleanly

⸻

15 security & privacy (v1 minimal)

15.1 no server credentials required; iCloud auth handled by OS
15.2 optional redact_paths (off by default) for context path fields in exported events
15.3 future: per-segment age encryption; not required for two-Mac iCloud

⸻

16 progress output

16.1 tk sync with no args uses default remote from config (icloud)
16.2 example progress output from tk sync icloud:

```
pull: +3 segments, 0 errors
ingest: +420 events (total 7,314)
push: +1 segment, index updated
status: clean
```

⸻

17 future-proofing

17.1 segment schema tk.event.v1 with schema field for versioned upgrades
17.2 remote trait interface so git/http/s3 can implement list/index/get/put
17.3 optional space in path now; can expand to work later without breaking anything

⸻

18 non-goals (v1)

18.1 no background daemon; sync is user-invoked
18.2 no bidirectional file-locking across iCloud; eventual consistency is acceptable
18.3 no server-side merge logic; client ingest handles all merging

⸻

19 quickstart checklist (two Macs)

19.1 on Mac A: tk push --all
19.2 on Mac A: tk remote add icloud folder ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events
19.3 on Mac A: tk sync icloud
19.4 wait for iCloud to finish uploading (Finder status dots)
19.5 on Mac B: same remote add, then tk sync icloud
19.6 verify: tk status sync shows clean on both; tk ls matches across machines

⸻
