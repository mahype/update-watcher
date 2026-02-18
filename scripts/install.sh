#!/usr/bin/env bash
# install.sh — Download and install update-watcher
# Usage: curl -sSL https://raw.githubusercontent.com/mahype/update-watcher/main/scripts/install.sh | bash
#
# Options:
#   --server      Run server setup automatically (dedicated user, permissions, sudoers)
#   --no-server   Skip server setup entirely

set -euo pipefail

REPO="mahype/update-watcher"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="update-watcher"
SERVICE_USER="update-watcher"

# Parse flags
SERVER_MODE=""
for arg in "$@"; do
    case "$arg" in
        --server)    SERVER_MODE="yes" ;;
        --no-server) SERVER_MODE="no" ;;
    esac
done

info()  { echo "  [*] $*"; }
warn()  { echo "  [!] $*"; }
error() { echo "  [!] $*" >&2; }

# Run command as root: directly if already root, via sudo otherwise
run_root() {
    if [ "$(id -u)" -eq 0 ]; then
        "$@"
    else
        sudo "$@"
    fi
}

# Check that we can run privileged commands
require_root() {
    if [ "$(id -u)" -eq 0 ]; then
        return 0
    fi
    if ! command -v sudo &>/dev/null; then
        error "This operation requires root privileges, but sudo is not installed."
        error "Please run this script as root."
        exit 1
    fi
    if ! sudo -v 2>/dev/null; then
        error "This operation requires root privileges, but sudo authentication failed."
        error "Please run this script as root or ensure your user has sudo access."
        exit 1
    fi
}

ask() {
    local prompt="$1" default="$2"
    if [ ! -t 0 ]; then
        # Non-interactive: use default
        [ "$default" = "y" ] && return 0 || return 1
    fi
    local hint="[y/N]"
    [ "$default" = "y" ] && hint="[Y/n]"
    read -p "  ${prompt} ${hint} " -n 1 -r
    echo
    if [ "$default" = "y" ]; then
        [[ ! $REPLY =~ ^[Nn]$ ]]
    else
        [[ $REPLY =~ ^[Yy]$ ]]
    fi
}

# --- Detect platform ---

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux|darwin) ;;
    *) error "Unsupported OS: $OS"; exit 1 ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l)        ARCH="armv7" ;;
    *) error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

info "Detected platform: ${OS}/${ARCH}"

# Set config directory based on OS
if [ "$OS" = "linux" ]; then
    CONFIG_DIR="/etc/update-watcher"
else
    CONFIG_DIR="${HOME}/.config/update-watcher"
fi

# --- Check prerequisites ---

for tool in curl tar; do
    if ! command -v "$tool" &> /dev/null; then
        error "Required tool not found: $tool"
        exit 1
    fi
done

# --- Download and install binary ---

info "Fetching latest release..."
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
    error "Failed to determine latest release"
    exit 1
fi
info "Latest version: ${LATEST}"

ARCHIVE="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"
info "Downloading ${URL}..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if ! curl -sSL -o "${TMPDIR}/${ARCHIVE}" "$URL"; then
    error "Download failed"
    exit 1
fi

info "Extracting..."
tar xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
if [ -w "$INSTALL_DIR" ]; then
    install -m 0755 "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    require_root
    run_root install -m 0755 "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

# --- Create config directory ---

if [ ! -d "$CONFIG_DIR" ]; then
    info "Creating config directory ${CONFIG_DIR}..."
    if [ -w "$(dirname "$CONFIG_DIR")" ]; then
        mkdir -p "$CONFIG_DIR"
    else
        run_root mkdir -p "$CONFIG_DIR"
    fi
fi

# --- Verify installation ---

VERSION=$("${INSTALL_DIR}/${BINARY_NAME}" version --short 2>/dev/null || echo "unknown")
echo ""
info "update-watcher ${VERSION} installed successfully!"
echo ""

# --- Linux server setup ---

