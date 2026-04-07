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
    && mkdir -p /usr/local/share/n98-magerun \
    && composer create-project n98/magerun /usr/local/share/n98-magerun --no-dev \
    && ln -s /usr/local/share/n98-magerun/bin/n98-magerun /usr/local/bin/n98-magerun \
    && ln -s /usr/local/share/n98-magerun/bin/n98-magerun /usr/local/bin/magerun \
    && ln -s /usr/local/share/n98-magerun/bin/n98-magerun /usr/local/bin/mr

USER www-data
