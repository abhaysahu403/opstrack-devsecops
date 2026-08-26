# syntax=docker/dockerfile:1

## ---- Builder stage ----
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/opstrack ./cmd/server

## ---- Runtime stage ----
FROM alpine:3.19 AS runtime

RUN apk add --no-cache ca-certificates curl && \
    addgroup -S opstrack && adduser -S opstrack -G opstrack

WORKDIR /app

COPY --from=builder /out/opstrack /app/opstrack
COPY --from=builder /src/migrations /app/migrations

ENV SERVER_PORT=8080
EXPOSE 8080

USER opstrack

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${SERVER_PORT}/health || exit 1

ENTRYPOINT ["/app/opstrack"]
