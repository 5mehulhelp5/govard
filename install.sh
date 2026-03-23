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
CLI_BINARY_NAME="govard"
DESKTOP_BINARY_NAME="govard-desktop"
REPO="ddtcorex/govard"
INSTALL_DIR="/usr/local/bin"
GOVARD_DIR="/opt/govard"
MIN_GO_VERSION="1.25.0"
SOURCE_MODE=false
FORCE_YES=false
SPECIFIC_VERSION=""
SOURCE_DIR=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-${0}}")" && pwd)"

desktop_build_tags() {
    local tags="desktop production"
    if [[ "$OS" == "linux" ]]; then
        tags="${tags} webkit2_41"
    fi
    echo "$tags"
}

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

run_as_user() {
    if [[ -n "${SUDO_USER:-}" && "$USER" == "root" ]]; then
        sudo -u "$SUDO_USER" "$@"
    else
        "$@"
    fi
}

usage() {
    cat <<EOF
Usage: install.sh [options]

Options:
  --source       Build from source instead of downloading binary
  --source-dir <path>  Build from a specific local source directory (requires --source)
  --local        Install to ~/.local/bin (no sudo)
  --dir <path>   Custom installation directory
  --version <v>  Install specific version (e.g. v1.9.0)
  -y, --yes      Assume yes to all prompts
  -h, --help     Show this help
EOF
    exit 0
}

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --source) SOURCE_MODE=true; shift ;;
        --source-dir) SOURCE_DIR="$2"; shift 2 ;;
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

    # certutil (for browser trust)
    if [[ "$OS" == "linux" ]]; then
        if command -v certutil >/dev/null 2>&1; then
            success "  certutil: Found"
        else
            warn "  certutil: Not found (Required for automatic browser SSL trust)"
            if [[ "$FORCE_YES" == true ]]; then
                info "Installing libnss3-tools (provides certutil) automatically..."
                sudo apt-get update && sudo apt-get install -y libnss3-tools
            else
            read -p "Do you want to install libnss3-tools (provides certutil) automatically? (y/N) " confirm </dev/tty
                if [[ $confirm =~ ^[Yy]$ ]]; then
                    sudo apt-get update && sudo apt-get install -y libnss3-tools
                fi
            fi
        fi
    fi
}

install_binary_file() {
    local source_path="$1"
    local binary_name="$2"
    local target_path="${INSTALL_DIR%/}/${binary_name}"

    info "Installing ${binary_name} to ${target_path}..."
    mkdir -p "$(dirname "$target_path")"

    if [ -w "$(dirname "$target_path")" ]; then
        mv "$source_path" "$target_path"
    else
        if command -v sudo >/dev/null 2>&1; then
            sudo mv "$source_path" "$target_path"
        else
            error "Cannot write to $(dirname "$target_path") and sudo is not available."
        fi
    fi

    if [ -w "$target_path" ]; then
        chmod +x "$target_path"
    else
        if command -v sudo >/dev/null 2>&1; then
            sudo chmod +x "$target_path"
        else
            error "Cannot chmod $target_path and sudo is not available."
        fi
    fi
}

warn_if_mixed_install_channels() {
    if [[ "${INSTALL_DIR%/}" != "/usr/local/bin" ]]; then
        return 0
    fi

    local mixed=false
    for binary_name in "$CLI_BINARY_NAME" "$DESKTOP_BINARY_NAME"; do
        local local_path="/usr/local/bin/${binary_name}"
        local system_path="/usr/bin/${binary_name}"
        if [[ -f "$local_path" && -f "$system_path" ]] && ! cmp -s "$local_path" "$system_path"; then
            warn "Detected conflicting ${binary_name} binaries:"
            echo "  - ${local_path}"
            echo "  - ${system_path}"
            mixed=true
        fi
    done

    if [[ "$mixed" == true ]]; then
        warn "Mixed install channels detected (.deb + local install). Use one channel only."
        info "If you want to keep /usr/local/bin as source of truth, remove the package-managed copy:"
        echo "  sudo apt remove govard    # or: sudo dpkg -r govard"
    fi
}

