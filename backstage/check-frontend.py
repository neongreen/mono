#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""Check what actually renders in the browser."""

from playwright.sync_api import sync_playwright
import sys

def check_frontend():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        
        # Collect console messages
        console_messages = []
        def handle_console(msg):
            console_messages.append(f"[{msg.type}] {msg.text}")
        page.on("console", handle_console)
        
        # Collect errors
        errors = []
        def handle_error(error):
            errors.append(str(error))
        page.on("pageerror", handle_error)
        
        print("Opening http://localhost:3000/...")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        
        # Wait a bit for React to render
        page.wait_for_timeout(3000)
        
        # Take screenshot
        page.screenshot(path="backstage-screenshot.png")
        print("✓ Screenshot saved to backstage-screenshot.png")
        
        # Check page title
        title = page.title()
        print(f"Page title: {title}")
        
        # Check body content
        body_text = page.locator("body").text_content()
        if body_text:
            print(f"Body has text content: {len(body_text)} characters")
            if body_text.strip():
                print(f"First 200 chars: {body_text[:200]}")
        else:
            print("Body appears empty")
        
        # Check for specific Backstage elements
        has_root = page.locator("#root").count() > 0
        print(f"Has #root div: {has_root}")
        
        if has_root:
            root_content = page.locator("#root").text_content()
            if root_content and root_content.strip():
                print(f"#root has content: {len(root_content)} characters")
            else:
                print("#root exists but appears empty (gray screen)")
        
        # Print console messages
        if console_messages:
            print(f"\nConsole messages ({len(console_messages)}):")
            for msg in console_messages[:10]:  # First 10
                print(f"  {msg}")
        
        # Print errors
        if errors:
            print(f"\nJavaScript errors ({len(errors)}):")
            for error in errors:
                print(f"  {error}")
        
        browser.close()
        
        # Return status
        if errors:
            print("\n❌ Found JavaScript errors")
            return 1
        elif not has_root or not root_content or not root_content.strip():
            print("\n❌ Gray screen detected - #root is empty")
            return 1
        else:
            print("\n✓ Page appears to have rendered content")
            return 0

if __name__ == "__main__":
    sys.exit(check_frontend())
