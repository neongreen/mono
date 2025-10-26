#!/usr/bin/env python3
"""
Automate migration from a multi-module Go workspace to a single root-level Go
module. The script expects a `go-projects.toml` file at the repository root
describing each existing Go module (directory plus current module path).

Steps performed:
1. Validate git state unless `--allow-dirty` is passed.
2. Load project metadata from `go-projects.toml`.
3. Delete `go.work`, `go.work.sum`, and every per-project `go.mod`/`go.sum`.
4. Initialise `go.mod` at the repository root.
5. Rewrite Go imports whose module path will change (e.g., `conf` → `github.com/neongreen/mono/conf`).
6. Update tooling/CI:
   * rewrite root `mise.toml` tasks that referenced `go.work`;
   * drop the legacy workspace lint script;
   * update `scripts/detect_go_projects.py` for the single-module layout;
   * adjust GitHub Actions workflows to use the root `go.sum` and add a new module tidy check.
7. Run `go fmt ./...` and `go mod tidy` (unless `--skip-go-commands`).

The script is idempotent for the pre-migration state: re-running after
resetting the repository to the original workspace layout will reapply the
same edits. It is not intended for partial post-migration runs.
"""

from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys
import textwrap
from pathlib import Path

try:
    import tomllib  # type: ignore
except ModuleNotFoundError:
    try:
        import tomli as tomllib  # type: ignore
    except ModuleNotFoundError as exc:
        print(
            "Error: Python 3.11+ (tomllib) or the tomli package is required to parse go-projects.toml.",
            file=sys.stderr,
        )
        raise SystemExit(1) from exc


ROOT = Path(__file__).resolve().parent.parent
ROOT_MODULE = "github.com/neongreen/mono"
GO_VERSION = "1.24.7"
GO_WORK_PATH = ROOT / "go.work"
GO_WORK_SUM_PATH = ROOT / "go.work.sum"
GO_PROJECTS_TOML = ROOT / "go-projects.toml"


def run_command(cmd, *, cwd=None):
    """Run a command and stream its output."""
    cwd = cwd or ROOT
    print(f"+ {' '.join(cmd)}")
    env = os.environ.copy()
    # Ensure Go never tries to rehydrate a workspace automatically.
    env.setdefault("GOWORK", "off")
    subprocess.run(cmd, cwd=cwd, env=env, check=True)


def ensure_clean_worktree(allow_dirty: bool) -> None:
    """Abort if the git worktree is dirty (unless explicitly allowed)."""
    if allow_dirty:
        return
    result = subprocess.run(
        ["git", "status", "--porcelain"],
        cwd=ROOT,
        check=True,
        text=True,
        capture_output=True,
    )
    if result.stdout.strip():
        print("Error: working tree is not clean. Commit/stash changes or pass --allow-dirty.", file=sys.stderr)
        sys.exit(1)


def load_projects_config():
    """Read go-projects.toml and return a list of (directory, module path)."""
    if not GO_PROJECTS_TOML.exists():
        print(f"Error: {GO_PROJECTS_TOML} not found. Please create it with project definitions.", file=sys.stderr)
        sys.exit(1)

    data = tomllib.loads(GO_PROJECTS_TOML.read_text(encoding="utf-8"))
    entries = []
    if "project" in data:
        entries = data["project"]
    elif "projects" in data:
        entries = data["projects"]

    if not entries:
        print(f"Error: no project entries found in {GO_PROJECTS_TOML}", file=sys.stderr)
        sys.exit(1)

    result = []
    for idx, entry in enumerate(entries, start=1):
        dir_name = entry.get("dir")
        module_path = entry.get("module")
        if not dir_name or not module_path:
            print(f"Error: project entry #{idx} in {GO_PROJECTS_TOML} must specify 'dir' and 'module'", file=sys.stderr)
            sys.exit(1)
        result.append((Path(dir_name), module_path))
    return result


def read_module_path(go_mod_path: Path) -> str:
    """Extract the module path from a go.mod file."""
    for raw_line in go_mod_path.read_text().splitlines():
        line = raw_line.strip()
        if line.startswith("module "):
            return line.split()[1]
    raise RuntimeError(f"Module path not found in {go_mod_path}")


def delete_workspace_files(module_dirs):
    """Delete go.work, go.work.sum, and per-module go.mod/go.sum files."""
    for path in (GO_WORK_PATH, GO_WORK_SUM_PATH):
        if path.exists():
            print(f"Removing {path}")
            path.unlink()

    for module_dir in module_dirs:
        go_mod = ROOT / module_dir / "go.mod"
        go_sum = ROOT / module_dir / "go.sum"
        if go_mod.exists():
            print(f"Removing {go_mod}")
            go_mod.unlink()
        if go_sum.exists():
            print(f"Removing {go_sum}")
            go_sum.unlink()


