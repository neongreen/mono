# Systematic Audit of All tk Commits (224 total)

Format: HASH | TYPE | VERDICT | WHY

Types: BUG-CRITICAL, BUG-HIGH, BUG-MEDIUM, BUG-LOW, FEATURE, REFACTOR, TEST, DOC, FORMAT
Verdict: For bugs - would tests have caught it? YES/NO/PARTIAL

---

## 570ecb72ac0b
**Description:** tk: Document legacy TaskID fields with tk-190 reference (PR #267 review)
**Files:**
  - tk/internal/reducer/reducer.go
  - tk/internal/types/payloads.go
**TYPE:** DOC
**VERDICT:** N/A
**WHY:** Added comments explaining legacy fields, not a bug fix

## df9ff245910f
**Description:** tk: Add error handling to test queries (PR review feedback)
**Files:**
  - tk/internal/tasks/create_test.go
**TYPE:** TEST
**VERDICT:** N/A
**WHY:** Test improvement, not a bug fix

## bcb27ccd8bcf
**Description:** tk: Remove temporary boundary analysis document - decision documented in READMEs
**Files:**
  - tk/internal/DATABASE_VS_TASKS.md
**TYPE:** DOC
**VERDICT:** N/A
**WHY:** Cleanup of temporary file

## cf6b3ba5c900
**Description:** tk: Add relation CRDT tests for OR-set semantics (tk-173)
**Files:**
  - tk/internal/reducer/relations_test.go
**TYPE:** TEST
**VERDICT:** N/A
**WHY:** Added new test coverage for CRDT relations

## d4d836e0120c
**Description:** tk: Add unit tests for internal/tasks operations (tk-185)
**Files:**
  - tk/internal/tasks/create_test.go
  - tk/internal/tasks/delete_test.go
  - tk/internal/tasks/note_test.go
**TYPE:** TEST
**VERDICT:** N/A
**WHY:** Added new test coverage

## e2ecf49b8df1
**Description:** tk: Extract Delete operation to internal/tasks (tk-184)
**Files:**
  - tk/cmd/rm.go
  - tk/internal/tasks/delete.go
**TYPE:** REFACTOR
**VERDICT:** N/A
**WHY:** Code organization, no behavior change

## 5e786374e21a
**Description:** tk: Extract AddNote operation to internal/tasks (tk-183)
**Files:**
  - tk/cmd/note.go
  - tk/internal/tasks/note.go
**TYPE:** REFACTOR
**VERDICT:** N/A
**WHY:** Code organization, no behavior change

## 86c37c9033a0
**Description:** tk: Extract Create operation to internal/tasks and document database/ vs tasks/ boundary (tk-182)
**Files:**
  - tk/cmd/new.go
  - tk/internal/DATABASE_VS_TASKS.md
  - tk/internal/database/README.md
**TYPE:** REFACTOR
**VERDICT:** N/A
**WHY:** Code organization + documentation, no behavior change

## d6ea68257713
**Description:** tk: Simplify internal/tasks README to <50 lines, focus on boundaries and principles
**Files:**
  - tk/internal/tasks/README.md
**TYPE:** DOC
**VERDICT:** N/A
**WHY:** Documentation update

## 736391efb42c
**Description:** tk: Add README to internal/tasks explaining package purpose and organization
**Files:**
  - tk/internal/tasks/README.md
**TYPE:** DOC
**VERDICT:** N/A
**WHY:** Added package documentation

## 44435adbb625
**Description:** tk: Add comprehensive unit tests for ResolveTaskReference edge cases
**Files:**
  - tk/internal/database/task_resolver_test.go
**TYPE:** TEST
**VERDICT:** N/A
**WHY:** Added 8 edge case tests for resolution

## a824d5e638e4
**Description:** tk: Add 'tk relate ls' command to list relations in copy-pasteable format
**Files:**
  - tk/cmd/relate/ls.go
  - tk/cmd/relate.go
**TYPE:** FEATURE
**VERDICT:** N/A
**WHY:** New command, not a bug fix

## e61d0310a764
**Description:** Auto-format code
**Files:**
  - tk/cmd/debug/rebuild.go
  - tk/internal/database/rebuild.go
  - tk/internal/remote/push_test.go
  - tk/internal/segment/filename.go
  - tk/internal/segment/filename_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3d9e857da721
**Description:** tk: Filed tasks for testability, shortcuts, and data improvements
**Files:**
  - tk/cmd/debug/rebuild.go
  - tk/internal/database/rebuild.go
  - tk/internal/remote/push_test.go
  - tk/internal/segment/filename.go
  - tk/internal/segment/filename_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 89ef999da8fb
**Description:** Auto-format code
**Files:**
  - tk/cmd/display.go
  - tk/cmd/new.go
  - tk/cmd/project/rm_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3c6161eb7126
**Description:** tk: Final QoL batch - note improvements (tk-31-16uq1v, tk-56, tk-58, tk-153)
**Files:**
  - tk/cmd/note.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 46849bfeaccb
**Description:** tk & tk-vscode: More QoL improvements (tk-31-16uq1v, tk-56, tk-58, tk-vsc-51)
**Files:**
  - tk/cmd/display.go
  - tk/cmd/mv.go
  - tk/cmd/new.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 637e3f9b28bd
**Description:** tk & tk-vscode: More quality of life improvements (tk-62, tk-67, tk-70, tk-94, tk-vsc-33)
**Files:**
  - tk/cmd/display.go
  - tk/internal/termutil/termutil.go
  - tk-vscode/src/extension.ts
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f5f65ffce12a
**Description:** tk & tk-vscode: Quality of life improvements (tk-61, tk-63, tk-75, tk-131, tk-vsc-13, tk-vsc-29)
**Files:**
  - tk/cmd/edit.go
  - tk/cmd/new.go
  - tk/cmd/show.go
  - tk-vscode/src/extension.ts
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 22fe51c23e7f
**Description:** tk: Add --force protection for project deletion (tk-146)
**Files:**
  - tk/cmd/project/rm.go
  - tk/cmd/project/rm_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 751c7dfe4991
**Description:** tk: Add -m flag to tk mark and fix UID deduplication (tk-138, tk-139)
**Files:**
  - tk/cmd/mark.go
  - tk/internal/database/task_resolver.go
  - tk/internal/database/tasks.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 16c420352d56
**Description:** tk: Fix resolution ambiguity bugs (tk-134, tk-135, tk-136, tk-137)
**Files:**
  - tk/cmd/project/alias_add.go
  - tk/cmd/project/create.go
  - tk/internal/database/helpers_test.go
  - tk/internal/database/resolution_ambiguity_test.go
  - tk/internal/database/task_resolver.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8e94ec36a57b
**Description:** tk-128: Add project_uuid to Task JSON output
**Files:**
  - tk/cmd/display.go
  - tk/internal/types/task.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 54bd4b305802
**Description:** tk-127: Add local_preferred_alias to project ls output
**Files:**
  - tk/cmd/project/ls.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c927032be7f1
**Description:** remove doc
**Files:**
  - tk/SYNC_STATE_CLEANUP.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b33edefd4260
**Description:** tk: store export state in database, move state dir to ~/.tk
**Files:**
  - tk/SYNC_STATE_CLEANUP.md
  - tk/internal/config/config.go
  - tk/internal/config/config_test.go
  - tk/internal/database/db_schema.go
  - tk/internal/database/export_state.go
  - tk/internal/remote/export.go
  - tk/internal/remote/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 22aa21b6664e
**Description:** fix compilation
**Files:**
  - tk/cmd/column_width_test.go
  - tk/cmd/relations_integration_test.go
  - tk/cmd/sort_test.go
  - tk/internal/database/projections_determinism_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c33fa8972391
**Description:** Revert "tk: rename TaskUUID to TaskID in Go structs, use 'id' in JSON"
**Files:**
  - tk/cmd/blockers.go
  - tk/cmd/column_width.go
  - tk/cmd/display.go
  - tk/cmd/graph.go
  - tk/cmd/ls.go
  - tk/cmd/ls_test.go
  - tk/cmd/relations_integration_test.go
  - tk/internal/query/filter.go
  - tk/internal/reducer/metadata.go
  - tk/internal/reducer/reducer.go
  - tk/internal/relations/graph.go
  - tk/internal/types/blocker.go
  - tk/internal/types/relations.go
  - tk/internal/types/task.go
  - tk/internal/utils/blocking.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0b4aa7795339
**Description:** tk: rename TaskUUID to TaskID in Go structs, use 'id' in JSON
**Files:**
  - tk/cmd/blockers.go
  - tk/cmd/column_width.go
  - tk/cmd/display.go
  - tk/cmd/graph.go
  - tk/cmd/ls.go
  - tk/cmd/ls_test.go
  - tk/cmd/relations_integration_test.go
  - tk/internal/query/filter.go
  - tk/internal/reducer/metadata.go
  - tk/internal/reducer/reducer.go
  - tk/internal/relations/graph.go
  - tk/internal/types/blocker.go
  - tk/internal/types/relations.go
  - tk/internal/types/task.go
  - tk/internal/utils/blocking.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 564f1ab9264b
**Description:** tk: improve JSON output - sort by UUID, rename fields, add UUID to Blocker
**Files:**
  - tk/cmd/blockers.go
  - tk/cmd/column_width.go
  - tk/cmd/conflicts.go
  - tk/cmd/display.go
  - tk/cmd/graph.go
  - tk/cmd/ls.go
  - tk/cmd/relations_integration_test.go
  - tk/cmd/show.go
  - tk/cmd/sort_test.go
  - tk/internal/reducer/reducer.go
  - tk/internal/types/blocker.go
  - tk/internal/types/task.go
  - tk/internal/types/task_sort.go
  - tk/internal/utils/blocking.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 279fb4666eca
**Description:** tk: make 'tk ls --json' output deterministic
**Files:**
  - tk/cmd/display.go
  - tk/internal/types/relations.go
  - tk/internal/types/task.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8ac6570ec2c7
**Description:** tk ls --json: deterministic order
**Files:**
  - tk/cmd/display.go
  - tk/internal/types/relations.go
  - tk/internal/types/task.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 201c4efe3aec
**Description:** Auto-format code
**Files:**
  - conf/cmd/root.go
  - lib/cli/cli.go
  - lib/ghrelease/ghrelease.go
  - tk/cmd/describe.go
  - tk/cmd/edit.go
  - tk/cmd/mark.go
  - tk/cmd/mv.go
  - tk/cmd/new.go
  - tk/cmd/note.go
  - tk/cmd/relate.go
  - tk/cmd/rm.go
  - tk/cmd/test_helpers.go
  - tk/internal/remote/export.go
  - tk/internal/remote/index.go
  - tk/internal/remote/ingest.go
  - tk/internal/types/ids.go
  - want/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 75ac13df140f
**Description:** Use cmd.RootCmd.Execute() directly in tk/main.go
**Files:**
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## e2900502cd85
**Description:** Auto-format code
**Files:**
  - tk/cmd/debug/rebuild.go
  - tk/internal/database/projections_determinism_test.go
  - tk/internal/database/rebuild.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8f340fb56ba0
**Description:** Merge branch 'main' into copilot/add-dagger-call-lint-ci-step
**Files:**
  - tk/cmd/root.go
  - want/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 32430546f183
**Description:** Fix lint issues: add nolint support and remove useless wrappers
**Files:**
  - conf/cmd/completion.go
  - conf/cmd/root.go
  - lib/cli/cli.go
  - lib/ghrelease/ghrelease.go
  - lib/linters/uselesswrapper/uselesswrapper.go
  - tk/cmd/describe.go
  - tk/cmd/edit.go
  - tk/cmd/ls_test.go
  - tk/cmd/mark.go
  - tk/cmd/mv.go
  - tk/cmd/new.go
  - tk/cmd/note.go
  - tk/cmd/project/alias_add.go
  - tk/cmd/project/alias_remove.go
  - tk/cmd/project/create.go
  - tk/cmd/project/helpers.go
  - tk/cmd/project/rm.go
  - tk/cmd/relate/add.go
  - tk/cmd/relate/helpers.go
  - tk/cmd/relate/remove.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6d7e1e61253b
**Description:** tk: Fix non-deterministic projection bug causing task ID mismatches after sync
**Files:**
  - tk/cmd/debug/rebuild.go
  - tk/cmd/debug.go
  - tk/internal/database/projections.go
  - tk/internal/database/projections_determinism_test.go
  - tk/internal/database/rebuild.go
  - tk/internal/remote/ingest.go
  - tk/internal/tasks/move.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 4a72908f2d53
**Description:** tk push: reimplement
**Files:**
  - tk/IMPLEMENTATION.md
  - tk/README.md
  - tk/cmd/debug/repair.go
  - tk/cmd/export.go
  - tk/cmd/root.go
  - tk/cmd/sync.go
  - tk/internal/remote/export.go
  - tk/internal/remote/push_test.go
  - tk/internal/remote/sync.go
  - tk/internal/remote/utils.go
  - tk/internal/segment/filename.go
  - tk/internal/segment/filename_test.go
  - tk/internal/segment/writer.go
  - tk/specs/spec-v1.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f561d787dcca
**Description:** tk: add --debug
**Files:**
  - AGENTS.md
  - tk/cmd/root.go
  - tk/internal/remote/export.go
  - tk/internal/remote/sync.go
  - tk/internal/segment/writer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 9f5da77b01f7
**Description:** Auto-format code
**Files:**
  - tk/cmd/ingest.go
  - tk/internal/database/db.go
  - tk/internal/database/db_reducer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c106222150b6
**Description:** Fix duplicate flag registration and rewrite edit/describe tests
**Files:**
  - tk/cmd/describe_test.go
  - tk/cmd/edit_cmd_status_test.go
  - tk/cmd/edit_test.go
  - tk/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ff0a142bf0cf
**Description:** Fix test build failures
**Files:**
  - tk/cmd/describe_test.go
  - tk/cmd/edit_cmd_status_test.go
  - tk/cmd/edit_test.go
  - tk/cmd/import_beads_test.go
  - tk/cmd/mv_test.go
  - tk/cmd/relations_integration_test.go
  - tk/internal/config/config_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 588c9dc2e3a8
**Description:** Fix code duplication and silent error handling
**Files:**
  - tk/ai-temp-audit-summary.md
  - tk/cmd/import_beads.go
  - tk/cmd/project/helpers.go
  - tk/cmd/relate/helpers.go
  - tk/cmd/utils.go
  - tk/internal/import/beads/importer.go
  - tk/internal/import/beads/types.go
  - tk/internal/utils/user.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5c852078924c
**Description:** Complete audit of silent error handling in tk codebase
**Files:**
  - tk/ai-temp-audit-summary.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 9e06e077db29
**Description:** Add error tracking and reporting for ingest operations
**Files:**
  - tk/cmd/ingest.go
  - tk/internal/remote/ingest.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c45affbe037c
**Description:** Refactor task operations and query logic into internal/ packages
**Files:**
  - tk/cmd/debug_helpers.go
  - tk/cmd/describe.go
  - tk/cmd/edit.go
  - tk/cmd/ingest.go
  - tk/cmd/ls.go
  - tk/cmd/mark.go
  - tk/cmd/mv.go
  - tk/internal/collision/collision.go
  - tk/internal/query/filter.go
  - tk/internal/remote/config.go
  - tk/internal/remote/ingest.go
  - tk/internal/remote/sync.go
  - tk/internal/tasks/edit.go
  - tk/internal/tasks/helpers.go
  - tk/internal/tasks/mark.go
  - tk/internal/tasks/move.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f90bf9def7b4
**Description:** Refactor beads import logic into modular internal/import/beads package
**Files:**
  - tk/cmd/import_beads.go
  - tk/internal/import/beads/importer.go
  - tk/internal/import/beads/project.go
  - tk/internal/import/beads/reader.go
  - tk/internal/import/beads/relation.go
  - tk/internal/import/beads/task.go
  - tk/internal/import/beads/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 4377a418d8e6
**Description:** Move export and ingest business logic to internal/remote
**Files:**
  - tk/cmd/debug_helpers.go
  - tk/cmd/export.go
  - tk/cmd/ingest.go
  - tk/cmd/relations_integration_test.go
  - tk/internal/collision/collision.go
  - tk/internal/config/config.go
  - tk/internal/config/config_test.go
  - tk/internal/config/types.go
  - tk/internal/database/db.go
  - tk/internal/database/db_reducer.go
  - tk/internal/database/projections.go
  - tk/internal/reducer/reducer.go
  - tk/internal/remote/config.go
  - tk/internal/remote/export.go
  - tk/internal/remote/ingest.go
  - tk/internal/remote/sync.go
  - tk/internal/remote/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a4d07e13b71c
**Description:** Fix remaining compilation errors after remote package refactoring
**Files:**
  - tk/cmd/status/sync.go
  - tk/cmd/sync_test.go
  - tk/internal/reducer/reducer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 32f69a279f3a
**Description:** Update remaining files to use internal/remote instead of internal/sync
**Files:**
  - tk/cmd/debug_helpers.go
  - tk/cmd/relations_integration_test.go
  - tk/cmd/status/sync.go
  - tk/internal/collision/collision_test.go
  - tk/internal/config/config_test.go
  - tk/internal/database/db.go
  - tk/internal/database/db_reducer.go
  - tk/internal/segment/writer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 63d94e7c0afd
**Description:** Refactor remote handling logic into internal/remote package
**Files:**
  - tk/cmd/export.go
  - tk/cmd/ingest.go
  - tk/cmd/remote/add.go
  - tk/cmd/remote/rm.go
  - tk/cmd/sync.go
  - tk/internal/collision/collision.go
  - tk/internal/config/config.go
  - tk/internal/remote/config.go
  - tk/internal/remote/convert.go
  - tk/internal/remote/export.go
  - tk/internal/remote/index.go
  - tk/internal/remote/ingest.go
  - tk/internal/remote/sync.go
  - tk/internal/remote/types.go
  - tk/internal/remote/utils.go
  - tk/internal/segment/reader.go
  - tk/internal/segment/types.go
  - tk/internal/segment/writer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 02eb82a884f4
**Description:** Complete tk command reorganization and fix compilation errors
**Files:**
  - AGENTS.md
  - tk/cmd/project/alias_add.go
  - tk/cmd/project/alias_remove.go
  - tk/cmd/project/create.go
  - tk/cmd/project/helpers.go
  - tk/cmd/project/rm.go
  - tk/cmd/relate/add.go
  - tk/cmd/relate/helpers.go
  - tk/cmd/relate/remove.go
  - tk/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 955adf2630c0
**Description:** Fix node and events commands - they are debug subcommands, not root commands
**Files:**
  - tk/cmd/debug.go
  - tk/cmd/events.go
  - tk/cmd/node.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 04ec8e22c943
**Description:** Reorganize tk commands into one file per command with subcommands in subdirectories
**Files:**
  - tk/cmd/conflicts.go
  - tk/cmd/events/list.go
  - tk/cmd/events/show.go
  - tk/cmd/events/stats.go
  - tk/cmd/events.go
  - tk/cmd/meta/claims.go
  - tk/cmd/meta/get.go
  - tk/cmd/meta/list.go
  - tk/cmd/meta/set.go
  - tk/cmd/meta.go
  - tk/cmd/node/regen.go
  - tk/cmd/node/show.go
  - tk/cmd/node.go
  - tk/cmd/project/alias.go
  - tk/cmd/project/alias_add.go
  - tk/cmd/project/alias_remove.go
  - tk/cmd/project/create.go
  - tk/cmd/project/ls.go
  - tk/cmd/project/rm.go
  - tk/cmd/project.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ad8e5b376fb5
**Description:** dagger: restructure
**Files:**
  - .dagger/README.md
  - .dagger/dagger.json
  - .dagger/go.mod
  - .dagger/internal/parallel/parallel.go
  - .dagger/main.go
  - .dagger/project_claudetrace.go
  - .dagger/project_conf.go
  - .dagger/project_dissect.go
  - .dagger/project_helpers.go
  - .dagger/project_ingest.go
  - .dagger/project_jjrun.go
  - .dagger/project_lib_ghclient.go
  - .dagger/project_lib_ghrelease.go
  - .dagger/project_lib_toml.go
  - .dagger/project_markdownformat.go
  - .dagger/project_namespace.go
  - .dagger/project_printpdf.go
  - .dagger/project_prrun.go
  - .dagger/project_tk.go
  - .dagger/project_want.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ff114dc449bb
**Description:** Update tk/README.md with renamed task syntax
**Files:**
  - tk/README.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 67636d110fd6
**Description:** Centralize all mise tasks to top-level mise.toml
**Files:**
  - .cursor/commands/mise.md
  - AGENTS.md
  - claude-trace/mise.toml
  - conf/mise.toml
  - dissect/mise.toml
  - ingest/mise.toml
  - jj-run/mise.toml
  - jj-run-py/mise.toml
  - lib/ghclient/mise.toml
  - lib/ghrelease/mise.toml
  - lib/toml/mise.toml
  - markdown-format/mise.toml
  - mdbook-comments/mise.toml
  - mise.toml
  - printpdf/TESTING.md
  - printpdf/mise.toml
  - prrun/mise.toml
  - tk/README.md
  - tk/mise.toml
  - tk-vscode/README.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a9f3153a5693
**Description:** Replace go fmt with golangci-lint fmt across all projects
**Files:**
  - AGENTS.md
  - GO_STYLE_GUIDE.md
  - claude-trace/AGENTS.md
  - claude-trace/mise.toml
  - conf/mise.toml
  - dissect/AGENTS.md
  - dissect/TESTING.md
  - dissect/mise.toml
  - ingest/mise.toml
  - jj-run/mise.toml
  - lib/ghclient/mise.toml
  - lib/ghrelease/AGENTS.md
  - lib/ghrelease/mise.toml
  - lib/toml/mise.toml
  - markdown-format/AGENTS.md
  - markdown-format/mise.toml
  - mise.toml
  - printpdf/AGENTS.md
  - printpdf/SUMMARY.md
  - printpdf/mise.toml
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6e70736d7d3c
**Description:** more fixes
**Files:**
  - .golangci.yml
  - conf/pkg/schemas/claude_parser.go
  - conf/pkg/schemas/jj_parser.go
  - conf/pkg/schemas/mise_json_parser.go
  - dissect/pkg/testutils/compare_directories.go
  - ingest/pkg/command/command.go
  - ingest/pkg/database/git.go
  - jj-run/cmd/main.go
  - jj-run/tests/integration_test.go
  - prrun/main.go
  - prrun/main_test.go
  - tk/internal/database/db_counters.go
  - tk/internal/database/db_node.go
  - tk/internal/database/db_schema.go
  - tk/internal/database/projections.go
  - tk/internal/database/task_resolver.go
  - want/mono.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 1d337f670b9b
**Description:** Remove useless tests that verify basic language features
**Files:**
  - .golangci.yml
  - AGENTS.md
  - BULLSHIT_TESTS.md
  - conf/cmd/main_test.go
  - conf/pkg/schemas/claude_parser_test.go
  - conf/pkg/schemas/jj_parser_test.go
  - conf/pkg/schemas/loader_test.go
  - conf/pkg/schemas/mise_json_parser_test.go
  - conf/pkg/tools/shims/shims_test.go
  - ingest/pkg/github/github_test.go
  - tk/internal/segment/writer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c9f2a69132ae
**Description:** bunch more lints
**Files:**
  - .dagger/main.go
  - .dagger/project_helpers.go
  - .dagger/projects.go
  - .golangci.yml
  - claude-trace/pkg/render/markdown_conversation.go
  - claude-trace/pkg/tui/model.go
  - conf/cmd/main.go
  - conf/pkg/config/config.go
  - conf/pkg/editors/toml.go
  - conf/pkg/schemas/integration_test.go
  - conf/pkg/schemas/mise.go
  - dagger.json
  - dissect/cmd/move.go
  - dissect/pkg/references/find_test.go
  - go.mod
  - ingest/pkg/integration/github_live_test.go
  - ingest/pkg/mcp/client.go
  - lib/ghclient/client.go
  - lib/ghrelease/ghrelease.go
  - lib/toml/apply.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 14d1d155a4a5
**Description:** omitzero
**Files:**
  - tk/internal/types/relations.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 64c5c948568a
**Description:** [AUTO] Format: golangci-lint fmt
**Files:**
  - beads-merge/main_test.go
  - claude-trace/cmd/extract.go
  - claude-trace/cmd/list.go
  - claude-trace/cmd/main.go
  - claude-trace/cmd/tui_cmd.go
  - claude-trace/cmd/view_cmd.go
  - claude-trace/pkg/parser/jsonl_parsing.go
  - claude-trace/pkg/render/markdown_conversation.go
  - claude-trace/pkg/render/render.go
  - claude-trace/pkg/storage/trace.go
  - claude-trace/pkg/tui/helpers.go
  - claude-trace/pkg/tui/model.go
  - claude-trace/pkg/tui/update_handlers.go
  - claude-trace/pkg/tui/view_renderers.go
  - claude-trace/pkg/viewer/server.go
  - conf/cmd/helpers.go
  - conf/cmd/integration_test.go
  - conf/cmd/main.go
  - conf/pkg/config/config.go
  - conf/pkg/config/config_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5ea5478ff8b6
**Description:** remove unused
**Files:**
  - .golangci.yml
  - conf/cmd/main.go
  - conf/pkg/config/config.go
  - conf/pkg/tools/mise/mise.go
  - ingest/pkg/githubmcp/ingest_test.go
  - ingest/pkg/mcp/client.go
  - jj-run/cmd/main.go
  - printpdf/pkg/converter/converter.go
  - printpdf/pkg/golden/golden_test.go
  - tk/cmd/column_width.go
  - tk/cmd/import_beads.go
  - want/handlers.go
  - want/tools.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 573239969f48
**Description:** [AUTO] lint modernize
**Files:**
  - claude-trace/cmd/list.go
  - claude-trace/pkg/parser/jsonl_parsing.go
  - claude-trace/pkg/parser/parser.go
  - claude-trace/pkg/parser/parser_test.go
  - claude-trace/pkg/render/render_test.go
  - claude-trace/pkg/render/tool_formatting.go
  - conf/cmd/display.go
  - conf/cmd/helpers.go
  - conf/cmd/main.go
  - conf/pkg/config/config.go
  - conf/pkg/config/config_test.go
  - conf/pkg/config/declarative_test.go
  - conf/pkg/diff/diff.go
  - conf/pkg/editors/json.go
  - conf/pkg/editors/json_test.go
  - conf/pkg/editors/toml.go
  - conf/pkg/editors/toml_test.go
  - conf/pkg/schemas/claude_parser.go
  - conf/pkg/schemas/claude_parser_test.go
  - conf/pkg/schemas/jj_parser.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## e6c83d4e1c79
**Description:** Add lib/version package and version subcommand to all Go CLI tools
**Files:**
  - .dagger/main.go
  - claude-trace/cmd/version.go
  - dissect/cmd/version.go
  - lib/version/version.go
  - printpdf/cmd/version.go
  - tk/cmd/version.go
  - want/cmd/version.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f60d266a1e0f
**Description:** docs: Document TK_DB_PATH environment variable
**Files:**
  - AGENTS.md
  - README.md
  - tk/README.md
  - tk/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0c48b04c0d73
**Description:** tk: Handle non-numeric beads IDs with fallback numbering
**Files:**
  - tk/cmd/import_beads.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 049ee309130b
**Description:** tk: Multi-prefix beads import with clash detection
**Files:**
  - tk/cmd/import_beads.go
  - tk/cmd/import_beads_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6aa353aafd85
**Description:** tk: Fix column width calculation to use full terminal width
**Files:**
  - tk/cmd/column_width.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 56d44cc2be96
**Description:** tk: Make all tables in ls use consistent column widths
**Files:**
  - tk/cmd/display.go
  - tk/cmd/ls.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b1743916fecb
**Description:** tk: Implement proper column width calculation system
**Files:**
  - tk/cmd/column_width.go
  - tk/cmd/column_width_test.go
  - tk/cmd/display.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 188657f1c310
**Description:** tk: Let ID, Status, and P columns auto-size in ls command
**Files:**
  - tk/cmd/display.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 86b2e1268305
**Description:** tk: Improve title column width calculation in ls command
**Files:**
  - tk/cmd/display.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 1b37c460753d
**Description:** tk: Fix terminal width detection by checking /dev/tty
**Files:**
  - tk/cmd/ls.go
  - tk/internal/termutil/termutil.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 4b7f2f15db36
**Description:** tk: Complete metadata implementation - display in ls, bd-mono alias
**Files:**
  - tk/cmd/display.go
  - tk/cmd/import_beads.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c848c7e6fe57
**Description:** tk: Reorganize debug commands, consolidate test helpers, add repair
**Files:**
  - tk/cmd/blockers.go
  - tk/cmd/conflicts.go
  - tk/cmd/conflicts_numbers.go
  - tk/cmd/conflicts_numbers_test.go
  - tk/cmd/debug/repair.go
  - tk/cmd/debug.go
  - tk/cmd/delete.go
  - tk/cmd/describe.go
  - tk/cmd/edit.go
  - tk/cmd/edit_cmd_status_test.go
  - tk/cmd/events.go
  - tk/cmd/export.go
  - tk/cmd/graph.go
  - tk/cmd/id.go
  - tk/cmd/import_beads.go
  - tk/cmd/ingest.go
  - tk/cmd/ls.go
  - tk/cmd/mark.go
  - tk/cmd/meta.go
  - tk/cmd/mv.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0370d1c2bdb2
**Description:** tk: Add metadata support with claims and authority resolution
**Files:**
  - go.mod
  - go.sum
  - tk/cmd/import_beads.go
  - tk/cmd/meta.go
  - tk/cmd/meta_integration_test.go
  - tk/cmd/view.go
  - tk/internal/reducer/metadata.go
  - tk/internal/reducer/metadata_test.go
  - tk/internal/reducer/reducer.go
  - tk/internal/types/ids.go
  - tk/internal/types/metadata.go
  - tk/internal/types/metadata_test.go
  - tk/internal/types/payloads.go
  - tk/internal/types/task.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 808931f6ec39
**Description:** tk: Fix event ordering to use Lamport timestamps (TDD)
**Files:**
  - tk/internal/database/db_events.go
  - tk/internal/database/db_events_test.go
  - tk/specs/TIMESTAMPS.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 888552b141e1
**Description:** tk: Add TIMESTAMPS.md spec documenting event ordering model
**Files:**
  - tk/specs/TIMESTAMPS.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 69e79b92c342
**Description:** tk: Add event-projection consistency check to debug doctor
**Files:**
  - tk/cmd/doctor.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 4e5a65b2c61a
**Description:** tk: Add TK_DB_PATH environment variable for custom database location
**Files:**
  - tk/internal/database/db.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 63b8256d7437
**Description:** tk: Fix beads import to handle actual bd dependency format
**Files:**
  - tk/cmd/import_beads.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 26157cb5b8a1
**Description:** version: Show build time in local timezone for human-readable output
**Files:**
  - tk/cmd/version.go
  - want/cmd/version.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 00261d01198d
**Description:** tk, want: Add --json flag to version commands
**Files:**
  - tk/cmd/version.go
  - want/main.go
  - want/version.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ed0d10d88a3b
**Description:** want: Add @local builds and version commands for tk and want
**Files:**
  - tk/cmd/import_beads.go
  - tk/cmd/version.go
  - want/main.go
  - want/mono.go
  - want/version.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c14b6d39221a
**Description:** tk: assign task numbers when importing from beads
**Files:**
  - tk/cmd/import_beads.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c65398f6eb12
**Description:** tk: implement import from beads
**Files:**
  - tk/cmd/import_beads.go
  - tk/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 14eb1ce34657
**Description:** tk view: rename to tk show
**Files:**
  - tk/IMPLEMENTATION.md
  - tk/README.md
  - tk/RELATIONS.md
  - tk/V4_IMPLEMENTATION_STATUS.md
  - tk/cmd/display.go
  - tk/cmd/root.go
  - tk/cmd/view.go
  - tk/specs/v4-migration.md
  - tk-vscode/package-lock.json
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8153fc56cd9a
**Description:** tk: Rename log file and fix security permissions
**Files:**
  - tk/internal/invlog/invlog.go
  - tk/internal/invlog/invlog_test.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## eb329922c3a0
**Description:** tk: Use t.Cleanup and proper pipe resource cleanup
**Files:**
  - tk/internal/invlog/invlog_test.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 7d4961ed5ec1
**Description:** tk: Use build constraints for platform-specific file locking
**Files:**
  - tk/internal/invlog/invlog.go
  - tk/internal/invlog/invlog_test.go
  - tk/internal/invlog/invlog_unix.go
  - tk/internal/invlog/invlog_windows.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## d30b01180f5e
**Description:** tk: Improve error handling and cross-platform compatibility
**Files:**
  - tk/internal/invlog/invlog.go
  - tk/internal/invlog/invlog_test.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2f1ddcbd1e73
**Description:** tk: Fix resource leak and add file locking for concurrent writes
**Files:**
  - tk/internal/invlog/invlog.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0a58f43f2831
**Description:** tk: Add proper error handling for pipe creation in logging
**Files:**
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c0226ccdebd7
**Description:** tk: Add invocation logging to track all command executions
**Files:**
  - tk/internal/invlog/invlog.go
  - tk/internal/invlog/invlog_test.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6f61c980299b
**Description:** tk ls: -p now only shows what it should
**Files:**
  - tk/cmd/ls.go
  - tk/cmd/ls_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c533a5a34ca6
**Description:** tk: implement TaskRef type and remove code duplication
**Files:**
  - tk/cmd/blockers.go
  - tk/cmd/conflicts.go
  - tk/cmd/debug_helpers.go
  - tk/cmd/delete.go
  - tk/cmd/edit.go
  - tk/cmd/export.go
  - tk/cmd/graph.go
  - tk/cmd/id.go
  - tk/cmd/ingest.go
  - tk/cmd/json_persistence.go
  - tk/cmd/mark.go
  - tk/cmd/mv.go
  - tk/cmd/note.go
  - tk/cmd/relate.go
  - tk/cmd/sync.go
  - tk/cmd/utils.go
  - tk/cmd/view.go
  - tk/internal/database/db_counters.go
  - tk/internal/database/db_taskids.go
  - tk/internal/database/task_resolver.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## eebdeafe5996
**Description:** Auto-format code
**Files:**
  - tk/cmd/debug.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f29b373e3d5c
**Description:** tk: fix staticcheck issues
**Files:**
  - tk/cmd/delete_test.go
  - tk/cmd/display.go
  - tk/cmd/relate.go
  - tk/cmd/test_helpers.go
  - tk/internal/reducer/reducer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 24b73adfa07d
**Description:** tk: remove temporary refactoring plan
**Files:**
  - tk/REFACTORING_PLAN.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b4805e45aa3e
**Description:** tk: clean up helper file names
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 214780ab3c18
**Description:** tk: reorganize debug commands
**Files:**
  - tk/cmd/helpers_debug.go
  - tk/cmd/root.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## d37019139944
**Description:** tk: rename helpers_display.go → display.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a070db5c0c89
**Description:** tk: move DB operations to internal/database
**Files:**
  - tk/cmd/helpers_display.go
  - tk/cmd/helpers_utils.go
  - tk/cmd/ls.go
  - tk/cmd/ls_test.go
  - tk/cmd/new.go
  - tk/internal/database/projects.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f95db956342c
**Description:** tk: move pure functions to internal/types
**Files:**
  - tk/cmd/helpers_display.go
  - tk/cmd/helpers_test.go
  - tk/cmd/helpers_utils.go
  - tk/cmd/ls.go
  - tk/cmd/sort_test.go
  - tk/internal/types/task_id.go
  - tk/internal/types/task_sort.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0bec5275c0c7
**Description:** tk: remove duplicate generateEventID/getNextLamportTimestamp
**Files:**
  - tk/REFACTORING_PLAN.md
  - tk/cmd/helpers_tasks.go
  - tk/cmd/project.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ca1aae6bb3ee
**Description:** tk: clean up cmd/ directory test organization
**Files:**
  - tk/cmd/helpers_test.go
  - tk/cmd/integration_test.go
  - tk/cmd/main_test.go
  - tk/cmd/relations_integration_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 7edf57eabdc1
**Description:** tk: restructure to follow Go conventions
**Files:**
  - tk/cmd/admin.go
  - tk/cmd/blockers.go
  - tk/cmd/conflicts.go
  - tk/cmd/conflicts_numbers.go
  - tk/cmd/conflicts_numbers_cmd_test.go
  - tk/cmd/delete.go
  - tk/cmd/delete_cmd_test.go
  - tk/cmd/describe.go
  - tk/cmd/describe_cmd_test.go
  - tk/cmd/doctor.go
  - tk/cmd/doctor_cmd_test.go
  - tk/cmd/edge_cases_test.go
  - tk/cmd/edit.go
  - tk/cmd/edit_cmd_status_test.go
  - tk/cmd/edit_cmd_test.go
  - tk/cmd/events.go
  - tk/cmd/export.go
  - tk/cmd/graph.go
  - tk/cmd/helpers_debug.go
  - tk/cmd/helpers_display.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## d55e9689cc86
**Description:** tk: organize tests into cmd/ and test/ directories
**Files:**
  - tk/cmd/helpers_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## baa0ee7791f1
**Description:** tk: move all code to cmd/ directory
**Files:**
  - tk/cmd/init.go
  - tk/cmd/ls.go
  - tk/cmd/main.go
  - tk/cmd/mark.go
  - tk/cmd/new.go
  - tk/cmd/note.go
  - tk/cmd/path.go
  - tk/cmd/view.go
  - tk/main.go
  - tk/mise.toml
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ca102d1dee39
**Description:** tk: organize files into internal packages
**Files:**
  - tk/admin_cmd.go
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/debug_events.go
  - tk/delete_cmd_test.go
  - tk/export_cmd.go
  - tk/graph_cmd.go
  - tk/ingest_cmd.go
  - tk/ls_cmd_test.go
  - tk/main.go
  - tk/remote_cmd.go
  - tk/status_cmd.go
  - tk/sync_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2b3c56c421a9
**Description:** tk: remove duplicate files after batch move to internal/database
**Files:**
  - tk/db.go
  - tk/db_counters.go
  - tk/db_events.go
  - tk/db_node.go
  - tk/db_schema.go
  - tk/db_taskids.go
  - tk/projections.go
  - tk/task_resolver.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 252632cd2dbe
**Description:** tk: move database functions
**Files:**
  - tk/admin_cmd.go
  - tk/blockers_cmd.go
  - tk/collision.go
  - tk/conflicts_numbers_cmd.go
  - tk/conflicts_numbers_cmd_test.go
  - tk/db.go
  - tk/db_counters.go
  - tk/db_events.go
  - tk/db_node.go
  - tk/db_schema.go
  - tk/db_taskids.go
  - tk/delete_cmd.go
  - tk/delete_cmd_test.go
  - tk/describe_cmd_test.go
  - tk/display.go
  - tk/doctor_cmd.go
  - tk/doctor_cmd_test.go
  - tk/edge_cases_test.go
  - tk/edit_cmd.go
  - tk/edit_cmd_status_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 02c7656605b7
**Description:** Add type safety for project references
**Files:**
  - tk/doctor_cmd.go
  - tk/id_cmd.go
  - tk/internal/types/ids.go
  - tk/project_cmd.go
  - tk/task_resolver.go
  - tk/tasks.go
  - tk/utils.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## bbe351ed6223
**Description:** tk: more tests
**Files:**
  - tk/internal/reducer/reducer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6e9fc3f0e6d2
**Description:** tk: fix project deletion?
**Files:**
  - tk/internal/reducer/reducer.go
  - tk/internal/reducer/reducer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## d3e2bf32291e
**Description:** Changes before error encountered
**Files:**
  - tk/internal/reducer/reducer.go
  - tk/projections.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f90a0952b10c
**Description:** tk: Fix project deletion to also delete tasks in reducer
**Files:**
  - tk/internal/reducer/reducer.go
  - tk/projections.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 46b0dc1099d1
**Description:** tk/tk-vsc: Implement delete features, rename commands, and add status filters
**Files:**
  - tk/delete_cmd.go
  - tk/ingest_cmd.go
  - tk/internal/types/events.go
  - tk/internal/types/ids.go
  - tk/main.go
  - tk/project_cmd.go
  - tk/projections.go
  - tk-vscode/AGENTS.md
  - tk-vscode/package.json
  - tk-vscode/src/extension.ts
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## d8f7df75978d
**Description:** Sort projects alphabetically in tk ls and JSON output
**Files:**
  - tk/display.go
  - tk/ls_cmd_test.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f057b5d33f88
**Description:** Remove unused variable and add project to VSCode group enum
**Files:**
  - tk/display.go
  - tk-vscode/package.json
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0f2564128a3c
**Description:** Fix JSON output to accept both project and prefix grouping modes
**Files:**
  - tk/display.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f16a6015aa19
**Description:** Show empty projects in tk ls and make new project button more visible
**Files:**
  - tk/display.go
  - tk/main.go
  - tk/utils.go
  - tk-vscode/package.json
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0f3329fc4b61
**Description:** tk: add comprehensive unit tests for delete functionality
**Files:**
  - tk/internal/reducer/reducer_test.go
  - tk/projection_test.go
  - tk/relations_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ee2fb424b8a4
**Description:** tk: add projection call after inserting delete event
**Files:**
  - tk/delete_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5b66792e01af
**Description:** tk: implement delete command with event sourcing and tests
**Files:**
  - tk/delete_cmd.go
  - tk/delete_cmd_test.go
  - tk/ingest_cmd.go
  - tk/internal/reducer/reducer.go
  - tk/internal/relations/graph.go
  - tk/internal/types/ids.go
  - tk/internal/types/payloads.go
  - tk/main.go
  - tk/projections.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 12acb1bb27fe
**Description:** tk: Remove v4 prefix from all files
**Files:**
  - tk/db_test.go
  - tk/v4_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 40627c258cbd
**Description:** tk: Move test files to be colocated with their code
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a9b479c1b853
**Description:** tk: Move reducer to internal/reducer/ and consolidate utilities
**Files:**
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/db.go
  - tk/db_reducer.go
  - tk/edit_cmd_status_test.go
  - tk/graph_cmd.go
  - tk/integration_test.go
  - tk/internal/utils/blocking.go
  - tk/internal/utils/ulid.go
  - tk/reducer_test.go
  - tk/relations.go
  - tk/relations_test.go
  - tk/sync_test.go
  - tk/ulid.go
  - tk/ulid_test.go
  - tk/v4_reducer.go
  - tk/v4_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## e48301ec1a04
**Description:** tk: Consolidate all payload types into internal/types/
**Files:**
  - tk/edit_cmd.go
  - tk/integration_test.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/project_cmd.go
  - tk/reducer.go
  - tk/reducer_test.go
  - tk/relate_cmd.go
  - tk/relations_test.go
  - tk/tasks.go
  - tk/test_helpers.go
  - tk/v4_projections.go
  - tk/v4_reducer.go
  - tk/v4_reducer_test.go
  - tk/v4_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## d07c5ac01a4d
**Description:** tk: fix aliases like "jj-run"
**Files:**
  - tk/edit_cmd.go
  - tk/internal/types/v4_types.go
  - tk/main.go
  - tk/relate_cmd.go
  - tk/v4_types_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b1b715323ab3
**Description:** tk: Move v4_types.go to internal/types/
**Files:**
  - tk/db_taskids.go
  - tk/edit_cmd.go
  - tk/id_cmd_test.go
  - tk/ingest_cmd.go
  - tk/integration_test.go
  - tk/mv_cmd.go
  - tk/project_cmd.go
  - tk/reducer_test.go
  - tk/relations_test.go
  - tk/task_resolver.go
  - tk/task_resolver_test.go
  - tk/tasks.go
  - tk/test_helpers.go
  - tk/v4_edge_cases_test.go
  - tk/v4_projections.go
  - tk/v4_reducer.go
  - tk/v4_reducer_test.go
  - tk/v4_test.go
  - tk/v4_types_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 1b79cf199349
**Description:** tk: Move Event and Task types to internal/types/
**Files:**
  - tk/blockers_cmd.go
  - tk/collision.go
  - tk/collision_test.go
  - tk/db_events.go
  - tk/debug_events.go
  - tk/display.go
  - tk/edit_cmd.go
  - tk/events_cmd.go
  - tk/export_cmd.go
  - tk/graph_cmd.go
  - tk/ingest_cmd.go
  - tk/integration_test.go
  - tk/internal/types/event.go
  - tk/internal/types/task.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/project_cmd.go
  - tk/reducer.go
  - tk/reducer_test.go
  - tk/relate_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5ecbc4f8f298
**Description:** tk: Extract RelationsGraph to internal/relations
**Files:**
  - tk/blockers_cmd.go
  - tk/integration_test.go
  - tk/internal/relations/graph.go
  - tk/reducer.go
  - tk/relations.go
  - tk/relations_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 7e8c738654b6
**Description:** tk: Rename 'tk project list' to 'tk project ls' and use consistent table styling
**Files:**
  - AGENTS.md
  - tk/project_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## cc9e802920e7
**Description:** tk: Clean up all orphaned comments from type moves
**Files:**
  - tk/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0ce5d9875e39
**Description:** tk: Clean up orphaned comments in types.go
**Files:**
  - tk/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 9387e298046b
**Description:** tk: Move relation types to internal/types/relations.go
**Files:**
  - tk/graph_cmd.go
  - tk/internal/types/relations.go
  - tk/relations.go
  - tk/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f01efbf22577
**Description:** tk: Move status types to internal/types/status.go
**Files:**
  - tk/internal/types/status.go
  - tk/reducer.go
  - tk/reducer_test.go
  - tk/relations_test.go
  - tk/types.go
  - tk/v4_reducer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 4a04e383a7cf
**Description:** tk: Move Blocker type to internal/types/blocker.go
**Files:**
  - tk/internal/types/blocker.go
  - tk/relations.go
  - tk/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0588049c46e6
**Description:** tk: Move Note type to internal/types/note.go
**Files:**
  - tk/internal/types/note.go
  - tk/reducer.go
  - tk/types.go
  - tk/v4_reducer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 1f0dc3c80f17
**Description:** tk: Extract payload types to internal/payloads
**Files:**
  - tk/edit_cmd.go
  - tk/integration_test.go
  - tk/internal/payloads/payloads.go
  - tk/main.go
  - tk/reducer.go
  - tk/reducer_test.go
  - tk/relate_cmd.go
  - tk/relations_test.go
  - tk/types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 13b0429976e2
**Description:** tk: Move segment_writer.go to internal/segment package
**Files:**
  - tk/export_cmd.go
  - tk/sync_cmd.go
  - tk/sync_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 50821534dc16
**Description:** tk: Move segment_reader.go to internal/segment package
**Files:**
  - tk/collision.go
  - tk/debug_events.go
  - tk/ingest_cmd.go
  - tk/sync_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## fba6e066f06d
**Description:** tk: Fix RenderTaskDisplayID to use project name instead of project ID
**Files:**
  - tk/task_resolver.go
  - tk/task_resolver_test.go
  - tk/test_helpers.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## be2ff6fc329e
**Description:** tk: Move sync_types.go to internal/sync package
**Files:**
  - tk/collision.go
  - tk/collision_test.go
  - tk/config.go
  - tk/config_test.go
  - tk/db.go
  - tk/db_reducer.go
  - tk/debug_events.go
  - tk/export_cmd.go
  - tk/ingest_cmd.go
  - tk/integration_test.go
  - tk/reducer.go
  - tk/remote_cmd.go
  - tk/segment_reader.go
  - tk/segment_writer.go
  - tk/segment_writer_test.go
  - tk/status_cmd.go
  - tk/sync_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## cd4e301fd9be
**Description:** tk: Export utility functions and add DB wrapper methods
**Files:**
  - tk/admin_cmd.go
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/conflicts_numbers_cmd.go
  - tk/db.go
  - tk/describe_cmd.go
  - tk/doctor_cmd.go
  - tk/doctor_cmd_test.go
  - tk/edit_cmd.go
  - tk/events_cmd.go
  - tk/export_cmd.go
  - tk/graph_cmd.go
  - tk/id_cmd.go
  - tk/ingest_cmd.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/node_cmd.go
  - tk/project_cmd.go
  - tk/relate_cmd.go
  - tk/sync_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## e78f3fa4fc95
**Description:** tk: Show project name instead of UID in `tk ls` for projects without aliases
**Files:**
  - tk/utils.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2bf945db08a8
**Description:** tk: Add --unset flag to mark command for clearing status
**Files:**
  - tk/edit_cmd.go
  - tk/main.go
  - tk-vscode/src/extension.ts
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c712fca471ab
**Description:** tk describe
**Files:**
  - tk/describe_cmd.go
  - tk/describe_cmd_test.go
  - tk/edit_cmd.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 308b244272c5
**Description:** tk: make ls project filter work
**Files:**
  - tk/README.md
  - tk/RELATIONS.md
  - tk/db_taskids.go
  - tk/db_test.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 694cc9050868
**Description:** tk: add -p shorthand for project flag
**Files:**
  - tk/README.md
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c15b1bf673df
**Description:** tk: Run go fmt on updated files
**Files:**
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/doctor_cmd.go
  - tk/events_cmd.go
  - tk/graph_cmd.go
  - tk/main.go
  - tk/node_cmd.go
  - tk/project_cmd.go
  - tk/remote_cmd.go
  - tk/status_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3dc1ce3aaeac
**Description:** tk: Add --json flag to most readonly commands
**Files:**
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/conflicts_numbers_cmd.go
  - tk/doctor_cmd.go
  - tk/events_cmd.go
  - tk/graph_cmd.go
  - tk/id_cmd.go
  - tk/main.go
  - tk/node_cmd.go
  - tk/project_cmd.go
  - tk/remote_cmd.go
  - tk/status_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2e2a953517b1
**Description:** tk: rename "tk status set" to "tk mark"
**Files:**
  - tk/IMPLEMENTATION.md
  - tk/README.md
  - tk/V4_IMPLEMENTATION_STATUS.md
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 72b3e01991b9
**Description:** tk: idk fix something about task IDs
**Files:**
  - tk/db_taskids.go
  - tk/db_test.go
  - tk/display.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6ac1ac58d8fc
**Description:** Split tk files into logical modules (main.go: 921→547, db.go: 843→152 lines)
**Files:**
  - tk/db.go
  - tk/db_counters.go
  - tk/db_events.go
  - tk/db_node.go
  - tk/db_reducer.go
  - tk/db_schema.go
  - tk/db_taskids.go
  - tk/display.go
  - tk/main.go
  - tk/tasks.go
  - tk/utils.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5daf03f0aa47
**Description:** tk: polish
**Files:**
  - tk/collision.go
  - tk/main.go
  - tk/sync_cmd.go
  - tk/task_resolver.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a95fbc452130
**Description:** tk remote rm
**Files:**
  - tk/remote_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## beb66a97193c
**Description:** Establish Go code style guide
**Files:**
  - AGENTS.md
  - GO_STYLE_GUIDE.md
  - tk/project_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 821c2afa7c79
**Description:** remove all v4 migration stuff
**Files:**
  - tk/admin_cmd.go
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/conflicts_numbers_cmd.go
  - tk/db.go
  - tk/db_test.go
  - tk/doctor_cmd.go
  - tk/edit_cmd.go
  - tk/events_cmd.go
  - tk/export_cmd.go
  - tk/graph_cmd.go
  - tk/id_cmd.go
  - tk/ingest_cmd.go
  - tk/integration_test.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/node_cmd.go
  - tk/prefix_cmd.go
  - tk/prefix_test.go
  - tk/project_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2f8902ac6c71
**Description:** finally v4 migration finally
**Files:**
  - tk/admin_cmd.go
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/conflicts_numbers_cmd.go
  - tk/db.go
  - tk/doctor_cmd.go
  - tk/edit_cmd.go
  - tk/events_cmd.go
  - tk/export_cmd.go
  - tk/graph_cmd.go
  - tk/id_cmd.go
  - tk/ingest_cmd.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/node_cmd.go
  - tk/prefix_cmd.go
  - tk/project_cmd.go
  - tk/relate_cmd.go
  - tk/sync_cmd.go
  - tk/sync_integration_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f7fcbabcb6cf
**Description:** fmt
**Files:**
  - tk/segment_writer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 616d5b6561de
**Description:** Add tests for segment_writer, main utils, v4_reducer and db functions
**Files:**
  - tk/db_test.go
  - tk/main_test.go
  - tk/segment_writer_test.go
  - tk/v4_reducer_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 15a4c936b9ed
**Description:** tk: fix duplicate tasks
**Files:**
  - tk/ingest_cmd.go
  - tk/main.go
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## adc9e1c3545c
**Description:** Add tests for v4_types, ulid, and config modules
**Files:**
  - tk/config_test.go
  - tk/ulid_test.go
  - tk/v4_types_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## f4e4b4020345
**Description:** Add code coverage metrics to tk CI workflow
**Files:**
  - .github/workflows/tk.yml
  - tk/.gitignore
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b547576e0e77
**Description:** Fix collision detection by removing local node ID filter
**Files:**
  - tk/collision.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 6f7b7cb16e38
**Description:** Add test demonstrating collision detection bug
**Files:**
  - tk/collision_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## df7f43795bc4
**Description:** tk: Remove backup/rollback tests as functionality was intentionally removed
**Files:**
  - tk/v4_edge_cases_test.go
  - tk/v4_migration.go
  - tk/v4_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 1c3d95f6afb6
**Description:** tk: Add missing RollbackV4 function and backup creation
**Files:**
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## baef5978a6c0
**Description:** tk: Document reducer cache config comparison behavior
**Files:**
  - tk/db.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 1b3150a1ba88
**Description:** tk: Implement reducer caching for performance
**Files:**
  - tk/blockers_cmd.go
  - tk/conflicts_cmd.go
  - tk/db.go
  - tk/graph_cmd.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## cea3d7e66324
**Description:** tk: Fix keep number policy and implement collision detection
**Files:**
  - tk/collision.go
  - tk/v4_projections.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c83c4c47afe2
**Description:** tk: Use proper ULID library and make space configurable
**Files:**
  - tk/debug_events.go
  - tk/ingest_cmd.go
  - tk/ulid.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## ab6b98b70b47
**Description:** tk: Fix panic and recursive string search (Phase 1)
**Files:**
  - tk/db.go
  - tk/debug_events.go
  - tk/main.go
  - tk/sync_integration_test.go
  - tk/ulid.go
  - tk/v4_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 78612618358e
**Description:** fix
**Files:**
  - tk/admin_cmd.go
  - tk/main.go
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 18c359afdcb7
**Description:** tk: add --json for tk events list
**Files:**
  - tk/events_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## cd34a5e77c32
**Description:** really fix?
**Files:**
  - tk/debug_events.go
  - tk/ingest_cmd.go
  - tk/project_cmd.go
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 48660149b44e
**Description:** Fix v3 event projection and add debug flag
**Files:**
  - tk/admin_cmd.go
  - tk/ingest_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2a642ca681ab
**Description:** Fix rebuild-from-remote to properly migrate v3 to v4
**Files:**
  - tk/admin_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b954d17f5592
**Description:** Add rebuild-from-remote command for corrupted databases
**Files:**
  - tk/admin_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## e43316f6b337
**Description:** Add validate-migration debug command
**Files:**
  - tk/admin_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 09a16c66438d
**Description:** Fix task lookup when TaskUUID contains task ID
**Files:**
  - tk/v4_migration.go
  - tk/v4_migration_task_lookup_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 4bf889508fdc
**Description:** Clarify that created_at stores nanoseconds
**Files:**
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2535623895bf
**Description:** Use deterministic timestamps for missing prefix projects
**Files:**
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2f82ed33f3d8
**Description:** Fix v4 migration to handle missing/removed prefixes
**Files:**
  - tk/v4_migration.go
  - tk/v4_migration_missing_prefix_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a725ff1f4ee8
**Description:** tk: Preserve original legacy UUID in task registration
**Files:**
  - tk/v4_migration.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a183978d3788
**Description:** tk: Address code review feedback - fix UUID consistency and error handling
**Files:**
  - tk/v4_migration.go
  - tk/v4_migration_task_lookup_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3de6b2577c41
**Description:** tk: Add comprehensive tests for v4 migration edge cases
**Files:**
  - tk/test_helpers.go
  - tk/v4_edge_cases_test.go
  - tk/v4_migration.go
  - tk/v4_migration_task_lookup_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 0e5e06abb939
**Description:** tk: Fix v4 migration to handle events referencing tasks not yet seen
**Files:**
  - tk/v4_migration.go
  - tk/v4_migration_task_lookup_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a4e4d2f47024
**Description:** tk: Fix duplicate task bug in ls command by preventing double event processing
**Files:**
  - tk/reducer.go
  - tk/v4_reducer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## dd3d0f3cf2e6
**Description:** tk: Document v4 implementation completion with comprehensive status report
**Files:**
  - tk/V4_IMPLEMENTATION_STATUS.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c7c071f80403
**Description:** tk: Add comprehensive v4 edge case and integration tests
**Files:**
  - tk/test_helpers.go
  - tk/v4_edge_cases_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8ae8aad7520d
**Description:** tk: Fix v4 ls command - make version-aware for task listing and grouping
**Files:**
  - tk/db.go
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 928ff5e168e0
**Description:** tkv4
**Files:**
  - tk/blockers_cmd.go
  - tk/conflicts_numbers_cmd.go
  - tk/conflicts_numbers_cmd_test.go
  - tk/db.go
  - tk/doctor_cmd.go
  - tk/doctor_cmd_test.go
  - tk/edit_cmd.go
  - tk/edit_cmd_status_test.go
  - tk/edit_cmd_test.go
  - tk/graph_cmd.go
  - tk/id_cmd.go
  - tk/id_cmd_test.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/mv_cmd_test.go
  - tk/mv_test.go
  - tk/relate_cmd.go
  - tk/task_resolver.go
  - tk/task_resolver_test.go
  - tk/test_helpers.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 203e1d998082
**Description:** tk: Fix event sourcing architecture - use proper projections
**Files:**
  - tk/ingest_cmd.go
  - tk/main.go
  - tk/project_cmd.go
  - tk/v4_migration.go
  - tk/v4_projections.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 981509cbfe15
**Description:** tk: Update README with v4 documentation
**Files:**
  - tk/README.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5e6d7d04aee8
**Description:** tk: Add comprehensive v4 tests (types, migration, events)
**Files:**
  - tk/v4_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 7b89c11a5f93
**Description:** tk: Add v4 task creation with project support
**Files:**
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c7e519c1a45e
**Description:** tk: Add project management CLI commands (v4)
**Files:**
  - tk/main.go
  - tk/project_cmd.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## c462544eb6e3
**Description:** tk: Add v4 reducer, migration trigger, and admin rollback command
**Files:**
  - tk/admin_cmd.go
  - tk/main.go
  - tk/reducer.go
  - tk/v4_migration.go
  - tk/v4_reducer.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 2714ad21d6a6
**Description:** tk: Add v4 type definitions and migration scaffolding
**Files:**
  - go.mod
  - go.sum
  - tk/v4_events.go
  - tk/v4_migration.go
  - tk/v4_types.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b64d688f6390
**Description:** Migrate to a single go.mod (#148)
**Files:**
  - .github/workflows/beads-merge.yml
  - .github/workflows/claude-trace.yml
  - .github/workflows/conf.yml
  - .github/workflows/dissect.yml
  - .github/workflows/go-workspace-lint.yml
  - .github/workflows/ingest.yml
  - .github/workflows/markdown-format.yml
  - .github/workflows/printpdf.yml
  - .github/workflows/prrun.yml
  - .github/workflows/tk.yml
  - .github/workflows/want.yml
  - beads-merge/go.mod
  - beads-merge/go.sum
  - claude-trace/cmd/extract.go
  - claude-trace/cmd/list.go
  - claude-trace/cmd/tui_cmd.go
  - claude-trace/cmd/view_cmd.go
  - claude-trace/go.mod
  - claude-trace/go.sum
  - claude-trace/pkg/render/markdown_conversation.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## de2eb1d02564
**Description:** tk: v4 migration strategy (prefixes → projects, mutable task numbers) (#145)
**Files:**
  - tk/specs/spec-v4-projects.md
  - tk/specs/v4-migration.md
  - tk/specs/v4-types.md
  - tk/specs/v4.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 343d8d76775e
**Description:** tk: spec v4
**Files:**
  - tk/specs/spec-v4-projects.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## b7de8fb07704
**Description:** tk: move specs
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3a51601f1c6e
**Description:** tk: Add sync integration tests and event debugging commands (#143)
**Files:**
  - tk/README.md
  - tk/events_cmd.go
  - tk/main.go
  - tk/prefix_cmd.go
  - tk/sync_integration_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8534884e313c
**Description:** tk: fix prefix event projection during sync (#142)
**Files:**
  - conf/go.mod
  - conf/go.sum
  - prrun/go.sum
  - tk/db.go
  - tk/ingest_cmd.go
  - tk/prefix_test.go
  - tk/sync_test.go
  - want/go.sum
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 8937dd25034b
**Description:** Implement task relations with OR-set CRDT for blocking and subtasks (#141)
**Files:**
  - tk/IMPLEMENTATION_SUMMARY.md
  - tk/README.md
  - tk/RELATIONS.md
  - tk/V3_ROLLUPS.md
  - tk/blockers_cmd.go
  - tk/config.go
  - tk/conflicts_cmd.go
  - tk/graph_cmd.go
  - tk/integration_test.go
  - tk/main.go
  - tk/reducer.go
  - tk/relate_cmd.go
  - tk/relations.go
  - tk/relations_test.go
  - tk/sync_types.go
  - tk/types.go
  - tk/ulid.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 35fc2db5c5d8
**Description:** lib/toml: use tomledit's ParseKey for TOML-compliant path parsing (#140)
**Files:**
  - beads-merge/go.sum
  - claude-trace/go.sum
  - conf/cmd/integration_test.go
  - conf/go.mod
  - conf/go.sum
  - conf/pkg/editors/toml.go
  - conf/pkg/editors/toml_test.go
  - dissect/go.sum
  - go.work.sum
  - ingest/go.sum
  - lib/ghrelease/go.sum
  - lib/toml/README.md
  - lib/toml/go.sum
  - lib/toml/path_parsing_test.go
  - lib/toml/toml.go
  - markdown-format/go.sum
  - printpdf/go.sum
  - prrun/go.mod
  - tk/go.sum
  - want/go.mod
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 13f77e63a1cc
**Description:** tk: Fix auto-number collision when moving tasks across prefixes (#139)
**Files:**
  - tk/mv_cmd.go
  - tk/mv_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3f77427ab1ae
**Description:** clean up go.sum
**Files:**
  - beads-merge/go.sum
  - claude-trace/go.sum
  - conf/go.sum
  - dissect/go.sum
  - ingest/go.sum
  - lib/ghclient/go.sum
  - lib/ghrelease/go.sum
  - lib/toml/go.sum
  - markdown-format/go.sum
  - mise.toml
  - printpdf/go.sum
  - prrun/go.sum
  - tk/go.sum
  - want/go.sum
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 713db56bbe25
**Description:** tk ls: add grouping
**Files:**
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 5815e37c054f
**Description:** tk ls: separate rows
**Files:**
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 7f12a3bc7809
**Description:** Fix Go workspace configuration and add consistency linter
**Files:**
  - .github/workflows/go-workspace-lint.yml
  - AGENTS.md
  - claude-trace/go.mod
  - claude-trace/go.sum
  - conf/go.mod
  - go.work
  - go.work.sum
  - ingest/go.mod
  - ingest/go.sum
  - lib/ghrelease/go.mod
  - lib/ghrelease/go.sum
  - mise.toml
  - printpdf/go.mod
  - printpdf/go.sum
  - prrun/go.mod
  - prrun/go.sum
  - scripts/lint-go-workspace.py
  - tk/go.mod
  - tk/go.sum
  - want/go.mod
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 3ca40f94f5ac
**Description:** tk: Add task movement with UUID-based identity and alias preservation (#129)
**Files:**
  - tk/README.md
  - tk/db.go
  - tk/main.go
  - tk/mv_cmd.go
  - tk/mv_test.go
  - tk/prefix_cmd.go
  - tk/reducer.go
  - tk/types.go
  - tk/ulid.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 00acabbea4a3
**Description:** Add event-sourced prefix support for task organization (#127)
**Files:**
  - go.work.sum
  - tk/README.md
  - tk/db.go
  - tk/go.mod
  - tk/go.sum
  - tk/ingest_cmd.go
  - tk/main.go
  - tk/prefix_cmd.go
  - tk/prefix_test.go
  - tk/reducer.go
  - tk/sync_test.go
  - tk/types.go
  - tk/ulid.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 7afe89078dd8
**Description:** tk: Display tasks as formatted table with colored statuses and word wrapping (#123)
**Files:**
  - go.work.sum
  - tk/go.mod
  - tk/go.sum
  - tk/main.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## a92c6c2bd498
**Description:** Implement tk v1 sync: offline-first event synchronization via immutable segments (#122)
**Files:**
  - tk/IMPLEMENTATION.md
  - tk/README.md
  - tk/collision.go
  - tk/config.go
  - tk/db.go
  - tk/export_cmd.go
  - tk/go.mod
  - tk/go.sum
  - tk/ingest_cmd.go
  - tk/main.go
  - tk/node_cmd.go
  - tk/remote_cmd.go
  - tk/segment_reader.go
  - tk/segment_writer.go
  - tk/status_cmd.go
  - tk/sync_cmd.go
  - tk/sync_test.go
  - tk/sync_types.go
  - tk/ulid.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 95684ac68cb6
**Description:** tk spec
**Files:**
  - tk/spec-v1.md
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

## 14dce560519d
**Description:** rename tak -> tk (#121)
**Files:**
  - .beads/issues.jsonl
  - README.md
  - claude-trace/go.sum
  - dissect/go.sum
  - go.work
  - go.work.sum
  - ingest/go.sum
  - mise.toml
  - release-mirror.toml
  - tak/.gitignore
  - tak/sort_test.go
  - tk/.gitignore
  - tk/sort_test.go
**TYPE:** [TO FILL]
**VERDICT:** [TO FILL]
**WHY:** [TO FILL]

