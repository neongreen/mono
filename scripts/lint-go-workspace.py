#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = []
# ///

"""
Go workspace consistency linter.

Validates that go.work replace directives match the versions required in go.mod files.
"""

import argparse
import re
import sys
from pathlib import Path
from typing import NamedTuple


class ReplaceDirective(NamedTuple):
    module: str
    version: str
    path: str


class Requirement(NamedTuple):
    module: str
    version: str
    source_file: Path


def parse_go_work(work_file: Path) -> list[ReplaceDirective]:
    """Parse replace directives from go.work file."""
    replaces = []
    content = work_file.read_text()

    # Match both single-line and block replace directives
    # Single-line: replace module version => path
    pattern = r"replace\s+(\S+)\s+(\S+)\s+=>\s+(\S+)"
    for match in re.finditer(pattern, content):
        module, version, path = match.groups()
        replaces.append(ReplaceDirective(module, version, path))

    # Block format: replace ( ... )
    # Match individual lines within replace blocks
    block_pattern = r"^\s+(\S+)\s+(\S+)\s+=>\s+(\S+)"
    for line in content.splitlines():
        match = re.match(block_pattern, line)
        if match:
            module, version, path = match.groups()
            replaces.append(ReplaceDirective(module, version, path))

    return replaces


def parse_go_mod(mod_file: Path) -> list[Requirement]:
    """Parse requirements from go.mod file."""
    requirements = []
    content = mod_file.read_text()

    # Match: module version in require blocks
    # Both single-line and multi-line require blocks
    in_require = False
    for line in content.splitlines():
        line = line.strip()

        if line.startswith("require ("):
            in_require = True
            continue
        elif line == ")":
            in_require = False
            continue

        # Single-line require
        if line.startswith("require "):
            line = line[8:]  # Remove "require "
            parts = line.split()
            if len(parts) >= 2:
                module, version = parts[0], parts[1]
                requirements.append(Requirement(module, version, mod_file))
        # Multi-line require
        elif in_require and line and not line.startswith("//"):
            parts = line.split()
            if len(parts) >= 2:
                module, version = parts[0], parts[1]
                requirements.append(Requirement(module, version, mod_file))

    return requirements


def fix_go_mod_version(mod_file: Path, module: str, old_version: str, new_version: str) -> None:
    """Fix version mismatch in a go.mod file."""
    content = mod_file.read_text()

    # Replace the version in the file
    # Handle both single-line and multi-line require blocks
    pattern = re.compile(
        rf"({re.escape(module)})\s+{re.escape(old_version)}\b"
    )

    new_content = pattern.sub(rf"\1 {new_version}", content)

    if new_content != content:
        mod_file.write_text(new_content)
        print(f"Fixed {mod_file}: {module} {old_version} -> {new_version}")


def main() -> int:
    """Main linter logic."""
    parser = argparse.ArgumentParser(
        description="Lint and fix Go workspace consistency"
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Automatically fix version mismatches",
    )
    args = parser.parse_args()

    repo_root = Path(__file__).parent.parent
    go_work = repo_root / "go.work"

    if not go_work.exists():
        print("Error: go.work not found", file=sys.stderr)
        return 1

    # Parse go.work replace directives
    replaces = parse_go_work(go_work)
    replace_map = {(r.module, r.version): r.path for r in replaces}

    # Create module -> version map for easier lookup
    module_to_version = {m: v for (m, v), _ in replace_map.items()}

    # Find all go.mod files
    go_mods = list(repo_root.rglob("go.mod"))

    # Filter out go.work.sum location if it exists
    go_mods = [f for f in go_mods if f != go_work]

    errors = []
    fixes = []

    # Check each go.mod file
    for mod_file in go_mods:
        requirements = parse_go_mod(mod_file)

        # Filter to only local monorepo requirements
        local_reqs = [
            r for r in requirements
            if r.module.startswith("github.com/neongreen/mono/")
        ]

        for req in local_reqs:
            # Check if there's a matching replace directive
            if (req.module, req.version) not in replace_map:
                # Check if there's a replace with different version
                matching_module_replaces = [
                    (v, p) for (m, v), p in replace_map.items() if m == req.module
                ]

                if matching_module_replaces:
                    for replace_version, replace_path in matching_module_replaces:
                        error_msg = (
                            f"{req.source_file.relative_to(repo_root)}: "
                            f"requires {req.module} {req.version}, "
                            f"but go.work replaces {replace_version}"
                        )
                        errors.append(error_msg)

                        if args.fix:
                            fixes.append((req.source_file, req.module, req.version, replace_version))
                else:
                    errors.append(
                        f"{req.source_file.relative_to(repo_root)}: "
                        f"requires {req.module} {req.version}, "
                        f"but no replace directive found in go.work"
                    )

    if args.fix and fixes:
        print("Fixing version mismatches:\n")
        for mod_file, module, old_version, new_version in fixes:
            fix_go_mod_version(mod_file, module, old_version, new_version)
        print("\nAll fixes applied. Re-running validation...\n")

        # Re-validate
        return main()

    if errors:
        print("Go workspace consistency errors found:\n", file=sys.stderr)
        for error in errors:
            print(f"  {error}", file=sys.stderr)
        print(
            "\nRun with --fix to automatically fix these issues.",
            file=sys.stderr
        )
        return 1

    print("Go workspace is consistent")
    return 0


if __name__ == "__main__":
    sys.exit(main())
