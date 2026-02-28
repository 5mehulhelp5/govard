#!/usr/bin/env bash
set -euo pipefail

# Govard Installer
# Unified script for binary download and source builds

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Defaults
BINARY_NAME="govard"
REPO="ddtcorex/govard"
INSTALL_DIR="/usr/local/bin"
GOVARD_DIR="/opt/govard"
MIN_GO_VERSION="1.24.0"
SOURCE_MODE=false
FORCE_YES=false
SPECIFIC_VERSION=""

# Banner
show_banner() {
    echo -e "${BLUE}"
    echo "  ____  ______     ___    ____  ____  "
    echo " / ___|/ _ \ \   / / \  |  _ \|  _ \ "
    echo "| |  _| | | \ \ / / _ \ | |_) | | | |"
    echo "| |_| | |_| |\ V / ___ \|  _ <| |_| |"
    echo " \____|\___/  \_/_/   \_\_| \_\____/ "
    echo -e "${NC}"
    echo -e "${BOLD}Go-based Versatile Runtime & Development${NC}"
    echo "========================================"
    echo ""
}

# Helpers
info()    { echo -e "${BLUE}info:${NC} $1"; }
success() { echo -e "${GREEN}success:${NC} $1"; }
warn()    { echo -e "${YELLOW}warning:${NC} $1"; }
error()   { echo -e "${RED}error:${NC} $1"; exit 1; }

usage() {
    cat <<EOF
Usage: install.sh [options]

Options:
  --source       Build from source instead of downloading binary
  --local        Install to ~/.local/bin (no sudo)
  --dir <path>   Custom installation directory
  --version <v>  Install specific version (e.g. v1.8.0)
  -y, --yes      Assume yes to all prompts
  -h, --help     Show this help
EOF
    exit 0
}

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --source) SOURCE_MODE=true; shift ;;
        --local)  INSTALL_DIR="${HOME}/.local/bin"; shift ;;
        --dir)    INSTALL_DIR="$2"; shift 2 ;;
        --version) SPECIFIC_VERSION="$2"; shift 2 ;;
        -y|--yes) FORCE_YES=true; shift ;;
        -h|--help) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

detect_env() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    info "Detected Environment: $OS/$ARCH"
}

check_dependencies() {
    info "Checking system dependencies..."
    
    # Docker
    if command -v docker >/dev/null 2>&1; then
        success "  Docker: $(docker --version | awk '{print $3}' | tr -d ',')"
    else
        warn "  Docker: Not found (Required to run Govard stacks)"
    fi

    # Docker Compose
    if docker compose version >/dev/null 2>&1; then
        success "  Docker Compose: $(docker compose version --short)"
    elif command -v docker-compose >/dev/null 2>&1; then
        success "  Docker Compose (v1): $(docker-compose --version | awk '{print $3}')"
        warn "  Docker Compose V2 (plugin) is recommended."
    else
        warn "  Docker Compose: Not found (Required to run Govard stacks)"
    fi

    # Git (for source mode)
    if command -v git >/dev/null 2>&1; then
        success "  Git: $(git --version | awk '{print $3}')"
    elif [[ "$SOURCE_MODE" == true ]]; then
        error "Git is required for --source mode."
    fi
}

