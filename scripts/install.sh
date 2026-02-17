#!/usr/bin/env bash
# install.sh — Download and install update-watcher
# Usage: curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash

set -euo pipefail

REPO="mahype/update-watcher"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="update-watcher"

info() { echo "  [*] $*"; }
error() { echo "  [!] $*" >&2; }

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux|darwin) ;;
    *) error "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l)  ARCH="armv7" ;;
    *) error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

info "Detected platform: ${OS}/${ARCH}"

# Set config directory based on OS
if [ "$OS" = "linux" ]; then
    CONFIG_DIR="/etc/update-watcher"
else
    CONFIG_DIR="${HOME}/.config/update-watcher"
fi

# Check for required tools
for tool in curl tar; do
    if ! command -v "$tool" &> /dev/null; then
        error "Required tool not found: $tool"
        exit 1
    fi
done

# Fetch latest release tag
info "Fetching latest release..."
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
    error "Failed to determine latest release"
    exit 1
fi
info "Latest version: ${LATEST}"

# Download binary
ARCHIVE="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"
info "Downloading ${URL}..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if ! curl -sSL -o "${TMPDIR}/${ARCHIVE}" "$URL"; then
    error "Download failed"
    exit 1
fi

# Extract
info "Extracting..."
tar xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# Install binary
info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
if [ -w "$INSTALL_DIR" ]; then
    install -m 0755 "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    sudo install -m 0755 "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

# Create config directory
if [ ! -d "$CONFIG_DIR" ]; then
    info "Creating config directory ${CONFIG_DIR}..."
    if [ -w "$(dirname "$CONFIG_DIR")" ]; then
        mkdir -p "$CONFIG_DIR"
    else
        sudo mkdir -p "$CONFIG_DIR"
    fi
fi

# Verify
info "Verifying installation..."
VERSION=$("${INSTALL_DIR}/${BINARY_NAME}" version --short 2>/dev/null || echo "unknown")
echo ""
echo "  update-watcher ${VERSION} installed successfully!"
echo ""
echo "  Quick start:"
echo "    update-watcher setup          # Interactive setup wizard"
echo "    update-watcher watch apt      # Add APT watcher"
echo "    update-watcher run --dry-run  # Test run without notifications"
echo ""

# Offer to run setup
if [ -t 0 ]; then
    read -p "  Run interactive setup now? [Y/n] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
        "${INSTALL_DIR}/${BINARY_NAME}" setup
    fi
fi
