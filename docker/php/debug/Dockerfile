ARG BASE_IMAGE=ddtcorex/govard-php:8.4
FROM ${BASE_IMAGE}

# Set to install a specific PECL Xdebug version instead of Govard's recommended
# default below (wired up via stack.xdebug_version in .govard.yml).
ARG XDEBUG_VERSION_OVERRIDE=""

USER root

# Xdebug 3.5.x is the release line supporting the full PHP 8.1-8.5 range
# (see xdebug.org/docs/compat). Pinning it avoids `pecl install xdebug`
# silently resolving to a build that hasn't been tested against a
# brand-new PHP release yet.
RUN set -eux; \
    apk add --no-cache $PHPIZE_DEPS linux-headers; \
    PHP_MAJOR=$(php -r 'echo PHP_MAJOR_VERSION;'); \
    PHP_MINOR=$(php -r 'echo PHP_MINOR_VERSION;'); \
    if [ -n "$XDEBUG_VERSION_OVERRIDE" ]; then \
        XDEBUG_VERSION="xdebug-${XDEBUG_VERSION_OVERRIDE}"; \
    elif [ "$PHP_MAJOR" -eq 8 ]; then \
        XDEBUG_VERSION="xdebug-3.5.3"; \
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
