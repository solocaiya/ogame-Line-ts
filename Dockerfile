# ============================================
# Stage 1: Build Go API Server
# ============================================
FROM golang:1.23-alpine AS go-builder

WORKDIR /build/server

# 安装 CGO 依赖（SQLite 需要）
RUN apk add --no-cache gcc musl-dev

# 先复制依赖文件，利用 Docker 缓存
COPY server/go.mod server/go.sum ./
RUN go config set proxy https://goproxy.cn,direct && \
    go mod download

# 复制服务器源码并编译
COPY server/ .
RUN CGO_ENABLED=1 GOOS=linux go build -o /ogame-api-server .

# ============================================
# Stage 2: Build Frontend
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build

# 安装 pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# 先复制依赖文件，利用 Docker 缓存
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制前端源码并构建
COPY . .
RUN pnpm run build

# ============================================
# Stage 3: Production (nginx + Go API)
# ============================================
FROM nginx:alpine

# 安装 dumb-init 用于正确处理信号
RUN apk add --no-cache dumb-init

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 清理默认的 nginx 静态文件
RUN rm -rf /usr/share/nginx/html/*

# 从前端构建阶段复制产物
COPY --from=frontend-builder /build/docs /usr/share/nginx/html

# 复制 Go API 服务器二进制
COPY --from=go-builder /ogame-api-server /usr/local/bin/ogame-api-server

# 创建数据目录（SQLite 数据库存放位置）
RUN mkdir -p /data

# 创建启动脚本：同时启动 API 服务器和 nginx
RUN printf '#!/bin/sh\nset -e\n/usr/local/bin/ogame-api-server -port 8081 &\nexec nginx -g "daemon off;"\n' > /start.sh && \
    chmod +x /start.sh

# SQLite 数据持久化卷
VOLUME ["/data"]

# 设置环境变量（API 服务器读取）
ENV DB_PATH=/data/ogame.db

EXPOSE 80

# Docker 健康检查：每30秒检测一次，连续3次失败视为不健康
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost/api/health || exit 1

# 使用 dumb-init 作为 PID 1 正确处理信号
ENTRYPOINT ["dumb-init", "--"]
CMD ["/start.sh"]
