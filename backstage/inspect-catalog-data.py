#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Inspect actual catalog data from API responses
"""

from playwright.sync_api import sync_playwright
import json

def inspect_catalog():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Intercept API responses
        catalog_data = {}

        async def handle_route(route, request):
            # Forward the request and get the response
            response = await route.fetch()
            body = await response.body()

            # Store catalog responses
            url = request.url
            if '/api/catalog/entities' in url and 'filter=kind%3Dcomponent' in url:
                try:
                    data = json.loads(body.decode('utf-8'))
                    catalog_data['components'] = data
                except:
                    pass

            # Continue with the response
            await route.fulfill(response=response)

        # Enable request interception
        page.route("**/api/catalog/**", handle_route)

        print("=== LOADING & SIGNING IN ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        # Sign in
        guest_button = page.locator("button:has-text('Enter')").first
        if guest_button.count() > 0:
            guest_button.click()
            page.wait_for_timeout(5000)

            if 'components' in catalog_data:
                data = catalog_data['components']
                items = data.get('items', [])
                print(f"\n=== CATALOG COMPONENTS ({len(items)}) ===")
                if len(items) == 0:
                    print("  (No components in catalog)")
                    print(f"\n  Total count in response: {data.get('totalItems', '?')}")
                else:
                    for item in items:
                        kind = item.get('kind', '?')
                        name = item.get('metadata', {}).get('name', '?')
                        print(f"  - {kind}: {name}")
            else:
                print("\n=== NO COMPONENT DATA CAPTURED ===")

        browser.close()

if __name__ == "__main__":
    inspect_catalog()
