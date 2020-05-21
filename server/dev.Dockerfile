ARG GOVERSION=1.14.3
FROM dependencies-dev AS builder

# RUN CGO_ENABLED=1 CC=/usr/bin/aarch64-linux-gnu-gcc-8 GOOS=linux GOARCH=arm64 \
# go build -o /go/bin/server /gospiga/server/cmd/server
RUN mkdir -p /home/go/src && mkdir /home/go/bin && mkdir /home/go/pkg
ENV GOPATH=/home/go
RUN cp -r /gospiga/ /home/go/src/

RUN xgo -go="go-$GOVERSION" -v -x --targets=linux/arm64 -out server gospiga

FROM balenalib/aarch64-alpine
RUN [ "cross-build-start" ]

COPY --from=builder /build/server-linux-arm64 /bin/
COPY --from=builder /gospiga/scripts /scripts
COPY --from=builder /gospiga/templates /templates
COPY /gql/schema.graphql /gql/
RUN chmod 766 /bin/server-linux-arm64

ENTRYPOINT ["/bin/server-linux-arm64"]
RUN [ "cross-build-end" ]
