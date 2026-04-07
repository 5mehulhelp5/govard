ARG PHP_VERSION=8.2
ARG GOVARD_IMAGE_REPOSITORY=ddtcorex/govard-

FROM ${GOVARD_IMAGE_REPOSITORY}php:${PHP_VERSION}
ARG PHP_VERSION

LABEL dev.govard.image.name="php-magento1"
LABEL dev.govard.php.version="${PHP_VERSION}"

USER root

RUN apk add --no-cache nodejs npm curl openssh-client

RUN npm install -g grunt-cli gulp yarn

RUN set -eux \
    && mkdir -p /usr/local/bin \
    && curl -sL --insecure https://files.magerun.net/n98-magerun.phar > /usr/local/bin/n98-magerun \
    && chmod +x /usr/local/bin/n98-magerun \
    && ln -s /usr/local/bin/n98-magerun /usr/local/bin/mr

USER www-data
