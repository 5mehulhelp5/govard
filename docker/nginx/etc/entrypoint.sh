#!/usr/bin/env sh
set -e

TEMPLATE="${NGINX_TEMPLATE:-magento2.conf}"
TEMPLATE_PATH="/etc/nginx/templates/${TEMPLATE}"

if [ -n "${PGID:-}" ]; then
  groupmod -g "${PGID}" nginx 2>/dev/null || true
fi
if [ -n "${PUID:-}" ]; then
  usermod -u "${PUID}" nginx 2>/dev/null || true
fi

if [ ! -f "$TEMPLATE_PATH" ]; then
  echo "nginx template not found: ${TEMPLATE_PATH}" >&2
  exit 1
fi

export NGINX_PUBLIC="${NGINX_PUBLIC:-}"
export XDEBUG_SESSION_PATTERN="${XDEBUG_SESSION_PATTERN:-PHPSTORM}"

envsubst '${NGINX_PUBLIC} ${XDEBUG_SESSION_PATTERN}' < "$TEMPLATE_PATH" > /etc/nginx/conf.d/default.conf

exec "$@"
