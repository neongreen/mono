#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
from pathlib import Path


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


def git_output(*args: str) -> str:
    result = subprocess.run(
        ["git", *args], check=True, text=True, capture_output=True
    )
    return result.stdout.strip()


def resolve_changed_files(event: str, before_sha: str, pr_base: str, ref_sha: str) -> list[str]:
    if event == "pull_request" and pr_base:
        diff_args = ["diff", "--name-only", f"origin/{pr_base}...HEAD"]
    else:
        if not before_sha or before_sha == "0" * 40:
            diff_args = ["diff-tree", "--no-commit-id", "--name-only", "-r", ref_sha]
        else:
            diff_args = ["diff", "--name-only", f"{before_sha}...{ref_sha}"]
    changed = git_output(*diff_args)
    return [line for line in changed.splitlines() if line]


def select_projects(
    event: str,
    candidates: list[str],
    manual: str,
    changed_files: list[str],
) -> list[str]:
    selected: list[str] = []

    if event == "workflow_dispatch" and manual.strip():
        chunks = manual.replace(",", "\n").splitlines()
        for chunk in chunks:
            name = chunk.strip()
            if not name:
                continue
            if name in candidates:
                selected.append(name)
            else:
                print(f"::warning::Requested project '{name}' not found or invalid")
        return selected

    for project in candidates:
        prefix = f"{project}/"
        if any(path.startswith(prefix) for path in changed_files):
            selected.append(project)
    return selected


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--event", required=True)
    parser.add_argument("--before-sha", default="")
    parser.add_argument("--pr-base", default="")
    parser.add_argument("--ref-sha", required=True)
    parser.add_argument("--manual-projects", default="")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    projects = list_projects()
    print(f"All Go projects: {', '.join(projects)}")

    changed_files = resolve_changed_files(
        args.event, args.before_sha, args.pr_base, args.ref_sha
    )
    if changed_files:
        print("Changed files:")
        print("\n".join(changed_files))

    selected = select_projects(
        args.event, projects, args.manual_projects, changed_files
    )

    with open(args.output, "a", encoding="utf-8") as fh:
        if selected:
            print(f"Selected projects: {', '.join(selected)}")
            fh.write("has_projects=true\n")
            fh.write(f"projects={json.dumps(selected)}\n")
        else:
            print("No projects selected for release")
            fh.write("has_projects=false\n")
            fh.write("projects=[]\n")


if __name__ == "__main__":
    main()
