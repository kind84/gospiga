FROM dependencies AS builder
FROM balenalib/aarch64-debian
RUN [ "cross-build-start" ]

RUN apt-get update && \
  apt-get install -y --no-install-recommends ca-certificates curl iputils-ping && \
  rm -rf /var/lib/apt/lists/*

ADD linux /usr/local/bin

EXPOSE 8080

RUN mkdir /server
WORKDIR /server

COPY --from=builder /gospiga/scripts /scripts
COPY --from=builder /gospiga/templates /templates
COPY --from=builder /gospiga/gql /gql

CMD ["server"]
RUN [ "cross-build-end" ]
