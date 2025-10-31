#!/bin/bash

# 项目初始化脚本
# 用途: 安装依赖、生成必要的文件、初始化数据库等

echo "========================================="
echo "  微服务项目初始化"
echo "========================================="

# 1. 检查 Go 环境
echo ""
echo "1. 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装，请先安装 Go 1.21+"
    exit 1
fi
echo "✓ Go 版本: $(go version)"

# 2. 安装 Go 依赖
echo ""
echo "2. 安装 Go 依赖..."
go mod download
go mod tidy
echo "✓ 依赖安装完成"

# 3. 生成 Proto 文件
echo ""
echo "3. 生成 Protocol Buffers 文件..."
bash scripts/generate_proto.sh
if [ $? -ne 0 ]; then
    echo "警告: Proto 文件生成失败，如果不使用 gRPC 可以忽略"
fi

# 4. 创建必要的目录
echo ""
echo "4. 创建必要的目录..."
mkdir -p logs
mkdir -p bin
echo "✓ 目录创建完成"

# 5. 复制配置文件
echo ""
echo "5. 检查配置文件..."
if [ ! -f ".env" ]; then
    echo "创建 .env 文件..."
    cp .env.example .env
    echo "⚠ 请编辑 .env 文件填写实际的配置信息"
fi

# 6. 启动 Docker 服务
echo ""
echo "6. 启动依赖服务 (PostgreSQL, Redis, RabbitMQ)..."
read -p "是否使用 Docker Compose 启动依赖服务? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
        echo "✓ 依赖服务启动成功"
        echo "  - PostgreSQL: localhost:5432"
        echo "  - Redis: localhost:6379"
        echo "  - RabbitMQ: localhost:5672 (管理界面: http://localhost:15672)"
    else
        echo "错误: docker-compose 未安装"
    fi
fi

# 7. 完成
echo ""
echo "========================================="
echo "  初始化完成!"
echo "========================================="
echo ""
echo "下一步:"
echo "  1. 编辑 config/config.yaml 填写配置信息"
echo "  2. 如果使用了 Docker，等待服务启动完成（约30秒）"
echo "  3. 运行服务:"
echo "     - 网关服务: make run-gateway"
echo "     - gRPC 服务: make run-grpc"
echo "     - 定时任务: make run-cron"
echo ""
echo "查看所有可用命令: make help"
echo ""

