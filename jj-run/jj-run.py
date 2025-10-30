#!/usr/bin/env python3
"""
Wrapper script for jj-run to maintain compatibility with tests.
This simply calls the main function from the jj_run package.
"""

import sys
from pathlib import Path

# Add src directory to Python path
src_dir = Path(__file__).parent / "src"
sys.path.insert(0, str(src_dir))

from jj_run.main import main

if __name__ == "__main__":
    main()
