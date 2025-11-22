#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

from playwright.sync_api import sync_playwright
import sys

with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page()
    
    try:
        print("Trying to access http://localhost:3000/ with no server running...")
        page.goto("http://localhost:3000/", timeout=5000)
        print("Page loaded somehow!")
    except Exception as e:
        print(f"Expected error: {type(e).__name__}")
        print("Good - nothing is serving on port 3000")
    
    browser.close()
