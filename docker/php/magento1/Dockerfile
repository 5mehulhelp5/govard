ARG PHP_VERSION=8.2
ARG GOVARD_IMAGE_REPOSITORY=ddtcorex/govard-

FROM ${GOVARD_IMAGE_REPOSITORY}php:${PHP_VERSION}
ARG PHP_VERSION

LABEL dev.govard.image.name="php-magento1"
LABEL dev.govard.php.version="${PHP_VERSION}"

USER root

RUN apk add --no-cache nodejs curl openssh-client \
    && if ! command -v npm >/dev/null 2>&1; then apk add --no-cache npm; fi

RUN npm install -g --unsafe-perm grunt-cli gulp yarn

RUN curl -sS -L https://github.com/netz98/n98-magerun/releases/download/1.101.1/n98-magerun.phar -o /usr/local/bin/n98-magerun \
    && chmod +x /usr/local/bin/n98-magerun \
    && ln -s /usr/local/bin/n98-magerun /usr/local/bin/magerun \
    && ln -s /usr/local/bin/n98-magerun /usr/local/bin/mr

USER www-data
