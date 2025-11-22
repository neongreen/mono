#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Check what entities are in the catalog
"""

from playwright.sync_api import sync_playwright
import json

def check_entities():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Track the entities response
        entities_response = None

        def handle_response(response):
            nonlocal entities_response
            if '/api/catalog/entities' in response.url and 'filter=kind' not in response.url:
                try:
                    entities_response = response.json()
                except:
                    pass

        page.on("response", handle_response)

        print("=== SIGNING IN ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        # Sign in as guest
        guest_button = page.locator("button:has-text('Enter')").first
        if guest_button.count() > 0:
            guest_button.click()
            page.wait_for_timeout(5000)

        # Make a direct API call to get all entities
        page.goto("http://localhost:3000/api/catalog/entities", wait_until="networkidle")
        page.wait_for_timeout(2000)

        body_text = page.locator("body").text_content()
        try:
            data = json.loads(body_text)
            if isinstance(data, dict) and 'items' in data:
                print(f"\n=== ENTITIES IN CATALOG ({len(data['items'])}) ===")
                for item in data['items']:
                    kind = item.get('kind', '?')
                    name = item.get('metadata', {}).get('name', '?')
                    namespace = item.get('metadata', {}).get('namespace', 'default')
                    print(f"  - {kind}: {namespace}/{name}")

                if len(data['items']) == 0:
                    print("  (No entities loaded)")
            else:
                print(f"\n=== RAW RESPONSE ===\n{body_text[:500]}")
        except:
            print(f"\n=== RAW RESPONSE ===\n{body_text[:500]}")

        browser.close()

if __name__ == "__main__":
    check_entities()
