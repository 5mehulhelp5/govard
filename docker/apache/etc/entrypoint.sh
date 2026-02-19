#!/usr/bin/env sh
set -e

DOCROOT="${APACHE_DOCUMENT_ROOT:-/var/www/html}"
XDEBUG_SESSION="${APACHE_XDEBUG_SESSION:-PHPSTORM}"

if [ -n "${PGID:-}" ]; then
  groupmod -g "${PGID}" www-data 2>/dev/null || groupmod -g "${PGID}" daemon 2>/dev/null || true
fi
if [ -n "${PUID:-}" ]; then
  usermod -u "${PUID}" www-data 2>/dev/null || usermod -u "${PUID}" daemon 2>/dev/null || true
fi

sed -i "s|@DOCROOT@|${DOCROOT}|g" /usr/local/apache2/conf/httpd.conf
sed -i "s|@XDEBUG_SESSION@|${XDEBUG_SESSION}|g" /usr/local/apache2/conf/httpd.conf

exec "$@"
