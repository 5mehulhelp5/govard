ARG BASE_IMAGE=ddtcorex/govard-php:8.4
FROM ${BASE_IMAGE}

USER root

RUN set -eux; \
    apk add --no-cache $PHPIZE_DEPS linux-headers; \
    pecl install xdebug; \
    docker-php-ext-enable xdebug; \
    apk del $PHPIZE_DEPS linux-headers

USER www-data
