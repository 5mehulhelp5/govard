#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-govard}"
REPO="${GOVARD_REPO:-ddtcorex/govard}"
INSTALL_DIR_DEFAULT="/usr/local/bin"
INSTALL_DIR="${INSTALL_DIR:-$INSTALL_DIR_DEFAULT}"
RELEASE_TAG="${GOVARD_VERSION:-}"

usage() {
  cat <<'EOF'
Install Govard from GitHub Releases.

Usage:
  install-release.sh [--version <tag>] [--local] [--dir <path>] [--repo <owner/repo>]
  curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/scripts/install-release.sh | bash -s -- [options]

Options:
  --version <tag>  Install a specific tag (e.g. v1.0.1). Defaults to latest release.
  --local          Install to ~/.local/bin (no sudo needed)
  --dir <path>     Install to a custom directory
  --repo <repo>    GitHub repository (default: ddtcorex/govard)
  -h, --help       Show this help

Environment variables:
  GOVARD_VERSION   Same as --version
  GOVARD_REPO      Same as --repo
  INSTALL_DIR      Same as --dir
  BINARY_NAME      Installed binary name (default: govard)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      if [[ $# -lt 2 ]]; then
        echo "Error: --version requires a value."
        usage
        exit 1
      fi
      RELEASE_TAG="$2"
      shift 2
      ;;
    --local)
      INSTALL_DIR="${HOME}/.local/bin"
      shift
      ;;
    --dir)
      if [[ $# -lt 2 ]]; then
        echo "Error: --dir requires a path."
        usage
        exit 1
      fi
      INSTALL_DIR="$2"
      shift 2
      ;;
    --repo)
      if [[ $# -lt 2 ]]; then
        echo "Error: --repo requires a value."
        usage
        exit 1
      fi
      REPO="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Error: unknown option '$1'"
      usage
      exit 1
      ;;
  esac
done

if command -v curl >/dev/null 2>&1; then
  downloader="curl"
elif command -v wget >/dev/null 2>&1; then
  downloader="wget"
else
  echo "Error: curl or wget is required."
  exit 1
fi

download_to() {
  local url="$1"
  local out="$2"
  if [[ "$downloader" == "curl" ]]; then
    curl -fsSL "$url" -o "$out"
  else
    wget -qO "$out" "$url"
  fi
}

download_text() {
  local url="$1"
  if [[ "$downloader" == "curl" ]]; then
    curl -fsSL "$url"
  else
    wget -qO- "$url"
  fi
}

can_write_dir() {
  local dir="$1"
  if [[ -d "$dir" ]]; then
    [[ -w "$dir" ]]
    return
  fi
  local parent
  parent="$(dirname "$dir")"
  [[ -d "$parent" && -w "$parent" ]]
}

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Linux) os="Linux" ;;
  Darwin) os="Darwin" ;;
  *)
    echo "Error: unsupported OS '$uname_s'."
    exit 1
    ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Error: unsupported architecture '$uname_m'."
    exit 1
    ;;
esac

if [[ -z "$RELEASE_TAG" ]]; then
  latest_json="$(download_text "https://api.github.com/repos/${REPO}/releases/latest")"
  RELEASE_TAG="$(printf '%s\n' "$latest_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
fi

if [[ -z "$RELEASE_TAG" ]]; then
  echo "Error: unable to determine release tag."
  exit 1
fi

version_no_v="${RELEASE_TAG#v}"
archive_name="${BINARY_NAME}_${version_no_v}_${os}_${arch}.tar.gz"
checksums_name="checksums.txt"
base_url="https://github.com/${REPO}/releases/download/${RELEASE_TAG}"
archive_url="${base_url}/${archive_name}"
checksums_url="${base_url}/${checksums_name}"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

archive_path="${tmp_dir}/${archive_name}"
checksums_path="${tmp_dir}/${checksums_name}"

echo "Downloading ${archive_name} from ${REPO}@${RELEASE_TAG}..."
download_to "$archive_url" "$archive_path"

if download_to "$checksums_url" "$checksums_path"; then
  checksum_line="$(grep -E "[[:space:]]${archive_name}$" "$checksums_path" || true)"
  if [[ -n "$checksum_line" ]]; then
    expected_checksum="$(printf '%s\n' "$checksum_line" | awk '{print $1}')"
    if command -v sha256sum >/dev/null 2>&1; then
      printf '%s\n' "$checksum_line" | (cd "$tmp_dir" && sha256sum -c - >/dev/null)
    elif command -v shasum >/dev/null 2>&1; then
      actual_checksum="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
      if [[ "$actual_checksum" != "$expected_checksum" ]]; then
        echo "Error: checksum verification failed."
        exit 1
      fi
    else
      echo "Warning: no sha256 tool found; skipping checksum verification."
    fi
  fi
fi

tar -xzf "$archive_path" -C "$tmp_dir"
binary_path="$(find "$tmp_dir" -type f -name "$BINARY_NAME" | head -n1)"
if [[ -z "$binary_path" ]]; then
  echo "Error: binary '${BINARY_NAME}' not found in archive."
  exit 1
fi

target_dir="$INSTALL_DIR"
target_path="${target_dir%/}/${BINARY_NAME}"

if can_write_dir "$target_dir"; then
  mkdir -p "$target_dir"
  install -m 0755 "$binary_path" "$target_path"
elif command -v sudo >/dev/null 2>&1; then
  sudo mkdir -p "$target_dir"
  sudo install -m 0755 "$binary_path" "$target_path"
else
  target_dir="${HOME}/.local/bin"
  target_path="${target_dir%/}/${BINARY_NAME}"
  mkdir -p "$target_dir"
  install -m 0755 "$binary_path" "$target_path"
  echo "Warning: no write access to '${INSTALL_DIR}' and sudo was not found."
fi

echo "Installed ${BINARY_NAME} to ${target_path}"
if ! command -v "${BINARY_NAME}" >/dev/null 2>&1; then
  echo "Add '${target_dir}' to your PATH if needed:"
  echo "  export PATH=\"${target_dir}:\$PATH\""
fi
echo "Run '${BINARY_NAME} version' to verify."
