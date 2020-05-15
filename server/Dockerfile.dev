FROM dependencies AS builder

WORKDIR /gospiga/server

COPY . .

RUN CGO_ENABLED=1 CC=/usr/bin/aarch64-linux-gnu-gcc-8 GOOS=linux GOARCH=arm64 \
go build -o /go/bin/server /gospiga/server/cmd/server


FROM alpine:latest

ENV GOSPIGA_SERVER_PORT=8080
COPY --from=builder /go/bin/server /bin/server
COPY --from=builder /gospiga/scripts /scripts
COPY --from=builder /gospiga/templates /templates
COPY /gql/schema.graphql /gql/

ENTRYPOINT ["/bin/server"]
