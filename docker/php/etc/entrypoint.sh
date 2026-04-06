#!/usr/bin/env sh
set -e

# ─── Synchronization ───────────────────────────────────────────────────────

# Ensure www-data group changes happen before UID changes. Once the current
# process rewrites its own UID mapping, later sudo calls may stop resolving.
if [ -n "${PGID:-}" ]; then
  CURRENT_GID=$(id -g www-data)
  if [ "${CURRENT_GID}" != "${PGID}" ]; then
    echo "Updating www-data GID to ${PGID}..."
    if ! sudo groupmod -g "${PGID}" www-data; then
      echo "Warning: could not update www-data GID to ${PGID}; continuing with GID ${CURRENT_GID}." >&2
    fi
  fi
fi

# Add a fallback group (uid 1000) for internal image files if needed
if [ -n "${PUID:-}" ] && [ "${PUID}" != "1000" ]; then
  if ! getent group govard-legacy >/dev/null; then
    sudo groupadd -g 1000 govard-legacy || true
    sudo usermod -aG govard-legacy www-data || true
  fi
fi

if [ -n "${PUID:-}" ]; then
  CURRENT_UID=$(id -u www-data)
  if [ "${CURRENT_UID}" != "${PUID}" ]; then
    echo "Updating www-data UID to ${PUID}..."
    if ! sudo usermod -u "${PUID}" www-data; then
      echo "Warning: could not update www-data UID to ${PUID}; continuing with UID ${CURRENT_UID}." >&2
    fi
  fi
fi

# Apply recursive chown if requested
if [ -n "${CHOWN_DIR_LIST:-}" ]; then
  for dir in ${CHOWN_DIR_LIST}; do
    if [ -d "${dir}" ]; then
      echo "Fixing permissions for ${dir}..."
      sudo chown -R www-data:www-data "${dir}"
    fi
  done
fi

# Ensure specific PHP directories are correct
sudo chown -R www-data:www-data /var/log/php /var/lib/php 2>/dev/null || true

# Start cron so Magento cron:install entries can execute inside the container.
if command -v crond >/dev/null 2>&1; then
  sudo crond 2>/dev/null || true
fi

# Ensure specific Composer version is active if requested
if [ -n "${COMPOSER_VERSION:-}" ] && [ "${COMPOSER_VERSION}" != "latest" ]; then
  # Exact version check can be tricky due to build dates in --version output
  # If it doesn't look like a direct match, we do a re-baseline install
  CURRENT_COMPOSER_VERSION=$(composer --version 2>/dev/null | cut -d' ' -f3)
  if [ "${CURRENT_COMPOSER_VERSION}" != "${COMPOSER_VERSION}" ]; then
    COMPOSER_BIN=$(which composer 2>/dev/null || echo "/usr/local/bin/composer")
    
    # Try pre-baked version first
    LOCAL_BIN=""
    case "${COMPOSER_VERSION}" in
      1) [ -f "/usr/local/bin/composer1" ] && LOCAL_BIN="/usr/local/bin/composer1" ;;
      2) [ -f "/usr/local/bin/composer2" ] && LOCAL_BIN="/usr/local/bin/composer2" ;;
      2.2) [ -f "/usr/local/bin/composer2lts" ] && LOCAL_BIN="/usr/local/bin/composer2lts" ;;
    esac

    if [ -n "${LOCAL_BIN}" ]; then
      echo "Using pre-baked Composer version ${COMPOSER_VERSION}..."
      sudo ln -sf "${LOCAL_BIN}" "${COMPOSER_BIN}"
      echo "Composer version $(composer --version | head -n1) is now active."
    else
      # Falling back to download for non-standard or specific point versions
      echo "Ensuring Composer version ${COMPOSER_VERSION} (downloading)..."
      DOWNLOAD_URL="https://getcomposer.org/composer-stable.phar"
      case "${COMPOSER_VERSION}" in
        1) DOWNLOAD_URL="https://getcomposer.org/composer-1.phar" ;;
        2) DOWNLOAD_URL="https://getcomposer.org/composer-2.phar" ;;
        2.2) DOWNLOAD_URL="https://getcomposer.org/download/latest-2.2.x/composer.phar" ;;
        *.*) DOWNLOAD_URL="https://getcomposer.org/download/${COMPOSER_VERSION}/composer.phar" ;;
      esac

      if sudo curl -sSfL "${DOWNLOAD_URL}" -o "${COMPOSER_BIN}.tmp"; then
        sudo chmod +x "${COMPOSER_BIN}.tmp"
        sudo mv "${COMPOSER_BIN}.tmp" "${COMPOSER_BIN}"
        echo "Composer version $(composer --version | head -n1) is now active."
      else
        echo "Warning: failed to download Composer version ${COMPOSER_VERSION}; falling back to default image version." >&2
      fi
    fi
  fi
fi

exec "$@"
