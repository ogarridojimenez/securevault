# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/server ./cmd/server/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/vault ./cmd/vault/

# ============================================================
# Stage 2: Runtime
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S securevault && adduser -S -G securevault securevault

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/vault .
COPY --from=builder /build/migrations ./migrations

USER securevault

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ./server || exit 1

ENTRYPOINT ["./server"]
