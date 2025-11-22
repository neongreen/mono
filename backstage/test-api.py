#!/usr/bin/env -S uv run --script

# /// script
# requires-python = ">=3.13"
# dependencies = ["requests", "rich"]
# ///

"""
Example script to test Backstage API locally.

Usage:
    ./test-api.py              # Run all API tests
    ./test-api.py --health     # Test health endpoint only
    ./test-api.py --catalog    # Test catalog API only
"""

import argparse
import sys
from typing import Any

import requests
from rich.console import Console
from rich.panel import Panel
from rich.syntax import Syntax

console = Console()

BASE_URL = "http://localhost:7007"


def get_guest_token() -> str:
    """Get a guest authentication token."""
    console.print("[bold blue]Getting guest token...[/bold blue]")
    response = requests.post(f"{BASE_URL}/api/auth/guest/refresh")
    response.raise_for_status()
    token = response.json()["backstageIdentity"]["token"]
    console.print("[green]✓ Token obtained[/green]")
    return token


def test_health() -> bool:
    """Test the health endpoint."""
    console.print("\n[bold]Testing Health Endpoint[/bold]")
    try:
        response = requests.get(f"{BASE_URL}/healthcheck", timeout=5)
        response.raise_for_status()
        data = response.json()
        console.print(Panel(f"[green]✓ Health check passed: {data}[/green]"))
        return True
    except Exception as e:
        console.print(Panel(f"[red]✗ Health check failed: {e}[/red]"))
        return False


def test_catalog(token: str) -> bool:
    """Test the catalog API."""
    console.print("\n[bold]Testing Catalog API[/bold]")
    try:
        headers = {"Authorization": f"Bearer {token}"}
        response = requests.get(
            f"{BASE_URL}/api/catalog/entities", headers=headers, timeout=10
        )
        response.raise_for_status()
        data = response.json()

        total = len(data)
        console.print(f"[green]✓ Catalog API working - Found {total} entities[/green]")

        # Show first entity as example
        if total > 0:
            import json

            example = json.dumps(data[0], indent=2)
            syntax = Syntax(example, "json", theme="monokai", line_numbers=True)
            console.print("\n[bold]Example entity:[/bold]")
            console.print(syntax)

        return True
    except Exception as e:
        console.print(Panel(f"[red]✗ Catalog API failed: {e}[/red]"))
        return False


def test_all() -> bool:
    """Run all API tests."""
    console.print(
        Panel.fit(
            "[bold cyan]Backstage API Test Suite[/bold cyan]",
            subtitle="Testing local instance",
        )
    )

    # Test health first (no auth needed)
    if not test_health():
        console.print("[red]Server is not running or not responding[/red]")
        return False

    # Get token and test authenticated endpoints
    try:
        token = get_guest_token()
        catalog_ok = test_catalog(token)

        console.print("\n[bold]Summary[/bold]")
        console.print(f"Health: [green]✓[/green]")
        console.print(f"Catalog: {'[green]✓[/green]' if catalog_ok else '[red]✗[/red]'}")

        return catalog_ok
    except Exception as e:
        console.print(f"[red]Failed to get authentication token: {e}[/red]")
        return False


def main() -> int:
    parser = argparse.ArgumentParser(description="Test Backstage API")
    parser.add_argument("--health", action="store_true", help="Test health endpoint")
    parser.add_argument("--catalog", action="store_true", help="Test catalog API")
    args = parser.parse_args()

    try:
        if args.health:
            success = test_health()
        elif args.catalog:
            token = get_guest_token()
            success = test_catalog(token)
        else:
            success = test_all()

        return 0 if success else 1
    except KeyboardInterrupt:
        console.print("\n[yellow]Interrupted by user[/yellow]")
        return 130
    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        return 1


if __name__ == "__main__":
    sys.exit(main())
