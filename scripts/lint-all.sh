#!/usr/bin/env bash
# Comprehensive linting script for the mono repository
# Runs all linters configured in mise.toml and .dagger/runlint.go
#
# Usage:
#   ./scripts/lint-all.sh              # Run all linters
#   ./scripts/lint-all.sh --go-only    # Run only Go linters
#   ./scripts/lint-all.sh --actions    # Run only GitHub Actions linters

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Flags
RUN_GO_LINTERS=true
RUN_ACTIONS_LINTERS=true
FAILED_LINTERS=()

# Parse arguments
for arg in "$@"; do
    case $arg in
        --go-only)
            RUN_ACTIONS_LINTERS=false
            ;;
        --actions)
            RUN_GO_LINTERS=false
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --go-only    Run only Go linters"
            echo "  --actions    Run only GitHub Actions linters"
            echo "  --help       Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $arg"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Change to repository root
cd "$(dirname "$0")/.."

echo -e "${BLUE}=== Mono Repository Linting ===${NC}"
echo ""

# Helper function to run a linter and track failures
run_linter() {
    local name="$1"
    shift
    echo -e "${YELLOW}Running: ${name}${NC}"
    if "$@"; then
        echo -e "${GREEN}✓ ${name} passed${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ ${name} failed${NC}"
        echo ""
        FAILED_LINTERS+=("$name")
        return 1
    fi
}

# Go Linters
if [ "$RUN_GO_LINTERS" = true ]; then
    echo -e "${BLUE}--- Go Linters ---${NC}"
    echo ""

    # 1. golangci-lint with config
    # Uses --new-from-rev to grandfather existing gosec violations at commit 0c23a5a5
    # (when gosec was first enabled). See tk-330 for details.
    run_linter "golangci-lint" \
        golangci-lint run --new-from-rev=0c23a5a5 || true

    # 2. go vet
    run_linter "go vet" \
        go vet ./... || true

    # 3. Custom linter: uselesswrapper
    run_linter "uselesswrapper" \
        go run ./linters/uselesswrapper/cmd/uselesswrapper ./... || true

    # 4. Custom linter: cobralint (tk-specific)
    run_linter "cobralint (tk)" \
        go run ./linters/cobralint/cmd/cobralint ./tk/... || true

    # 5. Custom linter: dircheck
    run_linter "dircheck" \
        go run ./linters/dircheck/cmd/dircheck ./... || true

    # 6. Go module tidiness check
    echo -e "${YELLOW}Running: go mod tidy check${NC}"
    go mod tidy
    if git diff --exit-code go.mod go.sum; then
        echo -e "${GREEN}✓ go mod tidy check passed${NC}"
        echo ""
    else
        echo -e "${RED}✗ go mod tidy check failed - go.mod or go.sum needs tidying${NC}"
        echo ""
        FAILED_LINTERS+=("go mod tidy")
    fi

    # 7. Lint .dagger module
    echo -e "${YELLOW}Running: golangci-lint (.dagger)${NC}"
    if (cd .dagger && golangci-lint run --new-from-rev=0c23a5a5); then
        echo -e "${GREEN}✓ golangci-lint (.dagger) passed${NC}"
        echo ""
    else
        echo -e "${RED}✗ golangci-lint (.dagger) failed${NC}"
        echo ""
        FAILED_LINTERS+=("golangci-lint (.dagger)")
    fi
fi

# GitHub Actions Linters
if [ "$RUN_ACTIONS_LINTERS" = true ]; then
    echo -e "${BLUE}--- GitHub Actions Linters ---${NC}"
    echo ""

    # 1. actionlint
    if command -v actionlint &> /dev/null; then
        run_linter "actionlint" \
            actionlint || true
    else
        echo -e "${YELLOW}⚠ actionlint not found, skipping${NC}"
        echo ""
    fi

    # 2. pinact (requires GITHUB_TOKEN)
    if command -v pinact &> /dev/null; then
        if gh auth token &> /dev/null; then
            run_linter "pinact" \
                env GITHUB_TOKEN="$(gh auth token)" pinact run --check --diff --verify || true
        else
            echo -e "${YELLOW}⚠ GitHub token not available (gh auth token failed), skipping pinact${NC}"
            echo ""
        fi
    else
        echo -e "${YELLOW}⚠ pinact not found, skipping${NC}"
        echo ""
    fi
fi

# Summary
echo -e "${BLUE}=== Linting Summary ===${NC}"
echo ""

if [ ${#FAILED_LINTERS[@]} -eq 0 ]; then
    echo -e "${GREEN}All linters passed!${NC}"
    exit 0
else
    echo -e "${RED}${#FAILED_LINTERS[@]} linter(s) failed:${NC}"
    for linter in "${FAILED_LINTERS[@]}"; do
        echo -e "${RED}  - $linter${NC}"
    done
    exit 1
fi