def rewrite_imports(import_map):
    """Rewrite Go imports that reference short module paths."""
    if not import_map:
        return

    go_files = [p for p in ROOT.rglob("*.go") if "vendor" not in p.parts]
    for path in go_files:
        original = path.read_text()
        lines = original.splitlines(keepends=True)
        in_block = False
        changed = False

        for idx, line in enumerate(lines):
            stripped = line.strip()
            if stripped.startswith("import "):
                if stripped == "import (":
                    in_block = True
                    continue
                else:
                    new_line = _rewrite_import_line(line, import_map)
                    if new_line != line:
                        lines[idx] = new_line
                        changed = True
            elif in_block:
                if stripped == ")":
                    in_block = False
                    continue
                new_line = _rewrite_import_line(line, import_map)
                if new_line != line:
                    lines[idx] = new_line
                    changed = True

        if changed:
            print(f"Rewriting imports in {path}")
            path.write_text("".join(lines))


def _rewrite_import_line(line, import_map):
    new_line = line
    for old, new in import_map.items():
        token = f'"{old}'
        if token in new_line:
            new_line = new_line.replace(token, f'"{new}')
    return new_line


def initialise_root_module(go_version: str) -> None:
    """Create go.mod in the repository root."""
    root_go_mod = ROOT / "go.mod"
    if root_go_mod.exists():
        print("Error: go.mod already exists at repository root. Aborting.", file=sys.stderr)
        sys.exit(1)

    run_command(["go", "mod", "init", ROOT_MODULE])

    content = root_go_mod.read_text()
    if f"go {go_version}" not in content:
        content = re.sub(r"go \d+\.\d+(?:\.\d+)?", f"go {go_version}", content)
        root_go_mod.write_text(content)


def update_mise_tasks() -> None:
    """Adjust root mise.toml tasks for the single-module setup."""
    mise_path = ROOT / "mise.toml"
    content = mise_path.read_text()

    # Update fmt-all run block.
    fmt_all_pattern = re.compile(
        r'(\[tasks\.fmt-all\]\n(?:[^\[]*?)run = """\n)(.*?)(\n""" ?)',
        re.DOTALL,
    )
    new_fmt_block = "mise run //...:fmt\nmise run //:go:fmt"
    content, count = fmt_all_pattern.subn(
        r"\1" + new_fmt_block + r"\3", content, count=1
    )
    if count == 0:
        raise RuntimeError("Unable to rewrite fmt-all task in mise.toml")

    # Remove old workspace lint tasks.
    content = re.sub(
        r'\n\[tasks\.\"lint:go-workspace(?::ci)?\"\][\s\S]*?(?=\n\[tasks|\Z)',
        "\n",
        content,
    )

    # Append new tasks if they are missing.
    additions = []
    if 'tasks."go:fmt"' not in content:
        additions.append(
            textwrap.dedent(
                """
                [tasks."go:fmt"]
                description = "Format Go code in the unified module"
                run = "go fmt ./..."
                """
            ).strip()
        )
    if 'tasks."go:tidy"' not in content:
        additions.append(
            textwrap.dedent(
                """
                [tasks."go:tidy"]
                description = "Synchronize Go dependencies"
                run = "go mod tidy"
                """
            ).strip()
        )
    if 'tasks."lint:go-module"' not in content:
        lint_block = (
            '[tasks."lint:go-module"]\n'
            'description = "Ensure go.mod and go.sum are tidy"\n'
            'run = """\n'
            'go mod tidy\n'
            'git diff --exit-code go.mod go.sum\n'
            '"""'
        )
        additions.append(lint_block)

    if additions:
        content = content.rstrip() + "\n\n" + "\n\n".join(additions) + "\n"

    mise_path.write_text(content)



def update_detect_script() -> None:
    """Update scripts/detect_go_projects.py to the single-module logic."""
    path = ROOT / "scripts" / "detect_go_projects.py"
    content = path.read_text()

    new_function = textwrap.dedent(
        """
        def list_projects() -> list[str]:
            projects: list[str] = []
            root = Path(".")
            for entry in sorted(root.iterdir()):
                if not entry.is_dir():
                    continue
                name = entry.name
                if name.startswith("."):
                    continue
                if name in {"diagram-dsl", "mdbook-comments", "scripts", "lib"}:
                    continue

                has_main = False
                for go_file in entry.rglob("*.go"):
                    if go_file.name.endswith("_test.go"):
                        continue
                    try:
                        for line in go_file.open("r", encoding="utf-8"):
                            if line.strip().startswith("package main"):
                                has_main = True
                                break
                    except UnicodeDecodeError:
                        continue
                    if has_main:
                        projects.append(name)
                        break
            return projects
        """
    ).strip()

    content = re.sub(
        r"def list_projects\(\) -> list\[str\]:[\s\S]*?return sorted\(projects\)",
        new_function,
        content,
        count=1,
    )
    path.write_text(content)


