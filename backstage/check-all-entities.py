#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Check all entities in catalog regardless of kind
"""

from playwright.sync_api import sync_playwright
import json

def check_all():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Track ALL entities API calls
        all_entities = None

        def handle_response(response):
            nonlocal all_entities
            url = response.url
            # Look for entities call without filters
            if '/api/catalog/entities/by-query' in url and 'limit=500' in url and 'filter=' not in url:
                try:
                    body = response.body()
                    data = json.loads(body)
                    if 'items' in data:
                        all_entities = data
                except Exception as e:
                    print(f"Error: {e}")

        page.on("response", handle_response)

        print("=== SIGNING IN ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        # Sign in
        guest_button = page.locator("button:has-text('Enter')").first
        if guest_button.count() > 0:
            guest_button.click()
            page.wait_for_timeout(5000)

            if all_entities:
                items = all_entities.get('items', [])
                print(f"\n=== ALL CATALOG ENTITIES ({len(items)}) ===")
                for item in items:
                    kind = item.get('kind', '?')
                    name = item.get('metadata', {}).get('name', '?')
                    namespace = item.get('metadata', {}).get('namespace', 'default')
                    print(f"  - {kind}: {namespace}/{name}")
                if len(items) == 0:
                    print("  (No entities)")
            else:
                print("\n⚠ Could not capture entities data")

        browser.close()

if __name__ == "__main__":
    check_all()
