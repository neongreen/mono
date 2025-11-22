#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page()

    # Collect console messages and errors
    console_msgs = []
    errors = []
    page.on("console", lambda msg: console_msgs.append(f"[{msg.type}] {msg.text}"))
    page.on("pageerror", lambda err: errors.append(str(err)))

    page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
    page.wait_for_timeout(5000)

    # Get visible text content
    body_text = page.locator("body").text_content()

    print("=== VISIBLE TEXT ON PAGE ===")
    print(body_text[:1000] if body_text else "(no text)")

    # Look for error messages
    error_elements = page.locator("text=/error/i, text=/could not/i, text=/failed/i").all()
    if error_elements:
        print("\n=== ERROR MESSAGES FOUND ===")
        for elem in error_elements[:5]:
            try:
                print(f"- {elem.text_content()}")
            except:
                pass

    # Check for specific catalog error
    if "catalog" in body_text.lower() and ("error" in body_text.lower() or "could not" in body_text.lower()):
        print("\n=== CATALOG ERROR DETECTED ===")
        # Find the catalog error section
        catalog_section = page.locator("text=/catalog/i").first
        try:
            parent = catalog_section.locator("xpath=ancestor::div[contains(@class, 'error') or contains(@class, 'alert')]").first
            print(parent.text_content())
        except:
            print("(Could not extract full error message)")

    # Print JavaScript errors
    if errors:
        print(f"\n=== JAVASCRIPT ERRORS ({len(errors)}) ===")
        for err in errors:
            print(err)

    # Print console errors
    error_logs = [msg for msg in console_msgs if 'error' in msg.lower()]
    if error_logs:
        print(f"\n=== CONSOLE ERRORS ({len(error_logs)}) ===")
        for msg in error_logs[:10]:
            print(msg)

    browser.close()
