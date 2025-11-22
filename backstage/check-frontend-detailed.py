#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""Detailed check of what's happening in the browser."""

from playwright.sync_api import sync_playwright
import sys

def check_frontend():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()
        
        # Collect ALL console messages
        console_messages = []
        def handle_console(msg):
            console_messages.append({
                'type': msg.type,
                'text': msg.text,
                'location': f"{msg.location.get('url', '')}:{msg.location.get('lineNumber', '')}"
            })
        page.on("console", handle_console)
        
        # Collect errors
        errors = []
        def handle_error(error):
            errors.append(str(error))
        page.on("pageerror", handle_error)
        
        # Collect network failures
        failed_requests = []
        def handle_response(response):
            if response.status >= 400:
                failed_requests.append({
                    'url': response.url,
                    'status': response.status,
                    'statusText': response.status_text
                })
        page.on("response", handle_response)
        
        print("Opening http://localhost:3000/...")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        
        # Wait longer for React
        print("Waiting 5 seconds for React to render...")
        page.wait_for_timeout(5000)
        
        # Check for loading indicators
        loading_indicators = page.locator("text=/loading/i").count()
        print(f"Loading indicators found: {loading_indicators}")
        
        # Check if there's any error text
        error_text = page.locator("text=/error/i").count()
        print(f"Error text found: {error_text}")
        
        # Get root content
        root_html = page.locator("#root").inner_html()
        print(f"\n#root HTML ({len(root_html)} chars):")
        print(root_html[:500] if root_html else "(empty)")
        
        # Print ALL console messages
        print(f"\n=== ALL Console Messages ({len(console_messages)}) ===")
        for msg in console_messages:
            print(f"[{msg['type']}] {msg['text']}")
            if msg['location']:
                print(f"    at {msg['location']}")
        
        # Print failed requests
        if failed_requests:
            print(f"\n=== Failed Requests ({len(failed_requests)}) ===")
            for req in failed_requests:
                print(f"  {req['status']} {req['statusText']}: {req['url']}")
        
        # Print errors
        if errors:
            print(f"\n=== JavaScript Errors ({len(errors)}) ===")
            for error in errors:
                print(f"  {error}")
        
        browser.close()
        
        return 1 if not root_html or not root_html.strip() else 0

if __name__ == "__main__":
    sys.exit(check_frontend())
