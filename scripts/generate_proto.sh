#!/bin/bash

# Proto 文件生成脚本
# 用途: 根据 .proto 文件生成 Go 代码

echo "开始生成 Protocol Buffers 文件..."

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装"
    echo "请访问 https://grpc.io/docs/protoc-installation/ 安装 protoc"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "安装 protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# 检查 protoc-gen-go-grpc 是否安装
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "安装 protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# 生成 proto 文件
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/service.proto

if [ $? -eq 0 ]; then
    echo "✓ Protocol Buffers 文件生成成功!"
else
    echo "✗ Protocol Buffers 文件生成失败!"
    exit 1
fi

