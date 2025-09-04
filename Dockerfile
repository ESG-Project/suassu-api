# ===== Build stage =====
FROM golang:1.25.1-alpine3.22 AS builder

# deps de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# deps
COPY go.mod go.sum ./
RUN go mod download

# fonte
COPY . .

# build estático, menor e reprodutível
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o main ./cmd/api

# ===== Runtime stage =====
FROM alpine:3.20

# HTTPS, timezone e curl p/ healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

# usuário não-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# binário
COPY --from=builder /app/main .

# (opcional) config-example
COPY --from=builder /app/internal/config/example.env ./config/

# perms
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

# Healthcheck com curl
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1

CMD ["./main"]
