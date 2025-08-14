# stage 1: build server
FROM golang:1.24.5 AS builder

WORKDIR /tools
RUN apt-get update \
 && apt-get install -y curl tar xz-utils \
 && curl -LO https://github.com/upx/upx/releases/download/v4.2.1/upx-4.2.1-amd64_linux.tar.xz \
 && tar -xf upx-4.2.1-amd64_linux.tar.xz \
 && mv upx-4.2.1-amd64_linux/upx /usr/local/bin/ \
 && rm -rf upx*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o qokl ./ \
  && upx --lzma --best qokl

# stage 2: runtime
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/qokl .

VOLUME ["/app/data"]

ENV APP_DATA_PATH=/app/data

ENTRYPOINT ["./qokl"]
