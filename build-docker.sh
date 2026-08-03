#!/bin/bash

# 本地 Docker 构建脚本
# 使用完整的源代码构建流程

echo "🚀 开始本地 Docker 构建..."

# 构建镜像
docker build -t ogame-vue-ts:local .

if [ $? -eq 0 ]; then
    echo "✅ Docker 镜像构建成功！"
    echo "📦 镜像标签: ogame-vue-ts:local"
    echo "🏃 运行命令: docker run -p 8080:80 -v ogame-data:/data ogame-vue-ts:local"
else
    echo "❌ Docker 镜像构建失败！"
    exit 1
fi