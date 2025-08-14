# stage 1: build
FROM golang:1.24.5 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o qokl .

# stage 2: minimal runtime
FROM gcr.io/distroless/base:nonroot

WORKDIR /app
COPY --from=builder /app/qokl .

VOLUME ["/app/data"]
ENV APP_DATA_PATH=/app/data

USER root
ENTRYPOINT ["/app/qokl"]

