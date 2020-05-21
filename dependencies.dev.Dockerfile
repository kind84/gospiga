ARG GOVERSION=1.14.3
FROM techknowlogick/xgo:go-${GOVERSION} AS dep

ENV GOPROXY=https://proxy.golang.org
ENV GOBIN=$GOPATH/bin

WORKDIR /gospiga

COPY go.mod .
COPY go.sum .

# RUN apt update && apt full-upgrade -y
# RUN apt install libc-dev gcc-aarch64-linux-gnu -y

RUN go mod download

# Add here shared packages
# COPY ./version.go .
# COPY ./pkg ./pkg
# COPY ./proto ./proto
# COPY ./scripts ./scripts
# COPY ./templates ./templates
# COPY ./include ./include
COPY . .