extract_binary_from_deb() {
    local deb_path="$1"
    local binary_name="$2"
    local output_path="$3"
    local extract_dir stage_dir

    extract_dir="$(mktemp -d)"
    stage_dir="${extract_dir}/stage"
    mkdir -p "$stage_dir"

    if command -v dpkg-deb >/dev/null 2>&1; then
        if ! dpkg-deb -x "$deb_path" "$stage_dir" >/dev/null 2>&1; then
            warn "Failed to extract ${deb_path} using dpkg-deb."
            rm -rf "$extract_dir"
            return 1
        fi
    else
        if ! command -v ar >/dev/null 2>&1; then
            warn "Cannot extract .deb package: neither dpkg-deb nor ar is available."
            rm -rf "$extract_dir"
            return 1
        fi

        local data_archive=""
        for candidate in data.tar.gz data.tar.xz data.tar.bz2 data.tar; do
            if ar p "$deb_path" "$candidate" > "${extract_dir}/${candidate}" 2>/dev/null; then
                data_archive="${extract_dir}/${candidate}"
                break
            fi
        done

        if [[ -z "$data_archive" ]]; then
            warn "Could not locate data.tar.* in ${deb_path}."
            rm -rf "$extract_dir"
            return 1
        fi

        case "$data_archive" in
            *.tar.gz) tar -xzf "$data_archive" -C "$stage_dir" ;;
            *.tar.xz) tar -xJf "$data_archive" -C "$stage_dir" ;;
            *.tar.bz2) tar -xjf "$data_archive" -C "$stage_dir" ;;
            *.tar) tar -xf "$data_archive" -C "$stage_dir" ;;
            *)
                warn "Unsupported Debian data archive format: $data_archive"
                rm -rf "$extract_dir"
                return 1
                ;;
        esac
    fi

    local candidate_path=""
    for candidate in "${stage_dir}/usr/local/bin/${binary_name}" "${stage_dir}/usr/bin/${binary_name}"; do
        if [[ -f "$candidate" ]]; then
            candidate_path="$candidate"
            break
        fi
    done

    if [[ -z "$candidate_path" ]]; then
        warn "Binary ${binary_name} not found in Debian package ${deb_path}."
        rm -rf "$extract_dir"
        return 1
    fi

    cp "$candidate_path" "$output_path"
    chmod +x "$output_path"
    rm -rf "$extract_dir"
    return 0
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
            read -p "Do you want to install Go $MIN_GO_VERSION automatically? (y/N) " confirm </dev/tty
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

resolve_version() {
    if [[ -z "$SPECIFIC_VERSION" ]]; then
        info "Fetching latest version..."
        SPECIFIC_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    info "Version: $SPECIFIC_VERSION"
    VERSION_NO_V="${SPECIFIC_VERSION#v}"
}

install_via_deb() {
    local deb_name="govard_${VERSION_NO_V}_linux_${ARCH}.deb"
    local deb_url="https://github.com/${REPO}/releases/download/${SPECIFIC_VERSION}/${deb_name}"
    local tmp_dir
    tmp_dir=$(mktemp -d)
    local deb_path="${tmp_dir}/${deb_name}"

    info "Downloading ${deb_name}..."
    if ! curl -fsSL "$deb_url" -o "$deb_path"; then
        warn "Failed to download ${deb_name}."
        rm -rf "$tmp_dir"
        return 1
    fi

    info "Installing via dpkg..."
    if sudo dpkg -i "$deb_path"; then
        rm -rf "$tmp_dir"
        success "Govard $SPECIFIC_VERSION installed via Debian package (CLI + Desktop)!"
        return 0
    fi

    warn "dpkg -i failed. Attempting to fix missing dependencies..."
    if sudo apt-get install -f -y 2>/dev/null; then
        rm -rf "$tmp_dir"
        success "Govard $SPECIFIC_VERSION installed via Debian package (CLI + Desktop)!"
        return 0
    fi

    warn "Debian package installation failed."
    rm -rf "$tmp_dir"
    return 1
}

install_binary() {
    info "Installing pre-built binary..."
    resolve_version

    # On Linux with dpkg available, prefer .deb package
    if [[ "$OS" == "linux" ]] && command -v dpkg >/dev/null 2>&1; then
        info "Debian-based system detected — using .deb package."
        if install_via_deb; then
            return 0
        fi
        warn "Falling back to archive-based installation."
    fi

    # Capitalize OS for archive names
    OS_CAP="$(echo "${OS:0:1}" | tr '[:lower:]' '[:upper:]')${OS:1}"

    TMP_DIR=$(mktemp -d)
    binaries=("$CLI_BINARY_NAME" "$DESKTOP_BINARY_NAME")
    extracted_entries=()

    for binary_name in "${binaries[@]}"; do
        archive_name="${binary_name}_${VERSION_NO_V}_${OS_CAP}_${ARCH}.tar.gz"
        archive_path="${TMP_DIR}/${archive_name}"
        download_url="https://github.com/${REPO}/releases/download/${SPECIFIC_VERSION}/${archive_name}"

        info "Downloading $download_url..."
        if curl -fsSL "$download_url" -o "$archive_path"; then
            tar -xzf "$archive_path" -C "$TMP_DIR"
            extracted_path="${TMP_DIR}/${binary_name}"
            if [[ ! -f "$extracted_path" ]]; then
                error "Archive ${archive_name} does not contain ${binary_name}."
            fi
            extracted_entries+=("${binary_name}:${extracted_path}")
            continue
        fi

        if [[ "$binary_name" == "$DESKTOP_BINARY_NAME" && "$OS" == "linux" ]]; then
            deb_name="govard_${VERSION_NO_V}_linux_${ARCH}.deb"
            deb_path="${TMP_DIR}/${deb_name}"
            deb_url="https://github.com/${REPO}/releases/download/${SPECIFIC_VERSION}/${deb_name}"
            warn "Desktop archive not found for ${SPECIFIC_VERSION}; falling back to Debian package ${deb_name}."
            info "Downloading $deb_url..."
            if ! curl -fsSL "$deb_url" -o "$deb_path"; then
                error "Failed to download ${deb_name} for desktop fallback."
            fi

            extracted_path="${TMP_DIR}/${binary_name}"
            if ! extract_binary_from_deb "$deb_path" "$binary_name" "$extracted_path"; then
                error "Failed to extract ${binary_name} from ${deb_name}."
            fi
            extracted_entries+=("${binary_name}:${extracted_path}")
            continue
        fi

        error "Failed to download ${binary_name} for version $SPECIFIC_VERSION."
    done

    for entry in "${extracted_entries[@]}"; do
        binary_name="${entry%%:*}"
        extracted_path="${entry#*:}"
        install_binary_file "$extracted_path" "$binary_name"
    done

    rm -rf "$TMP_DIR"
    success "Govard $SPECIFIC_VERSION installed (CLI + Desktop)!"
}