install_go() {
    info "Checking Go environment..."
    GO_BIN_DIR="/usr/local/go/bin"
    
    need_go_install=false
    if ! command -v go >/dev/null 2>&1; then
        info "Go is not installed."
        need_go_install=true
    else
        CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
        # Simple version comparison
        if [[ "$(printf '%s\n' "$MIN_GO_VERSION" "$CURRENT_GO" | sort -V | head -n1)" != "$MIN_GO_VERSION" ]]; then
            warn "Go version $CURRENT_GO is too old (Need $MIN_GO_VERSION+)"
            need_go_install=true
        else
            success "Go: $CURRENT_GO"
            return 0
        fi
    fi

    if [[ "$need_go_install" == true ]]; then
        if [[ "$FORCE_YES" == false ]]; then
            read -p "Do you want to install Go $MIN_GO_VERSION automatically? (y/N) " confirm
            if [[ ! $confirm =~ ^[Yy]$ ]]; then
                error "Go $MIN_GO_VERSION+ is required for source builds."
            fi
        fi

        info "Downloading Go $MIN_GO_VERSION..."
        GO_TAR="go${MIN_GO_VERSION}.${OS}-${ARCH}.tar.gz"
        GO_URL="https://go.dev/dl/${GO_TAR}"
        
        TMP_DIR=$(mktemp -d)
        curl -sL "$GO_URL" -o "${TMP_DIR}/${GO_TAR}"
        
        info "Extricating Go to /usr/local/go..."
        if [ -d /usr/local/go ]; then
            sudo rm -rf /usr/local/go
        fi
        sudo tar -C /usr/local -xzf "${TMP_DIR}/${GO_TAR}"
        rm -rf "$TMP_DIR"

        # Update PATH
        if ! echo "$PATH" | grep -q "$GO_BIN_DIR"; then
            export PATH="$PATH:$GO_BIN_DIR"
            SHELL_RC="${HOME}/.bashrc"
            [[ "$SHELL" == *"zsh"* ]] && SHELL_RC="${HOME}/.zshrc"
            
            echo "export PATH=\$PATH:$GO_BIN_DIR" >> "$SHELL_RC"
            info "Added $GO_BIN_DIR to $SHELL_RC"
        fi
        success "Go installed successfully."
    fi
}

install_binary() {
    info "Installing pre-built binary..."
    
    if [[ -z "$SPECIFIC_VERSION" ]]; then
        info "Fetching latest version..."
        SPECIFIC_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    
    info "Version: $SPECIFIC_VERSION"
    VERSION_NO_V="${SPECIFIC_VERSION#v}"
    
    # Capitalize OS for archive name
    OS_CAP="$(echo "${OS:0:1}" | tr '[:lower:]' '[:upper:]')${OS:1}"
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION_NO_V}_${OS_CAP}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${SPECIFIC_VERSION}/${ARCHIVE_NAME}"
    
    TMP_DIR=$(mktemp -d)
    info "Downloading $DOWNLOAD_URL..."
    if ! curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/bundle.tar.gz"; then
        error "Failed to download binary. It might not exist for version $SPECIFIC_VERSION."
    fi
    
    tar -xzf "${TMP_DIR}/bundle.tar.gz" -C "$TMP_DIR"
    
    TARGET_PATH="${INSTALL_DIR%/}/${BINARY_NAME}"
    info "Installing to $TARGET_PATH..."
    mkdir -p "$(dirname "$TARGET_PATH")"
    
    if [ -w "$(dirname "$TARGET_PATH")" ]; then
        mv "${TMP_DIR}/${BINARY_NAME}" "$TARGET_PATH"
    else
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "$TARGET_PATH"
    fi
    chmod +x "$TARGET_PATH"
    
    rm -rf "$TMP_DIR"
    success "Govard $SPECIFIC_VERSION installed!"
}

install_source() {
    install_go
    info "Building from source..."
    
    # Detect if we are in the repo
    if [[ -f "go.mod" && -f "cmd/govard/main.go" ]]; then
        info "Inside Govard repository. Building..."
    else
        info "Cloning repository..."
        TMP_SRC=$(mktemp -d)
        git clone "https://github.com/${REPO}.git" "$TMP_SRC"
        cd "$TMP_SRC"
    fi
    
    # Resolve version
    VERSION=$(git describe --tags --always 2>/dev/null || echo "source")
    LDFLAGS="-s -w -X govard/internal/cmd.Version=${VERSION#v}"
    
    go build -ldflags "$LDFLAGS" -o "$BINARY_NAME" cmd/govard/main.go
    
    TARGET_PATH="${INSTALL_DIR%/}/${BINARY_NAME}"
    info "Installing to $TARGET_PATH..."
    mkdir -p "$(dirname "$TARGET_PATH")"
    
    if [ -w "$(dirname "$TARGET_PATH")" ]; then
        mv "$BINARY_NAME" "$TARGET_PATH"
    else
        sudo mv "$BINARY_NAME" "$TARGET_PATH"
    fi
    chmod +x "$TARGET_PATH"
    success "Govard built and installed from source!"
}

main() {
    show_banner
    detect_env
    check_dependencies
    
    if [[ "$SOURCE_MODE" == true ]]; then
        install_source
    else
        install_binary
    fi
    
    echo ""
    info "Quick start:"
    echo "  govard --help"
    echo ""
    success "Installation complete!"
}

main
