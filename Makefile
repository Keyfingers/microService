.PHONY: help build run-gateway run-grpc run-cron proto clean test

help: ## 显示帮助信息
	@echo "可用的命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## 生成 protobuf 文件
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/service.proto

build: ## 编译所有服务
	@echo "编译网关服务..."
	go build -o bin/gateway cmd/gateway/main.go
	@echo "编译 gRPC 服务..."
	go build -o bin/grpc-server cmd/grpc-server/main.go
	@echo "编译定时任务服务..."
	go build -o bin/cron-server cmd/cron-server/main.go
	@echo "编译完成!"

run-gateway: ## 运行网关服务
	go run cmd/gateway/main.go

run-grpc: ## 运行 gRPC 服务
	go run cmd/grpc-server/main.go

run-cron: ## 运行定时任务服务
	go run cmd/cron-server/main.go

test: ## 运行测试
	go test -v ./...

clean: ## 清理编译文件
	rm -rf bin/
	rm -rf logs/

deps: ## 安装依赖
	go mod download
	go mod tidy

docker-up: ## 启动 Docker 服务
	docker-compose up -d

docker-down: ## 停止 Docker 服务
	docker-compose down

lint: ## 代码检查
	golangci-lint run

fmt: ## 代码格式化
	go fmt ./...

