#!/usr/bin/env sh
set -eu

BUILDER_NAME="${GOVARD_BUILDX_BUILDER:-govard-multiarch}"
PLATFORMS="${DOCKER_PLATFORMS:-}"

has_multiple_platforms() {
  count=0
  for platform in $(printf '%s' "${PLATFORMS}" | tr ',' ' '); do
    if [ -n "${platform}" ]; then
      count=$((count + 1))
    fi
    if [ "${count}" -gt 1 ]; then
      return 0
    fi
  done
  return 1
}

ensure_multiarch_builder() {
  if docker buildx inspect "${BUILDER_NAME}" >/dev/null 2>&1; then
    return 0
  fi

  echo "Creating Docker Buildx builder '${BUILDER_NAME}' for Govard multi-platform images..." >&2
  docker buildx create --name "${BUILDER_NAME}" --driver docker-container --bootstrap >/dev/null
}

builder_platforms() {
  docker buildx inspect "${BUILDER_NAME}" --bootstrap |
    awk '/^Platforms:/ { sub(/^Platforms:[[:space:]]*/, ""); print; exit }'
}

platform_is_supported() {
  platform="$1"
  platforms_without_spaces=$(printf '%s' "$2" | tr -d ' ')
  case ",${platforms_without_spaces}," in
    *",${platform},"*) return 0 ;;
    *) return 1 ;;
  esac
}

ensure_requested_platforms_supported() {
  supported_platforms=$(builder_platforms)
  for platform in $(printf '%s' "${PLATFORMS}" | tr ',' ' '); do
    if [ -z "${platform}" ]; then
      continue
    fi
    if ! platform_is_supported "${platform}" "${supported_platforms}"; then
      echo "Docker Buildx builder '${BUILDER_NAME}' does not support requested platform '${platform}'." >&2
      echo "Supported platforms: ${supported_platforms:-unknown}" >&2
      echo "Enable Docker Desktop/Rosetta emulation or install binfmt/QEMU support, then retry." >&2
      exit 1
    fi
  done
}

if has_multiple_platforms; then
  ensure_multiarch_builder
  ensure_requested_platforms_supported
  exec docker buildx bake --builder "${BUILDER_NAME}" "$@"
fi

exec docker buildx bake "$@"