install_source() {
    install_go
    info "Building from source..."

    local source_root=""

    if [[ -n "$SOURCE_DIR" ]]; then
        if ! source_root="$(cd "$SOURCE_DIR" 2>/dev/null && pwd)"; then
            error "Invalid --source-dir path: $SOURCE_DIR"
        fi
        if [[ ! -f "${source_root}/go.mod" || ! -f "${source_root}/cmd/govard/main.go" ]]; then
            error "--source-dir does not look like a Govard source tree: ${source_root}"
        fi
        info "Using requested source directory: ${source_root}"
    elif [[ -f "go.mod" && -f "cmd/govard/main.go" ]]; then
        source_root="$PWD"
        info "Using current directory source tree: ${source_root}"
    elif [[ -f "${SCRIPT_DIR}/go.mod" && -f "${SCRIPT_DIR}/cmd/govard/main.go" ]]; then
        source_root="$SCRIPT_DIR"
        info "Using installer directory source tree: ${source_root}"
    fi

    if [[ -n "$source_root" ]]; then
        cd "$source_root"
    else
        info "Local source tree not found. Cloning repository..."
        TMP_SRC=$(mktemp -d)
        git clone "https://github.com/${REPO}.git" "$TMP_SRC"
        cd "$TMP_SRC"
    fi
    
    # Resolve version
    VERSION=$(git describe --tags --always 2>/dev/null || echo "source")
    LDFLAGS="-s -w -X govard/internal/cmd.Version=${VERSION#v} -X govard/internal/desktop.Version=${VERSION#v}"

    TMP_BUILD_DIR=$(mktemp -d)
    go build -ldflags "$LDFLAGS" -o "${TMP_BUILD_DIR}/${CLI_BINARY_NAME}" cmd/govard/main.go
    DESKTOP_BUILD_TAGS="$(desktop_build_tags)"
    go build -tags "$DESKTOP_BUILD_TAGS" -ldflags "$LDFLAGS" -o "${TMP_BUILD_DIR}/${DESKTOP_BINARY_NAME}" cmd/govard-desktop/main.go

    install_binary_file "${TMP_BUILD_DIR}/${CLI_BINARY_NAME}" "$CLI_BINARY_NAME"
    install_binary_file "${TMP_BUILD_DIR}/${DESKTOP_BINARY_NAME}" "$DESKTOP_BINARY_NAME"

    rm -rf "$TMP_BUILD_DIR"
    success "Govard built and installed from source (CLI + Desktop, tags: ${DESKTOP_BUILD_TAGS})!"
}

main() {
    show_banner
    detect_env
    check_dependencies

    if [[ -n "$SOURCE_DIR" && "$SOURCE_MODE" != true ]]; then
        error "--source-dir requires --source"
    fi
    
    if [[ "$SOURCE_MODE" == true ]]; then
        install_source
    else
        install_binary
    fi
    
    echo ""
    info "Quick start:"
    echo "  govard --help"
    echo ""

    # Post-installation automation
    local govard_bin="${INSTALL_DIR%/}/${CLI_BINARY_NAME}"
    if [[ -x "$govard_bin" ]]; then
        info "Initializing global services..."
        run_as_user "$govard_bin" svc up -d --remove-orphans || warn "Failed to start global services automatically."

        info "Configuring SSL trust..."
        run_as_user "$govard_bin" doctor trust || warn "Failed to configure SSL trust automatically."
        echo ""
    fi

    warn_if_mixed_install_channels
    success "Installation complete!"
}

main
