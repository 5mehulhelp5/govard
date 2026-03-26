#!/usr/bin/env sh
set -e

DOCROOT="${APACHE_DOCUMENT_ROOT:-/var/www/html}"
XDEBUG_SESSION="${APACHE_XDEBUG_SESSION:-PHPSTORM}"
CONFIG_PATH="/usr/local/apache2/conf/httpd.conf"

if [ -n "${PGID:-}" ]; then
  groupmod -g "${PGID}" www-data 2>/dev/null || groupmod -g "${PGID}" daemon 2>/dev/null || true
fi
if [ -n "${PUID:-}" ]; then
  usermod -u "${PUID}" www-data 2>/dev/null || usermod -u "${PUID}" daemon 2>/dev/null || true
fi

if grep -q '@DOCROOT@\|@XDEBUG_SESSION@' "$CONFIG_PATH" 2>/dev/null; then
  if [ -w "$CONFIG_PATH" ]; then
    sed -i "s|@DOCROOT@|${DOCROOT}|g" "$CONFIG_PATH"
    sed -i "s|@XDEBUG_SESSION@|${XDEBUG_SESSION}|g" "$CONFIG_PATH"
  else
    echo "apache config contains unresolved placeholders but is not writable: ${CONFIG_PATH}" >&2
    exit 1
  fi
fi

exec "$@"
