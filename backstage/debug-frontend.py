#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Comprehensive frontend debugging - collect ALL data before diagnosing.
"""

from playwright.sync_api import sync_playwright
import json

def debug_frontend():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()
        
        # Collect ALL data
        console_msgs = []
        js_errors = []
        network_requests = []
        failed_requests = []
        
        page.on("console", lambda msg: console_msgs.append({
            'type': msg.type,
            'text': msg.text,
            'location': msg.location
        }))
        
        page.on("pageerror", lambda err: js_errors.append(str(err)))
        
        def handle_request(request):
            network_requests.append({
                'url': request.url,
                'method': request.method,
                'resourceType': request.resource_type
            })
        
        def handle_response(response):
            if response.status >= 400:
                failed_requests.append({
                    'url': response.url,
                    'status': response.status,
                    'statusText': response.status_text
                })
        
        page.on("request", handle_request)
        page.on("response", handle_response)
        
        print("=== LOADING PAGE ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(5000)
        
        print("\n=== HTML STRUCTURE ===")
        html = page.content()
        print(f"Total HTML length: {len(html)} chars")
        print(f"Contains #root: {'id=\"root\"' in html}")
        
        root_html = page.locator("#root").inner_html()
        print(f"#root innerHTML length: {len(root_html)} chars")
        if root_html:
            print(f"#root content preview: {root_html[:200]}")
        
        print("\n=== CONSOLE MESSAGES ===")
        for msg in console_msgs:
            print(f"[{msg['type']}] {msg['text']}")
        
        print(f"\n=== JAVASCRIPT ERRORS ({len(js_errors)}) ===")
        for err in js_errors:
            print(f"ERROR: {err}")
        
        print(f"\n=== NETWORK REQUESTS ({len(network_requests)}) ===")
        for req in network_requests[:20]:  # First 20
            print(f"{req['method']} {req['resourceType']}: {req['url']}")
        
        print(f"\n=== FAILED REQUESTS ({len(failed_requests)}) ===")
        for req in failed_requests:
            print(f"{req['status']} {req['statusText']}: {req['url']}")
        
        # Check what scripts loaded
        scripts = page.locator("script[src]").all()
        print(f"\n=== LOADED SCRIPTS ({len(scripts)}) ===")
        for script in scripts[:10]:  # First 10
            src = script.get_attribute("src")
            print(f"  {src}")
        
        browser.close()
        
        print("\n=== DIAGNOSIS ===")
        if js_errors:
            print("❌ JavaScript errors found - this is the likely cause")
        elif failed_requests:
            print("⚠️ Network requests failed - may be causing issues")
        elif not root_html or len(root_html) == 0:
            print("❌ #root is empty - app not rendering")
            print("   Possible causes:")
            print("   1. Import error (module not found)")
            print("   2. Configuration error preventing app init")
            print("   3. Silent error during app creation")
        else:
            print("✓ App appears to be rendering")

if __name__ == "__main__":
    debug_frontend()
