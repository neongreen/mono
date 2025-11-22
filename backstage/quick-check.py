#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page()

    messages = []
    page.on("console", lambda msg: messages.append(f"[{msg.type}] {msg.text}"))

    page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
    page.wait_for_timeout(5000)

    print("Console messages:")
    for msg in messages:
        if '[DEBUG]' in msg:
            print(msg)

    root_html = page.locator("#root").inner_html()
    print(f"\n#root length: {len(root_html)} chars")

    browser.close()
