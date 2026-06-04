# ============================================================
# people_page_backend — 企业级多阶段构建
# Go/Gin 服务，连接 MySQL + Redis，支持微信登录
# ============================================================

# ── Stage 1: 构建 Go 二进制 ─────────────────────────────────
FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ── Stage 2: 运行时镜像 ─────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache \
    ca-certificates \
    curl \
    tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 创建非 root 用户
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser

COPY --from=builder /app/server /app/server
COPY --from=builder /src/config/config.docker.yaml /app/config/config.yaml.template
COPY docker-entrypoint.sh /app/

RUN chmod +x /app/docker-entrypoint.sh && \
    mkdir -p /app/media/letters && chown -R appuser:appuser /app

WORKDIR /app
USER appuser

EXPOSE 8081

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

CMD ["/app/docker-entrypoint.sh"]