if [ "$OS" = "linux" ]; then
    setup_server() {
        echo ""
        info "=== Server Setup ==="
        echo ""

        # Verify root/sudo access before starting
        require_root

        # 1. Create dedicated system user
        if id "$SERVICE_USER" &>/dev/null; then
            info "User '${SERVICE_USER}' already exists, skipping creation."
        else
            info "Creating system user '${SERVICE_USER}'..."
            run_root useradd -r -s /usr/sbin/nologin -m -d /var/lib/${SERVICE_USER} "$SERVICE_USER"
            info "User '${SERVICE_USER}' created."
        fi

        # 2. Set config directory ownership and permissions
        info "Setting config directory permissions..."
        run_root chown "${SERVICE_USER}:${SERVICE_USER}" "$CONFIG_DIR"
        run_root chmod 755 "$CONFIG_DIR"

        if [ -f "${CONFIG_DIR}/config.yaml" ]; then
            run_root chown "${SERVICE_USER}:${SERVICE_USER}" "${CONFIG_DIR}/config.yaml"
            run_root chmod 600 "${CONFIG_DIR}/config.yaml"
        else
            run_root touch "${CONFIG_DIR}/config.yaml"
            run_root chown "${SERVICE_USER}:${SERVICE_USER}" "${CONFIG_DIR}/config.yaml"
            run_root chmod 600 "${CONFIG_DIR}/config.yaml"
        fi
        info "Config: ${CONFIG_DIR}/config.yaml (mode 0600, owner ${SERVICE_USER})"

        # 3. Log file
        if [ "$SERVER_MODE" = "yes" ] || ask "Set up log file at /var/log/${BINARY_NAME}.log?" "y"; then
            run_root touch "/var/log/${BINARY_NAME}.log"
            run_root chown "${SERVICE_USER}:${SERVICE_USER}" "/var/log/${BINARY_NAME}.log"
            run_root chmod 640 "/var/log/${BINARY_NAME}.log"
            info "Log file: /var/log/${BINARY_NAME}.log (mode 0640)"
        fi

        # 4. Sudoers for package manager refresh
        if [ "$SERVER_MODE" = "yes" ] || ask "Set up sudoers for package manager refresh?" "y"; then
            SUDOERS_FILE="/etc/sudoers.d/${SERVICE_USER}"
            SUDOERS_CONTENT="# update-watcher: allow package list refresh\n"
            SUDOERS_ADDED=false

            if command -v apt-get &>/dev/null; then
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/apt-get update\n"
                SUDOERS_ADDED=true
            fi
            if command -v dnf &>/dev/null; then
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/dnf check-update\n"
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/dnf updateinfo list --security\n"
                SUDOERS_ADDED=true
            fi
            if command -v pacman &>/dev/null; then
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/pacman -Sy\n"
                SUDOERS_ADDED=true
            fi
            if command -v zypper &>/dev/null; then
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive refresh\n"
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive list-patches --category security\n"
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /usr/bin/zypper --non-interactive list-updates\n"
                SUDOERS_ADDED=true
            fi
            if command -v apk &>/dev/null; then
                SUDOERS_CONTENT+="${SERVICE_USER} ALL=(root) NOPASSWD: /sbin/apk update\n"
                SUDOERS_ADDED=true
            fi

            if [ "$SUDOERS_ADDED" = true ]; then
                echo -e "$SUDOERS_CONTENT" | run_root tee "$SUDOERS_FILE" > /dev/null
                run_root chmod 440 "$SUDOERS_FILE"
                info "Sudoers file created: ${SUDOERS_FILE}"
            else
                warn "No supported package managers found, skipping sudoers setup."
            fi
        fi

        # 5. Docker group
        if command -v docker &>/dev/null; then
            if [ "$SERVER_MODE" = "yes" ] || ask "Add '${SERVICE_USER}' to docker group?" "n"; then
                run_root usermod -aG docker "$SERVICE_USER"
                info "User '${SERVICE_USER}' added to docker group."
            fi
        fi

        # 6. Cron job
        if [ "$SERVER_MODE" = "yes" ] || ask "Install daily cron job (7:00 AM)?" "y"; then
            (run_root crontab -u "$SERVICE_USER" -l 2>/dev/null | grep -v "${BINARY_NAME}"; \
             echo "0 7 * * * ${INSTALL_DIR}/${BINARY_NAME} run --quiet") | \
            run_root crontab -u "$SERVICE_USER" -
            info "Cron job installed for user '${SERVICE_USER}' (daily at 07:00)."
        fi

        echo ""
        info "=== Server setup complete ==="
        echo ""
        echo "  Next steps:"
        echo "    sudo -u ${SERVICE_USER} ${BINARY_NAME} setup    # Configure watchers & notifiers"
        echo "    sudo -u ${SERVICE_USER} ${BINARY_NAME} run --dry-run  # Test run"
        echo ""
    }

    if [ "$SERVER_MODE" = "yes" ]; then
        setup_server
    elif [ "$SERVER_MODE" != "no" ] && [ -t 0 ]; then
        echo "  This looks like a Linux server."
        if ask "Set up dedicated system user for production use?" "n"; then
            setup_server
        else
            echo ""
            echo "  Quick start:"
            echo "    ${BINARY_NAME} setup          # Interactive setup wizard"
            echo "    ${BINARY_NAME} watch apt      # Add APT watcher"
            echo "    ${BINARY_NAME} run --dry-run  # Test run without notifications"
            echo ""
            # Offer to run setup
            if ask "Run interactive setup now?" "y"; then
                "${INSTALL_DIR}/${BINARY_NAME}" setup
            fi
        fi
    else
        echo "  Quick start:"
        echo "    ${BINARY_NAME} setup          # Interactive setup wizard"
        echo "    ${BINARY_NAME} watch apt      # Add APT watcher"
        echo "    ${BINARY_NAME} run --dry-run  # Test run without notifications"
        echo ""
    fi

# --- macOS ---
else
    echo "  Quick start:"
    echo "    ${BINARY_NAME} setup          # Interactive setup wizard"
    echo "    ${BINARY_NAME} watch homebrew  # Add Homebrew watcher"
    echo "    ${BINARY_NAME} run --dry-run  # Test run without notifications"
    echo ""
    if [ -t 0 ]; then
        if ask "Run interactive setup now?" "y"; then
            "${INSTALL_DIR}/${BINARY_NAME}" setup
        fi
    fi
fi
