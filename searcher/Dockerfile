FROM dependencies AS builder

WORKDIR /gospiga/searcher

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o /go/bin/searcher /gospiga/searcher/cmd/searcher

ENTRYPOINT ["/go/bin/searcher"]


FROM alpine:latest

COPY --from=builder /go/bin/searcher /bin/searcher

ENTRYPOINT ["/bin/searcher"]
