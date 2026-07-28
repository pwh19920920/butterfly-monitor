# ===== Stage 1: builder =====
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 先复制依赖清单，利用层缓存
COPY go.mod go.sum ./
# 复制 vendor（离线构建，.dockerignore 不要排除 vendor/）
COPY vendor/ ./vendor/

# 复制源码
COPY . .

# 静态构建，使用 vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /out/dragonfly-monitor ./cmd

# ===== Stage 2: runtime =====
FROM alpine:latest

WORKDIR /app

# CA 证书与时区（连接 MySQL、日志时间正确）
RUN apk --no-cache add ca-certificates tzdata

# 拷贝二进制
COPY --from=builder /out/dragonfly-monitor ./

# 关键：configs 与二进制同目录层级，框架默认读取 configs/config.yml（相对 CWD=/app）
COPY configs/ ./configs/

# 日志目录（框架默认 ./logs，相对 CWD）
RUN mkdir -p /app/logs

EXPOSE 8088

ENTRYPOINT ["./dragonfly-monitor"]
