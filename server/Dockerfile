FROM dependencies AS builder

WORKDIR /gospiga/server

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o /go/bin/server /gospiga/server/cmd/server


FROM alpine:latest

ENV GOSPIGA_SERVER_PORT=8080
COPY --from=builder /go/bin/server /bin/server
COPY --from=builder /gospiga/scripts /scripts

ENTRYPOINT ["/bin/server"]
