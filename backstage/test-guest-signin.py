#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["playwright"]
# ///

"""
Test guest sign-in and catalog access
"""

from playwright.sync_api import sync_playwright

def test_guest_signin():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        context = browser.new_context()
        page = context.new_page()

        # Collect errors
        console_errors = []
        js_errors = []

        page.on("console", lambda msg: console_errors.append(msg.text) if msg.type == 'error' else None)
        page.on("pageerror", lambda err: js_errors.append(str(err)))

        print("=== LOADING PAGE ===")
        page.goto("http://localhost:3000/", wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        print("\n=== CHECKING FOR SIGN-IN PAGE ===")
        # Check if we're on the sign-in page
        if page.locator("text=Guest").count() > 0:
            print("✓ Guest sign-in option found")

            # Click the guest sign-in button
            print("\n=== CLICKING GUEST SIGN-IN ===")
            guest_button = page.locator("button:has-text('Enter')").first
            if guest_button.count() > 0:
                guest_button.click()
                print("✓ Clicked guest sign-in button")

                # Wait for navigation after signin
                page.wait_for_timeout(5000)

                # Check current URL
                current_url = page.url
                print(f"\n=== CURRENT URL ===\n{current_url}")

                # Check what's visible on the page
                print("\n=== PAGE CONTENT ===")
                visible_text = page.locator("body").text_content()
                print(f"First 500 chars: {visible_text[:500]}")

                # Check for specific content
                if "catalog" in visible_text.lower():
                    print("\n✓ Page mentions 'catalog'")

                if "error" in visible_text.lower() or "401" in visible_text:
                    print("\n❌ Error messages found on page")
                    # Find error messages
                    errors = page.locator("text=/error/i, text=/401/").all()
                    for err in errors[:3]:
                        try:
                            print(f"  - {err.text_content()}")
                        except:
                            pass
                else:
                    print("\n✓ No obvious error messages")

                # Check network requests
                print("\n=== CHECKING RECENT API CALLS ===")
                # This is a simplified check - in real scenario we'd track responses

            else:
                print("❌ Could not find Enter button")
        else:
            print("❌ Guest sign-in option not found")
            print(f"Page content: {page.locator('body').text_content()[:200]}")

        if console_errors:
            print(f"\n=== CONSOLE ERRORS ({len(console_errors)}) ===")
            for err in console_errors[:5]:
                print(f"  {err[:200]}")

        if js_errors:
            print(f"\n=== JS ERRORS ({len(js_errors)}) ===")
            for err in js_errors:
                print(f"  {err}")

        browser.close()

if __name__ == "__main__":
    test_guest_signin()
