ARG PHP_VERSION=8.4
ARG GOVARD_IMAGE_REPOSITORY=ddtcorex/govard-

FROM ${GOVARD_IMAGE_REPOSITORY}php:${PHP_VERSION}
ARG PHP_VERSION

LABEL dev.govard.image.name="php-magento2"
LABEL dev.govard.php.version="${PHP_VERSION}"

USER root

RUN set -eux; \
    PHP_VERSION_INT=$(echo $PHP_VERSION | cut -d. -f1,2 | tr -d .); \
    if [ "$PHP_VERSION_INT" -lt 72 ]; then COMPOSER_BIN="composer1"; else COMPOSER_BIN="composer"; fi; \
    USE_ZEND_ALLOC=0 php -d memory_limit=-1 -d pcre.jit=0 -d opcache.enable_cli=0 /usr/local/bin/$COMPOSER_BIN create-project --no-dev n98/magerun2 /usr/local/share/n98-magerun2 \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/n98-magerun \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/magerun \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/magerun2 \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/mr

USER www-data
