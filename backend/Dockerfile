# ===== Stage 1: builder =====
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 先复制依赖清单，利用层缓存
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 静态构建
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/butterfly-monitor ./cmd

# ===== Stage 2: runtime =====
FROM alpine:latest

WORKDIR /app

# CA 证书、时区、curl、mysql-client（entrypoint 需要）
RUN apk --no-cache add ca-certificates tzdata curl mysql-client

# 拷贝二进制
COPY --from=builder /out/butterfly-monitor ./

# 关键：configs 与二进制同目录层级，框架默认读取 configs/config.yml（相对 CWD=/app）
COPY configs/ ./configs/

# 日志目录（框架默认 ./logs，相对 CWD）
RUN mkdir -p /app/logs

EXPOSE 8088

ENTRYPOINT ["./butterfly-monitor"]
