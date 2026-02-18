#!/usr/bin/env bash
# uninstall.sh — Completely remove update-watcher
# Usage: curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/uninstall.sh | bash
#
# Options:
#   --yes   Skip all confirmation prompts

set -euo pipefail

BINARY_NAME="update-watcher"
INSTALL_DIR="/usr/local/bin"
SERVICE_USER="update-watcher"

# Parse flags
AUTO_YES=false
for arg in "$@"; do
    case "$arg" in
        --yes) AUTO_YES=true ;;
    esac
done

info()  { echo "  [*] $*"; }
warn()  { echo "  [!] $*"; }
error() { echo "  [!] $*" >&2; }

confirm() {
    local prompt="$1"
    if [ "$AUTO_YES" = true ]; then
        return 0
    fi
    if [ ! -t 0 ]; then
        return 0
    fi
    read -p "  ${prompt} [y/N] " -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]]
}

# --- Detect platform ---

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

if [ "$OS" = "linux" ]; then
    CONFIG_DIR="/etc/update-watcher"
else
    CONFIG_DIR="${HOME}/.config/update-watcher"
fi

# --- Detect installation type ---

HAS_SERVICE_USER=false
if [ "$OS" = "linux" ] && id "$SERVICE_USER" &>/dev/null; then
    HAS_SERVICE_USER=true
fi

HAS_BINARY=false
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    HAS_BINARY=true
fi

HAS_CONFIG=false
if [ -d "$CONFIG_DIR" ]; then
    HAS_CONFIG=true
fi

HAS_LOG=false
if [ -f "/var/log/${BINARY_NAME}.log" ]; then
    HAS_LOG=true
fi

HAS_SUDOERS=false
if [ -f "/etc/sudoers.d/${SERVICE_USER}" ]; then
    HAS_SUDOERS=true
fi

# --- Show what will be removed ---

echo ""
echo "  update-watcher uninstaller"
echo "  =========================="
echo ""

if [ "$HAS_BINARY" = false ] && [ "$HAS_CONFIG" = false ]; then
    warn "update-watcher does not appear to be installed."
    exit 0
fi

echo "  The following components were found:"
echo ""
[ "$HAS_BINARY" = true ]       && echo "    - Binary:   ${INSTALL_DIR}/${BINARY_NAME}"
[ "$HAS_CONFIG" = true ]       && echo "    - Config:   ${CONFIG_DIR}/"
[ "$HAS_LOG" = true ]          && echo "    - Log:      /var/log/${BINARY_NAME}.log"
[ "$HAS_SUDOERS" = true ]      && echo "    - Sudoers:  /etc/sudoers.d/${SERVICE_USER}"
[ "$HAS_SERVICE_USER" = true ] && echo "    - User:     ${SERVICE_USER}"
echo ""

if ! confirm "Remove all listed components?"; then
    info "Aborted."
    exit 0
fi

echo ""

# --- Remove cron job ---

if [ "$HAS_SERVICE_USER" = true ]; then
    if sudo crontab -u "$SERVICE_USER" -l &>/dev/null 2>&1; then
        info "Removing cron job for user '${SERVICE_USER}'..."
        sudo crontab -u "$SERVICE_USER" -r 2>/dev/null || true
    fi
elif [ "$HAS_BINARY" = true ]; then
    # Try removing cron via the binary itself
    if crontab -l 2>/dev/null | grep -q "${BINARY_NAME}"; then
        info "Removing cron job for current user..."
        "${INSTALL_DIR}/${BINARY_NAME}" uninstall-cron 2>/dev/null || \
            (crontab -l 2>/dev/null | grep -v "${BINARY_NAME}" | crontab - 2>/dev/null) || true
    fi
fi

# --- Remove binary ---

if [ "$HAS_BINARY" = true ]; then
    info "Removing binary ${INSTALL_DIR}/${BINARY_NAME}..."
    sudo rm -f "${INSTALL_DIR}/${BINARY_NAME}"
fi

# --- Remove config ---

if [ "$HAS_CONFIG" = true ]; then
    info "Removing config directory ${CONFIG_DIR}/..."
    if [ -w "$CONFIG_DIR" ] || [ -w "$(dirname "$CONFIG_DIR")" ]; then
        rm -rf "$CONFIG_DIR"
    else
        sudo rm -rf "$CONFIG_DIR"
    fi
fi

# --- Remove log file ---

if [ "$HAS_LOG" = true ]; then
    info "Removing log file /var/log/${BINARY_NAME}.log..."
    sudo rm -f "/var/log/${BINARY_NAME}.log"
fi

# --- Remove sudoers ---

if [ "$HAS_SUDOERS" = true ]; then
    info "Removing sudoers file /etc/sudoers.d/${SERVICE_USER}..."
    sudo rm -f "/etc/sudoers.d/${SERVICE_USER}"
fi

# --- Remove dedicated user ---

if [ "$HAS_SERVICE_USER" = true ]; then
    info "Removing system user '${SERVICE_USER}'..."
    sudo userdel -r "$SERVICE_USER" 2>/dev/null || sudo userdel "$SERVICE_USER" 2>/dev/null || true
fi

echo ""
info "update-watcher has been completely removed."
echo ""
