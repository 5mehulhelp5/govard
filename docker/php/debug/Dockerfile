ARG BASE_IMAGE=ddtcorex/govard-php:8.4
FROM ${BASE_IMAGE}

USER root

RUN set -eux; \
    apk add --no-cache $PHPIZE_DEPS linux-headers; \
    XDEBUG_VERSION=$(php -r 'echo (int)PHP_MAJOR_VERSION >= 8 ? "xdebug" : "xdebug-2.9.8";'); \
    pecl install "$XDEBUG_VERSION"; \
    docker-php-ext-enable xdebug; \
    apk del $PHPIZE_DEPS linux-headers

USER www-data
