#!/bin/bash
set -e

# Install script for Go tools from neongreen/mono
# Usage: ./install.sh <project> [version]
#   Example: ./install.sh dissect main.1
#   Example: ./install.sh markdown-format (installs latest)

REPO="neongreen/mono"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_info() {
    echo -e "${YELLOW}$1${NC}"
}

# Check arguments
if [ $# -lt 1 ]; then
    print_error "Usage: $0 <project> [version]"
    echo ""
    echo "Available projects: dissect, markdown-format"
    echo ""
    echo "Examples:"
    echo "  $0 dissect main.5          # Install specific version"
    echo "  $0 dissect pr-42.1         # Install PR version"
    echo "  $0 markdown-format         # Install latest main"
    exit 1
fi

PROJECT=$1
VERSION=${2:-}

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux)
        OS_NAME="linux"
        ;;
    darwin)
        OS_NAME="darwin"
        ;;
    mingw*|msys*|cygwin*)
        OS_NAME="windows"
        ;;
    *)
        print_error "Unsupported operating system: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH_NAME="amd64"
        ;;
    aarch64|arm64)
        ARCH_NAME="arm64"
        ;;
    *)
        print_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# If no version specified, try to get the latest from main
if [ -z "$VERSION" ]; then
    print_info "Finding latest version of $PROJECT from main branch..."
    
    # Use GitHub API to find latest release for this project
    API_URL="https://api.github.com/repos/$REPO/releases"
    
    # Find the latest main release for this project
    LATEST_TAG=$(curl -s "$API_URL" | grep -o "\"tag_name\": \"$PROJECT/main\.[0-9]*\"" | head -1 | sed 's/.*"\(.*\)".*/\1/')
    
    if [ -z "$LATEST_TAG" ]; then
        print_error "Could not find any releases for $PROJECT on main branch"
        echo "Please specify a version explicitly, e.g.: $0 $PROJECT main.1"
        exit 1
    fi
    
    VERSION=$(echo "$LATEST_TAG" | sed "s|$PROJECT/||")
    print_info "Found latest version: $VERSION"
fi

# Construct download URL
TAG="$PROJECT/$VERSION"
BINARY_NAME="$PROJECT-$VERSION-$OS_NAME-$ARCH_NAME"
if [ "$OS_NAME" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$BINARY_NAME"

print_info "Downloading $PROJECT version $VERSION for $OS_NAME/$ARCH_NAME..."
echo "URL: $DOWNLOAD_URL"

# Download to temporary location
TMP_FILE="/tmp/$BINARY_NAME"
if ! curl -L -f -o "$TMP_FILE" "$DOWNLOAD_URL"; then
    print_error "Failed to download binary"
    echo ""
    echo "Please check:"
    echo "  1. The project name is correct: $PROJECT"
    echo "  2. The version exists: $VERSION"
    echo "  3. Your platform is supported: $OS_NAME/$ARCH_NAME"
    echo ""
    echo "View available releases at: https://github.com/$REPO/releases"
    exit 1
fi

print_success "Download complete!"

# Make executable (not needed on Windows)
if [ "$OS_NAME" != "windows" ]; then
    chmod +x "$TMP_FILE"
fi

# Determine install location
INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="$INSTALL_DIR/$PROJECT"

# Check if we need sudo
if [ -w "$INSTALL_DIR" ]; then
    SUDO=""
else
    SUDO="sudo"
    print_info "Installing to $INSTALL_DIR (requires sudo)..."
fi

# Install
if $SUDO mv "$TMP_FILE" "$INSTALL_PATH"; then
    print_success "Successfully installed $PROJECT to $INSTALL_PATH"
    echo ""
    echo "You can now run: $PROJECT"
    
    # Verify installation
    if command -v "$PROJECT" >/dev/null 2>&1; then
        VERSION_OUTPUT=$("$PROJECT" --version 2>&1 || echo "version check not supported")
        echo "Installed version: $VERSION_OUTPUT"
    fi
else
    print_error "Failed to install binary to $INSTALL_PATH"
    echo ""
    echo "You can manually install by running:"
    echo "  sudo mv $TMP_FILE $INSTALL_PATH"
    exit 1
fi
