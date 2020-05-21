ARG GOVERSION=1.14.3
FROM golang:${GOVERSION} AS dep

ENV GOPROXY=https://proxy.golang.org

WORKDIR /gospiga

COPY go.mod .
COPY go.sum .

RUN apt update && apt full-upgrade -y
RUN go mod download

# Add here shared packages
COPY ./version.go .
COPY ./pkg ./pkg
COPY ./proto ./proto
COPY ./scripts ./scripts
COPY ./templates ./templates
COPY ./include ./include

