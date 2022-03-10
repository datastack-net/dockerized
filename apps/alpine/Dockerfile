ARG ALPINE_VERSION=""
FROM alpine:$ALPINE_VERSION
ARG ALPINE_PACKAGES=""
ARG BUILD_SCRIPT_ARGS=""

# Install packages
RUN apk add --no-cache $ALPINE_PACKAGES

ADD build.sh build.sh
RUN chmod +x build.sh
RUN ./build.sh $BUILD_SCRIPT_ARGS
RUN rm build.sh
