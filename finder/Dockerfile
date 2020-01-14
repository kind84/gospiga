FROM dependencies AS builder

WORKDIR /gospiga/finder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -o /go/bin/finder /gospiga/finder/cmd/finder

ENTRYPOINT ["/go/bin/finder"]


FROM alpine:latest

COPY --from=builder /go/bin/finder /bin/finder

ENTRYPOINT ["/bin/finder"]
