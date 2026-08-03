# Docker 构建说明

本项目支持两种 Docker 构建方式。Docker 镜像包含前端（nginx 托管）和后端（Go API 服务器）两部分。

## 🏠 本地构建

### 方式一：使用构建脚本（推荐）

**Linux/macOS:**
```bash
chmod +x build-docker.sh
./build-docker.sh
```

**Windows:**
```cmd
build-docker.bat
```

### 方式二：直接使用 Docker 命令

```bash
# 构建镜像
docker build -t ogame-vue-ts:local .

# 运行容器（带数据持久化）
docker run -p 8080:80 -v ogame-data:/data ogame-vue-ts:local
```

## ☁️ GitHub Actions 自动构建

当代码推送到 `main` 分支或创建 tag 时，GitHub Actions 会自动：

1. 在 Actions 环境中构建前端和 Go API 服务器
2. 使用构建产物创建 Docker 镜像
3. 推送到 GitHub Container Registry 和 Docker Hub

### 使用预构建镜像

```bash
# 从 GitHub Container Registry 拉取
docker pull ghcr.io/your-username/ogame-vue-ts:latest

# 运行（带数据持久化）
docker run -p 8080:80 -v ogame-data:/data ghcr.io/your-username/ogame-vue-ts:latest
```

## 📁 文件说明

- `Dockerfile` - 本地构建用，多阶段构建：Go API 服务器 + 前端 + nginx
- `Dockerfile.ci` - GitHub Actions 构建用，使用预构建前端产物 + 编译 Go API
- `.dockerignore` - 本地构建时排除的文件
- `.dockerignore.ci` - CI 构建时排除的文件
- `nginx.conf` - nginx 配置（静态文件 + API 反向代理）
- `build-docker.sh` / `build-docker.bat` - 本地构建便捷脚本

## 🔧 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_PATH` | SQLite 数据库文件路径 | `/data/ogame.db` |

### 数据持久化

SQLite 数据库存储在 `/data` 目录，建议挂载 Docker volume：

```bash
docker run -p 8080:80 -v ogame-data:/data ogame-vue-ts:local
```

### GitHub Actions 环境变量

需要在 GitHub 仓库设置中配置：

**Variables (公开):**
- `DOCKERHUB_USERNAME` - Docker Hub 用户名（可选）

**Secrets (私密):**
- `DOCKERHUB_TOKEN` - Docker Hub 访问令牌（可选）
- `GITHUB_TOKEN` - 自动提供，用于 GHCR

### 本地构建要求

- Docker
- 足够的磁盘空间（构建过程中会下载 Node.js 和 Go 依赖）

## 🚀 快速开始

1. **本地开发测试:**
   ```bash
   ./build-docker.sh
   docker run -p 8080:80 -v ogame-data:/data ogame-vue-ts:local
   ```

2. **访问应用:**
   打开浏览器访问 `http://localhost:8080`

3. **生产部署:**
   使用 GitHub Actions 自动构建的镜像进行部署

## 🖥️ 非 Docker 部署

如果不使用 Docker，可以分别运行前端和后端：

### 方式一：nginx + Go API（推荐）

```bash
# 1. 构建前端
pnpm run build

# 2. 将 docs/ 目录放到 nginx 静态文件目录
# 3. 复制 nginx.conf 到 nginx 配置目录
# 4. 启动 Go API 服务器
cd server && go build -o ogame-server . && ./ogame-server -port 8081

# 5. 启动 nginx
nginx
```

### 方式二：Go 独立二进制

```bash
# 1. 构建前端
pnpm run build

# 2. 构建根目录 Go 二进制（内嵌前端 + API 代理）
go build -o ogame .

# 3. 先启动 API 服务器
cd server && ./ogame-server -port 8081 &

# 4. 启动主程序（自动代理 /api/ 到 API 服务器）
./ogame -port 8080
```
