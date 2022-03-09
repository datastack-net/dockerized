# Docker image for protoc

ARG PROTOC_BASE=""
FROM $PROTOC_BASE
ARG PROTOC_VERSION="";
ARG PROTOC_ARCH="";

RUN apk add gcompat
ADD "https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-linux-$PROTOC_ARCH.zip" protoc.zip
RUN mkdir /usr/local/lib/protoc && unzip protoc.zip -d /usr/local/lib/protoc && rm protoc.zip
RUN ln -s /usr/local/lib/protoc/bin/protoc /usr/local/bin/protoc

ENTRYPOINT ["/usr/local/bin/protoc"]
CMD ["--version"]
