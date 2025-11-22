#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Test catalog API calls after guest sign-in
"""

from playwright.sync_api import sync_playwright

def test_catalog_api():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Track network requests
        api_calls = []
        failed_calls = []

        def handle_response(response):
            url = response.url
            if '/api/catalog/' in url:
                api_calls.append({
                    'url': url,
                    'status': response.status,
                    'method': response.request.method
                })
                if response.status >= 400:
                    failed_calls.append({
                        'url': url,
                        'status': response.status,
                        'statusText': response.status_text
                    })

        page.on("response", handle_response)

        print("=== LOADING PAGE & SIGNING IN ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        # Click guest sign-in
        guest_button = page.locator("button:has-text('Enter')").first
        if guest_button.count() > 0:
            guest_button.click()
            print("✓ Signed in as guest")

            # Wait for catalog page to load
            page.wait_for_timeout(5000)

            print(f"\n=== CATALOG API CALLS ({len(api_calls)}) ===")
            for call in api_calls:
                status_symbol = "✓" if call['status'] < 400 else "❌"
                print(f"{status_symbol} {call['method']} {call['status']} {call['url']}")

            if failed_calls:
                print(f"\n=== FAILED API CALLS ({len(failed_calls)}) ===")
                for call in failed_calls:
                    print(f"❌ {call['status']} {call['statusText']}: {call['url']}")
            else:
                print("\n✓ No failed API calls")

            # Check if we have any successful catalog calls
            successful_catalog_calls = [c for c in api_calls if c['status'] < 400]
            if successful_catalog_calls:
                print(f"\n✓ {len(successful_catalog_calls)} successful catalog API calls")
            else:
                print("\n❌ No successful catalog API calls")

        else:
            print("❌ Could not find guest sign-in button")

        browser.close()

if __name__ == "__main__":
    test_catalog_api()
