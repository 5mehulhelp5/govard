ARG PHP_VERSION=8.4
ARG GOVARD_IMAGE_REPOSITORY=ddtcorex/govard-

FROM ${GOVARD_IMAGE_REPOSITORY}php:${PHP_VERSION}
ARG PHP_VERSION

LABEL dev.govard.image.name="php-magento2"
LABEL dev.govard.php.version="${PHP_VERSION}"

USER root

RUN apk add --no-cache nodejs npm curl openssh-client

RUN npm install -g --unsafe-perm grunt-cli gulp yarn

RUN set -eux \
    && mkdir -p /usr/local/share/n98-magerun2 \
    && USE_ZEND_ALLOC=0 composer create-project n98/magerun2 /usr/local/share/n98-magerun2 --no-dev \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/n98-magerun \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/magerun \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/magerun2 \
    && ln -s /usr/local/share/n98-magerun2/bin/n98-magerun2 /usr/local/bin/mr

USER www-data
