#!/usr/bin/env sh
set -e

# ─── Synchronization ───────────────────────────────────────────────────────

# Ensure www-data matches the host PUID/PGID
if [ -n "${PUID:-}" ]; then
  CURRENT_UID=$(id -u www-data)
  if [ "${CURRENT_UID}" != "${PUID}" ]; then
    echo "Updating www-data UID to ${PUID}..."
    sudo usermod -u "${PUID}" www-data
  fi
fi

if [ -n "${PGID:-}" ]; then
  CURRENT_GID=$(id -g www-data)
  if [ "${CURRENT_GID}" != "${PGID}" ]; then
    echo "Updating www-data GID to ${PGID}..."
    sudo groupmod -g "${PGID}" www-data
  fi
fi

# Add a fallback group (uid 1000) for internal image files if needed
if [ -n "${PUID:-}" ] && [ "${PUID}" != "1000" ]; then
  if ! getent group govard-legacy >/dev/null; then
    sudo groupadd -g 1000 govard-legacy || true
    sudo usermod -aG govard-legacy www-data || true
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
  sudo crond
fi

exec "$@"
