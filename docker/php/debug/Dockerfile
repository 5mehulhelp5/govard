ARG BASE_IMAGE=ddtcorex/govard-php:8.4
FROM ${BASE_IMAGE}

USER root

RUN set -eux; \
    apk add --no-cache $PHPIZE_DEPS linux-headers; \
    PHP_MAJOR=$(php -r 'echo PHP_MAJOR_VERSION;'); \
    PHP_MINOR=$(php -r 'echo PHP_MINOR_VERSION;'); \
    XDEBUG_VERSION="xdebug"; \
    if [ "$PHP_MAJOR" -eq 7 ]; then \
        if [ "$PHP_MINOR" -ge 2 ]; then XDEBUG_VERSION="xdebug-3.1.6"; \
        elif [ "$PHP_MINOR" -eq 1 ]; then XDEBUG_VERSION="xdebug-2.9.8"; \
        else XDEBUG_VERSION="xdebug-2.6.1"; fi; \
    fi; \
    if [ "$PHP_MAJOR" -eq 5 ]; then XDEBUG_VERSION="xdebug-2.5.5"; fi; \
    pecl channel-update pecl.php.net; \
    pecl install "$XDEBUG_VERSION"; \
    docker-php-ext-enable xdebug; \
    apk del $PHPIZE_DEPS linux-headers

USER www-data