def update_workflows() -> None:
    """Adjust GitHub workflows for the single-module setup."""
    workflow_dir = ROOT / ".github" / "workflows"
    for path in workflow_dir.glob("*.yml"):
        original = path.read_text()
        updated = re.sub(
            r"(cache-dependency-path:\s+)[^\s]+/go\.sum",
            r"\1go.sum",
            original,
        )
        if updated != original:
            print(f"Updating cache path in {path}")
            path.write_text(updated)

    # Replace the workspace lint workflow with a go mod tidy check.
    old_name = workflow_dir / "go-workspace-lint.yml"
    new_name = workflow_dir / "go-module-lint.yml"
    if old_name.exists():
        print(f"Renaming {old_name} to {new_name.name}")
        old_name.rename(new_name)

    tidy_workflow = textwrap.dedent(
        """
        name: Go Module Lint

        on:
          push:
            paths:
              - "go.mod"
              - "go.sum"
              - "**/*.go"
              - ".github/workflows/go-module-lint.yml"
          pull_request:
            paths:
              - "go.mod"
              - "go.sum"
              - "**/*.go"
              - ".github/workflows/go-module-lint.yml"

        jobs:
          tidy:
            runs-on: ubuntu-latest
            steps:
              - name: Checkout code
                uses: actions/checkout@08c6903cd8c0fde910a37f88322edcfb5dd907a8 # v5.0.0
                with:
                  persist-credentials: false

              - name: Set up Go
                uses: actions/setup-go@44694675825211faa026b3c33043df3e48a5fa00 # v6.0.0
                with:
                  go-version: "1.24.7"
                  cache-dependency-path: go.sum

              - name: Check tidy state
                run: |
                  go mod tidy
                  git diff --exit-code go.mod go.sum
        """
    ).strip() + "\n"
    new_name.write_text(tidy_workflow)

    # Remove the old workspace linter script if it still exists.
    legacy_linter = ROOT / "scripts" / "lint-go-workspace.py"
    if legacy_linter.exists():
        print(f"Removing {legacy_linter}")
        legacy_linter.unlink()


def run_go_fmt_and_tidy() -> None:
    """Format Go code and tidy dependencies."""
    run_command(["go", "fmt", "./..."])
    run_command(["go", "mod", "tidy"])


def migrate(allow_dirty: bool, skip_go_commands: bool) -> None:
    ensure_clean_worktree(allow_dirty)
    project_entries = load_projects_config()

    module_info = []
    for module_dir, declared_module in project_entries:
        go_mod_path = ROOT / module_dir / "go.mod"
        module_path = declared_module
        if go_mod_path.exists():
            actual_module = read_module_path(go_mod_path)
            if actual_module != module_path:
                print(
                    f"Warning: {go_mod_path} declares {actual_module}, "
                    f"but {GO_PROJECTS_TOML} lists {module_path}. Using {actual_module}."
                )
                module_path = actual_module
        else:
            print(f"Warning: {go_mod_path} not found; continuing with configured module path {module_path}.")
        module_info.append((module_dir, module_path))

    known_dirs = {entry[0] for entry in module_info}
    extra_go_mods = []
    for go_mod in ROOT.rglob("go.mod"):
        try:
            rel_dir = go_mod.parent.relative_to(ROOT)
        except ValueError:
            continue
        if rel_dir == Path('.'):
            continue
        if rel_dir not in known_dirs:
            extra_go_mods.append(rel_dir.as_posix())
    if extra_go_mods:
        print("Warning: go.mod files outside go-projects.toml will be left untouched:")
        for rel in sorted(set(extra_go_mods)):
            print(f"  - {rel}")

    # Determine which module paths need rewriting.
    import_map = {}
    for module_dir, module_path in module_info:
        new_path = f"{ROOT_MODULE}/{module_dir.as_posix()}"
        if module_path == new_path:
            continue
        import_map[module_path] = new_path

    delete_workspace_files([info[0] for info in module_info])
    initialise_root_module(GO_VERSION)
    rewrite_imports(import_map)
    update_mise_tasks()
    update_detect_script()
    update_workflows()

    if not skip_go_commands:
        run_go_fmt_and_tidy()

    print("Migration complete.")


def main() -> None:
    parser = argparse.ArgumentParser(description="Migrate to a single Go module.")
    parser.add_argument(
        "--allow-dirty",
        action="store_true",
        help="Allow running with a dirty working tree.",
    )
    parser.add_argument(
        "--skip-go-commands",
        action="store_true",
        help="Skip go fmt and go mod tidy (useful for debugging).",
    )
    args = parser.parse_args()

    migrate(allow_dirty=args.allow_dirty, skip_go_commands=args.skip_go_commands)


if __name__ == "__main__":
    main()
