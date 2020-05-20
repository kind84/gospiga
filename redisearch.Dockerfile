# BUILD redisfab/redisearch-${OSNICK}:${VERSION}-${ARCH}

ARG REDIS_VER=5.0.7

# OSNICK=bionic|stretch|buster
ARG OSNICK=buster

# OS=debian:buster-slim|debian:stretch-slim|ubuntu:bionic
ARG OS=debian:buster-slim

# ARCH=x64|arm64v8|arm32v7
ARG ARCH=arm64v8

ARG GIT_DESCRIBE_VERSION

#----------------------------------------------------------------------------------------------
FROM redisfab/redis-${ARCH}-${OSNICK}-xbuild AS redis

RUN [ "cross-build-start" ]

WORKDIR /
RUN apt update && apt install git -y
RUN git clone https://github.com/RedisLabsModules/readies.git
RUN git clone https://github.com/RediSearch/RediSearch.git

RUN [ "cross-build-end" ]

#FROM balenalib/aarch64-debian AS builder
FROM redisfab/redis-${ARCH}-${OSNICK}-xbuild AS builder
RUN [ "cross-build-start" ]

ARG OSNICK
ARG OS
ARG ARCH
ARG REDIS_VER
ARG GIT_DESCRIBE_VERSION

RUN echo "Building for ${OSNICK} (${OS}) for ${ARCH}"
RUN apt update && apt install git make wget cmake build-essential -y

WORKDIR /build
COPY --from=redis /usr/local/ /usr/local/
COPY --from=redis /RediSearch/ .
COPY --from=redis /readies/ ./deps/readies/

RUN ./deps/readies/bin/getpy2
# RUN ./deps/readies/bin/system-setup.py
RUN /usr/local/bin/redis-server --version
RUN make fetch SHOW=1
RUN make build SHOW=1 CMAKE_ARGS="-DGIT_DESCRIBE_VERSION=${GIT_DESCRIBE_VERSION}"

# ARG PACK=0
ARG TEST=0

# RUN if [ "$PACK" = "1" ]; then make pack; fi
RUN if [ "$TEST" = "1" ]; then TEST= make test; fi

RUN [ "cross-build-end" ]
#----------------------------------------------------------------------------------------------
FROM redisfab/redis-${ARCH}-${OSNICK}-xbuild
RUN [ "cross-build-start" ]

ARG OSNICK
ARG OS
ARG ARCH
ARG REDIS_VER
# ARG PACK

WORKDIR /data

ENV LIBDIR /usr/lib/redis/modules
RUN mkdir -p "$LIBDIR";

COPY --from=builder /build/build/redisearch.so  "$LIBDIR"

CMD ["redis-server", "--loadmodule", "/usr/lib/redis/modules/redisearch.so"
RUN [ "cross-build-end" ]
