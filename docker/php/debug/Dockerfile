ARG BASE_IMAGE=ddtcorex/govard-php:8.4
FROM ${BASE_IMAGE}

# Set to install a specific PECL Xdebug version instead of Govard's recommended
# default below (wired up via stack.xdebug_version in .govard.yml).
ARG XDEBUG_VERSION_OVERRIDE=""

USER root

# Xdebug 3.5.x has a crash bug in its rewritten debugger.c: a real debug
# client connection combined with internal-function-call breakpoint checks
# (e.g. exception breakpoints) segfaults in zend_get_executed_lineno() via
# handle_breakpoints() (confirmed via gdb backtrace on PHP 8.1, both 3.5.1
# and 3.5.3). Default to the pre-rewrite 3.4.x line instead. PHP 8.5 has no
# 3.4.x option (see xdebug.org/docs/compat), so it stays on 3.5.x and is
# exposed to the same bug until Xdebug ships a fix.
RUN set -eux; \
    apk add --no-cache $PHPIZE_DEPS linux-headers; \
    PHP_MAJOR=$(php -r 'echo PHP_MAJOR_VERSION;'); \
    PHP_MINOR=$(php -r 'echo PHP_MINOR_VERSION;'); \
    if [ -n "$XDEBUG_VERSION_OVERRIDE" ]; then \
        XDEBUG_VERSION="xdebug-${XDEBUG_VERSION_OVERRIDE}"; \
    elif [ "$PHP_MAJOR" -eq 8 ]; then \
        if [ "$PHP_MINOR" -ge 5 ]; then XDEBUG_VERSION="xdebug-3.5.3"; \
        else XDEBUG_VERSION="xdebug-3.4.5"; fi; \
    elif [ "$PHP_MAJOR" -eq 7 ]; then \
        if [ "$PHP_MINOR" -ge 2 ]; then XDEBUG_VERSION="xdebug-3.1.6"; \
        elif [ "$PHP_MINOR" -eq 1 ]; then XDEBUG_VERSION="xdebug-2.9.8"; \
        else XDEBUG_VERSION="xdebug-2.6.1"; fi; \
    elif [ "$PHP_MAJOR" -eq 5 ]; then \
        XDEBUG_VERSION="xdebug-2.5.5"; \
    else \
        XDEBUG_VERSION="xdebug"; \
    fi; \
    pecl channel-update pecl.php.net; \
    pecl install "$XDEBUG_VERSION"; \
    docker-php-ext-enable xdebug; \
    apk del $PHPIZE_DEPS linux-headers

USER www-data
