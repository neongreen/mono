#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = []
# ///

import subprocess
import shutil
from pathlib import Path

# Build the executable
print("Building executable...")
subprocess.run(["swift", "build", "-c", "release"], check=True)

# Define paths
build_dir = Path(".build/release")
app_bundle = Path("Tiger.app")
contents_dir = app_bundle / "Contents"
macos_dir = contents_dir / "MacOS"

# Clean up old bundle if it exists
if app_bundle.exists():
    shutil.rmtree(app_bundle)

# Create bundle structure
print("Creating app bundle...")
macos_dir.mkdir(parents=True)

# Copy executable
executable_src = build_dir / "Tiger"
executable_dst = macos_dir / "Tiger"
shutil.copy2(executable_src, executable_dst)
executable_dst.chmod(0o755)

# Copy Info.plist
shutil.copy2("Info.plist", contents_dir / "Info.plist")

print(f"✓ App bundle created at {app_bundle.absolute()}")
print(f"  Run with: open {app_bundle}")
