# Docker 构建说明

本项目支持两种 Docker 构建方式：

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

# 运行容器
docker run -p 8080:80 ogame-vue-ts:local
```

## ☁️ GitHub Actions 自动构建

当代码推送到 `main` 分支或创建 tag 时，GitHub Actions 会自动：

1. 在 Actions 环境中构建项目
2. 使用构建产物创建 Docker 镜像
3. 推送到 GitHub Container Registry 和 Docker Hub

### 使用预构建镜像

```bash
# 从 GitHub Container Registry 拉取
docker pull ghcr.io/your-username/ogame-vue-ts:latest

# 从 Docker Hub 拉取（如果配置了）
docker pull your-dockerhub-username/ogame-vue-ts:latest

# 运行
docker run -p 8080:80 ghcr.io/your-username/ogame-vue-ts:latest
```

## 📁 文件说明

- `Dockerfile` - 本地构建用，包含完整的源代码构建流程
- `Dockerfile.ci` - GitHub Actions 构建用，使用预构建产物
- `.dockerignore` - 本地构建时排除的文件
- `.dockerignore.ci` - CI 构建时排除的文件
- `build-docker.sh` / `build-docker.bat` - 本地构建便捷脚本

## 🔧 配置说明

### GitHub Actions 环境变量

需要在 GitHub 仓库设置中配置：

**Variables (公开):**
- `DOCKERHUB_USERNAME` - Docker Hub 用户名（可选）

**Secrets (私密):**
- `DOCKERHUB_TOKEN` - Docker Hub 访问令牌（可选）
- `GITHUB_TOKEN` - 自动提供，用于 GHCR

### 本地构建要求

- Docker
- 足够的磁盘空间（构建过程中会下载 Node.js 依赖）

## 🚀 快速开始

1. **本地开发测试:**
   ```bash
   ./build-docker.sh
   docker run -p 8080:80 ogame-vue-ts:local
   ```

2. **访问应用:**
   打开浏览器访问 `http://localhost:8080`

3. **生产部署:**
   使用 GitHub Actions 自动构建的镜像进行部署