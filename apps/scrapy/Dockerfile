FROM alpine:latest

ARG SCRAPY_VERSION=""

ENV BUILD_DEPS gcc \
    cargo \
    musl-dev

RUN apk -U add \
        ${BUILD_DEPS} \
        libffi-dev \
        libxml2-dev \
        libxslt-dev \
        openssl-dev \
        libressl-dev \
        python3-dev \
        py-pip \
        curl \
        ca-certificates \
    && update-ca-certificates \
    && pip install --upgrade pip \
    && pip install scrapy==$SCRAPY_VERSION \
    && apk -U del ${BUILD_DEPS} \
    && rm -rf /var/cache/apk/*

ENTRYPOINT ["/usr/bin/scrapy"]