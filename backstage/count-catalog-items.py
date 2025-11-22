#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Count catalog items from API response
"""

from playwright.sync_api import sync_playwright
import json

def count_items():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Track responses
        component_count = None

        def handle_response(response):
            nonlocal component_count
            url = response.url
            # Look for the main components list API call
            if '/api/catalog/entities' in url and 'filter=kind%3Dcomponent' in url and 'facets' not in url:
                try:
                    body = response.body()
                    data = json.loads(body)
                    if 'items' in data:
                        component_count = len(data['items'])
                        print(f"\n=== COMPONENT API RESPONSE ===")
                        print(f"Total items: {data.get('totalItems', 'N/A')}")
                        print(f"Items in response: {len(data['items'])}")
                        if len(data['items']) > 0:
                            print("\nComponents:")
                            for item in data['items'][:5]:  # First 5
                                name = item.get('metadata', {}).get('name', '?')
                                print(f"  - {name}")
                except Exception as e:
                    print(f"Error parsing response: {e}")

        page.on("response", handle_response)

        print("=== LOADING & SIGNING IN ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        # Sign in
        guest_button = page.locator("button:has-text('Enter')").first
        if guest_button.count() > 0:
            guest_button.click()
            page.wait_for_timeout(5000)

            if component_count is not None:
                if component_count == 0:
                    print("\n✓ Catalog is working but empty (0 components)")
                else:
                    print(f"\n✓ Found {component_count} components in catalog")
            else:
                print("\n⚠ Could not capture component count")

        browser.close()

if __name__ == "__main__":
    count_items()
