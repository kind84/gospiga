# BUILD redisfab/redisearch-${OSNICK}:${VERSION}-${ARCH}

ARG REDIS_VER=5.0.7

# OSNICK=bionic|stretch|buster
ARG OSNICK=buster

# OS=debian:buster-slim|debian:stretch-slim|ubuntu:bionic
ARG OS=debian:buster

# ARCH=x64|arm64v8|arm32v7
ARG ARCH=arm64v8

ARG GIT_DESCRIBE_VERSION

#----------------------------------------------------------------------------------------------
FROM ${OS} AS builder
FROM ${ARCH}/redis:${REDIS_VER}-${OSNICK} AS redis

ARG OSNICK
ARG OS
ARG ARCH
ARG REDIS_VER
# ARG PACK

WORKDIR /data

ENV LIBDIR /usr/lib/redis/modules
RUN mkdir -p "$LIBDIR";

COPY --from=builder /build/build/redisearch.so  "$LIBDIR"

CMD ["redis-server", "--loadmodule", "/usr/lib/redis/modules/redisearch.so"]
